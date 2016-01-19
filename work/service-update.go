package work

import (
	"bufio"
	"fmt"
	"github.com/mgutz/ansi"
	"github.com/n0rad/go-erlog/logs"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func (s *Service) Update() error {
	logs.WithFields(s.fields).Info("Updating service")
	s.Generate()

	s.Lock("service/update", 1*time.Hour, "Updating")
	defer s.Unlock("service/update")

	if s.manifest.ConcurrentUpdater > 1 && !BuildFlags.Yes {
		logs.WithFields(s.fields).Fatal("Update concurrently require -y")
	}

	s.concurrentUpdater(s.ListUnits())
	return nil
}

func (s *Service) updateUnit(u Unit) {
ask:
	for {
		same, err := u.IsLocalContentSameAsRemote()
		if err != nil {
			logs.WithEF(err, s.fields).Warn("Cannot compare local and remote service")
		}
		if same {
			logs.WithFields(s.fields).Info("Remote service is already up to date")
			if !u.IsRunning() {
				logs.WithFields(s.fields).Info("But service is not running")
			} else if !BuildFlags.All {
				return
			}
		}

		if BuildFlags.Yes {
			break ask
		}
		action := askToProcessService(u)
		switch action {
		case ACTION_DIFF:
			u.DisplayDiff()
		case ACTION_QUIT:
			logs.WithFields(s.fields).Debug("User want to quit")
			if globalUpdater == 0 {
				s.Unlock("service/update")
			}
			os.Exit(1)
		case ACTION_SKIP:
			logs.WithFields(s.fields).Debug("User skip this service")
			return
		case ACTION_YES:
			break ask
		default:
			logs.WithFields(s.fields).Fatal("Should not be here")
		}
	}

	if atomic.LoadUint32(&globalUpdater) == 0 {
		atomic.AddUint32(&globalUpdater, 1)
		s.runHook(EARLY, "service/update", "update")
		defer s.runHook(LATE, "service/update", "update")
	} else {
		atomic.AddUint32(&globalUpdater, 1)
	}

	logs.WithFields(s.fields).Info("Updating unit")
	u.UpdateInside("service/update")
	time.Sleep(time.Second * 2)

}

var globalUpdater uint32 = 0

func (s *Service) concurrentUpdater(units []string) {
	wg := &sync.WaitGroup{}
	updateChan := make(chan Unit)
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
func askToProcessService(unit Unit) Action {
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
