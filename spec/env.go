package spec

import (
	"github.com/coreos/etcd/client"
	"github.com/n0rad/go-erlog/data"
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

type HookInfo struct {
	Command    string
	Service    Service
	Unit       Unit
	Action     string
	Attributes string
}

type Env interface {
	GetName() string
	GetFields() data.Fields
	GetAttributes() map[string]interface{}
	ListMachineNames() ([]string, error)
	RunFleetCmdGetOutput(args ...string) (string, string, error)
	RunFleetCmd(args ...string) error
	EtcdClient() client.KeysAPI
	RunEarlyHook(info HookInfo)
	RunLateHook(info HookInfo)
	ListUnits() map[string]UnitStatus
}
