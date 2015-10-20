package service

import (
	"bufio"
	"io/ioutil"
	"strings"
)

type Unit struct {
	path string
	name string
}

func NewUnit(path string, name string) *Unit {
	unit := new(Unit)
	unit.path = path
	unit.name = name
	return unit
}

func (u Unit) GetUnitContentAsFleeted() (string, error) {
	unitPath := u.path + "/units/" + u.name
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
		currentLine += strings.TrimRight(scanner.Text(), " ")
		if strings.HasSuffix(currentLine, "\\") {
			currentLine = currentLine[0:len(currentLine)-2] + "  "
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	return strings.Join(lines, "\n")
}
