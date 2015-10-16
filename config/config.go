package config

import (
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/utils"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var ggConfig GgConfig

type GgConfig struct {
	Path     string
	WorkPath string `yaml:"workPath,omitempty"`
}

func GetConfig() *GgConfig {
	return &ggConfig
}

func (c *GgConfig) Load() {
}

func init() {
	ggConfig = GgConfig{Path: utils.UserHomeOrFatal() + "/.config/green-garden"}

	if source, err := ioutil.ReadFile(ggConfig.Path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &ggConfig)
		if err != nil {
			panic(err)
		}
	}

	log.Get().Debug("Home folder is " + ggConfig.Path)
}
