package commands

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/ggn/config"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const FLEET_SUPPORTED_VERSION = "0.11.5"

func Execute() {
	config.GetConfig().Load()
	checkFleetVersion()

	var logLevel string
	var rootCmd = &cobra.Command{
		Use: "ggn",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := log.ParseLevel(logLevel)
			if err != nil {
				fmt.Printf("Unknown log level : %s\n", logLevel)
				os.Exit(1)
			}
			log.SetLevel(level)
		},
	}
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.AddCommand(versionCmd, generateCmd, genautocompleteCmd)

	loadEnvCommands(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	log.Debug("Victory !")
}

func checkFleetVersion() {
	output, err := utils.ExecCmdGetOutput("fleetctl")
	if err != nil {
		log.Fatal("fleetctl is required in PATH")
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
