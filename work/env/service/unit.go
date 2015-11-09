package service

import (
	"bufio"
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/spec"
	"io/ioutil"
	"strings"
)

type Unit struct {
	log      logrus.Entry
	path     string
	name     string
	unitPath string
	service  spec.Service
}

func NewUnit(path string, name string, service spec.Service) *Unit {
	l := service.GetLog()
	unit := &Unit{
		log:      *l.WithField("unit", name),
		service:  service,
		path:     path,
		name:     name,
		unitPath: path + "/" + name,
	}
	return unit
}

func (u Unit) GetUnitContentAsFleeted() (string, error) {
	unitFileContent, err := ioutil.ReadFile(u.unitPath)
	if err != nil {
		return "", err
	}
	return convertMultilineUnitToString(unitFileContent), nil
}

func (u Unit) Start() error {
	u.log.Info("Starting")
	_, err := u.service.GetEnv().RunFleetCmdGetOutput("start", u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot start unit")
		return err
	}
	return nil
}

func (u Unit) Destroy() error {
	u.log.Info("Destroying") // todo check that service exists before destroy
	_, err := u.service.GetEnv().RunFleetCmdGetOutput("destroy", u.name)
	if err != nil {
		logrus.WithError(err).Warn("Cannot destroy unit")
		return err
	}
	return nil
}

//func (u Unit) Stop() {
//	u.service.GetEnv().RunFleetCmdGetOutput("stop", u.name)
//}
//
//func (u Unit) Status() {
//	u.service.GetEnv().RunFleetCmdGetOutput("status ", u.name)
//}
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
