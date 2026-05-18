package handlers

import (
	"encoding/json"
	"net/http"
)

// Healthz is a liveness probe. It returns 200 OK with a tiny JSON body and
// makes no downstream calls. Future readiness checks (DB ping, gRPC dial)
// belong on a separate /readyz endpoint, not here.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
