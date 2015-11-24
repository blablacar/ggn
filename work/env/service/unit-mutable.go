package service

import (
	"github.com/Sirupsen/logrus"
	"time"
)

func (u Unit) Start() error {
	return u.runAction("start")
}

func (u Unit) Unload() error {
	return u.runAction("unload")
}

func (u Unit) Stop() error {
	return u.runAction("stop")
}

func (u Unit) Destroy() error {
	return u.runAction("destroy")
}

func (u Unit) Update() error {
	u.Log.Debug("Update")
	u.Service.GetEnv().RunEarlyHook(u.Name, "Update")
	defer u.Service.GetEnv().RunLateHook(u.Name, "Update")

	// destroy
	_, _, err := u.Service.GetEnv().RunFleetCmdGetOutput("destroy", u.Filename)
	if err != nil {
		logrus.WithError(err).Warn("Cannot destroy unit")
		return err
	}

	// start
	time.Sleep(time.Second * 2)
	_, _, err = u.Service.GetEnv().RunFleetCmdGetOutput("start", u.unitPath)
	if err != nil {
		logrus.WithError(err).Error("Cannot start unit")
		return err
	}
	return nil
}

/////////////////////////////

func (u Unit) runAction(action string) error {
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
