package work

import (
	"github.com/blablacar/ggn/spec"
	"sync"
)

func (e Env) Check() {
	e.log.Debug("Running check")

	info := spec.HookInfo{Command: "env/check", Action: "env/check"}
	e.RunEarlyHook(info)
	defer e.RunLateHook(info)

	e.concurrentChecker(e.ListServices())

	//	e.Generate()

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
	//		split := strings.Split(unitInfo[2], ".")
	//		e.LoadService(unitInfo[1]).LoadUnit(split[0]).Check("env/check")
	//	}
}

func (e Env) concurrentChecker(services []string) {
	wg := &sync.WaitGroup{}
	aChan := make(chan string)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			for service := range aChan {
				e.LoadService(service).Check()
			}
			wg.Done()
		}()
	}

	for _, service := range services {
		aChan <- service
	}
	close(aChan)
	wg.Wait()
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
