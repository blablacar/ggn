package commands

import "github.com/spf13/cobra"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate units for all envs",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
