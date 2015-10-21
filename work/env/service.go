package env

import (
	"bytes"
	"encoding/json"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/work/env/service"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type Service struct {
	path     string
	name     string
	manifest spec.ServiceManifest
}

func NewService(path string, name string) *Service {
	service := new(Service)
	service.path = path + "/" + name
	service.name = name
	service.loadManifest()
	return service
}

//////////////////////////////////////

func (s Service) LoadUnit(name string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", name)
	return unit
}

func (s Service) GenerateUnits(envAttributePath string, envName string) {
	tmpl := s.loadUnitTemplate()

	var acis string
	for _, aci := range s.manifest.Containers {
		aci, err := aci.FullyResolved()
		log.Error(aci)
		if err != nil {
			panic("Cannot resolve aci" + err.Error())
		}
		acis += aci.String() + " "
	}

	for i, node := range s.manifest.Nodes {

		if node[spec.NODE_HOSTNAME].(string) == "" {
			log.Error("hostname is mandatory in node informations :", s.manifestPath(), "node["+string(i)+"]")
			os.Exit(1)
		}
		log.Debug("Processing node : " + s.name + ":" + node[spec.NODE_HOSTNAME].(string))

		unitName := envName + "_" + s.name + "_" + node[spec.NODE_HOSTNAME].(string) + ".service"
		log.Trace("Unit name is :" + unitName)

		attributes := attributes.MergeAttributes(mergeAttributesDirectories(envAttributePath, s.path+spec.PATH_ATTRIBUTES))
		attributes["node"] = node //TODO this should be merged
		attributes["node"].(map[string]interface{})["acis"] = acis
		out, err := json.Marshal(attributes)
		attributes["attributes"] = string(out)

		var b bytes.Buffer
		err = tmpl.Execute(&b, attributes)
		if err != nil {
			log.Error("Failed to run templating for unit "+unitName, err)
			os.Exit(1)
		}
		ioutil.WriteFile(s.path+"/units"+"/"+unitName, b.Bytes(), 0644)

	}
}

///////////////////////////////////////

func mergeAttributesDirectories(envAttributesPath string, serviceAttributesPath string) []string {
	res := []string{}

	{
		in := attributes.NewInputs(envAttributesPath)
		// initialize input files list
		err := in.ListFiles()
		if err != nil {
			panic(err)
		}

		for _, file := range in.Files {
			res = append(res, in.Directory+file)
		}
	}
	{
		in := attributes.NewInputs(serviceAttributesPath)
		// initialize input files list
		err := in.ListFiles()
		if err != nil {
			panic(err)
		}

		for _, file := range in.Files {
			res = append(res, in.Directory+file)
		}
	}
	return res
}

func (s *Service) loadUnitTemplate() *Templating {
	path := s.path + spec.PATH_UNIT_TEMPLATE
	source, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warn("Cannot read unit template file '" + s.name + "' : " + path)
	}
	template := NewTemplating(s.name, string(source))
	template.Parse()
	return template
}

func (s Service) manifestPath() string {
	return s.path + spec.PATH_SERVICE_MANIFEST
}

func (s *Service) loadManifest() {
	manifest := spec.ServiceManifest{}
	path := s.manifestPath()
	log.Trace("Service manifest is at : " + path)
	source, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warn("Cannot find manifest for service '" + s.name + "' : " + path)
	}
	err = yaml.Unmarshal([]byte(source), &manifest)
	if err != nil {
		panic(err)
	}
	s.manifest = manifest
}
