package work

import (
	"bufio"
	"encoding/json"

	"github.com/coreos/fleet/unit"
	"github.com/juju/errors"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/logs"

	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/blablacar/dgr/bin-dgr/common"
)

type Unit struct {
	Fields         data.Fields
	Type           UnitType
	path           string
	Name           string
	hostname       string
	Filename       string
	unitPath       string
	Service        *Service
	generated      bool
	generatedMutex *sync.Mutex
}

func NewUnit(path string, hostname string, utype UnitType, service *Service) *Unit {
	l := service.GetFields()

	filename := service.GetEnv().GetName() + "_" + service.GetName() + "_" + hostname + utype.String()

	name := hostname
	if utype != TYPE_SERVICE {
		name += utype.String()
	}
	unit := &Unit{
		generatedMutex: &sync.Mutex{},
		Type:           utype,
		Fields:         l.WithField("unit", name),
		Service:        service,
		path:           path,
		hostname:       hostname,
		Name:           name,
		Filename:       filename,
		unitPath:       path + "/" + filename,
	}
	logs.WithFields(unit.Fields).WithField("hostname", hostname).Debug("New unit")

	return unit
}

func (u *Unit) GetType() UnitType {
	return u.Type
}

func (u *Unit) GetName() string {
	return u.Name
}

func (u *Unit) GetService() *Service {
	return u.Service
}

func (u *Unit) Check(command string) {
	if err := u.Service.Generate(); err != nil {
		logs.WithEF(err, u.Fields).Fatal("Generate failed")
	}
	logs.WithFields(u.Fields).Debug("Check")

	info := HookInfo{
		Service: u.Service,
		Unit:    u,
		Action:  "check",
		Command: command,
	}
	u.Service.GetEnv().RunEarlyHook(info)
	defer u.Service.GetEnv().RunLateHook(info)

	statuses := u.Service.GetEnv().ListUnits()
	var status UnitStatus
	if _, ok := statuses[u.Filename]; !ok {
		logs.WithFields(u.Fields).Warn("cannot find unit on fleet")
		return
	}
	status = statuses[u.Filename]
	logs.WithField("status", status).Debug("status")

	if status.Active != ACTIVE_ACTIVE {
		logs.WithFields(u.Fields).WithField("active", status.Active).Warn("unit status is not active")
		return
	}
	if status.Sub != SUB_RUNNING {
		logs.WithFields(u.Fields).WithField("sub", status.Sub).Warn("unit sub is not running")
		return
	}

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		logs.WithFields(u.Fields).Error("Cannot read unit")
		return
	}
	if !same {
		logs.WithFields(u.Fields).Warn("Unit is not up to date")
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
	if u.Type == TYPE_SERVICE && u.Service.HasTimer() {
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
	return common.ExecCmd("git", "diff", "--word-diff", remotePath, localPath)
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

func (u *Unit) runHook(isEarly bool, command string, action string) {
	out, err := json.Marshal(u.GenerateAttributes())
	if err != nil {
		logs.WithEF(err, u.Fields).Fatal("Cannot marshall attributes")
	}

	info := HookInfo{
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
		logs.WithEF(err, u.Fields).Debug("Cannot get local unit content")
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

func (u *Unit) IsRunning() bool {
	content, _, _ := u.Service.GetEnv().RunFleetCmdGetOutput("status", u.Filename)
	if strings.Contains(content, "Active: active (") {
		return true
	}
	return false
}

func (u *Unit) IsLoaded() bool {
	content, _, _ := u.Service.GetEnv().RunFleetCmdGetOutput("status", u.Filename)
	if strings.Contains(content, "Loaded: loaded") {
		return true
	}
	return false
}
