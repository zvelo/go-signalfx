/*
Package signalfx provides several mechanisms to easily send datapoints to SignalFx

Users of this package will primarily interact with the Reporter type. Reporter
is an object that tracks DataPoints and manages a Client.

For most cases, a reporter is created as follows (assuming $SFX_API_TOKEN is set
in the environment)

	reporter := signalfx.Newreporter(signalfx.NewConfig(), nil)

DataPoints are created and tracked via various methods on Reporter like:

	inc, _ := reporter.NewInc("SomeIncrementer", nil)
	inc.Inc(1)

And when ready to send the datapoints to SignalFx, just run

	_, err := reporter.Report(context.Background())

There are several types of metrics available, Gauge, Counter, CumulativeCounter
and Bucket. Bucket is provided to track multiple aspects of a single metric such
as number of times data has been added, the sum of each value added, the min and
max values added, etc.

The value for Gauge, Counter and CumulativeCounter may be any type of int,
float, string, nil or pointer or a Getter that returns one of those types. If
val is a pointer type, its value should not be changed, when tracked by a
Reporter, except within a PreReportCallback, for goroutine safety.

The Getter interface can also be used to as a value for a DataPoint. It may only
return any kind of int, float, string, nil or a pointer to one of those types.
Get will be called when Reporter.Report() is executed. Getters must implement
their own goroutine safety with Get as required. There are a few types that
implement Getter that can be used directly.

GetterFunc can wrap a function with the correct signature so that it satisfies
the Getter interface:

	f := signalfx.GetterFunc(func() (interface{}, error) {
		return 5, nil
	})
	reporter.NewGauge("GetterFunc", f, nil)

Value can wrap the basic types (ints, floats, strings, nil and pointers to them)
but is not goroutine safe. Changes should only be made in a PreReportCallback.

	i := 5
	reporter.NewCounter("SomeCounter", signalfx.Value(&i), nil)
	reporter.AddPreReportCallback(func() {
		i++
	})

There are also Int32, Int64, Uint32 and Uint64 types available whose methods are
goroutine safe and can also be safely changed if via atomic methods.

	val := signalfx.NewInt64(0)
	counter := reporter.NewCounter("SomeOtherCounter", val, nil)
	val.Inc(1)
	atomic.AddInt64((*int64)(val), 1)
*/
package signalfx
