package http

import "net/http"

func (h *Handler) addCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if len(h.DomainWhitelist) > 0 && !h.DomainWhitelist[origin] {
		h.Logger.Log("msg", "domain not in whitelist", "domain", origin)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}
