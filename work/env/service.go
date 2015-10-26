package env

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/discovery"
	"github.com/appc/spec/schema"
	"github.com/blablacar/attributes-merger/attributes"
	cntspec "github.com/blablacar/cnt/spec"
	"github.com/blablacar/green-garden/Godeps/_workspace/src/github.com/juju/errors"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/work/env/service"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type Service struct {
	path     string
	name     string
	manifest spec.ServiceManifest
	log      *log.Entry
}

func NewService(path string, name string) *Service {
	service := &Service{
		log:  log.WithField("service", name),
		path: path + "/" + name,
		name: name,
	}
	service.loadManifest()
	return service
}

//////////////////////////////////////

func (s Service) LoadUnit(name string) *service.Unit {
	unit := service.NewUnit(s.path+"/units", name)
	return unit
}

func (s Service) GenerateUnits(envAttributePath string, envName string) {
	s.log.Debug("Generating units")

	tmpl, err := s.loadUnitTemplate()
	if err != nil {
		s.log.WithError(err).Error("Cannot load units template")
		return
	}

	var acis string
	for _, aci := range s.manifest.Containers {
		if strings.HasPrefix(aci.ShortName(), "pod-") {
			logAci := s.log.WithField("aci", aci)

			app, err := discovery.NewAppFromString(aci.String())
			if app.Labels["os"] == "" {
				app.Labels["os"] = "linux"
			}
			if app.Labels["arch"] == "" {
				app.Labels["arch"] = "amd64"
			}

			endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
			if err != nil {
				logAci.WithError(err).Fatal("pod discovery failed")
			}

			url := endpoint.ACIEndpoints[0].ACI
			url = strings.Replace(url, "=aci", "=pod", 1) // TODO this is nexus specific

			logUrl := logAci.WithField("url", url)
			response, err := http.Get(url)
			if err != nil {
				logUrl.WithError(err).Fatal("Cannot get pod manifest content")
			} else {
				defer response.Body.Close()
				contents, err := ioutil.ReadAll(response.Body)
				if err != nil {
					logUrl.WithError(err).Fatal("Cannot read pod manifest content")
				}
				pod := schema.BlankPodManifest()
				pod.UnmarshalJSON(contents)

				for _, podAci := range pod.Apps {
					version, ok := podAci.Image.Labels.Get("version")
					if !ok {
						version = "latest"
					}
					fullname := cntspec.NewACFullName(podAci.Image.Name.String() + ":" + version)

					resolved, err := fullname.FullyResolved()
					if err != nil {
						logAci.WithError(err).Fatal("Cannot fully resolve ACI")
					}
					acis += resolved.String() + " "
				}
			}

		} else {
			aci, err := aci.FullyResolved()
			if err != nil {
				s.log.WithError(err).WithField("aci", aci).Error("Cannot resolve aci")
				return
			}
			acis += aci.String() + " "
		}
	}

	for i, node := range s.manifest.Nodes {

		if node[spec.NODE_HOSTNAME].(string) == "" {
			s.log.Error("hostname is mandatory in node informations :", s.manifestPath(), "node["+string(i)+"]")
			os.Exit(1)
		}
		s.log.Debug("Processing node :" + node[spec.NODE_HOSTNAME].(string))

		unitName := envName + "_" + s.name + "_" + node[spec.NODE_HOSTNAME].(string) + ".service"
		s.log.Debug("Unit name is :" + unitName)

		attributes := attributes.MergeAttributes(mergeAttributesDirectories(envAttributePath, s.path+spec.PATH_ATTRIBUTES))
		attributes["node"] = node //TODO this should be merged
		attributes["node"].(map[string]interface{})["acis"] = acis
		out, err := json.Marshal(attributes)
		attributes["attributes"] = strings.Replace(string(out), "\\\"", "\\\\\\\"", -1)

		var b bytes.Buffer
		err = tmpl.Execute(&b, attributes)
		if err != nil {
			s.log.Error("Failed to run templating for unit "+unitName, err)
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

func (s Service) manifestPath() string {
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
