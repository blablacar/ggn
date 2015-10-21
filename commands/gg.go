package commands

import (
	"bufio"
	"fmt"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/green-garden/builder"
	"github.com/blablacar/green-garden/config"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const FLEET_SUPPORTED_VERSION = "0.11.5"

func Execute() {

	config.GetConfig().Load()
	checkFleetVersion()

	var logLevel string
	var rootCmd = &cobra.Command{
		Use: "green-garden",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := log.LogLevel(logLevel)
			if err != nil {
				fmt.Printf("Unknown log level : %s", logLevel)
			}

			o, ok := log.Logger.(*logger.Logger)
			if ok {
				o.Level = *level
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.AddCommand(versionCmd, generateCmd)

	loadEnvCommands(rootCmd)

	rootCmd.Execute()
	log.Info("Victory !")
}

func checkFleetVersion() {
	output, err := utils.ExecCmdGetOutput("fleetctl")
	if err != nil {
		log.Error("fleetctl is required in PATH")
		os.Exit(1)
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "VERSION:" {
			scanner.Scan()
			versionString := strings.TrimSpace(scanner.Text())
			version, err := semver.NewVersion(versionString)
			if err != nil {
				log.Error("Cannot parse version of fleetctl", versionString)
				os.Exit(1)
			}
			supported, _ := semver.NewVersion(FLEET_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				log.Error("fleetctl version in your path is too old. Require >= " + FLEET_SUPPORTED_VERSION)
				os.Exit(1)
			}
			break
		}
	}
}
