package work

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/ggn"
	"io/ioutil"
	"os"
)

const PATH_ENV = "/env"

type Work struct {
	path string
	log  logrus.Entry
}

func NewWork(path string) *Work {
	return &Work{
		path: path,
		log:  *logrus.WithField("path", path),
	}
}

func (w Work) LoadEnv(name string) *Env {
	return NewEnvironment(w.path+PATH_ENV, name)
}

func (w Work) ListEnvs() []string {
	path := ggn.Home.Config.WorkPath + PATH_ENV
	if _, err := os.Stat(path); os.IsNotExist(err) {
		w.log.Error("env directory not found" + ggn.Home.Config.WorkPath)
		os.Exit(1)
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		w.log.Error("Cannot read env directory : "+path, err)
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
