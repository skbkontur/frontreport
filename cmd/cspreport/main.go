package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jessevdk/go-flags"
	"github.com/skbkontur/cspreport/amqp"
	"github.com/skbkontur/cspreport/http"
)

var version = "undefined"

func main() {
	var opts struct {
		Port           string `short:"p" long:"port" default:"8888" description:"port to listen" env:"CSPREPORT_PORT"`
		AMQPConnection string `short:"a" long:"amqp" default:"amqp://guest:guest@localhost:5672/" description:"AMQP connection string" env:"CSPREPORT_AMQP"`
		Version        bool   `short:"v" long:"version" description:"print version and exit"`
	}
	flags.Parse(&opts)

	if opts.Version {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)

	storage := &amqp.ReportStorage{
		MaxBatchSize:         10,
		BatchTimeout:         time.Second,
		PendingWorkCapacity:  100,
		Exchange:             "csp",
		RoutingKey:           "csp",
		AMQPConnectionString: opts.AMQPConnection,
		Logger:               log.NewContext(logger).With("component", "amqp"),
	}

	logger.Log("msg", "starting AMQP storage")
	if err := storage.Start(); err != nil {
		logger.Log("msg", "error starting AMQP storage", "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "started AMQP storage")

	handler := &http.Handler{
		BatchReportStorage: storage,
		Port:               opts.Port,
		Logger:             log.NewContext(logger).With("component", "http"),
	}

	logger.Log("msg", "starting HTTP handler")
	if err := handler.Start(); err != nil {
		logger.Log("msg", "error starting HTTP handler", "error", err)
		os.Exit(1)
	}
	logger.Log("msg", "started HTTP handler")

	signalChannel := make(chan os.Signal)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	logger.Log("msg", "received signal", "signal", <-signalChannel)

	logger.Log("msg", "stopping HTTP handler")
	if err := handler.Stop(); err != nil {
		logger.Log("msg", "error stopping HTTP handler", "error", err)
	}
	logger.Log("msg", "stopped HTTP handler")

	logger.Log("msg", "stopping AMQP storage")
	if err := storage.Stop(); err != nil {
		logger.Log("msg", "error stopping AMQP storage", "error", err)
	}
	logger.Log("msg", "stopped AMQP storage")
}
