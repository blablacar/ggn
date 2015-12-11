package env

import (
	"sync"
)

func (s *Service) Check() {
	s.log.Debug("Running check")
	s.runHook(EARLY, "service/check", "check")
	defer s.runHook(LATE, "service/check", "check")

	s.concurrentChecker(s.ListUnits())

	//	for _, unitName := range s.ListUnits() {
	//		s.LoadUnit(unitName).Check("service/check")
	//	}
}

func (s *Service) concurrentChecker(units []string) {
	wg := &sync.WaitGroup{}
	aChan := make(chan string)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			for unit := range aChan {
				s.LoadUnit(unit).Check("service/check")
			}
			wg.Done()
		}()
	}

	for _, unit := range units {
		aChan <- unit
	}
	close(aChan)
	wg.Wait()
}
