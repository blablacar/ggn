package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/work"
	"github.com/coreos/go-semver/semver"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const FLEET_SUPPORTED_VERSION = "0.11.5"

var CommitHash string
var GgnVersion string
var BuildDate string

func Execute(commitHash string, ggnVersion string, buildDate string) {
	CommitHash = commitHash
	GgnVersion = ggnVersion
	BuildDate = buildDate

	checkFleetVersion()

	ggn.Home = discoverHome()
	prepareLogs()
	work.BuildFlags.GenerateManifests = GenerateManifestPathFromArgs()

	rootCmd := &cobra.Command{
		Use: "ggn",
	}

	var useless string
	var useless2 []string
	rootCmd.PersistentFlags().StringVarP(&useless, "log-level", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&useless, "home-path", "H", ggn.DefaultHomeRoot()+"/ggn", "Set home folder")
	rootCmd.PersistentFlags().StringSliceVarP(&useless2, "generate-manifest", "M", []string{}, "Manifests used to generate. comma separated")
	rootCmd.AddCommand(versionCmd, genautocompleteCmd)

	newRoot := loadEnvCommandsReturnNewRoot(os.Args, rootCmd)
	rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		newRoot.PersistentFlags().AddFlag(flag)
	})

	args := []string{findEnv()}
	logs.WithField("args", args).Debug("Processing env with args")
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

	logs.Debug("Victory !")
}

func findEnv() string { //TODO this is fucking dirty
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "--" {
			return os.Args[i+1]
		}
		if os.Args[i] == "-h" || os.Args[i] == "--help" {
			return os.Args[i]
		}
		if os.Args[i] == "-M" ||
			os.Args[i] == "-L" ||
			os.Args[i] == "-H" {
			i++
			continue
		}
		if strings.HasPrefix(os.Args[i], "-") ||
			strings.HasPrefix(os.Args[i], "--generate-manifest=") ||
			strings.HasPrefix(os.Args[i], "--log-level=") ||
			strings.HasPrefix(os.Args[i], "--home-path=") {
			continue
		}
		return os.Args[i]
	}
	return ""
}

func prepareLogs() {
	// logs
	lvl := logLevelFromArgs()
	if lvl == "" {
		lvl = "info"
	}

	level, err := logs.ParseLevel(lvl)
	if err != nil {
		fmt.Printf("Unknown log level : %s\n", lvl)
		os.Exit(1)
	}
	logs.SetLevel(level)
}

func discoverHome() ggn.HomeStruct {
	homefolder := ggn.DefaultHomeRoot()
	if argsHome := homeFolderFromArgs(); argsHome != "" {
		homefolder = argsHome
	} else {
		_, err := os.Stat(homefolder + "/green-garden")
		_, err2 := os.Stat(homefolder + "/ggn")
		if os.IsNotExist(err2) && err == nil {
			logs.WithField("oldPath", homefolder+"/green-garden").
				WithField("newPath", homefolder+"/ggn").
				Warn("You are using the old home folder")
			homefolder += "/green-garden"
		} else {
			homefolder += "/ggn"
		}
	}
	return ggn.NewHome(homefolder)
}

func GenerateManifestPathFromArgs() []string {
	for i, arg := range os.Args {
		if arg == "--" {
			return []string{}
		}
		if arg == "-M" {
			return strings.Split(os.Args[i+1], ",")
		}
		if strings.HasPrefix(arg, "--generate-manifest=") {
			return strings.Split(arg[20:], ",")
		}
	}
	return []string{}
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
	output, err := common.ExecCmdGetOutput("fleetctl")
	if err != nil {
		logs.Fatal("fleetctl is required in PATH")
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "VERSION:" {
			scanner.Scan()
			versionString := strings.TrimSpace(scanner.Text())
			version, err := semver.NewVersion(versionString)
			if err != nil {
				logs.Error("Cannot parse version of fleetctl", versionString)
				os.Exit(1)
			}
			supported, _ := semver.NewVersion(FLEET_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				logs.Error("fleetctl version in your path is too old. Require >= " + FLEET_SUPPORTED_VERSION)
				os.Exit(1)
			}
			break
		}
	}
}
