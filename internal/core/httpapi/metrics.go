package httpapi

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// handleMetrics writes process counters in Prometheus text format.
//
// When a metrics token is configured it must be presented as a bearer token.
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if token := s.cfg.Metrics.Token; token != "" {
		presented, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(presented), []byte(token)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			WriteError(w, r, ErrUnauthorized)
			return
		}
	}
	s.metrics.WritePrometheus(w)
}
