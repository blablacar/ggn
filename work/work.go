package work

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/green-garden/config"
	"github.com/blablacar/green-garden/env"
	"io/ioutil"
	"os"
)

const PATH_ENV = "/env"

type Work struct {
	path string
}

func NewWork(path string) *Work {
	work := new(Work)
	work.path = path
	return work
}

func (w Work) LoadEnv(name string) *env.Env {
	return env.NewEnvironment(w.path+PATH_ENV, name)
}

func (w Work) ListEnvs() []string {
	path := config.GetConfig().WorkPath + PATH_ENV
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Get().Panic("env directory not found in " + config.GetConfig().WorkPath)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Get().Panic("Cannot read env directory : "+path, err)
	}

	var envs []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		envs = append(envs, file.Name())
	}
	return envs
}
