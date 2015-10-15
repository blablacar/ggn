package commands

import (
	"fmt"
	"github.com/blablacar/green-garden/application"
	"github.com/spf13/cobra"
	"os"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of gg",
	Long:  `Print the version number of cnt`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("gg\n\n")
		fmt.Printf("version    : %s\n", application.Version)
		if application.BuildDate != "" {
			fmt.Printf("build date : %s\n", application.BuildDate)
		}
		if application.CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", application.CommitHash)
		}
		os.Exit(0)
	},
}
