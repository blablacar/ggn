package ggn

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/ghodss/yaml"
	"io/ioutil"
)

type Config struct {
	Path     string
	WorkPath string `yaml:"workPath,omitempty"`
	User     string `yaml:"user,omitempty"`
}

type HomeStruct struct {
	path   string
	Config Config
}

func NewHome(path string) HomeStruct {
	logrus.WithField("path", path).Debug("loading home")

	var config Config
	if source, err := ioutil.ReadFile(path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			panic(err)
		}
	} else {
		logrus.WithField("path", path).WithField("file", "config.yml").Fatal("Cannot read configuration file")
	}
	return HomeStruct{
		path:   path,
		Config: config,
	}
}

func DefaultHomeRoot() string {
	return utils.UserHomeOrFatal() + "/.config"
}
