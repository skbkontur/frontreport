package http

import (
	"bytes"
	"net/http"
)

func (h *Handler) handleAsset(w http.ResponseWriter, r *http.Request) {
	data, err := Asset(r.URL.Path[1:])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		h.Logger.Log("msg", "cannot serve asset", "path", r.URL.Path, "error", err)
		return
	}

	info, _ := AssetInfo(r.URL.Path[1:])
	http.ServeContent(w, r, info.Name(), info.ModTime(), bytes.NewReader(data))
}
