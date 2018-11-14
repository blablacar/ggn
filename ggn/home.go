package ggn

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/ghodss/yaml"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
)

type Config struct {
	WorkPath   string `yaml:"workPath,omitempty"`
	User       string `yaml:"user,omitempty"`
	EnvVarUser string `yaml:"envVarUser,omitempty"`

	OverrideConfig map[string]struct {
		Fleet struct {
			Endpoint                 string `yaml:"endpoint,omitempty"`
			Username                 string `yaml:"username,omitempty"`
			Password                 string `yaml:"password,omitempty"`
			Strict_host_key_checking *bool  `yaml:"strict_host_key_checking,omitempty"`
			Sudo                     *bool  `yaml:"sudo,omitempty"`
			Driver                   string `yaml:"driver,omitempty"`
		} `yaml:"fleet,omitempty"`
	} `yaml:"overrideConfig,omitempty"`
}

type HomeStruct struct {
	fields data.Fields
	Path   string
	Config Config
}

func NewHome(path string) HomeStruct {
	fields := data.WithField("path", path)
	logs.WithFields(fields).Debug("loading home")

	var config Config
	if source, err := ioutil.ReadFile(path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &config)
		if err != nil {
			panic(err)
		}
	} else {
		logs.WithF(fields).WithField("file", "config.yml").Fatal("Cannot read configuration file")
	}
	return HomeStruct{
		fields: fields,
		Path:   path,
		Config: config,
	}
}

const PATH_LIST_MACHINES_CACHE = "/list-machines.cache"

func (h *HomeStruct) LoadMachinesCacheWithDate(env string) (string, time.Time) {
	logs.WithFields(h.fields).WithField("env", env).Debug("Loading list machines cache")
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
	logs.WithF(h.fields).WithField("env", env).Debug("save machines cache")
	if err := ioutil.WriteFile(h.Path+PATH_LIST_MACHINES_CACHE+"."+env, []byte(data), 0644); err != nil {
		logs.WithError(err).Warn("Cannot persist list-machines cache")
	}
}

func DefaultHomeRoot() string {
	home, err := homedir.Dir()
	if err != nil {
		logs.WithError(err).Fatal("Failed to find user home folder")
	}
	return home + "/.config"
}
