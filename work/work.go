package work

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/config"
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

func (w Work) LoadEnv(name string) *Env {
	return NewEnvironment(w.path+PATH_ENV, name)
}

func (w Work) ListEnvs() []string {
	path := config.GetConfig().WorkPath + PATH_ENV
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Error("env directory not found in " + config.GetConfig().WorkPath)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Error("Cannot read env directory : "+path, err)
		os.Exit(1)
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
