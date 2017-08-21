package commands

import (
	"os"
	"strings"
	"time"

	"github.com/blablacar/ggn/work"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
)

func prepareServiceCommands(service *work.Service) *cobra.Command {
	var ttl string

	serviceCmd := &cobra.Command{
		Use:   service.Name,
		Short: "run command for " + service.Name + " on env " + service.GetEnv().GetName(),
	}

	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "generate units for " + service.Name + " on env " + service.GetEnv().GetName(),
		Long:  `generate units using remote resolved or local pod/aci manifests`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.Generate(); err != nil {
				logs.WithE(err).Fatal("Generate failed")
			}
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check [manifest...]",
		Short: "Check units for " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := service.Check(); err != nil {
				logs.WithE(err).Fatal("Check failed")
			}
		},
	}

	diffCmd := &cobra.Command{
		Use:   "diff [manifest...]",
		Short: "diff units for " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.Diff()
		},
	}

	lockCmd := &cobra.Command{
		Use:   "lock [message...]",
		Short: "lock " + service.Name + " on env " + service.GetEnv().GetName(),
		Long: `Add a lock to the service in etcd to prevent somebody else to do modification actions on this service/units.` +
			`lock is ignored if set by the current user`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logs.Fatal("Please add a message to describe lock")
			}

			message := strings.Join(args, " ")
			ttl, err := time.ParseDuration(ttl)
			if err != nil {
				logs.WithError(err).Fatal("Wrong value for ttl")
			}

			service.Lock("service/lock", ttl, message)
		},
	}

	unlockCmd := &cobra.Command{
		Use:   "unlock",
		Short: "unlock " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.Unlock("service/unlock")
		},
	}

	listCmd := &cobra.Command{
		Use:   "list-units",
		Short: "list-units on " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.FleetListUnits("service/unlock")
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "update " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.Update()
			if err != nil {
				os.Exit(1)
			}
		},
	}

	serviceCmd.PersistentFlags().StringVarP(&work.BuildFlags.ManifestAttributes, "manifest-attributes", "A", "{}", "Attributes to template the service manifest with.")

	lockCmd.Flags().StringVarP(&ttl, "duration", "t", "1h", "lock duration")
	updateCmd.Flags().BoolVarP(&work.BuildFlags.All, "all", "a", false, "process all units, even up to date")
	updateCmd.Flags().BoolVarP(&work.BuildFlags.Yes, "yes", "y", false, "process units without asking")

	serviceCmd.AddCommand(generateCmd, lockCmd, unlockCmd, updateCmd, checkCmd, diffCmd, listCmd)

	for _, unitName := range service.ListUnits() {
		unit := service.LoadUnit(unitName)
		serviceCmd.AddCommand(prepareUnitCommands(unit))
	}

	return serviceCmd
}
