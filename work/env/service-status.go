package env

import (
	"strings"
)

func (s *Service) Status() {
	s.log.Debug("Running status")

	s.Generate(nil)

	units, _, err := s.env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-unit-files", "-no-legend", "-fields", "unit")
	if err != nil {
		s.log.WithError(err).Fatal("Cannot list unit files")
	}

	for _, unitName := range strings.Split(units, "\n") {
		unitInfo := strings.Split(unitName, "_")
		if len(unitInfo) != 3 {
			continue
		}
		if unitInfo[1] != s.Name {
			continue
		}
		split := strings.Split(unitInfo[2], ".")
		s.LoadUnit(split[0]).Diff()
	}
}
