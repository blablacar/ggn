package work

import (
	"os"
	"strconv"
	"time"

	"github.com/n0rad/go-erlog/logs"
)

func (u *Unit) Start(command string) error {
	if u.IsRunning() {
		logs.WithFields(u.Fields).Info("Service is already running")
		return nil
	}
	if !u.IsLoaded() {
		logs.WithFields(u.Fields).Debug("unit is not loaded yet")
		if err := u.Service.Generate(); err != nil {
			logs.WithEF(err, u.Fields).Fatal("Generate failed")
			return err
		}
		u.Load(command)
	} else {
		logs.WithFields(u.Fields).Debug("unit is already loaded")
	}
	return u.runAction(command, "start")
}

func (u *Unit) Unload(command string) error {
	return u.runAction(command, "unload")
}

func (u *Unit) Load(command string) error {
	if err := u.Service.Generate(); err != nil {
		logs.WithEF(err, u.Fields).Fatal("Generate failed")
	}
	return u.runAction(command, "load")
}

func (u *Unit) Stop(command string) error {
	return u.runAction(command, "stop")
}

func (u *Unit) Destroy(command string) error {
	return u.runAction(command, "destroy")
}

func (u *Unit) Restart(command string) error {
	logs.WithFields(u.Fields).Debug("restart")
	if u.Type == TYPE_SERVICE && u.Service.HasTimer() {
		logs.WithFields(u.Fields).Fatal("You cannot restart a service associated to a time")
	}

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
	if err := u.Service.Generate(); err != nil {
		logs.WithEF(err, u.Fields).Fatal("Generate failed")
	}
	logs.WithFields(u.Fields).Debug("Update")
	u.runHook(EARLY, command, "update")
	defer u.runHook(LATE, command, "update")

	u.Service.Lock(command, 1*time.Hour, "Update "+u.Name)
	defer u.Service.Unlock(command)

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		logs.WithEF(err, u.Fields).Warn("Cannot compare local and remote service")
	}
	if same {
		logs.WithFields(u.Fields).Info("Remote service is already up to date")
		if !u.IsRunning() {
			logs.WithFields(u.Fields).Info("But service is not running")
		} else if !BuildFlags.Force {
			return nil
		}
	}

	u.UpdateInside(command)

	return nil
}

func (u *Unit) Journal(command string, follow bool, lines int) {
	logs.WithFields(u.Fields).Debug("journal")
	u.runHook(EARLY, command, "journal")
	defer u.runHook(LATE, command, "journal")

	args := []string{"journal", "-lines", strconv.Itoa(lines)}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, u.Filename)

	err := u.Service.GetEnv().RunFleetCmd(args...)

	if err != nil && !follow {
		logs.WithEF(err, u.Fields).Fatal("Failed to run journal")
	}
}

func (u *Unit) Ssh(command string) {
	logs.WithFields(u.Fields).Debug("ssh")
	u.runHook(EARLY, command, "ssh")
	defer u.runHook(LATE, command, "ssh")

	err := u.Service.GetEnv().RunFleetCmd("ssh", u.Filename)
	if err != nil {
		logs.WithEF(err, u.Fields).Fatal("Failed to run status")
	}
}

func (u *Unit) Diff(command string) {
	logs.WithFields(u.Fields).Debug("diff")
	if err := u.Service.Generate(); err != nil {
		logs.WithEF(err, u.Fields).Fatal("Generate failed")
	}
	u.runHook(EARLY, command, "diff")
	defer u.runHook(LATE, command, "diff")

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		logs.WithFields(u.Fields).Warn("Cannot read unit")
	}
	if !same {
		u.DisplayDiff()
	}
}

func (u *Unit) Status(command string) {
	logs.WithFields(u.Fields).Debug("status")
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

	logs.WithFields(u.Fields).Debug(action)
	u.runHook(EARLY, command, action)
	defer u.runHook(LATE, command, action)

	_, _, err := u.Service.GetEnv().RunFleetCmdGetOutput(action, u.unitPath)
	if err != nil {
		logs.WithEF(err, u.Fields).Error("Cannot " + action + " unit")
		return err
	}
	return nil
}
