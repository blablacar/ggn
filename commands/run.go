package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func run(cmd *cobra.Command, args []string, work *work.Work, env string) {
	logrus.WithField("env", env).Debug("Running command")
	work.LoadEnv(env).Run(args)
}
