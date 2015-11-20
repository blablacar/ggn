package work

func (e Env) Generate() {
	e.log.Debug("Generating units")
	services := e.ListServices()

	for _, service := range services {
		service := e.LoadService(service)
		service.GenerateUnits(nil)
	}
}
