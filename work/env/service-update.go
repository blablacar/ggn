package env

import (
	"github.com/blablacar/ggn/builder"
	"github.com/juju/errors"
	"time"
)

func (s *Service) Update() error {
	s.log.Info("Updating service")
	s.Generate()

	s.Lock("service/update", 1*time.Hour, "Updating")
	lock := true
	defer func() {
		if lock {
			s.log.WithField("service", s.Name).Warn("!! Leaving but Service is still lock !!")
		}
	}()

	units, err := s.ListUnits()
	if err != nil {
		return errors.Annotate(err, "Cannot list units to update")
	}
units:
	for i, unit := range units {
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
					s.Unlock("service/update")
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

		if i == 0 {
			s.runHook(EARLY, "service/update", "update")
			defer s.runHook(LATE, "service/update", "update")
		}

		u.UpdateInside("service/update")
		time.Sleep(time.Second * 2)

	}
	s.Unlock("service/update")
	lock = false
	return nil
}
