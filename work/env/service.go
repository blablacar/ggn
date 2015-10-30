package env

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/utils"
	"github.com/blablacar/green-garden/work/env/service"
	"github.com/juju/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Service struct {
	env        spec.Env
	path       string
	name       string
	manifest   spec.ServiceManifest
	log        log.Entry
	attributes map[string]interface{}
}

func NewService(path string, name string, env spec.Env) *Service {
	l := env.GetLog()
	service := &Service{
		log:  *l.WithField("service", name),
		path: path + "/" + name,
		name: name,
		env:  env,
	}
	service.loadManifest()
	service.loadAttributes()
	return service
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

/////////////////////////////////////////////////

func (s *Service) LoadUnit(name string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", name, s)
	return unit
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
