package work

func (e Env) Generate() {
	services := e.ListServices()

	for _, service := range services {
		service := e.LoadService(service)
		service.GenerateUnits()
	}
}
