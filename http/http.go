package http

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/skbkontur/cspreport"
	"github.com/tylerb/graceful"
	"gopkg.in/tomb.v2"
)

// Handler processes incoming reports
type Handler struct {
	BatchReportStorage cspreport.BatchReportStorage
	Port               string
	Logger             cspreport.Logger
	tomb               tomb.Tomb
}

// Start initializes HTTP request handling
func (h *Handler) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleCSPReport)
	mux.HandleFunc("/csp", h.handleCSPReport)
	mux.HandleFunc("/pkp", h.handlePKPReport)

	server := &graceful.Server{
		Timeout: 10 * time.Second,
		Server: &http.Server{
			Addr:    fmt.Sprintf(":%s", h.Port),
			Handler: mux,
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
