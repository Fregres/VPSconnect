package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Fregres/VPSconnect/internal/metrics"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	statusHandler := http.HandlerFunc(s.handleStatus)
	protectedStatusHandler := s.auth(statusHandler)

	mux.Handle("GET /api/v1/status", protectedStatusHandler)
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("json encode error", err)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := metrics.CollectStatus()

	if err != nil {
		log.Printf("collect status: %v", err)
		http.Error(
			w,
			"failed to collect system status",
			http.StatusInternalServerError,
		)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")

	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("encode status response %v", err)
	}
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, "OK")
}
