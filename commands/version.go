package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of ggn",
	Long:  `Print the version number of cnt`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("ggn\n\n")
		fmt.Printf("version    : %s\n", GgnVersion)
		if BuildDate != "" {
			fmt.Printf("build date : %s\n", BuildDate)
		}
		if CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", CommitHash)
		}
		os.Exit(0)
	},
}
