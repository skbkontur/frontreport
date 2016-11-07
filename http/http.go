package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/tylerb/graceful"
	"gopkg.in/tomb.v2"

	"github.com/skbkontur/frontreport"
)

// Handler processes incoming reports
type Handler struct {
	BatchReportStorage frontreport.BatchReportStorage
	Port               string
	Logger             frontreport.Logger
	tomb               tomb.Tomb
}

// Start initializes HTTP request handling
func (h *Handler) Start() error {
	server := &graceful.Server{
		Timeout:          10 * time.Second,
		NoSignalHandling: true,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", h.Port),
			Handler: http.HandlerFunc(h.handleReport),
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

func (h *Handler) handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	switch r.URL.Path {
	case "/csp":
		report := &frontreport.CSPReport{}
		if err := h.processReport(r.Body, report); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "/hpkp":
		report := &frontreport.PKPReport{}
		if err := h.processReport(r.Body, report); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case "/stacktracejs":
		report := &frontreport.StacktraceJSReport{}
		if err := h.processReport(r.Body, report); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) processReport(body io.Reader, report frontreport.Reportable) error {
	dec := json.NewDecoder(body)
	if err := dec.Decode(report); err != nil {
		h.Logger.Log("msg", "cannot process JSON body", "report_type", report.GetType(), "error", err)
		return err
	}
	report.SetTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.999Z"))
	h.BatchReportStorage.AddReport(report)
	return nil
}
