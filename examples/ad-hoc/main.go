package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/zvelo/go-signalfx"
)

func main() {
	config := signalfx.NewConfig()
	//config.AuthToken = "<YOUR_SIGNALFX_API_TOKEN>" // if $SFX_API_TOKEN is set, this is unnecessary
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "no hostname:", err)
		os.Exit(1)
	}
	hostname += "-" + os.Args[1]

	reporter := signalfx.NewReporter(config, map[string]string{
		"process":  "ad-hoc",
		"hostname": hostname,
	})

	// a gauge is a point-in-time value; it can positive or negative
	gaugeWalker := newRandomWalker(0, 256)
	gauge := signalfx.NewInt64(0)
	g := reporter.NewGauge("test-gauge", gauge,
		map[string]string{
			"hostname": hostname,
			"process":  "ad-hoc",
		})
	_ = g
	cancelFunc := reporter.RunInBackground(time.Second)

	// a cumulative counter is a good choice to wrap an internal value
	var requestCount = signalfx.Uint64(0)
	reporter.NewCumulative("test-cumulative-counter",
		&requestCount,
		map[string]string{
			"hostname": hostname,
		})

	// a counter is a good choice for a value we want to create, then report

	maxInc := big.NewInt(1024)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := int64(0); i < 288; i++ {
			inc, err := rand.Int(rand.Reader, maxInc)
			if err != nil {
				fmt.Fprintln(os.Stderr, "failed to generate random increment:", err)
			} else {
				reporter.Inc("test-counter", map[string]string{"hostname": hostname}, inc.Int64())
				inc.Add(inc, big.NewInt(int64(requestCount)))
				requestCount.Set(inc.Uint64())
			}
			time.Sleep(time.Second * 2)
		}
	}()

	go func() {
		for {
			gaugeWalker.inc()
			gauge.Set(gaugeWalker.value.Int64())
			time.Sleep(time.Second * 2)
		}
	}()

	wg.Wait()
	cancelFunc()
}

type randomWalker struct {
	value        *big.Int
	maxIncrement *big.Int
	halfMax      *big.Int
}

func newRandomWalker(initial, maxIncrement int64) randomWalker {
	m := big.NewInt(maxIncrement)
	hm := &big.Int{}
	hm.Div(m, big.NewInt(2))
	return randomWalker{
		value:        big.NewInt(initial),
		maxIncrement: m,
		halfMax:      hm,
	}
}

func (r randomWalker) inc() {
	inc, err := rand.Int(rand.Reader, r.maxIncrement)
	if err != nil {
		panic(fmt.Sprint("error generating random data: ", err))
	}
	inc.Sub(inc, r.halfMax)
	r.value.Add(r.value, inc)
}
