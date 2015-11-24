package commands

import (
	"github.com/blablacar/green-garden/work/env/service"
	"github.com/spf13/cobra"
)

func prepareUnitCommands(unit *service.Unit) *cobra.Command {
	unitCmd := &cobra.Command{
		Use:   unit.Name,
		Short: getShortDescription(unit, "Run command for"),
	}

	startCmd := &cobra.Command{
		Use:   "start",
		Short: getShortDescription(unit, "Start"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Start()
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: getShortDescription(unit, "Stop"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Stop()
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: getShortDescription(unit, "Update"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Update()
		},
	}

	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: getShortDescription(unit, "Destroy"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Destroy()
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: getShortDescription(unit, "Get status of"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Status()
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: getShortDescription(unit, "Diff"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Diff()
		},
	}

	unloadCmd := &cobra.Command{
		Use:   "unload",
		Short: getShortDescription(unit, "Unload"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Unload()
		},
	}

	unitCmd.AddCommand(startCmd, stopCmd, updateCmd, destroyCmd, statusCmd, unloadCmd, diffCmd)
	return unitCmd
}

func getShortDescription(unit *service.Unit, action string) string {
	return action + " '" + unit.Name + "' from '" + unit.Service.GetName() + "' on env '" + unit.Service.GetEnv().GetName() + "'"
}
