// Package health serves the two probes Kubernetes needs: liveness says the
// process is alive, readiness says it may receive traffic.
//
// They ride on the admin listener, not on the gRPC port, because kubelet speaks
// HTTP and because a probe must still answer while gRPC is draining.
package health

import (
	"net/http"
	"sync/atomic"
)

// Readiness is flipped on once the gRPC server is serving and flipped off the
// moment SIGTERM arrives, so the pod leaves the Service endpoint list before
// the server starts refusing connections.
type Readiness struct {
	ready atomic.Bool
}

func NewReadiness() *Readiness {
	return &Readiness{}
}

func (r *Readiness) SetReady() {
	r.ready.Store(true)
}

func (r *Readiness) SetNotReady() {
	r.ready.Store(false)
}

// Register mounts /healthz and /readyz.
func Register(mux *http.ServeMux, readiness *Readiness) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writePlain(w, http.StatusOK, "ok")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !readiness.ready.Load() {
			writePlain(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		writePlain(w, http.StatusOK, "ready")
	})
}

func writePlain(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}
