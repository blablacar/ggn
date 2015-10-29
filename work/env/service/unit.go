package service

import (
	"bufio"
	"github.com/blablacar/green-garden/spec"
	"io/ioutil"
	"strings"
)

type Unit struct {
	path    string
	name    string
	service spec.Service
}

func NewUnit(path string, name string, service spec.Service) *Unit {
	unit := &Unit{
		path:    path,
		name:    name,
		service: service,
	}
	return unit
}

func (u Unit) GetUnitContentAsFleeted() (string, error) {
	unitPath := u.path + "/" + u.name
	unitFileContent, err := ioutil.ReadFile(unitPath)
	if err != nil {
		return "", err
	}
	return convertMultilineUnitToString(unitFileContent), nil
}

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
