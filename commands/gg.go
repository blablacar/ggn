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
	"io/ioutil"
	"os"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const FLEET_SUPPORTED_VERSION = "0.11.5"
const PATH_ENV = "/env"

func Execute() {
	log.Set(logger.NewLogger())
	config.GetConfig().Load()
	checkFleetVersion()

	var rootCmd = &cobra.Command{
		Use: "green-garden",
	}

	loadEnvCommands(rootCmd)

	rootCmd.AddCommand(versionCmd, listCmd)
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

var runner = func(cmd *cobra.Command, args []string) {
	log.Get().Info("Running command on " + cmd.Use)

	units, err := utils.ExecCmdGetOutput("fleetctl", "-strict-host-key-checking=false", "list-unit-files", "-no-legend", "-fields", "unit")
	if err != nil {
		log.Get().Panic("Cannot list unit files", err)
	}

	for _, unit := range strings.Split(units, "\n") {
		content, err := utils.ExecCmdGetOutput("fleetctl", "-strict-host-key-checking=false", "cat", unit)
		if err != nil {
			log.Get().Panic("Fleetctl failed to cat service content : ", unit)
		}
		unitInfo := strings.Split(unit, "_")
		if unitInfo[0] != cmd.Use {
			log.Get().Warn("Unknown unit" + unit)
			continue
		}

		unitPath := config.GetConfig().EnvPath + PATH_ENV + "/" + cmd.Use + "/services/" + unitInfo[1] + "/units/" + unit
		unitFileContent, err := ioutil.ReadFile(unitPath)
		if err != nil {
			log.Get().Warn("Cannot read unit file : " + unitPath)
			continue
		}

		res := convertMultilineUnitToString(unitFileContent)
		if res != content {
			log.Get().Info("Unit '" + unit + "' is not up to date")
			log.Get().Trace(content)
			//			log.Get().Trace(fmt.Sprintf("%x", content))
			log.Get().Trace(res)
			//			log.Get().Trace(fmt.Sprintf("%x", res))
		}
	}
}

func convertMultilineUnitToString(content []byte) string {
	var lines []string
	var currentLine string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		currentLine += strings.TrimRight(scanner.Text(), " ")
		if strings.HasSuffix(currentLine, "\\") {
			currentLine = currentLine[0:len(currentLine)-2] + "  "
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	return strings.Join(lines, "\n")
}

func loadEnvCommands(rootCmd *cobra.Command) {
	if _, err := os.Stat(config.GetConfig().EnvPath + PATH_ENV); os.IsNotExist(err) {
		log.Get().Panic("env directory not found in " + config.GetConfig().EnvPath)
	}

	files, err := ioutil.ReadDir(config.GetConfig().EnvPath + PATH_ENV)
	if err != nil {
		log.Get().Panic("Cannot read env directory", err)
	}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}

		var envCmd = &cobra.Command{
			Use: f.Name(),
			Run: runner,
		}
		rootCmd.AddCommand(envCmd)
	}
}
