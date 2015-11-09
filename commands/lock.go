package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
	"strings"
	"time"
)

func unLock(cmd *cobra.Command, args []string, work *work.Work, env string, service string) {
	work.LoadEnv(env).LoadService(service).Unlock()
}

func lock(cmd *cobra.Command, args []string, work *work.Work, env string, service string, duration string) {
	if len(args) == 0 {
		logrus.Fatal("Please add a message to describe lock")
	}

	message := strings.Join(args, " ")
	ttl, err := time.ParseDuration(duration)
	if err != nil {
		logrus.WithError(err).Fatal("Wrong value for ttl")
	}

	work.LoadEnv(env).LoadService(service).Lock(ttl, message)
}
