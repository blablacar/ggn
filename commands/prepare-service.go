package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/builder"
	"github.com/blablacar/green-garden/work/env"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

func prepareServiceCommands(service *env.Service) *cobra.Command {
	var serviceCmd = &cobra.Command{
		Use:   service.Name,
		Short: "run command for " + service.Name + " on env " + service.GetEnv().GetName(),
	}

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Check local units matches what is running on " + service.GetEnv().GetName() + " for " + service.GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.Check()
		},
	}

	var generateCmd = &cobra.Command{
		Use:   "generate [manifest...]",
		Short: "generate units for " + service.Name + " on env " + service.GetEnv().GetName(),
		Long:  `generate units using remote resolved or local pod/aci manifests`,
		Run: func(cmd *cobra.Command, args []string) {
			service.Generate(args)
		},
	}

	var diffCmd = &cobra.Command{
		Use:   "diff [manifest...]",
		Short: "diff units for " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.Diff()
		},
	}

	var ttl string
	var lockCmd = &cobra.Command{
		Use:   "lock [message...]",
		Short: "lock " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				logrus.Fatal("Please add a message to describe lock")
			}

			message := strings.Join(args, " ")
			ttl, err := time.ParseDuration(ttl)
			if err != nil {
				logrus.WithError(err).Fatal("Wrong value for ttl")
			}

			service.Lock(ttl, message)
		},
	}
	lockCmd.Flags().StringVarP(&ttl, "duration", "t", "1h", "lock duration")

	var unlockCmd = &cobra.Command{
		Use:   "unlock",
		Short: "unlock " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			service.Unlock()
		},
	}

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "update " + service.Name + " on env " + service.GetEnv().GetName(),
		Run: func(cmd *cobra.Command, args []string) {
			err := service.Update()
			if err != nil {
				os.Exit(1)
			}
		},
	}
	updateCmd.Flags().BoolVarP(&builder.BuildFlags.All, "all", "a", false, "process all units, even up to date")
	updateCmd.Flags().BoolVarP(&builder.BuildFlags.Yes, "yes", "y", false, "process units without asking")

	serviceCmd.AddCommand(generateCmd, checkCmd, lockCmd, unlockCmd, updateCmd, diffCmd)

	for _, unitName := range service.ListUnits() {
		unit := service.LoadUnit(unitName)
		serviceCmd.AddCommand(prepareUnitCommands(unit))
	}

	return serviceCmd
}
