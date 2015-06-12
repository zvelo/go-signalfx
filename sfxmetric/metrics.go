package sfxmetric

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

type Metrics struct {
	metrics map[*Metric]interface{}
	lock    sync.Mutex
}

func NewMetrics(l int) *Metrics {
	return &Metrics{
		metrics: make(map[*Metric]interface{}, l),
	}
}

func (ms *Metrics) Add(m *Metric) *Metrics {
	if m != nil {
		ms.lock.Lock()
		defer ms.lock.Unlock()

		ms.metrics[m] = nil
	}

	return ms
}

func (ms *Metrics) Remove(vals ...*Metric) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, m := range vals {
		delete(ms.metrics, m)
	}
}

func (ms *Metrics) RemoveMetrics(val *Metrics) {
	val.lock.Lock()
	defer val.lock.Unlock()

	for m := range val.metrics {
		ms.Remove(m)
	}
}

func (ms *Metrics) DataPoints() (*sfxproto.DataPoints, error) {
	ret := sfxproto.NewDataPoints(ms.Len())

	ms.lock.Lock()
	defer ms.lock.Unlock()

	for m := range ms.metrics {
		if err := m.update(); err != nil {
			return nil, err
		}

		ret.Add(m.dp)
	}

	return ret, nil
}

func (ms *Metrics) Len() int {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	return len(ms.metrics)
}
