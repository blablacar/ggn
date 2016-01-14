package spec

import (
	cntspec "github.com/blablacar/cnt/spec"
	"github.com/n0rad/go-erlog/data"
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
	ConcurrentUpdater int                  `yaml:"concurrentUpdater"`
	Containers        []cntspec.ACFullname `yaml:"containers"`
	ExecStartPre      []string             `yaml:"execStartPre"`
	ExecStart         []string             `yaml:"execStart"`
	Nodes             interface{}          `yaml:"nodes"`
}

type Service interface {
	HasTimer() bool
	PrepareAciList() string
	NodeAttributes(hostname string) map[string]interface{}
	GetAttributes() map[string]interface{}
	Generate()
	Unlock(command string)
	Lock(command string, ttl time.Duration, message string)
	GetName() string
	GetEnv() Env
	GetFields() data.Fields
	GetFleetUnitContent(unit string) (string, error)
}
