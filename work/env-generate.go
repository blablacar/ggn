package work

func (e Env) Generate() {
	services := e.listServices()

	for _, service := range services {
		service := e.LoadService(service)
		service.GenerateUnits(e.attributesDir(), e.name)
	}
}
