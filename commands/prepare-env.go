package commands

import (
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func prepareEnvCommands(env *work.Env) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   env.GetName(),
		Short: "Run command for " + env.GetName(),
	}

	checkCmd := &cobra.Command{
		Use:   env.GetName(),
		Short: "Check local units with what is running on fleet on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Check()
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Status of " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Status()
		},
	}

	fleetctlCmd := &cobra.Command{
		Use:   "fleetctl",
		Short: "Run fleetctl command on " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Fleetctl(args)
		},
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate units for " + env.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			env.Generate()
		},
	}
	envCmd.AddCommand(generateCmd, fleetctlCmd, checkCmd, statusCmd)

	for _, serviceName := range env.ListServices() {
		service := env.LoadService(serviceName)
		envCmd.AddCommand(prepareServiceCommands(service))
	}

	return envCmd
}
