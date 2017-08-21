package work

import (
	"sync"

	"github.com/n0rad/go-erlog/logs"
)

func (e Env) Check() {
	e.Generate()
	logs.WithFields(e.fields).Debug("Running check")

	info := HookInfo{Command: "env/check", Action: "env/check"}
	e.RunEarlyHookFatal(info)
	defer e.RunLateHookFatal(info)

	e.concurrentChecker(e.ListServices())
}

func (e Env) concurrentChecker(services []string) {
	wg := &sync.WaitGroup{}
	aChan := make(chan string)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			for service := range aChan {
				if err := e.LoadService(service).Check(); err != nil {
					logs.WithE(err).WithField("service", service).Error("Check failed")
				}
			}
			wg.Done()
		}()
	}

	for _, service := range services {
		aChan <- service
	}
	close(aChan)
	wg.Wait()
}
