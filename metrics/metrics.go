package metrics

import (
	"net"
	"time"

	"github.com/cyberdelia/go-metrics-graphite"
	"github.com/rcrowley/go-metrics"

	"github.com/skbkontur/frontreport"
)

// MetricStorage is a Graphite implementation of frontreport.MetricStorage interface
type MetricStorage struct {
	GraphiteConnectionString string
	GraphitePrefix           string
	Logger                   frontreport.Logger
	registry                 metrics.Registry
	histograms               map[string]metrics.Histogram
	counters                 map[string]metrics.Counter
}

// Start initializes Graphite reporter
func (ms *MetricStorage) Start() error {
	ms.registry = metrics.NewRegistry()
	ms.histograms = make(map[string]metrics.Histogram)
	ms.counters = make(map[string]metrics.Counter)

	if ms.GraphiteConnectionString != "" {
		addr, _ := net.ResolveTCPAddr("tcp", ms.GraphiteConnectionString)
		go graphite.Graphite(ms.registry, time.Minute, ms.GraphitePrefix, addr)
	}

	return nil
}

// Stop does nothing - there is no way to gracefully flush Graphite reporter
func (ms *MetricStorage) Stop() error {
	return nil
}

// RegisterHistogram creates a uniform-sampled histogram of integers
func (ms *MetricStorage) RegisterHistogram(name string) {
	ms.histograms[name] = metrics.GetOrRegisterHistogram(name, ms.registry, metrics.NewUniformSample(1000))
}

// UpdateHistogram adds new value to histogram or complains that it is not registered
// NOTE this is very much NOT thread-safe; you should register all your histograms before trying to update any of them
func (ms *MetricStorage) UpdateHistogram(name string, value int) {
	h, ok := ms.histograms[name]
	if ok {
		h.Update(int64(value))
	} else {
		ms.Logger.Log("msg", "somebody tried to use non-registered histogram", "name", name)
	}
}

// RegisterCounter creates a counter
func (ms *MetricStorage) RegisterCounter(name string) {
	ms.counters[name] = metrics.GetOrRegisterCounter(name, ms.registry)
}

// IncCounter adds value to counter or complains that it is not registered
// NOTE this is very much NOT thread-safe; you should register all your counters before trying to increase any of them
func (ms *MetricStorage) IncCounter(name string, value int) {
	c, ok := ms.counters[name]
	if ok {
		c.Inc(int64(value))
	} else {
		ms.Logger.Log("msg", "somebody tried to use non-registered counter", "name", name)
	}
}
