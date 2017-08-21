package commands

import (
	"github.com/blablacar/ggn/work"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

func prepareUnitCommands(unit *work.Unit) *cobra.Command {
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
			unit.Start("unit/start")
		},
	}

	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: getShortDescription(unit, "Stop"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Stop("unit/stop")
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: getShortDescription(unit, "Update"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Update("unit/update")
		},
	}

	restartCmd := &cobra.Command{
		Use:   "restart",
		Short: getShortDescription(unit, "Restart"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Restart("unit/restart")
		},
	}

	destroyCmd := &cobra.Command{
		Use:   "destroy",
		Short: getShortDescription(unit, "Destroy"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Destroy("unit/destroy")
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: getShortDescription(unit, "Get status of"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Status("unit/status")
		},
	}

	journalCmd := &cobra.Command{
		Use:   "journal",
		Short: getShortDescription(unit, "Get journal of"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Journal("unit/journal", follow, lines)
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff",
		Short: getShortDescription(unit, "Diff"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Diff("unit/diff")
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: getShortDescription(unit, "check"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Check("unit/check")
		},
	}

	unloadCmd := &cobra.Command{
		Use:   "unload",
		Short: getShortDescription(unit, "Unload"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Unload("unit/unload")
		},
	}
	loadCmd := &cobra.Command{
		Use:   "load",
		Short: getShortDescription(unit, "load"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Load("unit/load")
		},
	}

	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: getShortDescription(unit, "ssh"),
		Run: func(cmd *cobra.Command, args []string) {
			unit.Ssh("unit/ssh")
		},
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: getShortDescription(unit, "generate"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := unit.Service.Generate(); err != nil {
				logs.WithE(err).Error("Generate failed")
			}
		},
	}

	unitCmd.PersistentFlags().StringVarP(&work.BuildFlags.ManifestAttributes, "manifest-attributes", "A", "{}", "Attributes to template the service manifest with.")
	journalCmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow")
	journalCmd.Flags().IntVarP(&lines, "lines", "l", 10, "lines")
	updateCmd.Flags().BoolVarP(&work.BuildFlags.Force, "force", "f", false, "force update even if up to date")

	unitCmd.AddCommand(startCmd, stopCmd, updateCmd, destroyCmd, statusCmd, unloadCmd,
		diffCmd, checkCmd, journalCmd, loadCmd, sshCmd, restartCmd, generateCmd)
	return unitCmd
}

func getShortDescription(unit *work.Unit, action string) string {
	return action + " '" + unit.Name + "' from '" + unit.Service.GetName() + "' on env '" + unit.Service.GetEnv().GetName() + "'"
}
