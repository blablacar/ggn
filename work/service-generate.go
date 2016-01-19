package work

import (
	"github.com/appc/spec/discovery"
	"github.com/appc/spec/schema"
	cntspec "github.com/blablacar/cnt/spec"
	"github.com/blablacar/ggn/utils"
	"github.com/n0rad/go-erlog/logs"
	"io/ioutil"
	"net/http"
	"strings"
)

func (s *Service) Generate() {
	s.generatedMutex.Lock()
	defer s.generatedMutex.Unlock()

	if s.generated {
		return
	}

	logs.WithFields(s.fields).Debug("Generating units")

	serviceTmpl, err := s.loadUnitTemplate(PATH_UNIT_SERVICE_TEMPLATE)
	if err != nil {
		logs.WithEF(err, s.fields).Fatal("Cannot load service template")
	}

	var timerTmpl *utils.Templating
	if s.hasTimer {
		timerTmpl, err = s.loadUnitTemplate(PATH_UNIT_TIMER_TEMPLATE)
		if err != nil {
			logs.WithEF(err, s.fields).Fatal("Cannot load timer template")
		}
	}

	if len(s.nodesAsJsonMap) == 0 {
		logs.WithFields(s.fields).Fatal("No node to process in manifest")
		return
	}

	for _, unitName := range s.ListUnits() {
		unit := s.LoadUnit(unitName)
		if unit.GetType() == TYPE_SERVICE {
			unit.Generate(serviceTmpl)
		} else if unit.GetType() == TYPE_TIMER {
			unit.Generate(timerTmpl)
		} else {
			logs.WithFields(s.fields).WithField("type", unit.GetType()).Fatal("Unknown unit type")
		}
	}
	s.generated = true
}

func (s Service) NodeAttributes(hostname string) map[string]interface{} {
	for _, node := range s.nodesAsJsonMap {
		host := node.(map[string]interface{})[NODE_HOSTNAME].(string)
		if host == hostname {
			return node.(map[string]interface{})
		}
	}
	logs.WithFields(s.fields).WithField("hostname", hostname).Fatal("Cannot find host in service list")
	return nil
}

func (s Service) podManifestToMap(result map[string][]cntspec.ACFullname, contents []byte) error {
	pod := schema.BlankPodManifest()
	err := pod.UnmarshalJSON(contents)
	if err != nil {
		return err
	}

	var podname string
	var acis []cntspec.ACFullname
	for i, podAci := range pod.Apps {
		version, _ := podAci.Image.Labels.Get("version")
		fullname := cntspec.NewACFullName(podAci.Image.Name.String() + ":" + version)
		if i == 0 {
			nameSplit := strings.SplitN(fullname.ShortName(), "_", 2)
			podname = fullname.DomainName() + "/" + nameSplit[0]
		}

		//		resolved, err := fullname.FullyResolved() // TODO should not be resolved for local test ??
		//		if err != nil {
		//			logs.WithEF(err, s.fields).Fatal("Cannot fully resolve ACI")
		//		}
		acis = append(acis, *fullname)
	}

	result[podname] = acis
	return nil
}

func (s Service) aciManifestToMap(result map[string][]cntspec.ACFullname, contents []byte) error {
	aci := schema.BlankImageManifest()
	err := aci.UnmarshalJSON(contents)
	if err != nil {
		return err
	}

	version, _ := aci.Labels.Get("version")
	fullname := cntspec.NewACFullName(aci.Name.String() + ":" + version)
	result[fullname.Name()] = []cntspec.ACFullname{*fullname}
	return nil
}

func (s Service) sources(sources []string) map[string][]cntspec.ACFullname {
	res := make(map[string][]cntspec.ACFullname)
	for _, source := range sources {
		content, err := ioutil.ReadFile(source)
		if err != nil {
			logs.WithEF(err, s.fields).Warn("Cannot read source file")
			continue
		}
		if err := s.aciManifestToMap(res, content); err != nil {
			if err2 := s.podManifestToMap(res, content); err2 != nil {
				logs.WithEF(err, s.fields).WithField("error2", err2).Error("Cannot read manifest file as aci nor pod")
			}
		}
	}
	return res
}

func (s Service) discoverPod(name cntspec.ACFullname) []cntspec.ACFullname {
	podFields := s.fields.WithField("pod", name)

	app, err := discovery.NewAppFromString(name.String())
	if app.Labels["os"] == "" {
		app.Labels["os"] = "linux"
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = "amd64"
	}

	endpoint, _, err := discovery.DiscoverEndpoints(*app, nil, false)
	if err != nil {
		logs.WithEF(err, podFields).Fatal("pod discovery failed")
	}

	if len(endpoint.ACIEndpoints) == 0 {
		logs.WithF(podFields).Fatal("Discovery does not contain a url")
	}
	url := endpoint.ACIEndpoints[0].ACI
	url = strings.Replace(url, "=aci", "=pod", 1) // TODO this is nexus specific

	logUrl := podFields.WithField("url", url)
	response, err := http.Get(url)
	if err != nil {
		logs.WithEF(err, logUrl).Fatal("Cannot get pod manifest content")
		return nil
	} else {
		if response.StatusCode != 200 {
			logs.WithFields(logUrl).WithField("status_code", response.StatusCode).
				WithField("status_message", response.Status).
				Fatal("Receive response error for discovery")
		}
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.WithEF(err, logUrl).Fatal("Cannot read pod manifest content")
		}
		tmpMap := make(map[string][]cntspec.ACFullname, 1)
		if err := s.podManifestToMap(tmpMap, content); err != nil {
			logs.WithEF(err, logUrl).Fatal("Cannot read pod content")
		}
		acis := tmpMap[name.Name()]
		if acis == nil {
			logs.WithFields(logUrl).Fatal("Discovered pod name does not match requested")
		}
		return acis
	}
}

func (s *Service) PrepareAciList() string {

	if len(s.manifest.Containers) == 0 {
		return ""
	}

	s.aciListMutex.Lock()
	defer s.aciListMutex.Unlock()

	if s.aciList != "" {
		return s.aciList
	}

	override := s.sources(BuildFlags.GenerateManifests)
	logs.WithFields(s.fields).WithField("data", override).Debug("Local resolved sources")

	var acis string
	for _, aci := range s.manifest.Containers {
		containerLog := s.fields.WithField("container", aci.String())
		logs.WithFields(containerLog).Debug("Processing container")
		if strings.HasPrefix(aci.ShortName(), "pod-") && !strings.Contains(aci.ShortName(), "_") { // TODO this is CNT specific
			var podAcis []cntspec.ACFullname
			if override[aci.Name()] != nil {
				logs.WithFields(containerLog).Debug("Using local source to resolve")
				podAcis = override[aci.Name()]
			} else {
				logs.WithFields(containerLog).Debug("Using remote source to resolve")
				podAcis = s.discoverPod(aci)
			}
			for _, aci := range podAcis {
				acis += aci.String() + " "
			}
		} else {
			var taci cntspec.ACFullname
			if override[aci.Name()] != nil {
				logs.WithFields(containerLog).Debug("Using local source to resolve")
				taci = override[aci.Name()][0]
			} else {
				logs.WithFields(containerLog).Debug("Using remote source to resolve")
				aciTmp, err := aci.FullyResolved()
				taci = *aciTmp
				if err != nil {
					logs.WithEF(err, containerLog).Fatal("Cannot resolve aci")
					return ""
				}
			}
			acis += taci.String() + " "
		}
	}
	if acis == "" {
		logs.WithFields(s.fields).Error("Cannot resolve aci")
	}
	s.aciList = acis
	return acis
}
