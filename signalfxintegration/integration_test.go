package signalfxintegration

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zvelo/go-signalfx"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func TestIntegration(t *testing.T) {
	Convey("Integration Test", t, func(c C) {
		So(len(os.Getenv("SFX_API_TOKEN")), ShouldBeGreaterThan, 0)
		reporter := signalfx.NewReporter(signalfx.NewConfig(), nil)

		gauge1Val := int64(0)
		gauge1 := reporter.NewGauge("TestReporterIT",
			signalfx.Value(&gauge1Val),
			sfxproto.Dimensions{"metric": "1"})

		gauge2Val := float64(.5)
		gauge2 := reporter.NewGauge("TestReporterIT",
			signalfx.GetterFunc(func() (interface{}, error) { return gauge2Val, nil }),
			sfxproto.Dimensions{"metric": "2"})

		counter1Val := int64(2)
		counter1 := reporter.NewGauge("TestReporterIT",
			signalfx.Value(&counter1Val),
			sfxproto.Dimensions{"metric": "3"})

		ccounter1Val := int64(3)
		ccounter1 := reporter.NewGauge("TestReporterIT",
			signalfx.Value(&ccounter1Val),
			sfxproto.Dimensions{"metric": "3"})

		// For 3 sec, send a point every sec
		for i := 0; i < 3; i++ {
			dps, err := reporter.Report(context.Background())
			So(err, ShouldBeNil)
			So(dps.Len(), ShouldEqual, 4)

			gauge1Val++
			gauge2Val++
			counter1Val++
			ccounter1Val++

			time.Sleep(time.Second)
		}

		// metric 1 sends 0, 1, 2 for average of 1
		// metric 2 sends 0.5, 1.5, 2.5 for average of 1.5
		// metric 3 sends 2, 3, 3, 4, 4, 5 for average of 3.5

		reporter.RemoveDataPoint(gauge1, gauge2, counter1, ccounter1)
	})
}
