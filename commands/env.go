package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/config"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
	"strings"
)

func loadEnvCommands(rootCmd *cobra.Command) {
	log.WithField("path", config.GetConfig().WorkPath).Debug("Loading envs")
	work := work.NewWork(config.GetConfig().WorkPath)

	for _, f := range work.ListEnvs() {
		var env = f
		var envCmd = &cobra.Command{
			Use:   env,
			Short: "Run command for " + env,
			Run: func(cmd *cobra.Command, args []string) {
				runner(cmd, args, work)
			},
		}

		var runCmd = &cobra.Command{
			Use:   "run",
			Short: "run fleetctl command on " + env,
			Run: func(cmd *cobra.Command, args []string) {
				run(cmd, args, work, env)
			},
		}
		envCmd.AddCommand(runCmd)

		var generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "Generate units for " + env,
			Run: func(cmd *cobra.Command, args []string) {
				generate(cmd, args, work, env)
			},
		}
		envCmd.AddCommand(generateCmd)

		rootCmd.AddCommand(envCmd)
	}
}

func run(cmd *cobra.Command, args []string, work *work.Work, env string) {
	log.WithField("env", env).Debug("Running command")
	work.LoadEnv(env).Run(args)
}

func generate(cmd *cobra.Command, args []string, work *work.Work, env string) {
	log.WithField("env", env).Debug("Generating units")
	work.LoadEnv(env).Generate()
}

func runner(cmd *cobra.Command, args []string, work *work.Work) {
	logEnv := log.WithField("env", cmd.Use)
	logEnv.Info("Running command")

	env := work.LoadEnv(cmd.Use)

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
