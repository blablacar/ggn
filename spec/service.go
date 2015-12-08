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
	Containers   []cntspec.ACFullname `yaml:"containers"`
	ExecStartPre []string             `yaml:"execStartPre"`
	ExecStart    []string             `yaml:"execStart"`
	Nodes        interface{}          `yaml:"nodes"`
}

type Service interface {
	PrepareAciList(sources []string) string
	NodeAttributes(hostname string) map[string]interface{}
	GetAttributes() map[string]interface{}
	Generate(sources []string)
	Unlock(command string)
	Lock(command string, ttl time.Duration, message string)
	GetName() string
	GetEnv() Env
	GetLog() logrus.Entry
	GetFleetUnitContent(unit string) (string, error)
}
