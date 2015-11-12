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
	u.Log.Info("Diff")

	local, remote, err := u.serviceLocalAndRemoteContent()
	if err != nil {
		return err
	}

	ioutil.WriteFile("/tmp/ggn-local", []byte(local), 0644)
	defer os.Remove("/tmp/ggn-local")
	ioutil.WriteFile("/tmp/ggn-remote", []byte(remote), 0644)
	defer os.Remove("/tmp/ggn-remote")
	utils.ExecCmd("git", "diff", "/tmp/ggn-remote", "/tmp/ggn-local")
	return nil
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
		u.Log.WithError(err).Error("Cannot read unit file")
		return "", "", err
	}

	remoteContent, err := u.service.GetFleetUnitContent(u.Name)
	if err != nil {
		u.Log.WithError(err).Error("Cannot read unit file")
		return "", "", err
	}
	return localContent, remoteContent, nil
}

func (u Unit) Start() error {
	u.Log.Info("Starting")
	_, err := u.service.GetEnv().RunFleetCmdGetOutput("start", u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot start unit")
		return err
	}
	return nil
}

func (u Unit) Destroy() error {
	u.Log.Info("Destroying") // todo check that service exists before destroy
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
			line = "\n"
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
