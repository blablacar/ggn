package env

import (
	"bufio"
	"fmt"
	"github.com/blablacar/ggn/builder"
	"github.com/blablacar/ggn/work/env/service"
	"github.com/juju/errors"
	"github.com/mgutz/ansi"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func (s *Service) Update() error {
	s.log.Info("Updating service")
	s.Generate()

	s.Lock("service/update", 1*time.Hour, "Updating")
	defer s.Unlock("service/update")

	units, err := s.ListUnits()
	if err != nil {
		return errors.Annotate(err, "Cannot list units to update")
	}

	if s.manifest.ConcurrentUpdater > 1 && !builder.BuildFlags.Yes {
		s.log.Fatal("Update concurrently require -y")
	}

	s.concurrentUpdater(units)
	return nil
}

func (s *Service) updateUnit(u service.Unit) {
ask:
	for {
		same, err := u.IsLocalContentSameAsRemote()
		if err != nil {
			u.Log.WithError(err).Warn("Cannot compare local and remote service")
		}
		if same {
			u.Log.Info("Remote service is already up to date")
			if !builder.BuildFlags.All {
				return
			}
		}
		if builder.BuildFlags.Yes {
			break ask
		}
		action := askToProcessService(u)
		switch action {
		case ACTION_DIFF:
			u.DisplayDiff()
		case ACTION_QUIT:
			u.Log.Debug("User want to quit")
			if globalUpdater == 0 {
				s.Unlock("service/update")
			}
			os.Exit(1)
		case ACTION_SKIP:
			u.Log.Debug("User skip this service")
			return
		case ACTION_YES:
			break ask
		default:
			u.Log.Fatal("Should not be here")
		}
	}

	if atomic.LoadUint32(&globalUpdater) == 0 {
		atomic.AddUint32(&globalUpdater, 1)
		s.runHook(EARLY, "service/update", "update")
		defer s.runHook(LATE, "service/update", "update")
	} else {
		atomic.AddUint32(&globalUpdater, 1)
	}

	u.UpdateInside("service/update")
	time.Sleep(time.Second * 2)

}

var globalUpdater uint32 = 0

func (s *Service) concurrentUpdater(units []string) {
	wg := &sync.WaitGroup{}
	updateChan := make(chan service.Unit)
	for i := 0; i < s.manifest.ConcurrentUpdater; i++ {
		wg.Add(1)
		go func() {
			for unit := range updateChan {
				s.updateUnit(unit)
			}
			wg.Done()
		}()
	}

	for _, unit := range units {
		u := s.LoadUnit(unit)
		updateChan <- *u
	}
	close(updateChan)
	wg.Wait()
}

//////////////////////////////////////////////////////////
func askToProcessService(unit service.Unit) Action {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Update unit " + ansi.LightGreen + unit.Name + ansi.Reset + " ? : (yes,skip,diff,quit) ")
		text, _ := reader.ReadString('\n')
		t := strings.ToLower(strings.Trim(text, " \n"))
		if t == "o" || t == "y" || t == "ok" || t == "yes" {
			return ACTION_YES
		}
		if t == "s" || t == "skip" {
			return ACTION_SKIP
		}
		if t == "d" || t == "diff" {
			return ACTION_DIFF
		}
		if t == "q" || t == "quit" {
			return ACTION_QUIT
		}
		continue
	}
	return ACTION_QUIT
}
