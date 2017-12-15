package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/skbkontur/frontreport"
)

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
	if len(h.ServiceWhitelist) > 0 && !h.ServiceWhitelist[report.GetService()] {
		h.Logger.Log("msg", "service not in whitelist", "service", report.GetService(), "report_type", report.GetType())
		h.metrics.errors[report.GetType()].Inc(1)
		return errors.New("service not in whitelist")
	}
	report.SetTimestamp(time.Now().UTC().Format("2006-01-02T15:04:05.999Z"))
	report.SetHost(host)

	switch report.(type) {
	case *frontreport.StacktraceJSReport:
		report.(*frontreport.StacktraceJSReport).Stack = h.SourcemapProcessor.ProcessStack(report.(*frontreport.StacktraceJSReport).Stack)
	}

	h.ReportStorage.AddReport(report)
	return nil
}
