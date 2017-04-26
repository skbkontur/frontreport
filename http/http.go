package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/tylerb/graceful"
	"gopkg.in/tomb.v2"

	"github.com/skbkontur/frontreport"
)

// Handler processes incoming reports
type Handler struct {
	ReportStorage frontreport.ReportStorage
	Port          string
	Logger        frontreport.Logger
	MetricStorage frontreport.MetricStorage
	tomb          tomb.Tomb
	metrics       struct {
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
	switch r.Method {
	case http.MethodGet:
		h.handleAsset(w, r)

	case http.MethodPost:
		h.handleReport(w, r)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleAsset(w http.ResponseWriter, r *http.Request) {
	data, err := Asset(r.URL.Path[1:])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		h.Logger.Log("msg", "cannot serve asset", "path", r.URL.Path, "error", err)
		return
	}
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8") // only JS assets served for now
	w.Write(data)
}

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "csp"):
		report := &frontreport.CSPReport{}
		if err := h.processReport(r.Body, report, r.Host); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case strings.Contains(r.URL.Path, "pkp"):
		report := &frontreport.PKPReport{}
		if err := h.processReport(r.Body, report, r.Host); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case strings.Contains(r.URL.Path, "stacktracejs"):
		report := &frontreport.StacktraceJSReport{}
		if err := h.processReport(r.Body, report, r.Host); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) processReport(body io.Reader, report frontreport.Reportable, host string) error {
	h.metrics.total[report.GetType()].Inc(1)

	dec := json.NewDecoder(body)
	if err := dec.Decode(report); err != nil {
		h.Logger.Log("msg", "cannot process JSON body", "report_type", report.GetType(), "error", err)
		h.metrics.errors[report.GetType()].Inc(1)
		return err
	}
	report.SetTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.999Z"))
	report.SetHost(host)
	h.ReportStorage.AddReport(report)
	return nil
}
