package service

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/builder"
	"time"
)

func (u Unit) Start() error {
	u.Service.Generate(nil)
	return u.runAction("start")
}

func (u Unit) Unload() error {
	return u.runAction("unload")
}

func (u Unit) Load() error {
	return u.runAction("load")
}

func (u Unit) Stop() error {
	return u.runAction("stop")
}

func (u Unit) Destroy() error {
	return u.runAction("destroy")
}

func (u Unit) Update(lock bool) error {
	u.Service.Generate(nil)
	u.Log.Debug("Update")

	if lock {
		u.Service.Lock(1*time.Hour, "Update "+u.Name)
		u.Service.Unlock()
	}

	same, err := u.IsLocalContentSameAsRemote()
	if err != nil {
		u.Log.WithError(err).Warn("Cannot compare local and remote service")
	}
	if same && !builder.BuildFlags.Force {
		u.Log.Info("Remote service is already up to date, not updating")
		return nil
	}

	u.Service.GetEnv().RunEarlyHook(u.Name, "update")
	defer u.Service.GetEnv().RunLateHook(u.Name, "update")

	// destroy
	_, _, err = u.Service.GetEnv().RunFleetCmdGetOutput("destroy", u.Filename)
	if err != nil {
		logrus.WithError(err).Warn("Cannot destroy unit")
	}

	// start
	time.Sleep(time.Second * 2)
	_, _, err = u.Service.GetEnv().RunFleetCmdGetOutput("start", u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot start unit")
		return err
	}

	//		status, err2 := u.Status()
	//		u.Log.WithField("status", status).Debug("Log status")
	//		if err2 != nil {
	//			log.WithError(err2).WithField("status", status).Panic("Unit failed just after start")
	//			return err2
	//		}
	//		if status == "inactive" {
	//			log.WithField("status", status).Panic("Unit failed just after start")
	//			return errors.New("unit is inactive just after start")
	//		}
	//
	//		s.checkServiceRunning()

	// TODO ask deploy pod version ()
	// TODO YES/NO
	// TODO check running tmux
	// TODO running as root ??
	// TODO notify slack
	// TODO store old version
	// TODO !!!!! check that service is running well before going to next server !!!
	return nil
}

/////////////////////////////

func (u Unit) runAction(action string) error {
	u.Service.Lock(1*time.Hour, action+" "+u.Name)
	u.Service.Unlock()
	u.Log.Debug(action)
	u.Service.GetEnv().RunEarlyHook(u.Name, action)
	defer u.Service.GetEnv().RunLateHook(u.Name, action)

	_, _, err := u.Service.GetEnv().RunFleetCmdGetOutput(action, u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot " + action + " unit")
		return err
	}
	return nil
}
