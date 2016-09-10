package main

import (
	"fmt"
	"io/ioutil"
	"os"
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
		CACertPath     string `short:"c" long:"cacert" default:"" description:"custom CA cert file for SSL connections" env:"CSPREPORT_CACERT"`
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

	if opts.CACertPath != "" {
		ca, err := ioutil.ReadFile(opts.CACertPath)
		if err != nil {
			logger.Log("msg", "error opening cacert", "error", err)
			os.Exit(1)
		}

		storage.CACert = ca
	}

	if err := storage.Start(); err != nil {
		logger.Log("msg", "error starting AMQP storage", "error", err)
		os.Exit(1)
	}
	defer storage.Stop()

	handler := &http.Handler{
		BatchReportStorage: storage,
		Port:               opts.Port,
		Logger:             log.NewContext(logger).With("component", "http"),
	}

	if err := handler.Start(); err != nil {
		logger.Log("msg", "error starting HTTP handler", "error", err)
		os.Exit(1)
	}
}
