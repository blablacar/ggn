package work

import "github.com/n0rad/go-erlog/logs"

func (e Env) Fleetctl(args []string) {
	logs.WithFields(e.fields).Debug("Running fleetctl")
	err := e.RunFleetCmd(args...)
	if err != nil {
		logs.WithEF(err, e.fields).Error("Fleetctl command failed")
	}
}
