package env

import (
	"bufio"
	"fmt"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/attributes-merger/attributes"
	"github.com/blablacar/green-garden/builder"
	"github.com/blablacar/green-garden/spec"
	"github.com/blablacar/green-garden/utils"
	"github.com/blablacar/green-garden/work/env/service"
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

func (s *Service) GetEnv() spec.Env {
	return s.env
}

func (s *Service) GetLog() logrus.Entry {
	return s.log
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
	for _, unitName := range unitNames {
		s.LoadUnit(unitName).Check()
	}
}

func (s *Service) GetFleetUnitContent(unit string) (string, error) { //TODO this method should be in unit
	stdout, stderr, err := s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "cat", unit)
	if err != nil && stderr == "Unit "+unit+" not found" {
		return "", nil
	}
	return stdout, err
}

func (s *Service) Unlock() {
	s.log.Info("Unlocking")

	kapi := s.env.EtcdClient()
	_, err := kapi.Delete(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		s.log.WithError(cerr).Panic("Cannot unlock service")
	}
}

func (s *Service) Lock(ttl time.Duration, message string) {
	s.log.WithField("ttl", ttl).WithField("message", message).Info("locking")

	kapi := s.env.EtcdClient()
	resp, err := kapi.Get(context.Background(), s.lockPath, nil)
	if cerr, ok := err.(*client.ClusterError); ok {
		s.log.WithError(cerr).Fatal("Server error reading on fleet")
	} else if err != nil {
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

func (s *Service) Update() error {
	s.log.Info("Updating service")
	s.GenerateUnits(nil)

	hostname, _ := os.Hostname()
	s.Lock(time.Hour*1, "["+os.Getenv("USER")+"@"+hostname+"] Updating")
	lock := true
	defer func() {
		if lock {
			s.log.WithField("service", s.name).Warn("!! Leaving but Service is still lock !!")
		}
	}()
units:
	for i, unit := range s.ListUnits() {
		u := s.LoadUnit(unit)

	ask:
		for {
			same, err := u.IsLocalContentSameAsRemote()
			if err != nil {
				u.Log.WithError(err).Warn("Cannot compare local and remote service")
			}
			if same {
				u.Log.Info("Remote service is already up to date")
				if !builder.BuildFlags.All {
					continue units
				}
			}
			if builder.BuildFlags.Yes {
				break ask
			}
			action := s.askToProcessService(i, u)
			switch action {
			case ACTION_DIFF:
				u.DisplayDiff()
			case ACTION_QUIT:
				u.Log.Debug("User want to quit")
				if i == 0 {
					s.Unlock()
					lock = false
				}
				return errors.New("User want to quit")
			case ACTION_SKIP:
				u.Log.Debug("User skip this service")
				continue units
			case ACTION_YES:
				break ask
			default:
				u.Log.Fatal("Should not be here")
			}
		}

		u.Destroy()
		time.Sleep(time.Second * 2)
		err := u.Start()
		if err != nil {
			log.WithError(err).Error("Failed to start service. Keeping lock")
			return err
		}
		time.Sleep(time.Second * 2)
		//		status, err2 := u.Status()
		//		u.Log.WithField("status", status).Debug("Log status")
		//		if err2 != nil {
		//			log.WithError(err2).WithField("status", status).Panic("Unit failed just after start")
		//			return err2
		//		}
		//		if status == "inactive" {
		//			log.WithField("status", status).Panic("Unit failed just after start")
		//			return errors.New("unit is inactive just after start")
		//		}
		//
		//		s.checkServiceRunning()

		// TODO ask deploy pod version ()
		// TODO YES/NO
		// TODO check running tmux
		// TODO running as root ??
		// TODO notify slack
		// TODO store old version
		// TODO !!!!! check that service is running well before going to next server !!!

	}
	s.Unlock()
	lock = false
	return nil
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

func (s *Service) checkServiceRunning() {

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
