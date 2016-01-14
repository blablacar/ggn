package commands

import (
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/work"
	"github.com/n0rad/go-erlog/logs"
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
		if autocompleteTarget == "" {
			autocompleteTarget = ggn.Home.Path + "/ggn_completion.sh"
		}
		if autocompleteType != "bash" {
			logs.WithField("type", autocompleteType).Fatal("Only Bash is supported for now")
		}

		work := work.NewWork(ggn.Home.Config.WorkPath)

		for _, command := range cmd.Root().Commands() {
			if command.Use != "genautocomplete" && command.Use != "version" && command.Use != "help [command]" { // TODO sux
				command.AddCommand(prepareEnvCommands(work.LoadEnv(command.Use)))
			}
		}
		err := cmd.Root().GenBashCompletionFile(autocompleteTarget)
		if err != nil {
			logs.WithE(err).Fatal("Failed to generate shell completion file")
		} else {
			logs.WithField("path", autocompleteTarget).Info("Bash completion saved")
		}
		os.Exit(0)
	},
}

func init() {
	genautocompleteCmd.PersistentFlags().StringVarP(&autocompleteTarget, "completionfile", "", "", "Autocompletion file")
	genautocompleteCmd.PersistentFlags().StringVarP(&autocompleteType, "type", "", "bash", "Autocompletion type (currently only bash supported)")
}
