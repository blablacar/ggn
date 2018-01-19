package work

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	txttmpl "text/template"
	"time"

	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/utils"
	"github.com/coreos/etcd/client"
	"github.com/juju/errors"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	"gopkg.in/yaml.v2"
)

const PATH_SERVICES = "/services"

type Config struct {
	Fleet struct {
		Endpoint                 string `yaml:"endpoint,omitempty"`
		Username                 string `yaml:"username,omitempty"`
		Password                 string `yaml:"password,omitempty"`
		Strict_host_key_checking bool   `yaml:"strict_host_key_checking,omitempty"`
		Sudo                     bool   `yaml:"sudo,omitempty"`
		Driver                   string `yaml:"driver,omitempty"`
	} `yaml:"fleet,omitempty"`
	EnvName string `yaml:"envName,omitempty"`
}

type Env struct {
	path          string
	dirName       string
	fields        data.Fields
	attributes    map[string]interface{}
	config        Config
	services      map[string]*Service
	servicesMutex *sync.Mutex
	Partials      *txttmpl.Template
}

func NewEnvironment(root string, name string) *Env {
	fields := data.WithField("env", name)
	path := root + "/" + name
	_, err := ioutil.ReadDir(path)
	if err != nil {
		logs.WithEF(err, fields).Fatal("Cannot read env directory")
	}

	env := &Env{
		services:      map[string]*Service{},
		servicesMutex: &sync.Mutex{},
		path:          path,
		dirName:       name,
		fields:        fields,
		config:        Config{},
	}
	env.loadAttributes()
	env.loadConfig()
	env.loadPartials()
	return env
}

func (e Env) GetDirName() string {
	return e.dirName
}

func (e Env) GetEnvName() string {
	return e.config.EnvName
}

func (e Env) GetFields() data.Fields {
	return e.fields
}

func (e Env) GetAttributes() map[string]interface{} {
	return e.attributes
}

func (e Env) FleetctlListUnits() {
	stdout, _, err := e.RunFleetCmdGetOutput("list-units", "--full", "--no-legend")
	if err != nil {
		logs.WithEF(err, e.fields).Fatal("Failed to list-units")
	}

	unitStatuses := strings.Split(stdout, "\n")
	for _, unitStatus := range unitStatuses {
		fmt.Println(unitStatus)
	}
}

func (e Env) FleetctlListMachines() {
	stdout, _, err := e.RunFleetCmdGetOutput("list-machines", "--full", "--no-legend")
	if err != nil {
		logs.WithEF(err, e.fields).Fatal("Failed to list-machines")
	}

	machines := strings.Split(stdout, "\n")
	for _, machine := range machines {
		fmt.Println(machine)
	}
}

func (e Env) LoadService(name string) *Service {
	e.servicesMutex.Lock()
	defer e.servicesMutex.Unlock()

	if val, ok := e.services[name]; ok {
		return val
	}

	service := NewService(e.path+"/services", name, e)
	e.services[name] = service
	return service
}

func (e Env) attributesDir() string {
	return e.path + PATH_ATTRIBUTES
}

func (e *Env) loadConfig() {
	if source, err := ioutil.ReadFile(e.path + "/config.yml"); err == nil {
		err = yaml.Unmarshal([]byte(source), &e.config)
		if err != nil {
			panic(err)
		}
	}

	// backward compatibility with fleet < 1.0.0 : etcd as default driver
	if e.config.Fleet.Driver == "" {
		e.config.Fleet.Driver = "etcd"
	}
	if e.config.EnvName == "" {
		e.config.EnvName = e.dirName
	}

	src := strings.Split(e.config.Fleet.Endpoint, ",")
	dest := make([]string, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	e.config.Fleet.Endpoint = strings.Join(dest, ",")
}

func (e *Env) loadPartials() {
	if ok, err := common.IsDirEmpty(e.path + PATH_TEMPLATES); ok || err != nil {
		return
	}
	tmplDir, err := template.NewTemplateDir(e.path+PATH_TEMPLATES, "", false)
	if err != nil {
		logs.WithEF(err, e.fields).WithField("path", e.path+PATH_ATTRIBUTES).Fatal("Failed to load partial templating")
	}
	e.Partials = tmplDir.Partials
}

func (e *Env) loadAttributes() {
	files, err := utils.AttributeFiles(e.path + PATH_ATTRIBUTES)
	if err != nil {
		logs.WithEF(err, e.fields).WithField("path", e.path+PATH_ATTRIBUTES).Fatal("Cannot load attribute files")
	}
	files, err = e.addIncludeFiles(files)
	if err != nil {
		logs.WithEF(err, e.fields).WithField("path", e.path+PATH_ATTRIBUTES).Fatal("Cannot load include files")
	}

	e.attributes, err = utils.MergeAttributesFiles(files)
	if err != nil {
		logs.WithEF(err, e.fields).WithField("path", e.path+PATH_ATTRIBUTES).Fatal("Failed to merge attributes")
	}
	logs.WithFields(e.fields).WithField("attributes", e.attributes).Debug("Attributes loaded")
}

func (e *Env) addIncludeFiles(files []string) ([]string, error) {
	type includeFiles struct {
		Include []string
	}
	for _, file := range files {
		var f includeFiles
		yml, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(yml, &f)
		for _, inclusion := range f.Include {
			sepCount := strings.Count(inclusion, ":")
			if sepCount == 2 {
				fields := strings.Split(inclusion, ":")
				includeFile := strings.Replace(fields[2], ".", "/", -1) + ".yml"
				if fields[1] == "" {
					logs.WithField("include", inclusion).Fatal("Trying to include environment attributes from itself")
				} else { // env:prod-dc1:some.include
					includeFile = fmt.Sprintf("%v%v/%v%v/%v",
						ggn.Home.Config.WorkPath,
						PATH_ENV,
						fields[1],
						PATH_COMMON_ATTRIBUTES,
						includeFile,
					)
					files = append(files, includeFile)
				}
			} else { // some.global.include
				includeFile := strings.Replace(inclusion, ".", "/", -1) + ".yml"
				includeFile = fmt.Sprintf("%v%v/%v", ggn.Home.Config.WorkPath, PATH_COMMON_ATTRIBUTES, includeFile)
				files = append(files, includeFile)

			}
		}
	}
	return files, nil

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

		if _, err := os.Stat(path + "/" + file.Name() + PATH_SERVICE_MANIFEST); os.IsNotExist(err) {
			continue
		}

		services = append(services, file.Name())
	}
	return services
}

var inMemoryNames []string

func (e Env) ListMachineNames() ([]string, error) {
	logs.WithFields(e.fields).Debug("list machines")
	if inMemoryNames != nil {
		return inMemoryNames, nil
	}

	data, modification := ggn.Home.LoadMachinesCacheWithDate(e.dirName)
	if data == "" || modification.Add(12*time.Hour).Before(time.Now()) {
		logs.WithFields(e.fields).Debug("reloading list machines cache")
		datatmp, _, err := e.RunFleetCmdGetOutput("list-machines", "--fields=metadata", "--no-legend")
		if err != nil {
			return nil, errors.Annotate(err, "Cannot list-machines")
		}
		data = datatmp
		ggn.Home.SaveMachinesCache(e.dirName, data)
	}

	var names []string

	machines := strings.Split(data, "\n")
	for _, machine := range machines {
		metas := strings.Split(machine, ",")
		for _, meta := range metas {
			elem := strings.Split(meta, "=")
			if elem[0] == "dirName" {
				// TODO this is specific to blablacar's metadata ??
				names = append(names, elem[1])
			}
		}
	}
	inMemoryNames = names
	return names, nil
}

const PATH_HOOKS = "/hooks"

func (e Env) RunEarlyHookFatal(info HookInfo) {
	if err := e.runHookAndGetNumRun("/early", info); err != nil {
		logs.WithE(err).Fatal("hook failed")
	}
}

func (e Env) RunLateHookFatal(info HookInfo) {
	if err := e.runHookAndGetNumRun("/late", info); err != nil {
		logs.WithE(err).Fatal("hook failed")
	}
}

func (e Env) RunHook(info HookInfo) error {
	return e.runHookAndGetNumRun("/command", info)
}

func (e Env) runHookAndGetNumRun(path string, info HookInfo) error {
	logs.WithFields(e.fields).WithField("path", path).WithField("info", info).Debug("Running hook")
	files, err := ioutil.ReadDir(e.path + PATH_HOOKS + path)
	if err != nil {
		logs.WithEF(err, e.fields).Debug("Cannot read hook directory")
		return nil
	}

	envs := map[string]string{}
	envs["ENVDIR"] = e.dirName
	envs["ENV"] = e.config.EnvName
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
	envs["GGN_HOME_PATH"] = ggn.Home.Path
	envs["LOG_LEVEL"] = logs.GetLevel().String()

	for _, f := range files {
		if !f.IsDir() {
			hookFields := data.WithField("dirName", f.Name())

			args := []string{e.path + PATH_HOOKS + path + "/" + f.Name()}
			for key, val := range envs {
				args = append([]string{key + "='" + strings.Replace(val, "'", "'\"'\"'", -1) + "'"}, args...)
			}

			logs.WithFields(hookFields).Debug("Running Hook")
			if err := common.ExecCmd("bash", "-c", strings.Join(args, " ")); err != nil {
				return errs.WithF(hookFields, "Hook status is failed")
			}
		}
	}
	return nil
}

const FLEETCTL_ENDPOINT = "FLEETCTL_ENDPOINT"
const FLEETCTL_SSH_USERNAME = "FLEETCTL_SSH_USERNAME"
const FLEETCTL_STRICT_HOST_KEY_CHECKING = "FLEETCTL_STRICT_HOST_KEY_CHECKING"
const FLEETCTL_SUDO = "FLEETCTL_SUDO"
const FLEETCTL_DRIVER = "FLEETCTL_DRIVER"

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
		logs.WithEF(err, e.fields).Fatal("Failed to create etcd client")
	}
	kapi := client.NewKeysAPI(c)
	return kapi
}

func (e Env) runFleetCmdInternal(getOutput bool, args []string) (string, string, error) {
	logs.WithF(e.fields).WithField("command", strings.Join(args, " ")).Debug("Running command on fleet")
	if e.config.Fleet.Endpoint == "" {
		return "", "", errors.New("Cannot find fleet.endpoint env config to call fleetctl")
	}

	envs := map[string]string{}
	envs[FLEETCTL_DRIVER] = e.config.Fleet.Driver
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
		stdout, stderr, err = common.ExecCmdGetStdoutAndStderr("bash", "-c", strings.Join(args, " "))
	} else {
		err = common.ExecCmd("bash", "-c", strings.Join(args, " "))
	}
	return stdout, stderr, err
}
