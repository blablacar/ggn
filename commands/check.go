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

	for _, unit := range strings.Split(units, "\n") {
		logUnit := logEnv.WithField("unit", unit)

		content, err := env.RunFleetCmdGetOutput("-strict-host-key-checking=false", "cat", unit)
		if err != nil {
			logUnit.WithError(err).Fatal("Fleetctl failed to cat service content")
		}
		unitInfo := strings.Split(unit, "_")
		if unitInfo[0] != cmd.Use {
			logUnit.Warn("Unknown unit")
			continue
		}

		res, err := env.LoadService(unitInfo[1]).LoadUnit(unit).GetUnitContentAsFleeted()
		if err != nil {
			logUnit.WithError(err).Warn("Cannot read unit file")
			continue
		}
		if res != content {
			logUnit.Info("Unit is not up to date")
			logUnit.WithField("source", "fleet").Debug(content)
			logUnit.WithField("source", "file").Debug(res)
		}
	}
}

func checkService(cmd *cobra.Command, args []string, work *work.Work, env string, serviceName string) {
	service := work.LoadEnv(env).LoadService(serviceName)
	service.Check()
}
