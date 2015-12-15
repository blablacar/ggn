package commands

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/utils"
	"github.com/blablacar/ggn/builder"
	"github.com/blablacar/ggn/ggn"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

const FLEET_SUPPORTED_VERSION = "0.11.5"

func Execute() {
	checkFleetVersion()

	homefolder := ggn.DefaultHomeRoot()
	if argsHome := homeFolderFromArgs(); argsHome != "" {
		homefolder = argsHome
	} else {
		_, err := os.Stat(homefolder + "/green-garden")
		_, err2 := os.Stat(homefolder + "/ggn")
		if os.IsNotExist(err2) && err == nil {
			log.WithField("oldPath", homefolder+"/green-garden").
				WithField("newPath", homefolder+"/ggn").
				Warn("You are using the old home folder")
			homefolder += "/green-garden"
		} else {
			homefolder += "/ggn"
		}
	}
	ggn.Home = ggn.NewHome(homefolder)

	// logs
	lvl := logLevelFromArgs()
	if lvl == "" {
		lvl = "info"
	}

	level, err := log.ParseLevel(lvl)
	if err != nil {
		fmt.Printf("Unknown log level : %s", lvl)
		os.Exit(1)
	}
	log.SetLevel(level)

	rootCmd := &cobra.Command{
		Use: "ggn",
	}

	var useless string
	var useless2 string

	rootCmd.PersistentFlags().StringVarP(&useless2, "log-level", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&useless, "home-path", "H", ggn.DefaultHomeRoot()+"/ggn", "Set home folder")
	rootCmd.PersistentFlags().StringSliceVarP(&builder.BuildFlags.GenerateManifests, "generate-manifest", "M", []string{}, "Manifests used to generate. comma separated")
	rootCmd.AddCommand(versionCmd, generateCmd, genautocompleteCmd)

	loadEnvCommands(rootCmd)

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	log.Debug("Victory !")
}

func homeFolderFromArgs() string {
	for i, arg := range os.Args {
		if arg == "--" {
			return ""
		}
		if arg == "-H" {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--home-path=") {
			return arg[12:]
		}
	}
	return ""
}

func logLevelFromArgs() string {
	for i, arg := range os.Args {
		if arg == "--" {
			return ""
		}
		if arg == "-L" {
			return os.Args[i+1]
		}
		if strings.HasPrefix(arg, "--log-level=") {
			return arg[12:]
		}
	}
	return ""
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
