package work

import (
	"bufio"
	"github.com/blablacar/ggn/spec"
	"strings"
	"sync"
)

var statusCache map[string]spec.UnitStatus
var mutex = &sync.Mutex{}

func (e Env) ListUnits() map[string]spec.UnitStatus {
	mutex.Lock()
	defer mutex.Unlock()

	if statusCache != nil {
		return statusCache
	}

	stdout, _, err := e.RunFleetCmdGetOutput("list-units", "-no-legend")
	if err != nil {
		e.log.WithError(err).Fatal("Cannot list units")
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	status := make(map[string]spec.UnitStatus)
	for _, line := range lines {
		split := strings.Fields(line)
		status[split[0]] = spec.UnitStatus{Unit: split[0], Machine: split[1], Active: split[2], Sub: split[3]}
	}

	statusCache = status
	return status
}
