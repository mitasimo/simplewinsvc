package simplewinsvc

import (
	"time"

	"golang.org/x/sys/windows/svc"
)

// Starter is
type Starter interface {
	Start() error
}

// Stopper is
type Stopper interface {
	Stop() error
}

// Run -
func Run(svcName string, starter Starter, stopper Stopper) error {

	return svc.Run(svcName, wrapperHandler{starter, stopper})
}

type wrapperHandler struct {
	starter Starter
	stopper Stopper
}

func (wh wrapperHandler) Execute(args []string, changes <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {

	status <- svc.Status{State: svc.StartPending}

	err := wh.starter.Start()
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
			err = wh.stopper.Stop()
			break loop
		default:
			status <- c.CurrentStatus
		}
	}

	status <- svc.Status{State: svc.Stopped}
	return
}
