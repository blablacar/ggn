package service

import (
	"bufio"
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/green-garden/spec"
	"github.com/coreos/fleet/unit"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Unit struct {
	Log      logrus.Entry
	path     string
	Name     string
	unitPath string
	service  spec.Service
}

func NewUnit(path string, name string, service spec.Service) *Unit {
	l := service.GetLog()
	unit := &Unit{
		Log:      *l.WithField("unit", name),
		service:  service,
		path:     path,
		Name:     name,
		unitPath: path + "/" + name,
	}
	return unit
}

func (u Unit) Check() {
	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.WithError(err).Warn("Cannot diff with remote")
	}
	if !same {
		u.Log.Info("Unit is not up to date")
		u.DisplayDiff()
	}
}

func (u Unit) GetUnitContentAsFleeted() (string, error) {
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

func (u Unit) DisplayDiff() error {
	u.Log.Debug("Diff")

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

func (u Unit) IsLocalContentSameAsRemote() (bool, error) {
	local, remote, err := u.serviceLocalAndRemoteContent()
	if err != nil {
		return false, err
	}
	return local == remote, nil
}

func (u Unit) serviceLocalAndRemoteContent() (string, string, error) {
	localContent, err := u.GetUnitContentAsFleeted()
	if err != nil {
		u.Log.WithError(err).Error("Cannot read local unit file")
		return "", "", err
	}

	remoteContent, err := u.service.GetFleetUnitContent(u.Name)
	if err != nil {
		u.Log.WithError(err).Error("Cannot read remote unit file")
		return localContent, "", err
	}
	return localContent, remoteContent, nil
}

func (u Unit) Start() error {
	u.Log.Debug("Starting")
	_, err := u.service.GetEnv().RunFleetCmdGetOutput("start", u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot start unit")
		return err
	}
	return nil
}

func (u Unit) Destroy() error {
	u.Log.Debug("Destroying") // todo check that service exists before destroy
	_, err := u.service.GetEnv().RunFleetCmdGetOutput("destroy", u.Name)
	if err != nil {
		logrus.WithError(err).Warn("Cannot destroy unit")
		return err
	}
	return nil
}

func (u Unit) Status() (string, error) {
	content, err := u.service.GetEnv().RunFleetCmdGetOutput("status", u.Name)
	if err != nil {
		return "", err
	}

	reg, err := regexp.Compile(`Active: (active|inactive|deactivating|activating)`)
	if err != nil {
		u.Log.Panic("Cannot compule regex")
	}
	matched := reg.FindStringSubmatch(content)

	//	if !strings.Contains(content, "Active: %s ") { // Active: failed
	//		return content, errors.New("unit is not in running state")
	//	}

	return matched[1], err
}

//func (u Unit) Stop() {
//	u.service.GetEnv().RunFleetCmdGetOutput("stop", u.name)
//}
//
//
//func (u Unit) Cat() {
//	u.service.GetEnv().RunFleetCmdGetOutput("cat ", u.name)
//}

///////////////////////////////////////////

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
