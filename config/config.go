package config

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var ggnConfig GgnConfig

type GgnConfig struct {
	Path     string
	WorkPath string `yaml:"workPath,omitempty"`
	User     string `yaml:"user,omitempty"`
}

func GetConfig() *GgnConfig {
	return &ggnConfig
}

func (c *GgnConfig) Load() {
}

func init() {
	ggnConfig = GgnConfig{Path: utils.UserHomeOrFatal() + "/.config/green-garden"}

	if source, err := ioutil.ReadFile(ggnConfig.Path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &ggnConfig)
		if err != nil {
			panic(err)
		}
	}

	log.Debug("Home folder is " + ggnConfig.Path)
}
