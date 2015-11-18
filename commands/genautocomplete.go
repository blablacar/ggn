package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/config"
	"github.com/spf13/cobra"
	"os"
)

var autocompleteTarget string

var autocompleteType string

var genautocompleteCmd = &cobra.Command{
	Use:   "genautocomplete",
	Short: "Generate shell autocompletion script",
	Long: `Generates a shell autocompletion script.
NOTE: The current version supports Bash only.
      This should work for *nix systems with Bash installed.
By default, the file is written directly to ~/.config/green-garden/
Add ` + "`--completionfile=/path/to/file`" + ` flag to set alternative
file-path and name.
Logout and in again to reload the completion scripts,
or just source them in directly:
	$ . /etc/bash_completion`,

	Run: func(cmd *cobra.Command, args []string) {

		if autocompleteType != "bash" {
			logrus.WithField("type", autocompleteType).Fatalln("Only Bash is supported for now")
		}
		err := cmd.Root().GenBashCompletionFile(autocompleteTarget)
		if err != nil {
			logrus.WithError(err).Fatalln("Failed to generate shell completion file")
		} else {
			logrus.WithField("path", autocompleteTarget).Println("Bash completion saved")
		}
		os.Exit(0)
	},
}

func init() {

	genautocompleteCmd.PersistentFlags().StringVarP(&autocompleteTarget, "completionfile", "", config.GetConfig().Path+"/ggn_completion.sh", "Autocompletion file")
	genautocompleteCmd.PersistentFlags().StringVarP(&autocompleteType, "type", "", "bash", "Autocompletion type (currently only bash supported)")
}