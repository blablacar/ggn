package commands

import (
	"github.com/blablacar/cnt/log"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "init files-tree",
	Long:  `init files-tree`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Get().Warn("here")
	},
}
