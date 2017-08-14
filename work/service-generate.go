package work

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/appc/spec/discovery"
	"github.com/appc/spec/schema"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/blablacar/dgr/bin-templater/template"
	"github.com/blablacar/ggn/ggn"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
)

func (s *Service) Generate() error {
	s.generatedMutex.Lock()
	defer s.generatedMutex.Unlock()

	if s.generated {
		return nil
	}

	if s.isTemplatedManifest {
		logs.WithFields(s.fields).Debug("Reloading templated service manifest")
		s.reloadService()
	}

	logs.WithFields(s.fields).Debug("Generating units")

	serviceTmpl, err := s.loadUnitTemplate(PATH_UNIT_SERVICE_TEMPLATE)
	if err != nil {
		return errs.WithEF(err, s.fields, "Cannot load service template")
	}

	var timerTmpl *template.Templating
	if s.hasTimer {
		timerTmpl, err = s.loadUnitTemplate(PATH_UNIT_TIMER_TEMPLATE)
		if err != nil {
			logs.WithEF(err, s.fields).Fatal("Cannot load timer template")
		}
	}

	if len(s.manifest.GgnMinimalVersion) > 0 {
		var currentVersion = common.Version(ggn.GgnVersion)
		logs.WithField("version", s.manifest.GgnMinimalVersion).Debug("Found ggn minimal version")
		if s.manifest.GgnMinimalVersion.GreaterThan(currentVersion) {
			logs.WithFields(s.fields).
				WithField("minimalversion", s.manifest.GgnMinimalVersion).
				WithField("version", currentVersion).
				Fatal("You don't use the minimal required version of ggn for this service")
		}
	}
	if len(s.nodesAsJsonMap) == 0 {
		logs.WithFields(s.fields).Fatal("No node to process in manifest")
		return nil
	}

	for _, unitName := range s.ListUnits() {
		unit := s.LoadUnit(unitName)
		if unit.GetType() == TYPE_SERVICE {
			if err := unit.Generate(serviceTmpl); err != nil {
				return err
			}
		} else if unit.GetType() == TYPE_TIMER {
			if err := unit.Generate(timerTmpl); err != nil {
				return err
			}
		} else {
			logs.WithFields(s.fields).WithField("type", unit.GetType()).Fatal("Unknown unit type")
		}
	}
	s.generated = true
	return nil
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

func (s Service) podManifestToMap(result map[string][]common.ACFullname, contents []byte) error {
	pod := schema.BlankPodManifest()
	err := pod.UnmarshalJSON(contents)
	if err != nil {
		return err
	}

	var podname string
	var acis []common.ACFullname
	for i, podAci := range pod.Apps {
		version, _ := podAci.Image.Labels.Get("version")
		fullname := common.NewACFullName(podAci.Image.Name.String() + ":" + version)
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

func (s Service) aciManifestToMap(result map[string][]common.ACFullname, contents []byte) error {
	aci := schema.BlankImageManifest()
	err := aci.UnmarshalJSON(contents)
	if err != nil {
		return err
	}

	version, _ := aci.Labels.Get("version")
	fullname := common.NewACFullName(aci.Name.String() + ":" + version)
	result[fullname.Name()] = []common.ACFullname{*fullname}
	return nil
}

func (s Service) sources(sources []string) map[string][]common.ACFullname {
	res := make(map[string][]common.ACFullname)
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

func (s Service) discoverPod(name common.ACFullname) ([]common.ACFullname, error) {
	podFields := s.fields.WithField("pod", name)

	app, err := discovery.NewAppFromString(name.String())
	if app.Labels["os"] == "" {
		app.Labels["os"] = "linux"
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = "amd64"
	}

	insecure := discovery.InsecureTLS | discovery.InsecureHTTP
	endpoint, _, err := discovery.DiscoverACIEndpoints(*app, nil, insecure)
	if err != nil {
		logs.WithEF(err, podFields).Fatal("pod discovery failed")
	}

	if len(endpoint) == 0 {
		logs.WithF(podFields).Fatal("Discovery does not contain a url")
	}
	url := endpoint[0].ACI
	url = strings.Replace(url, "=aci", "=pod", 1) // TODO this is nexus specific

	logUrl := podFields.WithField("url", url)
	response, err := http.Get(url)
	if err != nil {
		return nil, errs.WithEF(err, logUrl, "Cannot get pod manifest content")
	} else {
		if response.StatusCode != 200 {
			return nil, errs.WithF(logUrl.WithField("status_code", response.StatusCode).
				WithField("status_message", response.Status), "Receive response error for discovery")
		}
		defer response.Body.Close()
		content, err := ioutil.ReadAll(response.Body)
		if err != nil {
			logs.WithEF(err, logUrl).Fatal("Cannot read pod manifest content")
		}
		tmpMap := make(map[string][]common.ACFullname, 1)
		if err := s.podManifestToMap(tmpMap, content); err != nil {
			logs.WithEF(err, logUrl).Fatal("Cannot read pod content")
		}
		acis := tmpMap[name.Name()]
		if acis == nil {
			logs.WithFields(logUrl).Fatal("Discovered pod name does not match requested")
		}
		return acis, nil
	}
}

func (s *Service) PrepareAcis() ([]string, error) {

	if len(s.manifest.Containers) == 0 {
		return []string{}, nil
	}

	s.aciListMutex.Lock()
	defer s.aciListMutex.Unlock()

	if len(s.aciList) > 0 {
		return s.aciList, nil
	}

	override := s.sources(BuildFlags.GenerateManifests)
	logs.WithFields(s.fields).WithField("data", override).Debug("Local resolved sources")

	var acis []string
	for _, aci := range s.manifest.Containers {
		containerLog := s.fields.WithField("container", aci.String())
		logs.WithFields(containerLog).Debug("Processing container")
		if strings.HasPrefix(aci.ShortName(), "pod-") && !strings.Contains(aci.ShortName(), "_") { // TODO this is CNT specific
			var podAcis []common.ACFullname
			if override[aci.Name()] != nil {
				logs.WithFields(containerLog).Debug("Using local source to resolve")
				podAcis = override[aci.Name()]
			} else {
				logs.WithFields(containerLog).Debug("Using remote source to resolve")
				pAcis, err := s.discoverPod(aci)
				if err != nil {
					return []string{}, err
				}
				podAcis = pAcis
			}
			for _, aci := range podAcis {
				acis = append(acis, aci.String())
			}
		} else {
			var taci common.ACFullname
			if override[aci.Name()] != nil {
				logs.WithFields(containerLog).Debug("Using local source to resolve")
				taci = override[aci.Name()][0]
			} else {
				logs.WithFields(containerLog).Debug("Using remote source to resolve")
				aciTmp, err := aci.FullyResolved()
				taci = *aciTmp
				if err != nil {
					logs.WithEF(err, containerLog).Fatal("Cannot resolve aci")
					return []string{}, nil
				}
			}
			acis = append(acis, taci.String())
		}
	}
	if len(acis) == 0 {
		logs.WithFields(s.fields).Error("Cannot resolve aci")
	}
	s.aciList = acis
	return acis, nil
}
