# GGN

[![GoDoc](https://godoc.org/blablacar/ggn?status.png)](https://godoc.org/github.com/blablacar/ggn) [![Build Status](https://travis-ci.org/blablacar/ggn.svg?branch=master)](https://travis-ci.org/blablacar/ggn)


GGN uses a tree structure to describe envs and services in envs. It will generate systemd units based on information taken in the directories and send them to the environment using fleet.

# directory structure

```
env
|-- development
|   |-- attributes
|   |   `-- dns.yml                      # Attributes of this env (dns suffix, dns servers IPs, zookeeper IPs, ...)
|   |-- services                         # list of services in this env
|   |   |-- loadbalancer                 
|   |   |   |-- attributes               # loadbalancer attributes in this env
|   |   |   |   `-- nginx.yml            # any structure configuration
|   |   |   |-- unit.tmpl                # template uses to generate the systemd's units for loadbalancer
|   |   |   `-- service-manifest.yml     # manifest for this service
|   |   |-- cassandra
|   |   |   |-- attributes
|   |   |   |   |-- cassandra.yml        # cassandra configuration for this env (DC name, seeds nodes, cluster name)
|   |   |   |   `-- datastax-agent.yml   # another configuration file that will be merged with the other one
|   |   |   `-- service-manifest.yml    
|   |   ...
|   `-- config.yml                       # configuration of this env (fleet)
|-- prod-DC1
|   ...
|-- prod-DC2
|   ...
|-- preprod
|   ...
```

# commands

Some command example : 

```bash
ggn dev redis srv1 start           start redis server1 unit
ggn preprod check                  check that all units of all services in prepod are running and are up to date
ggn prod cassandra cass1 journal   see journal of cass1 systemd unit 
ggn prod lb lb1 stop               stop lb1 unit in prod
ggn prod cassandra update          rolling update of cassandra servers in prod
```

Envs commands :
```
- check                            check that all units are up to date and running in this env
- fleetctl                         run custom fleetctl command
- list-units                       list all units in this env
- generate                         generate all units in this env
```

Services commands:
```
- generate                         generate units of this service
- check                            check that all units of this service are up to date and running
- diff                             display diff of units between when is generated and what is running in this env
- lock                             lock this service. nobody will be able to run mutable actions
- unlock                           unlock the service
- list-units                       list-units of this service in fleet
- update                           run a update of all units for this service
```

Units commands:
```
- start                            start the service
- stop                             stop the service
- update                           update the service
- restart                          restart the service
- destroy                          destroy the service
- status                           display fleet status of the unit
- journal                          display unit's journal
- diff                             display the diff between genetated and what is running
- check                            check that the service is running and is up to date
- unload                           unload the unit
- load                             load the unit
- ssh                              ssh on the server's running this unit
```


# global configuration file

The configuration file is located at ~/.config/green-garden/config.yml

```
workPath: /home/myuser/build-tools           # root directory of environments. all envs have to be in an env/ directory in it
```


# service manifest structure

```yaml
concurrentUpdater: 2                            # concurrent run when updating the service
containers:
  - aci.example.com/pod-cassandra               # list of aci or pod

nodes:                                          # list of nodes for this service
  - hostname: cass1                             # hostname of the service
    ip: 10.2.135.136                            # any other property used in the template
    node-id: 1     # node special attribute
    fleet:
      - MachineMetadata="rack=113" "pos=4"

  - hostname: cass2
    ip: 10.2.143.136
    node-id: 2
    fleet:
      - MachineMetadata="rack=213" "pos=4"
```

# example of unit template

```
[Unit]
Description=pod-cassandra {{.hostname}}
After=mnt-sda9-{{.hostname}}-mount-sdb1.mount \
      mnt-sda9-{{.hostname}}-mount-sdc1.mount \
      mnt-sda9-{{.hostname}}-mount-sdd1.mount \
      mnt-sda9-{{.hostname}}-mount-sde1.mount \

[Service]
{{.environmentAttributes}}
ExecStartPre=/opt/bin/rkt gc --grace-period=0s --expire-prepared=0s
ExecStartPre=-/opt/bin/rkt image gc
ExecStart=/opt/bin/rkt --insecure-skip-verify run \
    --private-net='bond0:IP={{.ip}}' \
    --volume=cassandra-mount-1,kind=host,source=/mnt/sda9/{{.hostname}}/mount/sdb1 \
    --volume=cassandra-mount-2,kind=host,source=/mnt/sda9/{{.hostname}}/mount/sdc1 \
    --volume=cassandra-mount-3,kind=host,source=/mnt/sda9/{{.hostname}}/mount/sdd1 \
    --volume=cassandra-mount-4,kind=host,source=/mnt/sda9/{{.hostname}}/mount/sde1 \
    --volume=cassandra-commitlog,kind=host,source=/mnt/sda9/{{.hostname}}/commitlog \
    --volume=cassandra-savedcaches,kind=host,source=/mnt/sda9/{{.hostname}}/saved_caches \
    --set-env=CONFD_OVERRIDE='{{.environmentAttributesVars}}' \
    --set-env=HOSTNAME={{.hostname}} \
    --set-env=DOMAINNAME="{{.domainname}}" \
    {{.acis}}

[X-Fleet]
{{range .fleet -}}
    {{- . }}
{{end -}}
```

