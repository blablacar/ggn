package work

import (
	"github.com/blablacar/ggn/spec"
	"strings"
)

func (e Env) Check() {
	e.log.Debug("Running check")

	info := spec.HookInfo{Command: "env/check", Action: "env/check"}
	e.RunEarlyHook(info)
	defer e.RunLateHook(info)

	e.Generate()

	units, _, err := e.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-unit-files", "-no-legend", "-fields", "unit")
	if err != nil {
		e.log.WithError(err).Fatal("Cannot list unit files")
	}

	for _, unitName := range strings.Split(units, "\n") {
		unitInfo := strings.Split(unitName, "_")
		if len(unitInfo) != 3 {
			e.log.WithField("unit", unitName).Warn("Unknown unit format for GGN")
			continue
		}
		split := strings.Split(unitInfo[2], ".")
		e.LoadService(unitInfo[1]).LoadUnit(split[0]).Check("env/check")
	}
}

//import (
//	"strings"
//)
//
//func (e Env) Check() {
//	e.log.Debug("Running command")
//
//	e.Generate()
//
//	units, _, err := e.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-unit-files", "-no-legend", "-fields", "unit")
//	if err != nil {
//		e.log.WithError(err).Fatal("Cannot list unit files")
//	}
//
//	for _, unitName := range strings.Split(units, "\n") {
//		unitInfo := strings.Split(unitName, "_")
//		if len(unitInfo) != 3 {
//			e.log.WithField("unit", unitName).Warn("Unknown unit format for GGN")
//			continue
//		}
//		e.LoadService(unitInfo[1]).LoadUnit(unitName).Check()
//	}
//}
