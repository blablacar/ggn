package commands

import (
	"fmt"
	"os"

	"github.com/blablacar/ggn/ggn"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ggn",
	Long:  `Print the version number of cnt`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("ggn\n\n")
		fmt.Printf("version    : %s\n", ggn.GgnVersion)
		if ggn.BuildDate != "" {
			fmt.Printf("build date : %s\n", ggn.BuildDate)
		}
		if ggn.CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", ggn.CommitHash)
		}
		os.Exit(0)
	},
}
