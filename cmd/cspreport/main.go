package main

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jessevdk/go-flags"

	"github.com/skbkontur/cspreport"
	"github.com/skbkontur/cspreport/amqp"
	"github.com/skbkontur/cspreport/http"
)

var logger log.Logger
var version = "undefined"

func main() {
	var opts struct {
		Port           string `short:"p" long:"port" default:"8888" description:"port to listen" env:"CSPREPORT_PORT"`
		AMQPConnection string `short:"a" long:"amqp" default:"amqp://guest:guest@localhost:5672/" description:"AMQP connection string" env:"CSPREPORT_AMQP"`
		Logfile        string `short:"l" long:"logfile" description:"log file name (writes to stdout if not specified)" env:"CSPREPORT_LOGFILE"`
		Version        bool   `short:"v" long:"version" description:"print version and exit"`
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

	storage := &amqp.ReportStorage{
		MaxBatchSize:         10,
		MaxConcurrentBatches: 10,
		BatchTimeout:         time.Second,
		PendingWorkCapacity:  100,
		ExchangeName:         "csp",
		RoutingKey:           "csp",
		AMQPConnectionString: opts.AMQPConnection,
		Logger:               log.NewContext(logger).With("component", "amqp"),
	}

	handler := &http.Handler{
		BatchReportStorage: storage,
		Port:               opts.Port,
		Logger:             log.NewContext(logger).With("component", "http"),
	}

	mustStart(storage)
	mustStart(handler)

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	logger.Log("msg", "received signal", "signal", <-signalChannel)

	mustStop(handler)
	mustStop(storage)
}

func mustStart(service cspreport.Service) {
	name := reflect.TypeOf(service)

	logger.Log("msg", "starting service", "name", name)
	if err := service.Start(); err != nil {
		logger.Log("msg", "error starting service", "name", name, "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "started service", "name", name)
}

func mustStop(service cspreport.Service) {
	name := reflect.TypeOf(service)

	logger.Log("msg", "stopping service", "name", name)
	if err := service.Stop(); err != nil {
		logger.Log("msg", "error stopping service", "name", name, "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "stopped service", "name", name)
}
