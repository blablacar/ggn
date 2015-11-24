package spec

import (
	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
)

const PATH_ATTRIBUTES = "/attributes"

type Env interface {
	GetName() string
	GetLog() logrus.Entry
	GetAttributes() map[string]interface{}
	ListMachineNames() []string
	RunFleetCmdGetOutput(args ...string) (string, string, error)
	RunFleetCmd(args ...string) error
	EtcdClient() client.KeysAPI
	RunEarlyHook(service string, action string)
	RunLateHook(service string, action string)
}
