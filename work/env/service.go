package env

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/spec"
	"github.com/blablacar/ggn/utils"
	"github.com/blablacar/ggn/work/env/service"
	"github.com/coreos/etcd/client"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"
)

type Service struct {
	env            spec.Env
	path           string
	Name           string
	hasTimer       bool
	manifest       spec.ServiceManifest
	nodesAsJsonMap []interface{}
	log            log.Entry
	lockPath       string
	attributes     map[string]interface{}
	generated      bool
	generatedMutex *sync.Mutex
	units          map[string]*service.Unit
	unitsMutex     *sync.Mutex
	aciList        string
	aciListMutex   *sync.Mutex
}

func NewService(path string, name string, env spec.Env) *Service {
	l := env.GetLog()

	hasTimer := false
	if _, err := os.Stat(path + "/" + name + spec.PATH_UNIT_TIMER_TEMPLATE); err == nil {
		hasTimer = true
	}

	service := &Service{
		units:          map[string]*service.Unit{},
		unitsMutex:     &sync.Mutex{},
		aciListMutex:   &sync.Mutex{},
		generatedMutex: &sync.Mutex{},
		hasTimer:       hasTimer,
		log:            *l.WithField("service", name),
		path:           path + "/" + name,
		Name:           name,
		env:            env,
		lockPath:       "/ggn-lock/" + name + "/lock",
	}

	service.log.Debug("New Service")

	service.loadManifest()
	service.loadAttributes()
	service.prepareNodesAsJsonMap()
	return service
}

func (s *Service) prepareNodesAsJsonMap() {
	tmpRes, err := utils.TransformYamlToJson(s.manifest.Nodes)
	var res []interface{} = tmpRes.([]interface{})
	if err != nil {
		s.log.WithError(err).Fatal("Cannot transform yaml to json")
	}

	if res[0].(map[string]interface{})[spec.NODE_HOSTNAME].(string) == "*" {
		if len(res) > 1 {
			s.log.Fatal("You cannot mix all nodes with single node. Yet ?")
		}

		newNodes := *new([]interface{})
		machines, err := s.env.ListMachineNames()
		if err != nil {
			s.log.WithError(err).Fatal("Cannot list machines to generate units")
		}
		for _, machine := range machines {
			node := utils.CopyMap(res[0].(map[string]interface{}))
			node[spec.NODE_HOSTNAME] = machine
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

func (s *Service) GetEnv() spec.Env {
	return s.env
}

func (s *Service) GetLog() logrus.Entry {
	return s.log
}

func (s *Service) LoadUnit(name string) *service.Unit {
	s.unitsMutex.Lock()
	defer s.unitsMutex.Unlock()

	if val, ok := s.units[name]; ok {
		return val
	}
	var unit *service.Unit
	if strings.HasSuffix(name, spec.TYPE_TIMER.String()) {
		unit = service.NewUnit(s.path+"/units", name[:len(name)-len(spec.TYPE_TIMER.String())], spec.TYPE_TIMER, s)
	} else {
		unit = service.NewUnit(s.path+"/units", name, spec.TYPE_SERVICE, s)
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
		res = append(res, node.(map[string]interface{})[spec.NODE_HOSTNAME].(string))
		if s.hasTimer {
			res = append(res, node.(map[string]interface{})[spec.NODE_HOSTNAME].(string)+".timer")
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
		s.log.WithError(err).Fatal("Failed to list-units")
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
	s.log.Info("Unlocking")
	s.runHook(EARLY, command, "unlock")
	defer s.runHook(LATE, command, "unlock")

	kapi := s.env.EtcdClient()
	_, err := kapi.Delete(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		s.log.WithError(cerr).Panic("Cannot unlock service")
	}
}

func (s *Service) Lock(command string, ttl time.Duration, message string) {
	userAndHost := "[" + ggn.GetUserAndHost() + "] "
	message = userAndHost + message

	s.log.WithField("ttl", ttl).WithField("message", message).Info("locking")
	s.runHook(EARLY, command, "lock")
	defer s.runHook(LATE, command, "lock")

	kapi := s.env.EtcdClient()
	resp, err := kapi.Get(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		s.log.WithError(cerr).Fatal("Server error reading on fleet")
	} else if err != nil {
		_, err := kapi.Set(context.Background(), s.lockPath, message, &client.SetOptions{TTL: ttl})
		if err != nil {
			s.log.WithError(err).Fatal("Cannot write lock")
		}
	} else if strings.HasPrefix(resp.Node.Value, userAndHost) {
		_, err := kapi.Set(context.Background(), s.lockPath, message, &client.SetOptions{TTL: ttl})
		if err != nil {
			s.log.WithError(err).Fatal("Cannot write lock")
		}
	} else {
		s.log.WithField("message", resp.Node.Value).
			WithField("ttl", resp.Node.TTLDuration().String()).
			Fatal("Service is already locked")
	}
}

/////////////////////////////////////////////////

type Action int

const (
	ACTION_YES Action = iota
	ACTION_SKIP
	ACTION_DIFF
	ACTION_QUIT
)

const EARLY = true
const LATE = false

func (s *Service) runHook(isEarly bool, command string, action string) {
	out, err := json.Marshal(s.attributes)
	if err != nil {
		s.log.WithError(err).Panic("Cannot marshall attributes")
	}

	info := spec.HookInfo{
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
	files, err := utils.AttributeFiles(s.path + spec.PATH_ATTRIBUTES)
	if err != nil {
		s.log.WithError(err).WithField("path", s.path+spec.PATH_ATTRIBUTES).Panic("Cannot load Attributes files")
	}
	attr = attributes.MergeAttributesFilesForMap(attr, files)
	s.attributes = attr
	s.log.WithField("attributes", s.attributes).Debug("Attributes loaded")
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
	return s.path + spec.PATH_SERVICE_MANIFEST
}

func (s *Service) loadManifest() {
	manifest := spec.ServiceManifest{}
	path := s.manifestPath()
	source, err := ioutil.ReadFile(path)
	if err != nil {
		s.log.WithError(err).WithField("path", path).Warn("Cannot find manifest for service")
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		s.log.WithError(err).Fatal("Cannot Read service manifest")
	}

	if manifest.ConcurrentUpdater == 0 {
		manifest.ConcurrentUpdater = 1
	}

	s.log.WithField("manifest", manifest).Debug("Manifest loaded")
	s.manifest = manifest
}
