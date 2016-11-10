package metrics

import (
	"fmt"
	"net"
	"os"
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
}

// Start initializes Graphite reporter
func (ms *MetricStorage) Start() error {
	ms.registry = metrics.NewRegistry()

	if ms.GraphiteConnectionString != "" {
		addr, err := net.ResolveTCPAddr("tcp", ms.GraphiteConnectionString)
		if err != nil {
			ms.Logger.Log("msg", "error resolving Graphite connection string", "error", err)
		} else {
			prefix := ms.GraphitePrefix
			hostname, err := os.Hostname()
			if err == nil {
				prefix = fmt.Sprintf("%s.%s", prefix, hostname)
			}
			go graphite.Graphite(ms.registry, time.Minute, prefix, addr)
		}
	}

	return nil
}

// Stop does nothing - there is no way to gracefully flush Graphite reporter
func (ms *MetricStorage) Stop() error {
	return nil
}

// RegisterHistogram creates a uniform-sampled histogram of integers
func (ms *MetricStorage) RegisterHistogram(name string) frontreport.MetricHistogram {
	return metrics.NewRegisteredHistogram(name, ms.registry, metrics.NewUniformSample(1000))
}

// RegisterCounter creates a counter
func (ms *MetricStorage) RegisterCounter(name string) frontreport.MetricCounter {
	return metrics.NewRegisteredCounter(name, ms.registry)
}
