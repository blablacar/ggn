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

		var compare = &cobra.Command{
			Use:   "compare",
			Short: "Compare local units with what is running on fleet on " + env,
			Run: func(cmd *cobra.Command, args []string) {
				compareEnv(cmd, args, work, env)
			},
		}

		var runCmd = &cobra.Command{
			Use:   "fleetctl",
			Short: "Run fleetctl command on " + env,
			Run: func(cmd *cobra.Command, args []string) {
				fleetctl(cmd, args, work, env)
			},
		}
		envCmd.AddCommand(runCmd, compare)

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

			var compareCmd = &cobra.Command{
				Use:   "compare",
				Short: "Compare local units with what is running on fleet on " + env + "for " + service,
				Run: func(cmd *cobra.Command, args []string) {
					compareService(cmd, args, work, env, service)
				},
			}

			var generateCmd = &cobra.Command{
				Use:   "generate",
				Short: "generate units for " + service + " on env :" + env,
				Run: func(cmd *cobra.Command, args []string) {
					generateService(cmd, args, work, env, service)
				},
			}

			serviceCmd.AddCommand(generateCmd, compareCmd)

			envCmd.AddCommand(serviceCmd)
		}
	}
}
