# go-signalfx

[![Circle CI](https://circleci.com/gh/zvelo/go-signalfx.svg?style=svg)](https://circleci.com/gh/zvelo/go-signalfx) [![Coverage Status](https://coveralls.io/repos/zvelo/go-signalfx/badge.svg?branch=master)](https://coveralls.io/r/zvelo/go-signalfx?branch=master) [![GoDoc](https://godoc.org/zvelo.io/go-signalfx?status.svg)](https://godoc.org/zvelo.io/go-signalfx)

Provides a simple way to send DataPoints to SignalFx

Fully documented via [godoc](https://godoc.org/zvelo.io/go-signalfx).

## Changes

This release greatly changes the API.  It cleanly separates metrics
and their data points; this change also extends to ownership semantics
and goroutine cleanliness.

### Separation of single data points and metric time series

Added `Reporter.Inc`, `Reporter.Sample` and `Reporter.Record` for
one-shot counter, cumulative-counter and gauge values.

#### Metrics

Metrics are a new concept: they represent metric time series, which
have internal state and may be converted to a data point at a
particular point in time.

Client code may define its own metrics, however, convenient-to-use
Counter, WrappedCounter, CumulativeCounter, WrappedCumulativeCounter,
Gauge and WrappedGauge types are provided.

To track metrics over time, use `Reporter.Track` to start tracking
them and `Reporter.Untrack` to stop tracking them.

### No need for sfxproto

Client code should no longer need to know about the `sfxproto`
library, which is used internally by `signalfx`.

### Argument order

Function arguments should go from most general to most specific,
e.g. from metric name, to dimensions, to value.

## Simple usage

1. Create a `Config` object. If `$SFX_API_TOKEN` is set in the
   environment, it will be used within the `Config`. Other default
   values are generally acceptable for most uses.

    ```go
    config := signalfx.NewConfig()
    config.AuthToken = "<YOUR_SIGNALFX_API_TOKEN>" // if $SFX_API_TOKEN is set, this is unnecessary
    ```

2. Create a `Reporter` object and set any dimensions that should be
   set on every metric it sends to SignalFx.  Optionally, call
   `Reporter.SetPrefix` to set a metric prefix which will be prepended
   to every metric that reporter reports (this can be used to enforce
   hard environment separation).

    ```go
    reporter := signalfx.NewReporter(config, map[string]string{
        "dim0": "val0",
        "dim1": "val1",
    })
    ```

3. Add static DataPoints as needed, the value will be sent to SignalFx
   later when `reporter.Report` is called. All operations on the
   DataPoint are goroutine safe.

    ```go
    reporter.Record(
                     "SomeGauge",
                     sfxproto.Dimensions{
                       "gauge_dim0": "gauge_val0",
                       "gauge_dim1": "gauge_val1",
                     },
                     5,
                   )
    // will be reported on Metric "SomeGauge" with integer value 5
    ```

5. To track a metric over time, use a Metric:

    ```go
    counter := NewCounter("counter-metric-name", nil, 0)
    reporter.Track(counter)
    ⋮
    counter.Inc(1)
    ⋮
    counter.Inc(3)
    Reporter.Report(…) // will report a counter value of 4
    ```

6. `Bucket` is also provided to help with reporting multiple aspects of a Metric simultaneously. All operations on `Bucket` are goroutine safe.

    ```go
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
    ```

7. When ready to send the DataPoints to SignalFx, just `Report` them.

    ```go
    reporter.Report(context.Background())
    ```
