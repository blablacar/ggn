package spec

import "github.com/Sirupsen/logrus"

const PATH_ATTRIBUTES = "/attributes"

type Env interface {
	GetName() string
	GetLog() logrus.Entry
	GetAttributes() map[string]interface{}
	ListMachineNames() []string
	RunFleetCmdGetOutput(args ...string) (string, error)
}
