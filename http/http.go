package http

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/tylerb/graceful"
	"gopkg.in/tomb.v2"

	"github.com/skbkontur/frontreport"
)

// Handler processes incoming reports
type Handler struct {
	ReportStorage    frontreport.ReportStorage
	Port             string
	ServiceWhitelist map[string]bool
	DomainWhitelist  map[string]bool
	Logger           frontreport.Logger
	MetricStorage    frontreport.MetricStorage
	tomb             tomb.Tomb
	metrics          struct {
		total  map[string]frontreport.MetricCounter
		errors map[string]frontreport.MetricCounter
	}
}

// Start initializes HTTP request handling
func (h *Handler) Start() error {
	h.metrics.total = make(map[string]frontreport.MetricCounter)
	h.metrics.errors = make(map[string]frontreport.MetricCounter)
	for _, reportType := range []string{"csp", "pkp", "stacktracejs"} {
		h.metrics.total[reportType] = h.MetricStorage.RegisterCounter(fmt.Sprintf("http.report_decoding.%s.total", reportType))
		h.metrics.errors[reportType] = h.MetricStorage.RegisterCounter(fmt.Sprintf("http.report_decoding.%s.errors", reportType))
	}

	server := &graceful.Server{
		Timeout:          10 * time.Second,
		NoSignalHandling: true,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", h.Port),
			Handler: http.HandlerFunc(h.handleRequest),
		},
	}

	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return err
	}

	h.tomb.Go(func() error {
		err := server.Serve(listener)
		select {
		case <-h.tomb.Dying():
			return nil
		default:
			return err
		}
	})

	h.tomb.Go(func() error {
		<-h.tomb.Dying()
		return listener.Close()
	})

	return nil
}

// Stop finishes listening to HTTP
func (h *Handler) Stop() error {
	h.tomb.Kill(nil)
	return h.tomb.Wait()
}

func (h *Handler) handleRequest(w http.ResponseWriter, r *http.Request) {
	h.addCORSHeaders(w, r)

	switch r.Method {
	case http.MethodGet:
		h.handleAsset(w, r)

	case http.MethodPost:
		h.handleReport(w, r)

	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
