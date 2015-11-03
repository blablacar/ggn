package env

import (
	"bytes"
	"encoding/json"
	"github.com/appc/spec/discovery"
	"github.com/appc/spec/schema"
	cntspec "github.com/blablacar/cnt/spec"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func (s Service) GenerateUnits() {
	s.log.Debug("Generating units")

	tmpl, err := s.loadUnitTemplate()
	if err != nil {
		s.log.WithError(err).Error("Cannot load units template")
		return
	}

	if len(s.manifest.Nodes) == 0 {
		s.log.Error("No node to process in manifest")
		return
	}

	nodes := s.manifest.Nodes
	if s.manifest.Nodes[0][spec.NODE_HOSTNAME].(string) == "*" {
		if len(s.manifest.Nodes) > 1 {
			s.log.Error("You cannot mix all nodes with single node. Yet ?")
			return
		}

		newNodes := *new([]map[string]interface{})
		machines := s.env.ListMachineNames()
		for _, machine := range machines {
			node := utils.CopyMap(s.manifest.Nodes[0])
			node[spec.NODE_HOSTNAME] = machine
			newNodes = append(newNodes, node)
		}

		nodes = newNodes
	}

	acis, err := s.prepareAciList()
	if err != nil {
		s.log.WithError(err).Error("Cannot prepare aci list")
		return
	}

	for i, node := range nodes {
		s.writeUnit(i, node, tmpl, acis)
	}
}

func (s Service) writeUnit(i int, node map[string]interface{}, tmpl *Templating, acis string) {
	if node[spec.NODE_HOSTNAME].(string) == "" {
		s.log.WithField("index", i).Error("hostname is mandatory in node informations")
	}
	s.log.Debug("Processing node :" + node[spec.NODE_HOSTNAME].(string))

	unitName := s.env.GetName() + "_" + s.name + "_" + node[spec.NODE_HOSTNAME].(string) + ".service"
	s.log.Debug("Unit name is :" + unitName)

	data := make(map[string]interface{})
	data["attribute"] = utils.CopyMap(s.attributes)
	out, err := json.Marshal(data["attribute"])
	if err != nil {
		s.log.WithError(err).Panic("Cannot marshall attributes")
	}
	data["attributes"] = strings.Replace(string(out), "\\\"", "\\\\\\\"", -1)
	data["node"] = node
	data["node"].(map[string]interface{})["acis"] = acis

	var b bytes.Buffer
	err = tmpl.Execute(&b, data)
	if err != nil {
		s.log.Error("Failed to run templating for unit "+unitName, err)
	}
	ok, err := utils.Exists(s.path + "/units")
	if !ok || err != nil {
		os.Mkdir(s.path+"/units", 0755)
	}
	err = ioutil.WriteFile(s.path+"/units"+"/"+unitName, b.Bytes(), 0644)
	if err != nil {
		s.log.WithError(err).WithField("path", s.path+"/units"+"/"+unitName).Error("Cannot writer unit")
	}
}

func (s Service) prepareAciList() (string, error) {
	if len(s.manifest.Containers) == 0 {
		return "", nil
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
				return "", err
			}
			acis += aci.String() + " "
		}
	}
	if acis == "" {
		s.log.Error("Aci list is empty after discovery")
	}
	return acis, nil
}
