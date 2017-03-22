package commands

import (
	"github.com/blablacar/ggn/work"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

func prepareEnvCommands(env *work.Env) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   env.GetName(),
		Short: "Run command for " + env.GetName(),
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check of " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Check()
		},
	}

	fleetctlCmd := &cobra.Command{
		Use:   "fleetctl",
		Short: "Run fleetctl command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Fleetctl(args)
		},
	}

	listUnitsCmd := &cobra.Command{
		Use:   "list-units",
		Short: "Run list-units command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.FleetctlListUnits()
		},
	}

	listMachinesCmd := &cobra.Command{
		Use:   "list-machines",
		Short: "Run list-machines command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.FleetctlListMachines()
		},
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate units for " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Generate()
		},
	}
	envCmd.AddCommand(generateCmd, fleetctlCmd, checkCmd, listUnitsCmd, listMachinesCmd)

	unitNames := make(map[string]struct{})
	conflictUnits := make(map[string]struct{})
	for _, serviceName := range env.ListServices() {
		service := env.LoadService(serviceName)
		envCmd.AddCommand(prepareServiceCommands(service))

		for _, unitName := range service.ListUnits() {
			if _, ok := unitNames[unitName]; ok {
				conflictUnits[unitName] = struct{}{}
			}
			unitNames[unitName] = struct{}{}
		}
	}

	for _, serviceName := range env.ListServices() {
		service := env.LoadService(serviceName)
		for _, unitName := range service.ListUnits() {
			unit := service.LoadUnit(unitName)
			if _, ok := conflictUnits[unitName]; ok {
				envCmd.AddCommand(&cobra.Command{
					Use:   unit.Name,
					Short: getShortDescription(unit, "Run command for"),
					Run: func(cmd *cobra.Command, args []string) {
						logs.WithField("unit", unitName).Fatal("unit name in conflict. please specify service")
					},
				})
			} else {
				envCmd.AddCommand(prepareUnitCommands(unit))
			}

		}
	}

	return envCmd
}
