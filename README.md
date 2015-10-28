# green-garden



# commands
```bash
ggn generate                        generate all units for all envs
ggn prod-XXX                        compare local units with what is running in this env
ggn prod-XXX run list-units         run fleetctl command on this env
ggn prod-XXX generate               generate units for this env
```

# configuration file

The configuration file is located at ~/.config/green-garden/config.yml

```
workPath: /home/myuser/build-tools
```

# env directory structure

```bash
env
└── XXX
    ├── attributes                      # Attributes file for this env
    │   └── fleet.yml                   # a list of files defining default attributes for this env
    └── services                        # services running in this env
        ├── cassandra
        │   ├── attributes              # attributes directory for this service
        │   ├── unit.tml                # systemd unit template for this service
        │   ├── service-manifest.yml    # manifest that give information for this service
        │   ├── units                   # generated units
        └── elasticsearch               # a list of files defining default attributes for this env
            ├── attributes              # attributes directory for this service
            ├── unit.tml                # systemd unit template for this service
            ├── service-manifest.yml    # manifest that give information for this service
            └── units                   # generated units
```

# manifest structure

```yaml
containers:
  - aci.example.com/pod-cassandra               # list of aci or pod

nodes:                                          # list of nodes for this service
  - hostname: cass1                             # hostname of the service
    ip: 10.2.135.136                            # any other property used in the template
    fleet:
      - MachineMetadata="rack=113" "pos=4"

  - hostname: cass2
    ip: 10.2.143.136
    fleet:
      - MachineMetadata="rack=213" "pos=4"
```

# example of unit template

```
[Unit]
Description=pod-cassandra {{.node.hostname}}
After=mnt-sda9-{{.node.hostname}}-mount-sdb1.mount \
      mnt-sda9-{{.node.hostname}}-mount-sdc1.mount \
      mnt-sda9-{{.node.hostname}}-mount-sdd1.mount \
      mnt-sda9-{{.node.hostname}}-mount-sde1.mount \

[Service]
ExecStartPre=/opt/bin/rkt gc --grace-period=0s --expire-prepared=0s
ExecStartPre=-/opt/bin/rkt image gc
ExecStart=/opt/bin/rkt --insecure-skip-verify run \
    --private-net='bond0:IP={{.node.ip}}' \
    --volume=cassandra-mount-1,kind=host,source=/mnt/sda9/{{.node.hostname}}/mount/sdb1 \
    --volume=cassandra-mount-2,kind=host,source=/mnt/sda9/{{.node.hostname}}/mount/sdc1 \
    --volume=cassandra-mount-3,kind=host,source=/mnt/sda9/{{.node.hostname}}/mount/sdd1 \
    --volume=cassandra-mount-4,kind=host,source=/mnt/sda9/{{.node.hostname}}/mount/sde1 \
    --volume=cassandra-commitlog,kind=host,source=/mnt/sda9/{{.node.hostname}}/commitlog \
    --volume=cassandra-savedcaches,kind=host,source=/mnt/sda9/{{.node.hostname}}/saved_caches \
    --set-env=CONFD_OVERRIDE='{{.attributes}}' \
    --set-env=HOSTNAME={{.node.hostname}} \
    --set-env=DOMAINNAME="{{.domainname}}" \
    {{.node.acis}}

[X-Fleet]
{{range .node.fleet -}}
    {{- . }}
{{end -}}
```

