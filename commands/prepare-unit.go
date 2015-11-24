package commands

import (
	"github.com/blablacar/green-garden/work/env/service"
	"github.com/spf13/cobra"
)

func prepareUnitCommands(unit *service.Unit) *cobra.Command {
	var unitCmd = &cobra.Command{
		Use:   unit.Name,
		Short: "run command for " + unit.Name + " on " + unit.Service.GetName() + " on env " + unit.Service.GetEnv().GetName(),
	}

	var startCmd = &cobra.Command{
		Use:   "start",
		Short: "start " + unit.Name + " from " + unit.Service.GetName() + " on env " + unit.Service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Start()
		},
	}

	unitCmd.AddCommand(startCmd)

	return unitCmd
}
