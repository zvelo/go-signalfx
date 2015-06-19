package main

import (
	"sync/atomic"
	"time"

	"github.com/zvelo/go-signalfx"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

func main() {
	config := signalfx.NewConfig()
	config.AuthToken = "<YOUR_SIGNALFX_API_TOKEN>" // if $SFX_API_TOKEN is set, this is unnecessary

	reporter := signalfx.NewReporter(config, sfxproto.Dimensions{
		"dim0": "val0",
		"dim1": "val1",
	})

	gauge := reporter.NewGauge("SomeGauge", 5, sfxproto.Dimensions{
		"gauge_dim0": "gauge_val0",
		"gauge_dim1": "gauge_val1",
	})
	// will be reported on Metric "SomeGauge" with integer value 5

	// the timestamp defaults to the time the datapoint was created
	// do this to change it to something specific
	gauge.SetTime(time.Now())

	gauge.Set(9)
	// will now be reported with integer value 9

	inc, incDP := reporter.NewInc("SomeIncrementer", nil)
	inc.Inc(1)
	inc.Inc(5)
	// will be reported on Metric "SomeIncrementer" with integer value 6

	cval := int64(0)
	counter := reporter.NewCounter("SomeCounter", signalfx.Value(&cval), nil)
	reporter.AddPreReportCallback(func() {
		// add 1 to cval just before it is reported
		cval++
	})

	// safely set the value to 1
	atomic.AddInt64(&cval, 1)
	// "SomeCounter" will be reported with value 2 (after the PreReportCallback is executed)

	bucket := reporter.NewBucket("SomeBucket", nil)
	bucket.Add(3)
	bucket.Add(5)
	// 5 DataPoints will be sent.
	// * Metric "SomeBucket" value of 2 with appended dimension "rollup" = "count"
	// * Metric "SomeBucket" value of 8 with appended dimension "rollup" = "sum"
	// * Metric "SomeBucket" value of 34 with appended dimension "rollup" = "sumsquare"
	// * Metric "SomeBucket.min" value of 3 with appended dimension "rollup" = "min"
	// * Metric "SomeBucket.max" value of 5 with appended dimension "rollup" = "max"
	// Min and Max are reset each time bucket is reported

	reporter.Report(context.Background())

	reporter.RemoveDataPoint(gauge, incDP, counter)
	reporter.RemoveBucket(bucket)
}
