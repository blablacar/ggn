package commands

import (
	"github.com/blablacar/green-garden/builder"
	"github.com/blablacar/green-garden/work/env/service"
	"github.com/spf13/cobra"
)

func prepareUnitCommands(unit *service.Unit) *cobra.Command {
	var follow bool
	var lines int

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
			unit.Update(true)
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

	journalCmd := &cobra.Command{
		Use:   "journal",
		Short: getShortDescription(unit, "Get journal of"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Journal(follow, lines)
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: getShortDescription(unit, "Diff"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Diff()
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: getShortDescription(unit, "check"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Check()
		},
	}

	unloadCmd := &cobra.Command{
		Use:   "unload",
		Short: getShortDescription(unit, "Unload"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Unload()
		},
	}
	loadCmd := &cobra.Command{
		Use:   "load",
		Short: getShortDescription(unit, "load"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Load()
		},
	}

	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: getShortDescription(unit, "ssh"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Ssh()
		},
	}

	journalCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow")
	journalCmd.Flags().IntVarP(&lines, "lines", "l", 10, "lines")
	updateCmd.Flags().BoolVarP(&builder.BuildFlags.Force, "force", "f", false, "force update even if up to date")

	unitCmd.AddCommand(startCmd, stopCmd, updateCmd, destroyCmd, statusCmd, unloadCmd,
		diffCmd, checkCmd, journalCmd, loadCmd, sshCmd)
	return unitCmd
}

func getShortDescription(unit *service.Unit, action string) string {
	return action + " '" + unit.Name + "' from '" + unit.Service.GetName() + "' on env '" + unit.Service.GetEnv().GetName() + "'"
}
