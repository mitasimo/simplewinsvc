package simplewinsvc

import (
	"time"

	"golang.org/x/sys/windows/svc"
)

// Service is an interface that must be implemented by an object to run as a Windows service
//
// Start must start the service in a non-blocking way
// Stop is intended for stopping the service
type Service interface {
	Start() error
	Stop() error
}

// Run accesses the Windows service Manager to start the service
func Run(svcName string, service Service) error {

	return svc.Run(svcName, serviceHandler{service})
}

type serviceHandler struct {
	Service
}

// Execute implements the svc.Handler interface
func (wh serviceHandler) Execute(args []string, changes <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {

	status <- svc.Status{State: svc.StartPending}

	err := wh.Start()
	if err != nil {
		status <- svc.Status{State: svc.Stopped}
		return
	}

	status <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

loop:
	for {
		c := <-changes
		switch c.Cmd {
		case svc.Interrogate:
			status <- c.CurrentStatus
			// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
			time.Sleep(100 * time.Millisecond)
			status <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			status <- svc.Status{State: svc.StopPending}
			err = wh.Stop()
			break loop
		default:
			status <- c.CurrentStatus
		}
	}

	status <- svc.Status{State: svc.Stopped}
	return
}
