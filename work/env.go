package work

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	cntUtils "github.com/blablacar/cnt/utils"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/spec"
	"github.com/blablacar/ggn/utils"
	"github.com/blablacar/ggn/work/env"
	"github.com/coreos/etcd/client"
	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

const PATH_SERVICES = "/services"

type Config struct {
	Fleet struct {
		Endpoint                 string `yaml:"endpoint,omitempty"`
		Username                 string `yaml:"username,omitempty"`
		Password                 string `yaml:"password,omitempty"`
		Strict_host_key_checking bool   `yaml:"strict_host_key_checking,omitempty"`
		Sudo                     bool   `yaml:"sudo,omitempty"`
	} `yaml:"fleet,omitempty"`
}

type Env struct {
	path          string
	name          string
	log           logrus.Entry
	attributes    map[string]interface{}
	config        Config
	services      map[string]*env.Service
	servicesMutex *sync.Mutex
}

func NewEnvironment(root string, name string) *Env {
	log := *log.WithField("env", name)
	path := root + "/" + name
	_, err := ioutil.ReadDir(path)
	if err != nil {
		log.WithError(err).Error("Cannot read env directory")
	}

	env := &Env{
		services:      map[string]*env.Service{},
		servicesMutex: &sync.Mutex{},
		path:          path,
		name:          name,
		log:           log,
		config:        Config{},
	}
	env.loadAttributes()
	env.loadConfig()
	return env
}

func (e Env) GetName() string {
	return e.name
}

func (e Env) GetLog() logrus.Entry {
	return e.log
}

func (e Env) GetAttributes() map[string]interface{} {
	return e.attributes
}

func (e Env) FleetctlListUnits() {
	stdout, _, err := e.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-units", "--full", "--no-legend")
	if err != nil {
		e.log.WithError(err).Fatal("Failed to list-units")
	}

	unitStatuses := strings.Split(stdout, "\n")
	for _, unitStatus := range unitStatuses {
		fmt.Println(unitStatus)
	}
}

func (e Env) LoadService(name string) *env.Service {
	e.servicesMutex.Lock()
	defer e.servicesMutex.Unlock()

	if val, ok := e.services[name]; ok {
		return val
	}

	service := env.NewService(e.path+"/services", name, e)
	e.services[name] = service
	return service
}

func (e Env) attributesDir() string {
	return e.path + spec.PATH_ATTRIBUTES
}

func (e *Env) loadConfig() {
	if source, err := ioutil.ReadFile(e.path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &e.config)
		if err != nil {
			panic(err)
		}
	}
}

func (e *Env) loadAttributes() {
	files, err := utils.AttributeFiles(e.path + spec.PATH_ATTRIBUTES)
	if err != nil {
		e.log.WithError(err).WithField("path", e.path+spec.PATH_ATTRIBUTES).Panic("Cannot load Attributes files")
	}
	e.attributes = attributes.MergeAttributesFiles(files)
	e.log.WithField("attributes", e.attributes).Debug("Attributes loaded")
}

func (e Env) ListServices() []string {
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
		if _, err := os.Stat(path + "/" + file.Name() + spec.PATH_SERVICE_MANIFEST); os.IsNotExist(err) {
			continue
		}

		services = append(services, file.Name())
	}
	return services
}

var inMemoryNames []string

func (e Env) ListMachineNames() ([]string, error) {
	e.log.Debug("list machines")
	if inMemoryNames != nil {
		return inMemoryNames, nil
	}

	data, modification := ggn.Home.LoadMachinesCacheWithDate(e.name)
	if data == "" || modification.Add(12*time.Hour).Before(time.Now()) {
		e.log.Debug("reloading list machines cache")
		datatmp, _, err := e.RunFleetCmdGetOutput("list-machines", "--fields=metadata", "--no-legend")
		if err != nil {
			return nil, errors.Annotate(err, "Cannot list-machines")
		}
		data = datatmp
		ggn.Home.SaveMachinesCache(e.name, data)
	}

	var names []string

	machines := strings.Split(data, "\n")
	for _, machine := range machines {
		metas := strings.Split(machine, ",")
		for _, meta := range metas {
			elem := strings.Split(meta, "=")
			if elem[0] == "name" { // TODO this is specific to blablacar's metadata ??
				names = append(names, elem[1])
			}
		}
	}
	inMemoryNames = names
	return names, nil
}

const PATH_HOOKS = "/hooks"

func (e Env) RunEarlyHook(info spec.HookInfo) {
	e.runHook("/early", info)
}

func (e Env) RunLateHook(info spec.HookInfo) {
	e.runHook("/late", info)
}

func (e Env) runHook(path string, info spec.HookInfo) {
	e.log.WithField("path", path).WithField("info", info).Debug("Running hook")
	files, err := ioutil.ReadDir(e.path + PATH_HOOKS + path)
	if err != nil {
		log.WithError(err).Debug("Cannot read hood directory")
		return
	}

	envs := map[string]string{}
	envs["ENV"] = e.name
	envs["COMMAND"] = info.Command
	if info.Unit != nil {
		envs["UNIT"] = info.Unit.GetName()
	}
	if info.Service != nil {
		envs["SERVICE"] = info.Service.GetName()
	}
	envs["WHO"] = ggn.GetUserAndHost()
	envs["ACTION"] = info.Action
	envs["ATTRIBUTES"] = info.Attributes

	for _, f := range files {
		if !f.IsDir() {
			hookLog := log.WithField("name", f.Name())

			args := []string{e.path + PATH_HOOKS + path + "/" + f.Name()}
			for key, val := range envs {
				args = append([]string{key + "='" + strings.Replace(val, "'", "'\"'\"'", -1) + "'"}, args...)
			}

			hookLog.Debug("Running Hook")
			if err := cntUtils.ExecCmd("bash", "-c", strings.Join(args, " ")); err != nil {
				hookLog.Fatal("Hook status is failed")
			}
		}
	}
}

const FLEETCTL_ENDPOINT = "FLEETCTL_ENDPOINT"
const FLEETCTL_SSH_USERNAME = "FLEETCTL_SSH_USERNAME"
const FLEETCTL_STRICT_HOST_KEY_CHECKING = "FLEETCTL_STRICT_HOST_KEY_CHECKING"
const FLEETCTL_SUDO = "FLEETCTL_SUDO"

func (e Env) RunFleetCmd(args ...string) error {
	_, _, err := e.runFleetCmdInternal(false, args)
	return err
}

func (e Env) RunFleetCmdGetOutput(args ...string) (string, string, error) {
	return e.runFleetCmdInternal(true, args)
}

func (e Env) EtcdClient() client.KeysAPI {
	cfg := client.Config{
		Endpoints:               strings.Split(e.config.Fleet.Endpoint, ","),
		Username:                e.config.Fleet.Username,
		Password:                e.config.Fleet.Password,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		e.log.WithError(err).Fatal("Failed to create etcd client")
	}
	kapi := client.NewKeysAPI(c)
	return kapi
}

func (e Env) runFleetCmdInternal(getOutput bool, args []string) (string, string, error) {
	e.log.WithField("command", strings.Join(args, " ")).Debug("Running command on fleet")
	if e.config.Fleet.Endpoint == "" {
		return "", "", errors.New("Cannot find fleet.endpoint env config to call fleetctl")
	}

	envs := map[string]string{}
	envs[FLEETCTL_ENDPOINT] = e.config.Fleet.Endpoint
	if e.config.Fleet.Username != "" {
		envs[FLEETCTL_SSH_USERNAME] = e.config.Fleet.Username
	}
	envs[FLEETCTL_STRICT_HOST_KEY_CHECKING] = fmt.Sprintf("%t", e.config.Fleet.Strict_host_key_checking)
	envs[FLEETCTL_SUDO] = fmt.Sprintf("%t", e.config.Fleet.Sudo)

	args = append([]string{"fleetctl"}, args...)
	for key, val := range envs {
		args = append([]string{key + "='" + val + "'"}, args...)
	}

	var stdout string
	var stderr string
	var err error
	if getOutput {
		stdout, stderr, err = cntUtils.ExecCmdGetStdoutAndStderr("bash", "-c", strings.Join(args, " "))
	} else {
		err = cntUtils.ExecCmd("bash", "-c", strings.Join(args, " "))
	}
	return stdout, stderr, err
}
