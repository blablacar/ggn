package work

func (e Env) Fleetctl(args []string) {
	e.log.Debug("Running fleetctl")
	err := e.RunFleetCmd(args...)
	if err != nil {
		e.log.WithError(err).Error("Fleetctl command failed")
	}
}
