package spec

import (
	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
)

const PATH_ATTRIBUTES = "/attributes"

const ACTIVE_ACTIVE = "active"
const SUB_RUNNING = "running"

type UnitStatus struct {
	Unit    string
	Machine string
	Active  string
	Sub     string
}

type Env interface {
	GetName() string
	GetLog() logrus.Entry
	GetAttributes() map[string]interface{}
	ListMachineNames() []string
	RunFleetCmdGetOutput(args ...string) (string, string, error)
	RunFleetCmd(args ...string) error
	EtcdClient() client.KeysAPI
	RunEarlyHookUnit(unit Unit, action string)
	RunLateHookUnit(unit Unit, action string)
	RunEarlyHookService(service Service, action string)
	RunLateHookService(service Service, action string)
	ListUnits() map[string]UnitStatus
}
