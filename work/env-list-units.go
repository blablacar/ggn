package work

import (
	"bufio"
	"strings"
	"sync"

	"github.com/n0rad/go-erlog/logs"
)

var statusCache map[string]UnitStatus
var mutex = &sync.Mutex{}

func (e Env) ListUnits() map[string]UnitStatus {
	mutex.Lock()
	defer mutex.Unlock()

	if statusCache != nil {
		return statusCache
	}

	stdout, _, err := e.RunFleetCmdGetOutput("list-units", "-no-legend")
	if err != nil {
		logs.WithEF(err, e.fields).Debug("Cannot list units")
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(stdout))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	status := make(map[string]UnitStatus)
	for _, line := range lines {
		split := strings.Fields(line)
		status[split[0]] = UnitStatus{Unit: split[0], Machine: split[1], Active: split[2], Sub: split[3]}
	}

	statusCache = status
	return status
}
