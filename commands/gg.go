package commands

import (
	"bufio"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/green-garden/builder"
	"github.com/blablacar/green-garden/config"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const FLEET_SUPPORTED_VERSION = "0.11.5"

func Execute() {
	log.Set(logger.NewLogger())
	config.GetConfig().Load()
	checkFleetVersion()

	var rootCmd = &cobra.Command{
		Use: "green-garden",
	}

	loadEnvCommands(rootCmd)

	rootCmd.AddCommand(versionCmd, generateCmd)
	rootCmd.Execute()

	log.Get().Info("Victory !")
}

func checkFleetVersion() {
	output, err := utils.ExecCmdGetOutput("fleetctl")
	if err != nil {
		log.Get().Panic("fleetctl is required in PATH")
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "VERSION:" {
			scanner.Scan()
			versionString := strings.TrimSpace(scanner.Text())
			version, err := semver.NewVersion(versionString)
			if err != nil {
				log.Get().Panic("Cannot parse version of fleetctl", versionString)
			}
			supported, _ := semver.NewVersion(FLEET_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				log.Get().Panic("fleetctl version in your path is too old. Require >= " + FLEET_SUPPORTED_VERSION)
			}
			break
		}
	}
}
