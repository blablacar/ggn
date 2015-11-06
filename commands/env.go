package commands

import (
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/config"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func loadEnvCommands(rootCmd *cobra.Command) {
	log.WithField("path", config.GetConfig().WorkPath).Debug("Loading envs")
	work := work.NewWork(config.GetConfig().WorkPath)

	for _, f := range work.ListEnvs() {
		var env = f
		var envCmd = &cobra.Command{
			Use:   env,
			Short: "Run command for " + env,
		}

		var check = &cobra.Command{
			Use:   "check",
			Short: "Check local units with what is running on fleet on " + env,
			Run: func(cmd *cobra.Command, args []string) {
				checkEnv(cmd, args, work, env)
			},
		}

		var runCmd = &cobra.Command{
			Use:   "fleetctl",
			Short: "Run fleetctl command on " + env,
			Run: func(cmd *cobra.Command, args []string) {
				fleetctl(cmd, args, work, env)
			},
		}
		envCmd.AddCommand(runCmd, check)

		var generateCmd = &cobra.Command{
			Use:   "generate",
			Short: "Generate units for " + env,
			Run: func(cmd *cobra.Command, args []string) {
				generateEnv(cmd, args, work, env)
			},
		}
		envCmd.AddCommand(generateCmd)

		rootCmd.AddCommand(envCmd)

		for _, g := range work.LoadEnv(env).ListServices() {
			var service = g
			var serviceCmd = &cobra.Command{
				Use:   service,
				Short: "run command for " + service + " on env " + env,
			}

			var checkCmd = &cobra.Command{
				Use:   "check",
				Short: "Check local units matches what is running on " + env + " for " + service,
				Run: func(cmd *cobra.Command, args []string) {
					checkService(cmd, args, work, env, service)
				},
			}

			var generateCmd = &cobra.Command{
				Use:   "generate [manifest...]",
				Short: "generate units for " + service + " on env :" + env,
				Long:  `generate units using remote resolved or local pod/aci manifests`,
				Run: func(cmd *cobra.Command, args []string) {
					generateService(cmd, args, work, env, service)
				},
			}

			serviceCmd.AddCommand(generateCmd, checkCmd)

			envCmd.AddCommand(serviceCmd)
		}
	}
}
