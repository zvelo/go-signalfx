package sfxreporter

import (
	"sync"
	"time"

	"github.com/zvelo/go-signalfx/sfxclient"
	"github.com/zvelo/go-signalfx/sfxconfig"
	"github.com/zvelo/go-signalfx/sfxproto"
	"golang.org/x/net/context"
)

// DataPointCallback is a functional callback that can be passed to DataPointCallback as a way
// to have the caller calculate and return their own datapoints
type DataPointCallback func(defaultDims sfxproto.Dimensions) *sfxproto.DataPoints

type Reporter struct {
	client              *sfxclient.Client
	defaultDimensions   sfxproto.Dimensions
	datapoints          sfxproto.DataPoints
	buckets             []*Bucket
	preCollectCallbacks []func()
	dataPointCallbacks  []DataPointCallback
	lock                sync.Mutex
}

func New(config *sfxconfig.Config, defaultDimensions sfxproto.Dimensions) *Reporter {
	return &Reporter{
		client:            sfxclient.New(config),
		defaultDimensions: defaultDimensions,
	}
}

func (r *Reporter) Bucket(metricName string, dimensions sfxproto.Dimensions) *Bucket {
	ret := NewBucket(metricName, dimensions)

	r.lock.Lock()
	defer r.lock.Unlock()

	r.buckets = append(r.buckets, ret)
	return ret
}

func (r *Reporter) Cumulative(metricName string, value interface{}) *sfxproto.DataPoint {
	ret := sfxproto.NewDataPoint(sfxproto.MetricType_CUMULATIVE_COUNTER, metricName, value, time.Now(), r.defaultDimensions)
	r.datapoints.Add(ret)
	return ret
}

func (r *Reporter) Gauge(metricName string, value interface{}) *sfxproto.DataPoint {
	ret := sfxproto.NewDataPoint(sfxproto.MetricType_GAUGE, metricName, value, time.Now(), r.defaultDimensions)
	r.datapoints.Add(ret)
	return ret
}

func (r *Reporter) Counter(metricName string, value interface{}) *sfxproto.DataPoint {
	ret := sfxproto.NewDataPoint(sfxproto.MetricType_COUNTER, metricName, value, time.Now(), r.defaultDimensions)
	r.datapoints.Add(ret)
	return ret
}

func (r *Reporter) RemoveDataPoint(dps ...*sfxproto.DataPoint) {
	for _, dp := range dps {
		r.datapoints.Remove(dp)
	}
}

func (r *Reporter) RemoveBucket(bs ...*Bucket) {
	if len(bs) != 1 {
		for _, val := range bs {
			r.RemoveBucket(val)
		}
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for i, b := range r.buckets {
		if bs[0] == b {
			r.buckets = append(r.buckets[:i], r.buckets[i+1:]...)
		}
	}
}

// AddPreCollectCallback adds a function that is called before Report().  This is useful for refetching
// things like runtime.Memstats() so they are only fetched once per report() call
func (r *Reporter) AddPreCollectCallback(f func()) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.preCollectCallbacks = append(r.preCollectCallbacks, f)
}

// AddDataPointCallback adds a callback that itself will generate datapoints to report
func (r *Reporter) AddDataPointCallback(f DataPointCallback) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.dataPointCallbacks = append(r.dataPointCallbacks, f)
}

func (r *Reporter) Report(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for _, f := range r.preCollectCallbacks {
		f()
	}

	dps := r.datapoints.ShallowCopy()

	for _, f := range r.dataPointCallbacks {
		dps.Append(f(r.defaultDimensions))
	}

	for _, b := range r.buckets {
		dps.Append(b.DataPoints(r.defaultDimensions))
	}

	return r.client.Submit(ctx, dps)
}
