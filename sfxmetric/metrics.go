package sfxmetric

import (
	"sync"

	"github.com/zvelo/go-signalfx/sfxproto"
)

// Metrics represents a set of multiple Metrics
type Metrics struct {
	metrics map[*Metric]interface{}
	lock    sync.Mutex
}

// NewMetrics returns a new Metrics object with an expected, length, of l
func NewMetrics(l int) *Metrics {
	return &Metrics{
		metrics: make(map[*Metric]interface{}, l),
	}
}

// Add a Metric to the set
func (ms *Metrics) Add(m *Metric) *Metrics {
	if m != nil {
		ms.lock.Lock()
		defer ms.lock.Unlock()

		ms.metrics[m] = nil
	}

	return ms
}

// Remove Metric(s) from the set. The match is by testing for pointer
// equality, not Metric equality.
func (ms *Metrics) Remove(vals ...*Metric) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	for _, m := range vals {
		delete(ms.metrics, m)
	}
}

// RemoveMetrics removes Metric(s) from the set
func (ms *Metrics) RemoveMetrics(val *Metrics) {
	val.lock.Lock()
	defer val.lock.Unlock()

	for m := range val.metrics {
		ms.Remove(m)
	}
}

// DataPoints returns a sfxproto.DataPoints object representing the underlying
// DataPoints contained in the Metrics object. If a Metric has a Getter, the
// value will be updated before returning.
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

// Len returns the number of metrics the Metrics object contains
func (ms *Metrics) Len() int {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	return len(ms.metrics)
}
