package spec

import (
	"github.com/Sirupsen/logrus"
	cntspec "github.com/blablacar/cnt/spec"
	"time"
)

const PATH_SERVICE_MANIFEST = "/service-manifest.yml"

//type NodeManifest struct {
//	Hostname string   `yaml:"hostname"`
//	Ip       string   `yaml:"ip"`
//	Fleet    []string `yaml:"fleet"`
//}

const NODE_HOSTNAME = "hostname"

type ServiceManifest struct {
	Containers   []cntspec.ACFullname     `yaml:"containers"`
	ExecStartPre []string                 `yaml:"execStartPre"`
	ExecStart    []string                 `yaml:"execStart"`
	Nodes        []map[string]interface{} `yaml:"nodes"`
}

type Service interface {
	Unlock()
	Lock(ttl time.Duration, message string)
	GetName() string
	GetEnv() Env
	GetLog() logrus.Entry
	GetFleetUnitContent(unit string) (string, error)
}
