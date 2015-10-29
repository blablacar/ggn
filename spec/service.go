package spec

import "github.com/blablacar/cnt/spec"

const PATH_SERVICE_MANIFEST = "/service-manifest.yml"

//type NodeManifest struct {
//	Hostname string   `yaml:"hostname"`
//	Ip       string   `yaml:"ip"`
//	Fleet    []string `yaml:"fleet"`
//}

const NODE_HOSTNAME = "hostname"

type ServiceManifest struct {
	Containers   []spec.ACFullname        `yaml:"containers"`
	ExecStartPre []string                 `yaml:"execStartPre"`
	ExecStart    []string                 `yaml:"execStart"`
	Nodes        []map[string]interface{} `yaml:"nodes"`
}

type Service interface {
}
