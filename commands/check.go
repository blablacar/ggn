package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
	"strings"
)

func checkEnv(cmd *cobra.Command, args []string, work *work.Work, envString string) {
	logEnv := logrus.WithField("env", envString)
	logEnv.Info("Running command")

	env := work.LoadEnv(envString)

	units, err := env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "list-unit-files", "-no-legend", "-fields", "unit")
	if err != nil {
		logEnv.WithError(err).Fatal("Cannot list unit files")
	}

	for _, unitName := range strings.Split(units, "\n") {
		unitInfo := strings.Split(unitName, "_")
		if len(unitInfo) != 3 {
			logEnv.WithField("unit", unitName).Warn("Unknown unit format for GGN")
			continue
		}
		env.LoadService(unitInfo[1]).LoadUnit(unitName).Check()
	}
}

func checkService(cmd *cobra.Command, args []string, work *work.Work, env string, serviceName string) {
	service := work.LoadEnv(env).LoadService(serviceName)
	service.Check()
}
