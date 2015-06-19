# go-signalfx

[![Circle CI](https://circleci.com/gh/zvelo/go-signalfx.svg?style=svg)](https://circleci.com/gh/zvelo/go-signalfx) [![Coverage Status](https://coveralls.io/repos/zvelo/go-signalfx/badge.svg?branch=master)](https://coveralls.io/r/zvelo/go-signalfx?branch=master) [![GoDoc](https://godoc.org/github.com/zvelo/go-signalfx?status.svg)](https://godoc.org/github.com/zvelo/go-signalfx)

Provides a simple way to send DataPoints to SignalFx

Fully documented via [godoc](https://godoc.org/github.com/zvelo/go-signalfx).

1. Create a `Config` object. If `$SFX_API_TOKEN` is set in the environment, it will be used within the `Config`. Other default values are generally acceptable for most uses.

    ```go
    config := signalfx.NewConfig()
    config.AuthToken = "<YOUR_SIGNALFX_API_TOKEN>" // if $SFX_API_TOKEN is set, this is unnecessary 
    ```

2. Create a `Reporter` object and set any dimensions that should be set on evey metric it sends to SignalFx.

    ```go
    reporter := signalfx.NewReporter(config, sfxproto.Dimensions{
        "dim0": "val0",
        "dim1": "val1",
    })
    ```

3. Add static DataPoints as needed, the value will be sent to SignalFx later when `reporter.Report` is called. All operations on the DataPoint are goroutine safe.

    ```go
    gauge := reporter.NewGauge("SomeGauge", 5, sfxproto.Dimensions{
        "gauge_dim0": "gauge_val0",
        "gauge_dim1": "gauge_val1",
    })
    // will be reported on Metric "SomeGauge" with integer value 5

    // the timestamp defaults to the time the datapoint was created
    // do this to change it to something specific
    gauge.SetTime(time.Now())
    ```

4. Even the value can even be changed before reporting.

    ```go
    gauge.Set(9)
    // will now be reported with integer value 9
    ```

5. Incrementers are a special case, there is one for Counter Metric Types, and another for Cumulative Counters. All operations on Incrementers are goroutine safe.

    ```go
    inc, _ := reporter.NewInc("SomeIncrementer", nil)
    inc.Inc(1)
    inc.Inc(5)
    // will be reported on Metric "SomeIncrementer" with integer value 6
    ```

6. Sometimes it's necessary that a DataPoint retrieve the value at the time of reporting, not when it is created. This can be done by using the `signalfx.Getter` interface.
   `signalfx.Value` implements this interface and can wrap any kind of `int`, `float`, `string`, `nil` or pointer to those.
   It's most useful with pointers though as the value of the pointer at the time of the report is what is sent.
   It is important to note that **changing the value of a pointer should only be done (a) atomically or (b) in a `PreReportCallback`**.

    ```go
    cval := int64(0)
    reporter.NewCounter("SomeCounter", signalfx.Value(&cval), nil)
    reporter.AddPreReportCallback(func() {
        // add 1 to cval just before it is reported
        cval++
    })

    // safely set the value to 1
    atomic.AddInt64(&cval, 1)
    // "SomeCounter" will be reported with value 2 (after the PreReportCallback is executed)
    ```

8. `Bucket` is also provided to help with reporting multiple aspects of a Metric simultaneously. All operations on `Bucket` are goroutine safe.

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
