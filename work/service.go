package work

import (
	"encoding/json"
	"fmt"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/utils"
	"github.com/coreos/etcd/client"
	"github.com/juju/errors"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type Service struct {
	fields         data.Fields
	env            Env
	path           string
	Name           string
	hasTimer       bool
	manifest       ServiceManifest
	nodesAsJsonMap []interface{}
	lockPath       string
	attributes     map[string]interface{}
	generated      bool
	generatedMutex *sync.Mutex
	units          map[string]*Unit
	unitsMutex     *sync.Mutex
	aciList        string
	aciListMutex   *sync.Mutex
}

func NewService(path string, name string, env Env) *Service {
	l := env.GetFields()

	hasTimer := false
	if _, err := os.Stat(path + "/" + name + PATH_UNIT_TIMER_TEMPLATE); err == nil {
		hasTimer = true
	}

	service := &Service{
		units:          map[string]*Unit{},
		unitsMutex:     &sync.Mutex{},
		aciListMutex:   &sync.Mutex{},
		generatedMutex: &sync.Mutex{},
		hasTimer:       hasTimer,
		fields:         l.WithField("service", name),
		path:           path + "/" + name,
		Name:           name,
		env:            env,
		lockPath:       "/ggn-lock/" + name + "/lock",
	}

	logs.WithFields(service.fields).Debug("New service")

	service.loadManifest()
	service.loadAttributes()
	service.prepareNodesAsJsonMap()
	return service
}

func (s *Service) prepareNodesAsJsonMap() {
	if s.manifest.Nodes == nil || len(s.manifest.Nodes.([]interface{})) == 0 {
		logs.WithFields(s.fields).Warn("No nodes defined in service")
		return
	}
	tmpRes, err := utils.TransformYamlToJson(s.manifest.Nodes)
	var res []interface{} = tmpRes.([]interface{})
	if err != nil {
		logs.WithEF(err, s.fields).Fatal("Cannot transform yaml to json")
	}

	if res[0].(map[string]interface{})[NODE_HOSTNAME].(string) == "*" {
		if len(res) > 1 {
			logs.WithFields(s.fields).Fatal("You cannot mix all nodes with single node. Yet ?")
		}

		newNodes := *new([]interface{})
		machines, err := s.env.ListMachineNames()
		if err != nil {
			logs.WithEF(err, s.fields).Fatal("Cannot list machines to generate units")
		}
		for _, machine := range machines {
			node := utils.CopyMap(res[0].(map[string]interface{}))
			node[NODE_HOSTNAME] = machine
			newNodes = append(newNodes, node)
		}
		res = newNodes
	}
	s.nodesAsJsonMap = res
}

func (s *Service) HasTimer() bool {
	return s.hasTimer
}

func (s *Service) GetAttributes() map[string]interface{} {
	return s.attributes
}

func (s *Service) GetName() string {
	return s.Name
}

func (s *Service) GetEnv() Env {
	return s.env
}

func (s *Service) GetFields() data.Fields {
	return s.fields
}

func (s *Service) LoadUnit(name string) *Unit {
	s.unitsMutex.Lock()
	defer s.unitsMutex.Unlock()

	if val, ok := s.units[name]; ok {
		return val
	}
	var unit *Unit
	if strings.HasSuffix(name, TYPE_TIMER.String()) {
		unit = NewUnit(s.path+"/units", name[:len(name)-len(TYPE_TIMER.String())], TYPE_TIMER, s)
	} else {
		unit = NewUnit(s.path+"/units", name, TYPE_SERVICE, s)
	}
	s.units[name] = unit
	return unit
}

func (s *Service) Diff() {
	s.Generate()
	for _, unitName := range s.ListUnits() {
		unit := s.LoadUnit(unitName)
		unit.Diff("service/diff")
	}
}

func (s *Service) ListUnits() []string {
	res := []string{}
	for _, node := range s.nodesAsJsonMap {
		res = append(res, node.(map[string]interface{})[NODE_HOSTNAME].(string))
		if s.hasTimer {
			res = append(res, node.(map[string]interface{})[NODE_HOSTNAME].(string)+".timer")
		}
	}
	return res
}

func (s *Service) GetFleetUnitContent(unit string) (string, error) { //TODO this method should be in unit
	stdout, stderr, err := s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "cat", unit)
	if err != nil && stderr == "Unit "+unit+" not found" {
		return "", nil
	}
	return stdout, err
}

func (s *Service) FleetListUnits(command string) {
	stdout, _, err := s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-units", "--full", "--no-legend")
	if err != nil {
		logs.WithEF(err, s.fields).Fatal("Failed to list-units")
	}

	unitStatuses := strings.Split(stdout, "\n")
	prefix := s.env.GetName() + "_" + s.Name + "_"
	for _, unitStatus := range unitStatuses {
		if strings.HasPrefix(unitStatus, prefix) {
			fmt.Println(unitStatus)
		}
	}
}

func (s *Service) Unlock(command string) {
	logs.WithFields(s.fields).Info("Unlocking")
	s.runHook(EARLY, command, "unlock")
	defer s.runHook(LATE, command, "unlock")

	kapi := s.env.EtcdClient()
	_, err := kapi.Delete(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		logs.WithEF(cerr, s.fields).Fatal("Cannot unlock service")
	}
}

func (s *Service) Lock(command string, ttl time.Duration, message string) {
	userAndHost := "[" + ggn.GetUserAndHost() + "] "
	message = userAndHost + message

	logs.WithFields(s.fields).WithField("ttl", ttl).WithField("message", message).Info("Locking")
	s.runHook(EARLY, command, "lock")
	defer s.runHook(LATE, command, "lock")

	kapi := s.env.EtcdClient()
	resp, err := kapi.Get(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		logs.WithEF(cerr, s.fields).Fatal("Server error reading on fleet")
	} else if err != nil {
		_, err := kapi.Set(context.Background(), s.lockPath, message, &client.SetOptions{TTL: ttl})
		if err != nil {
			logs.WithEF(cerr, s.fields).Fatal("Cannot write lock")
		}
	} else if strings.HasPrefix(resp.Node.Value, userAndHost) {
		_, err := kapi.Set(context.Background(), s.lockPath, message, &client.SetOptions{TTL: ttl})
		if err != nil {
			logs.WithEF(cerr, s.fields).Fatal("Cannot write lock")
		}
	} else {
		logs.WithFields(s.fields).WithField("message", resp.Node.Value).
			WithField("ttl", resp.Node.TTLDuration().String()).
			Fatal("Service is already locked")
	}
}

/////////////////////////////////////////////////

func (s *Service) runHook(isEarly bool, command string, action string) {
	out, err := json.Marshal(s.attributes)
	if err != nil {
		logs.WithEF(err, s.fields).Fatal("Cannot marshall attributes")
	}

	info := HookInfo{
		Service:    s,
		Action:     "service/" + action,
		Command:    command,
		Attributes: string(out),
	}
	if isEarly {
		s.GetEnv().RunEarlyHook(info)
	} else {
		s.GetEnv().RunLateHook(info)
	}

}

func (s *Service) loadAttributes() {
	attr := utils.CopyMap(s.env.GetAttributes())
	files, err := utils.AttributeFiles(s.path + PATH_ATTRIBUTES)
	if err != nil {
		logs.WithEF(err, s.fields).WithField("path", s.path+PATH_ATTRIBUTES).Fatal("Cannot load Attributes files")
	}
	attr = attributes.MergeAttributesFilesForMap(attr, files)
	s.attributes = attr
	logs.WithFields(s.fields).WithField("attributes", s.attributes).Debug("Attributes loaded")
}

func (s *Service) loadUnitTemplate(filename string) (*utils.Templating, error) {
	path := s.path + filename
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot read unit template file")
	}
	template := utils.NewTemplating(s.Name, string(source))
	return template, template.Parse()
}

func (s *Service) manifestPath() string {
	return s.path + PATH_SERVICE_MANIFEST
}

func (s *Service) loadManifest() {
	manifest := ServiceManifest{}
	path := s.manifestPath()
	source, err := ioutil.ReadFile(path)
	if err != nil {
		logs.WithEF(err, s.fields).WithField("path", path).Warn("Cannot find manifest for service")
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		logs.WithEF(err, s.fields).Fatal("Cannot Read service manifest")
	}

	if manifest.ConcurrentUpdater == 0 {
		manifest.ConcurrentUpdater = 1
	}

	logs.WithFields(s.fields).WithField("manifest", manifest).Debug("Manifest loaded")
	s.manifest = manifest
}
