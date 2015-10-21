package work

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/work/env"
	"io/ioutil"
	"os"
)

const PATH_SERVICES = "/services"

type Env struct {
	path string
	name string
}

func NewEnvironment(root string, name string) *Env {
	path := root + "/" + name
	_, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error("Cannot read env directory : "+path, err)
		os.Exit(1)
	}

	env := new(Env)
	env.path = path
	env.name = name
	return env
}

func (e Env) LoadService(name string) *env.Service {
	return env.NewService(e.path+"/services", name)
}

func (e Env) attributesDir() string {
	return e.path + spec.PATH_ATTRIBUTES
}

func (e Env) listServices() []string {
	path := e.path + PATH_SERVICES
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return []string{}
	}

	var services []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		services = append(services, file.Name())
	}
	return services
}
