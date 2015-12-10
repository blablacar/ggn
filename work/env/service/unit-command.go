package service

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/builder"
	"os"
	"strconv"
	"time"
)

func (u *Unit) Start(command string) error {
	if u.isRunning() {
		u.Log.Info("Service is already running")
		return nil
	}
	if !u.isLoaded() {
		u.Log.Debug("unit is not loaded yet")
		u.Service.Generate(nil)
		u.Load(command)
	} else {
		u.Log.Debug("unit is already loaded")
	}
	return u.runAction(command, "start")
}

func (u *Unit) Unload(command string) error {
	return u.runAction(command, "unload")
}

func (u *Unit) Load(command string) error {
	u.Service.Generate(nil)
	return u.runAction(command, "load")
}

func (u *Unit) Stop(command string) error {
	return u.runAction(command, "stop")
}

func (u *Unit) Destroy(command string) error {
	return u.runAction(command, "destroy")
}

func (u *Unit) Restart(command string) error {
	u.Log.Debug("restart")
	u.runHook(EARLY, command, "restart")
	defer u.runHook(LATE, command, "restart")

	u.Service.Lock(command, 1*time.Hour, "Restart "+u.Name)
	defer u.Service.Unlock(command)

	u.Stop(command)
	time.Sleep(time.Second * 2)
	u.Start(command)

	return nil
}

func (u *Unit) Update(command string) error {
	u.Service.Generate(nil)
	u.Log.Debug("Update")
	u.runHook(EARLY, command, "update")
	defer u.runHook(LATE, command, "update")

	u.Service.Lock(command, 1*time.Hour, "Update "+u.Name)
	defer u.Service.Unlock(command)

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.WithError(err).Warn("Cannot compare local and remote service")
	}
	if same && !builder.BuildFlags.Force {
		u.Log.Info("Remote service is already up to date, not updating")
		return nil
	}

	u.UpdateInside(command)

	return nil
}

func (u *Unit) Journal(command string, follow bool, lines int) {
	u.Log.Debug("journal")
	u.runHook(EARLY, command, "journal")
	defer u.runHook(LATE, command, "journal")

	args := []string{"journal", "-lines", strconv.Itoa(lines)}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, u.Filename)

	err := u.Service.GetEnv().RunFleetCmd(args...)

	if err != nil && !follow {
		logrus.WithError(err).Fatal("Failed to run journal")
	}
}

func (u *Unit) Ssh(command string) {
	u.Log.Debug("ssh")
	u.runHook(EARLY, command, "ssh")
	defer u.runHook(LATE, command, "ssh")

	err := u.Service.GetEnv().RunFleetCmd("ssh", u.Filename)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to run status")
	}
}

func (u *Unit) Diff(command string) {
	u.Log.Debug("diff")
	u.Service.Generate(nil)
	u.runHook(EARLY, command, "diff")
	defer u.runHook(LATE, command, "diff")

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.Warn("Cannot read unit")
	}
	if !same {
		u.DisplayDiff()
	}
}

func (u *Unit) Status(command string) {
	u.Log.Debug("status")
	u.runHook(EARLY, command, "status")
	defer u.runHook(LATE, command, "status")

	err := u.Service.GetEnv().RunFleetCmd("status", u.Filename)
	if err != nil {
		os.Exit(1)
	}
}

////////////////////////////////////////////

func (u *Unit) runAction(command string, action string) error {
	if command == action {
		u.Service.Lock(command, 1*time.Hour, action+" "+u.Name)
		defer u.Service.Unlock(command)
	}

	u.Log.Debug(action)
	u.runHook(EARLY, command, action)
	defer u.runHook(LATE, command, action)

	_, _, err := u.Service.GetEnv().RunFleetCmdGetOutput(action, u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot " + action + " unit")
		return err
	}
	return nil
}
