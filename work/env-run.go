package work

func (e Env) Run(args []string) {
	err := e.RunFleetCmd(args...)
	if err != nil {
		e.log.WithError(err).Error("Fleetctl command failed")
	}
}
