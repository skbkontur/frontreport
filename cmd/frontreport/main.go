package main

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jessevdk/go-flags"

	"github.com/skbkontur/frontreport"
	"github.com/skbkontur/frontreport/amqp"
	"github.com/skbkontur/frontreport/http"
	"github.com/skbkontur/frontreport/metrics"
)

var logger log.Logger
var version = "undefined"

func main() {
	var opts struct {
		Port               string `short:"p" long:"port" default:"8888" description:"port to listen" env:"FRONTREPORT_PORT"`
		AMQPConnection     string `short:"a" long:"amqp" default:"amqp://guest:guest@localhost:5672/" description:"AMQP connection string" env:"FRONTREPORT_AMQP"`
		ServiceWhitelist   string `short:"s" long:"service-whitelist" description:"allow reports only from this comma-separated list of services (allows all if not specified)" env:"FRONTREPORT_SERVICE_WHITELIST"`
		DomainWhitelist    string `short:"d" long:"domain-whitelist" description:"allow CORS requests only from this comma-separated list of domains (allows all if not specified)" env:"FRONTREPORT_DOMAIN_WHITELIST"`
		Logfile            string `short:"l" long:"logfile" description:"log file name (writes to stdout if not specified)" env:"FRONTREPORT_LOGFILE"`
		GraphiteConnection string `short:"g" long:"graphite" description:"Graphite connection string for internal metrics" env:"FRONTREPORT_GRAPHITE"`
		GraphitePrefix     string `short:"r" long:"graphite-prefix" description:"prefix for Graphite metrics" env:"FRONTREPORT_GRAPHITE_PREFIX"`
		Version            bool   `short:"v" long:"version" description:"print version and exit"`
	}
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	if opts.Version {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	if opts.Logfile == "" {
		logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stdout))
	} else {
		logfile, err := os.OpenFile(opts.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open logfile %s: %s", opts.Logfile, err)
			os.Exit(1)
		}
		defer logfile.Close()
		logger = log.NewLogfmtLogger(log.NewSyncWriter(logfile))
	}
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)

	logger.Log("msg", "starting program", "pid", os.Getpid())

	metrics := &metrics.MetricStorage{
		GraphiteConnectionString: opts.GraphiteConnection,
		GraphitePrefix:           opts.GraphitePrefix,
		Logger:                   log.NewContext(logger).With("component", "metrics"),
	}

	storage := &amqp.ReportStorage{
		MaxBatchSize:         10,
		MaxConcurrentBatches: 10,
		BatchTimeout:         time.Second,
		PendingWorkCapacity:  100,
		ExchangeName:         "csp",
		RoutingKey:           "csp",
		AMQPConnectionString: opts.AMQPConnection,
		Logger:               log.NewContext(logger).With("component", "amqp"),
		MetricStorage:        metrics,
	}

	handler := &http.Handler{
		ReportStorage: storage,
		Port:          opts.Port,
		Logger:        log.NewContext(logger).With("component", "http"),
		MetricStorage: metrics,
	}
	if opts.ServiceWhitelist != "" {
		serviceWhitelist := strings.Split(opts.ServiceWhitelist, ",")
		handler.ServiceWhitelist = make(map[string]bool, len(serviceWhitelist))
		for _, service := range serviceWhitelist {
			handler.ServiceWhitelist[strings.TrimSpace(service)] = true
		}
	}
	if opts.DomainWhitelist != "" {
		domainWhitelist := strings.Split(opts.DomainWhitelist, ",")
		handler.DomainWhitelist = make(map[string]bool, len(domainWhitelist))
		for _, domain := range domainWhitelist {
			handler.DomainWhitelist[strings.TrimSpace(domain)] = true
		}
	}

	mustStart(metrics)
	mustStart(storage)
	mustStart(handler)

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	logger.Log("msg", "received signal", "signal", <-signalChannel)

	mustStop(handler)
	mustStop(storage)
	mustStop(metrics)
}

func mustStart(service frontreport.Service) {
	name := reflect.TypeOf(service)

	logger.Log("msg", "starting service", "name", name)
	if err := service.Start(); err != nil {
		logger.Log("msg", "error starting service", "name", name, "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "started service", "name", name)
}

func mustStop(service frontreport.Service) {
	name := reflect.TypeOf(service)

	logger.Log("msg", "stopping service", "name", name)
	if err := service.Stop(); err != nil {
		logger.Log("msg", "error stopping service", "name", name, "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "stopped service", "name", name)
}
