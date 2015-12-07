package env

import (
	"bufio"
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
	"github.com/mgutz/ansi"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Service struct {
	env        spec.Env
	path       string
	Name       string
	manifest   spec.ServiceManifest
	log        log.Entry
	lockPath   string
	attributes map[string]interface{}
}

func NewService(path string, name string, env spec.Env) *Service {
	l := env.GetLog()
	service := &Service{
		log:      *l.WithField("service", name),
		path:     path + "/" + name,
		Name:     name,
		env:      env,
		lockPath: "/ggn-lock/" + name + "/lock",
	}
	service.loadManifest()
	service.loadAttributes()
	return service
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

func (s *Service) LoadUnit(hostname string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", hostname, s)
	return unit
}

func (s *Service) Diff() {
	s.Generate(nil)
	units, err := s.ListUnits()
	if err != nil {
		s.log.WithError(err).Fatal("Cannot list units to run diff")
	}
	for _, unitName := range units {
		unit := s.LoadUnit(unitName)
		unit.Diff("service/diff")
	}
}

func (s *Service) ListUnits() ([]string, error) {
	res := []string{}
	if len(s.manifest.Nodes) == 0 {
		return res, nil
	}

	if s.manifest.Nodes[0][spec.NODE_HOSTNAME].(string) == "*" {
		machines, err := s.env.ListMachineNames()
		if err != nil {
			return nil, errors.Annotate(err, "Cannot generate unit list from list of machines")
		}
		for _, node := range machines {
			res = append(res, node)
		}
	} else {
		for _, node := range s.manifest.Nodes {
			res = append(res, node[spec.NODE_HOSTNAME].(string))
		}
	}
	return res, nil
}

func (s *Service) GetFleetUnitContent(unit string) (string, error) { //TODO this method should be in unit
	stdout, stderr, err := s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "cat", unit)
	if err != nil && stderr == "Unit "+unit+" not found" {
		return "", nil
	}
	return stdout, err
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

func (s *Service) askToProcessService(index int, unit *service.Unit) Action {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Update unit " + ansi.LightGreen + unit.Name + ansi.Reset + " ? : (yes,skip,diff,quit) ")
		text, _ := reader.ReadString('\n')
		t := strings.ToLower(strings.Trim(text, " \n"))
		if t == "o" || t == "y" || t == "ok" || t == "yes" {
			return ACTION_YES
		}
		if t == "s" || t == "skip" {
			return ACTION_SKIP
		}
		if t == "d" || t == "diff" {
			return ACTION_DIFF
		}
		if t == "q" || t == "quit" {
			return ACTION_QUIT
		}
		continue
	}
	return ACTION_QUIT
}

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

func (s *Service) loadUnitTemplate() (*utils.Templating, error) {
	path := s.path + spec.PATH_UNIT_TEMPLATE
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot read unit template file")
	}
	template := utils.NewTemplating(s.Name, string(source))
	template.Parse()
	return template, nil
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
	s.manifest = manifest
}
