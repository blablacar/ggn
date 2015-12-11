package service

import (
	"bufio"
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/ggn/spec"
	"github.com/coreos/fleet/unit"
	"github.com/juju/errors"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Unit struct {
	Log       logrus.Entry
	Type      spec.UnitType
	path      string
	Name      string
	hostname  string
	Filename  string
	unitPath  string
	Service   spec.Service
	generated bool
}

func NewUnit(path string, hostname string, utype spec.UnitType, service spec.Service) *Unit {
	l := service.GetLog()

	filename := service.GetEnv().GetName() + "_" + service.GetName() + "_" + hostname + utype.String()

	name := hostname
	if utype != spec.TYPE_SERVICE {
		name += utype.String()
	}
	unit := &Unit{
		Type:     utype,
		Log:      *l.WithField("unit", hostname),
		Service:  service,
		path:     path,
		hostname: hostname,
		Name:     name,
		Filename: filename,
		unitPath: path + "/" + filename,
	}
	unit.Log.Debug("New unit")

	return unit
}

func (u *Unit) GetType() spec.UnitType {
	return u.Type
}

func (u *Unit) GetName() string {
	return u.Name
}

func (u *Unit) GetService() spec.Service {
	return u.Service
}

func (u *Unit) Check(command string) {
	u.Log.Debug("Check")

	info := spec.HookInfo{
		Service: u.Service,
		Unit:    u,
		Action:  "check",
		Command: command,
	}
	u.Service.GetEnv().RunEarlyHook(info)
	defer u.Service.GetEnv().RunLateHook(info)

	statuses := u.Service.GetEnv().ListUnits()
	var status spec.UnitStatus
	if _, ok := statuses[u.Filename]; !ok {
		u.Log.Warn("cannot find unit on fleet")
		return
	}
	status = statuses[u.Filename]
	logrus.WithField("status", status).Debug("status")

	if status.Active != spec.ACTIVE_ACTIVE {
		u.Log.WithField("active", status.Active).Warn("unit status is not active")
		return
	}
	if status.Sub != spec.SUB_RUNNING {
		u.Log.WithField("sub", status.Sub).Warn("unit sub is not running")
		return
	}

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.Error("Cannot read unit")
		return
	}
	if !same {
		u.Log.Warn("Unit is not up to date")
		return
	}
}

func (u *Unit) GetUnitContentAsFleeted() (string, error) {
	unitFileContent, err := ioutil.ReadFile(u.unitPath)
	if err != nil {
		return "", err
	}

	fleetunit, err := unit.NewUnitFile(string(unitFileContent))
	if err != nil {
		return "", err
	}
	return convertMultilineUnitToString([]byte(fleetunit.String())), nil
}

func (u *Unit) UpdateInside(command string) {
	u.Destroy(command)
	time.Sleep(time.Second * 2)
	if u.Type == spec.TYPE_SERVICE && u.Service.HasTimer() {
		u.Load(command)
	} else {
		u.Start(command)
	}
}

func (u *Unit) DisplayDiff() error {
	local, remote, err := u.serviceLocalAndRemoteContent()
	if err != nil {
		return err
	}

	localPath := "/tmp/" + u.Name + "__local"
	remotePath := "/tmp/" + u.Name + "__remote"

	ioutil.WriteFile(localPath, []byte(local), 0644)
	defer os.Remove(localPath)
	ioutil.WriteFile(remotePath, []byte(remote), 0644)
	defer os.Remove(remotePath)
	return utils.ExecCmd("git", "diff", "--word-diff", remotePath, localPath)
}

func (u *Unit) IsLocalContentSameAsRemote() (bool, error) {
	local, remote, err := u.serviceLocalAndRemoteContent()
	if local != "" && err != nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return local == remote, nil
}

///////////////////////////////////////////

const EARLY = true
const LATE = false

func (u *Unit) runHook(isEarly bool, command string, action string) {
	out, err := json.Marshal(u.GenerateAttributes())
	if err != nil {
		u.Log.WithError(err).Panic("Cannot marshall attributes")
	}

	info := spec.HookInfo{
		Service:    u.Service,
		Unit:       u,
		Action:     "unit/" + action,
		Command:    command,
		Attributes: string(out),
	}
	if isEarly {
		u.Service.GetEnv().RunEarlyHook(info)
	} else {
		u.Service.GetEnv().RunLateHook(info)
	}

}

func (u *Unit) serviceLocalAndRemoteContent() (string, string, error) {
	localContent, err := u.GetUnitContentAsFleeted()
	if err != nil {
		u.Log.WithError(err).Debug("Cannot get local unit content")
		return "", "", errors.Annotate(err, "Cannot read local unit file")
	}

	remoteContent, err := u.Service.GetFleetUnitContent(u.Filename)
	if err != nil {
		return localContent, "", errors.Annotate(err, "CanCannot read remote unit file")
	}
	return localContent, remoteContent, nil
}

func convertMultilineUnitToString(content []byte) string {
	var lines []string
	var currentLine string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" && currentLine != "" {
			currentLine = strings.TrimRight(currentLine, " ")
		}
		currentLine += strings.TrimRight(line, " ")
		if strings.HasSuffix(currentLine, "\\") {
			currentLine = currentLine[0:len(currentLine)-2] + "  "
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	return strings.Join(lines, "\n")
}

func (u *Unit) isRunning() bool {
	content, _, _ := u.Service.GetEnv().RunFleetCmdGetOutput("status", u.Filename)
	if strings.Contains(content, "Active: active (running)") {
		return true
	}
	return false
}

func (u *Unit) isLoaded() bool {
	content, _, _ := u.Service.GetEnv().RunFleetCmdGetOutput("status", u.Filename)
	if strings.Contains(content, "Loaded: loaded") {
		return true
	}
	return false
}
