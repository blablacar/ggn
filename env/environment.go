package env

import (
	"bufio"
	"github.com/blablacar/cnt/log"
	"io/ioutil"
	"strings"
)

const PATH_SERVICES = "/services"
const PATH_UNITS = "/units"

type Env struct {
	Path string
}

func NewEnvironment(root string, name string) *Env {
	path := root + "/" + name
	_, err := ioutil.ReadDir(path)
	if err != nil {
		log.Get().Panic("Cannot read env directory : "+path, err)
	}

	env := new(Env)
	env.Path = path
	return env
}

func (e Env) GetUnitContentAsFleeted(service string, unitName string) (string, error) {
	unitPath := e.Path + "/services/" + service + "/units/" + unitName
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
