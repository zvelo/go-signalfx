package signalfxintegration

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"zvelo.io/go-signalfx"
	"zvelo.io/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func TestIntegration(t *testing.T) {
	Convey("Integration Test", t, func(c C) {
		So(len(os.Getenv("SFX_API_TOKEN")), ShouldBeGreaterThan, 0)
		reporter := signalfx.NewReporter(signalfx.NewConfig(), nil)

		gauge1Val := int64(0)
		gauge1 := signalfx.WrapGauge(
			"TestReporterIT",
			map[string]string{
				"metric": "1",
			},
			signalfx.Value(&gauge1Val),
		)

		/* TODO(buhl): confirm that SignalFX support floating-point values in gauges */
		// gauge2Val := float64(.5)
		// gauge2 := signalfx.WrapGauge(
		// 	"TestReporterIT",
		// 	map[string]string{"metric": "2"},
		// 	signalfx.GetterFunc(func() (interface{}, error) { return gauge2Val, nil }),
		// )

		counter1Val := int64(2)
		// FIXME: is there a reason that this counter is a gauge?
		counter1 := signalfx.WrapGauge(
			"TestReporterIT",
			sfxproto.Dimensions{"metric": "3"},
			signalfx.Value(&counter1Val),
		)

		ccounter1Val := int64(3)
		// FIXME: is there a reason that this counter is a gauge?
		ccounter1 := signalfx.WrapGauge(
			"TestReporterIT",
			sfxproto.Dimensions{"metric": "3"},
			signalfx.Value(&ccounter1Val),
		)
		// TODO(buhl): add gauge2 back when line 29 is resolved
		reporter.Track(gauge1, counter1, ccounter1)

		// For 3 sec, send a point every sec
		for i := 0; i < 3; i++ {
			dps, err := reporter.Report(context.Background())
			So(err, ShouldBeNil)
			// TODO(buhl): change this to 3 when the above TODO on line 29 is resolved
			So(len(dps), ShouldEqual, 3)

			gauge1Val++
			// TODO(buhl): uncomment when line 29 is resolved
			// gauge2Val++
			counter1Val++
			ccounter1Val++

			time.Sleep(time.Second)
		}

		// metric 1 sends 0, 1, 2 for average of 1
		// metric 2 sends 0.5, 1.5, 2.5 for average of 1.5
		// metric 3 sends 2, 3, 3, 4, 4, 5 for average of 3.5

		// TODO(buhl): add gauge2 back when line 29 is resolved
		reporter.Untrack(gauge1, counter1, ccounter1)
	})
}
