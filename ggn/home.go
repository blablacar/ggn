package ggn

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"os"
	"time"
)

type Config struct {
	WorkPath string `yaml:"workPath,omitempty"`
	User     string `yaml:"user,omitempty"`
}

type HomeStruct struct {
	log    logrus.Entry
	Path   string
	Config Config
}

func NewHome(path string) HomeStruct {
	log := logrus.WithField("path", path)
	log.Debug("loading home")

	var config Config
	if source, err := ioutil.ReadFile(path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			panic(err)
		}
	} else {
		log.WithField("file", "config.yml").Fatal("Cannot read configuration file")
	}
	return HomeStruct{
		log:    *log,
		Path:   path,
		Config: config,
	}
}

const PATH_LIST_MACHINES_CACHE = "/list-machines.cache"

func (h *HomeStruct) LoadMachinesCacheWithDate(env string) (string, time.Time) {
	h.log.WithField("env", env).Debug("Loading list machines cache")
	info, err := os.Stat(h.Path + PATH_LIST_MACHINES_CACHE + "." + env)
	if err != nil {
		return "", time.Now()
	}
	content, err := ioutil.ReadFile(h.Path + PATH_LIST_MACHINES_CACHE + "." + env)
	if err != nil {
		return "", time.Now()
	}
	return string(content), info.ModTime()
}

func (h *HomeStruct) SaveMachinesCache(env string, data string) {
	h.log.WithField("env", env).Debug("save machines cache")
	if err := ioutil.WriteFile(h.Path+PATH_LIST_MACHINES_CACHE+"."+env, []byte(data), 0644); err != nil {
		logrus.WithError(err).Warn("Cannot persist list-machines cache")
	}
}

func DefaultHomeRoot() string {
	return utils.UserHomeOrFatal() + "/.config"
}
