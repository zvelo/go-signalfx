package signalfx

import (
	"errors"
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// A BackgroundReporter is a reporter which automatically reports
// periodically.
type BackgroundReporter struct {
	Reporter
	Interval    time.Duration
	controlChan chan controlCommand
}

type controlCommand int

const (
	stopCommand controlCommand = iota
	pauseCommand
	resumeCommand
)

var (
	ErrBackgroundReporterNotStarted = errors.New("Background reporter not started")
)

// NewBackgroundReporter creates a new background reporter, but does not start it; it must be started with BackgroundReporter.Start()
func NewBackgroundReporter(config *Config, defaultDimensions sfxproto.Dimensions, interval time.Duration) *BackgroundReporter {
	return &BackgroundReporter{
		Reporter: *NewReporter(config, defaultDimensions),
		Interval: interval,
	}
}

func (br *BackgroundReporter) Start() {
	br.controlChan = make(chan controlCommand)
	go func() {
		ticker := time.NewTicker(br.Interval)
		for {
			select {
			case <-ticker.C:
				if _, err := br.Reporter.Report(context.Background()); err != nil {
					log.Printf("[ERR] background reporter: %s", err)
				}
			case command := <-br.controlChan:
				switch command {
				case stopCommand:
					ticker.Stop()
					br.controlChan = nil
					return
				case pauseCommand:
					ticker.Stop()
				case resumeCommand:
					ticker = time.NewTicker(br.Interval)
				default:
					log.Printf("[ERR] background reporter: unknown command %d", command)
				}
			}
		}
	}()
}

func (br BackgroundReporter) Pause() error {
	if br.controlChan == nil {
		return ErrBackgroundReporterNotStarted
	}
	br.controlChan <- pauseCommand
	return nil
}

func (br BackgroundReporter) Resume() error {
	// TODO: add better error checking, e.g. for started-but-not-paused
	if br.controlChan == nil {
		return ErrBackgroundReporterNotStarted
	}
	br.controlChan <- resumeCommand
	return nil
}

func (br BackgroundReporter) Report(ctx context.Context) (*DataPoints, error) {
	return br.Reporter.Report(ctx)
}

func (br BackgroundReporter) Stop() error {
	if br.controlChan == nil {
		return ErrBackgroundReporterNotStarted
	}
	br.controlChan <- stopCommand
	return nil
}
