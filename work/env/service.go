package env

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/utils"
	"github.com/blablacar/green-garden/work/env/service"
	"github.com/coreos/etcd/client"
	"github.com/juju/errors"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type Service struct {
	env        spec.Env
	path       string
	name       string
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
		name:     name,
		env:      env,
		lockPath: "/ggn-lock/" + name + "/lock",
	}
	service.loadManifest()
	service.loadAttributes()
	return service
}

func (s *Service) LoadUnit(name string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", name, s)
	return unit
}

func (s *Service) ListUnits() []string {
	res := []string{}
	if s.manifest.Nodes[0][spec.NODE_HOSTNAME].(string) == "*" {
		machines := s.env.ListMachineNames()
		for _, node := range machines {
			res = append(res, s.UnitName(node))
		}
	} else {
		for _, node := range s.manifest.Nodes {
			res = append(res, s.UnitName(node[spec.NODE_HOSTNAME].(string)))
		}
	}
	return res
}

func (s *Service) Check() {
	unitNames := s.ListUnits()
	for _, unit := range unitNames {
		logUnit := s.log.WithField("unit", unit)
		localContent, err := s.LoadUnit(unit).GetUnitContentAsFleeted()
		if err != nil {
			logUnit.WithError(err).Error("Cannot read unit file")
			continue
		}
		remoteContent, err := s.GetFleetUnitContent(unit)
		if err != nil {
			logUnit.WithError(err).Error("Cannot read unit file")
			continue
		}

		if localContent != remoteContent {
			logUnit.Error("Unit is not up to date")
			logUnit.WithField("source", "fleet").Debug(remoteContent)
			logUnit.WithField("source", "file").Debug(localContent)
		}
	}

}

func (s *Service) GetFleetUnitContent(unit string) (string, error) {
	return s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "cat", unit)
}

/////////////////////////////////////////////////

func (s *Service) LockRelease() {
	kapi := s.env.EtcdClient()
	kapi.Delete(context.Background(), s.lockPath, nil)
}

func (s *Service) LockService(ttlSecond time.Duration, message string) {
	kapi := s.env.EtcdClient()
	resp, err := kapi.Get(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		s.log.WithError(cerr).Fatal("Server error reading on fleet")
	} else if err != nil {
		_, err := kapi.Set(context.Background(), s.lockPath, message, &client.SetOptions{TTL: ttlSecond})
		if err != nil {
			s.log.WithError(err).Fatal("Cannot write lock")
		}
	} else {
		s.log.WithField("message", resp.Node.Value).
			WithField("ttl", resp.Node.TTLDuration().String()).
			Fatal("Service is already locked")
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

func (s *Service) loadUnitTemplate() (*Templating, error) {
	path := s.path + spec.PATH_UNIT_TEMPLATE
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Annotate(err, "Cannot read unit template file")
	}
	template := NewTemplating(s.name, string(source))
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
