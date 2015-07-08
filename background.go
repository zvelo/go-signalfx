package signalfx

import (
	"errors"
	"log"
	"time"
)

type BackgroundJob chan backgroundCommand

type backgroundCommand int

const (
	pauseCommand backgroundCommand = iota
	resumeCommand
	doCommand
)

var (
	// ErrBackgroundJobNotStarted indicates that a background job
	// has not been started (i.e., it's nil).
	ErrBackgroundJobNotStarted = errors.New("Background job not started")
)

func Background(interval time.Duration, doFunc func()) BackgroundJob {
	controlChan := make(BackgroundJob)
	go func() {
		paused := false
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				doFunc()
			case command, ok := <-controlChan:
				if !ok {
					ticker.Stop()
					return
				}
				switch command {
				case pauseCommand:
					ticker.Stop()
					paused = true
				case resumeCommand:
					if paused {
						ticker = time.NewTicker(interval)
						paused = false
					}
				case doCommand:
					ticker.Stop()
					doFunc()
					ticker = time.NewTicker(interval)
				default:
					log.Printf("[ERR] background reporter: unknown command %d", command)
				}
			}
		}
	}()
	return controlChan
}

func (b BackgroundJob) Pause() error {
	if b == nil {
		return ErrBackgroundJobNotStarted
	}
	b <- pauseCommand
	return nil
}

func (b BackgroundJob) Resume() error {
	// TODO: add better error checking, e.g. for started-but-not-paused
	if b == nil {
		return ErrBackgroundJobNotStarted
	}
	b <- resumeCommand
	return nil
}

// Stop stops a background job.  It is an error to attempt to stop an
// already-stopped job.
func (b *BackgroundJob) Stop() error {
	if b == nil || *b == nil {
		return ErrBackgroundJobNotStarted
	}
	close(*b)
	*b = nil
	return nil
}

func (b BackgroundJob) Do() error {
	if b == nil {
		return ErrBackgroundJobNotStarted
	}
	b <- doCommand
	return nil
}
