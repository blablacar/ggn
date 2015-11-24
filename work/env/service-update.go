package env

import (
	"errors"
	"github.com/blablacar/green-garden/builder"
	"os"
	"time"
)

func (s *Service) Update() error {
	s.log.Info("Updating service")
	s.Generate(nil)

	hostname, _ := os.Hostname()
	s.Lock(time.Hour*1, "["+os.Getenv("USER")+"@"+hostname+"] Updating")
	lock := true
	defer func() {
		if lock {
			s.log.WithField("service", s.Name).Warn("!! Leaving but Service is still lock !!")
		}
	}()
units:
	for i, unit := range s.ListUnits() {
		u := s.LoadUnit(unit)

	ask:
		for {
			same, err := u.IsLocalContentSameAsRemote()
			if err != nil {
				u.Log.WithError(err).Warn("Cannot compare local and remote service")
			}
			if same {
				u.Log.Info("Remote service is already up to date")
				if !builder.BuildFlags.All {
					continue units
				}
			}
			if builder.BuildFlags.Yes {
				break ask
			}
			action := s.askToProcessService(i, u)
			switch action {
			case ACTION_DIFF:
				u.DisplayDiff()
			case ACTION_QUIT:
				u.Log.Debug("User want to quit")
				if i == 0 {
					s.Unlock()
					lock = false
				}
				return errors.New("User want to quit")
			case ACTION_SKIP:
				u.Log.Debug("User skip this service")
				continue units
			case ACTION_YES:
				break ask
			default:
				u.Log.Fatal("Should not be here")
			}
		}

		u.Destroy()
		time.Sleep(time.Second * 2)
		err := u.Start()
		if err != nil {
			s.log.WithError(err).Error("Failed to start service. Keeping lock")
			return err
		}
		time.Sleep(time.Second * 2)
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

	}
	s.Unlock()
	lock = false
	return nil
}
