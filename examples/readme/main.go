package main

import (
	"sync/atomic"

	"zvelo.io/go-signalfx"
	"golang.org/x/net/context"
)

func main() {
	config := signalfx.NewConfig()
	config.AuthToken = "<YOUR_SIGNALFX_API_TOKEN>" // if $SFX_API_TOKEN is set, this is unnecessary

	reporter := signalfx.NewReporter(config, map[string]string{
		"dim0": "val0",
		"dim1": "val1",
	})

	gaugeVal := 5
	gauge := signalfx.WrapGauge(
		"SomeGauge",
		map[string]string{
			"gauge_dim0": "gauge_val0",
			"gauge_dim1": "gauge_val1",
		},
		signalfx.Value(&gaugeVal),
	)
	reporter.Track(gauge)
	// would be reported on Metric "SomeGauge" with integer value 5

	gaugeVal = 9
	// would now be reported with integer value 9

	f := signalfx.GetterFunc(func() (interface{}, error) {
		return 5, nil
	})
	gauge2 := signalfx.WrapGauge("GetterFunc", nil, f)
	reporter.Track(gauge2)
	// will be reported on Metric "GetterFunc" with integer value 5

	cval := 5
	counter := signalfx.WrapGauge("SomeCounter", nil, signalfx.Value(&cval))
	reporter.AddPreReportCallback(func() {
		// add 1 to cval just before it is reported
		cval++
	})
	reporter.Track(counter)
	// "SomeCounter" will be reported with value 6

	i := signalfx.NewInt64(0)
	iMetric := signalfx.WrapCounter("SomeInt64", nil, i)
	i.Set(7)
	atomic.AddInt64((*int64)(i), 2)
	i.Inc(1)
	i.Inc(5)
	// will be reported on Metric "SomeInt64" with integer value 15

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

	reporter.Untrack(gauge, iMetric, counter)
	reporter.RemoveBucket(bucket)
}
