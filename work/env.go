package work

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/work/env"
	"io/ioutil"
	"os"
)

const PATH_SERVICES = "/services"

type Env struct {
	path       string
	name       string
	log        logrus.Entry
	attributes map[string]interface{}
}

func NewEnvironment(root string, name string) *Env {
	log := *log.WithField("env", name)
	path := root + "/" + name
	_, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithError(err).Error("Cannot read env directory")
	}

	env := &Env{
		path: path,
		name: name,
		log:  log,
	}
	env.attributes = env.loadAttributes()
	return env
}

func (e Env) LoadService(name string) *env.Service {
	return env.NewService(e.path+"/services", name)
}

func (e Env) attributesDir() string {
	return e.path + spec.PATH_ATTRIBUTES
}

func (e Env) loadAttributes() map[string]interface{} {
	attributes := attributes.MergeAttributes(e.attributeFiles())
	e.log.WithField("attributes", attributes).Debug("Attributes loaded")
	return attributes
}

func (e Env) attributeFiles() []string {
	res := []string{}

	in := attributes.NewInputs(e.path + spec.PATH_ATTRIBUTES)
	// initialize input files list
	err := in.ListFiles()
	if err != nil {
		panic(err)
	}

	for _, file := range in.Files {
		res = append(res, in.Directory+file)
	}
	return res
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

const FLEETCTL_ENDPOINT = "FLEETCTL_ENDPOINT"
const FLEETCTL_SSH_USERNAME = "FLEETCTL_SSH_USERNAME"
const FLEETCTL_STRICT_HOST_KEY_CHECKING = "FLEETCTL_STRICT_HOST_KEY_CHECKING"
const FLEETCTL_SUDO = "FLEETCTL_SUDO"

func (e Env) RunFleetCmd(args ...string) error {
	_, err := e.runFleetCmdInternal(false, args...)
	return err
}

func (e Env) RunFleetCmdGetOutput(args ...string) (string, error) {
	return e.runFleetCmdInternal(true, args...)
}

func (e Env) runFleetCmdInternal(getOutput bool, args ...string) (string, error) {
	endpoint := "http://" + e.attributes["fleet"].(map[string]interface{})["endpoint"].(string)
	username := e.attributes["fleet"].(map[string]interface{})["username"].(string)
	strict_host_key_checking := e.attributes["fleet"].(map[string]interface{})["strict_host_key_checking"].(bool)
	sudo := e.attributes["fleet"].(map[string]interface{})["sudo"].(bool)

	os.Setenv(FLEETCTL_ENDPOINT, endpoint)
	os.Setenv(FLEETCTL_SSH_USERNAME, username)
	os.Setenv(FLEETCTL_STRICT_HOST_KEY_CHECKING, fmt.Sprintf("%t", strict_host_key_checking))
	os.Setenv(FLEETCTL_SUDO, fmt.Sprintf("%t", sudo))

	var out string
	var err error
	if getOutput {
		out, err = utils.ExecCmdGetOutput("fleetctl", args...)
	} else {
		err = utils.ExecCmd("fleetctl", args...)
	}
	//	if err != nil {
	//		e.log.WithError(err).
	//		WithField("FLEETCTL_ENDPOINT", endpoint).
	//		WithField("FLEETCTL_SSH_USERNAME", username).
	//		WithField("FLEETCTL_STRICT_HOST_KEY_CHECKING", strict_host_key_checking).
	//		WithField("FLEETCTL_SUDO", sudo).
	//		Error("Fleetctl command failed")
	//	}

	os.Setenv(FLEETCTL_ENDPOINT, "")
	os.Setenv(FLEETCTL_SSH_USERNAME, "")
	os.Setenv(FLEETCTL_STRICT_HOST_KEY_CHECKING, "")
	os.Setenv(FLEETCTL_SUDO, "")
	return out, err
}
