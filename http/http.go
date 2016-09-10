package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/skbkontur/cspreport"
)

// Handler processes incoming reports
type Handler struct {
	BatchReportStorage cspreport.BatchReportStorage
	Port               string
	Logger             cspreport.Logger
}

// Start initializes HTTP request handling
func (h *Handler) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleCSPReport)
	mux.HandleFunc("/csp", h.handleCSPReport)
	mux.HandleFunc("/pkp", h.handlePKPReport)
	return http.ListenAndServe(fmt.Sprintf(":%s", h.Port), mux)
}

func (h *Handler) handleCSPReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	bodyParsed := &cspreport.CSPReport{}
	if err := dec.Decode(bodyParsed); err != nil {
		h.Logger.Log("msg", "malformed JSON body", "report_type", "CSP", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bodyParsed.Body.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.999Z")
	h.BatchReportStorage.AddCSPReport(*bodyParsed)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handlePKPReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	bodyParsed := &cspreport.PKPReport{}
	if err := dec.Decode(bodyParsed); err != nil {
		h.Logger.Log("msg", "malformed JSON body", "report_type", "PKP", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.BatchReportStorage.AddPKPReport(*bodyParsed)
	w.WriteHeader(http.StatusNoContent)
}
