// Package httpapi exposes pipelines and their runs over HTTP.
//
// Handlers depend on small interfaces (Catalog, Triggerer, run.Reader,
// Readiness) rather than on the concrete service, so each can be faked
// in isolation and the service can grow without dragging the API
// along (Interface Segregation + Dependency Inversion).
package httpapi

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/service"
)

// Catalog lists pipelines.
type Catalog interface {
	Pipelines() []service.Info
	Pipeline(name string) (service.Info, bool)
}

// Triggerer starts runs asynchronously.
type Triggerer interface {
	Trigger(name string, trigger run.Trigger) (run.Run, error)
}

// Readiness reports whether the process can do useful work (its
// databases are reachable).
type Readiness interface {
	Ready(ctx context.Context) error
}

// Deps are the handler's collaborators.
type Deps struct {
	Catalog   Catalog
	Triggerer Triggerer
	Runs      run.Reader
	Readiness Readiness
	Logger    *slog.Logger
	Version   string
	// AuthToken, when non-empty, is required as a Bearer token on /v1.
	AuthToken string
}

// NewHandler returns the API's http.Handler.
func NewHandler(d Deps) http.Handler {
	if d.Logger == nil {
		d.Logger = slog.New(slog.DiscardHandler)
	}
	h := &handler{deps: d}

	v1 := http.NewServeMux()
	v1.HandleFunc("GET /v1/pipelines", h.listPipelines)
	v1.HandleFunc("GET /v1/pipelines/{name}", h.getPipeline)
	v1.HandleFunc("POST /v1/pipelines/{name}/runs", h.triggerRun)
	v1.HandleFunc("GET /v1/runs", h.listRuns)
	v1.HandleFunc("GET /v1/runs/{id}", h.getRun)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)
	mux.HandleFunc("GET /version", h.version)
	mux.Handle("/v1/", requireToken(d.AuthToken, v1))

	return recoverer(d.Logger, accessLog(d.Logger, mux))
}

type handler struct{ deps Deps }

func (h *handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.deps.Readiness != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := h.deps.Readiness.Ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": err.Error()})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (h *handler) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": h.deps.Version})
}

func (h *handler) listPipelines(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"pipelines": h.deps.Catalog.Pipelines()})
}

func (h *handler) getPipeline(w http.ResponseWriter, r *http.Request) {
	info, ok := h.deps.Catalog.Pipeline(r.PathValue("name"))
	if !ok {
		writeError(w, http.StatusNotFound, service.ErrNotFound)
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *handler) triggerRun(w http.ResponseWriter, r *http.Request) {
	started, err := h.deps.Triggerer.Trigger(r.PathValue("name"), run.TriggerAPI)
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, service.ErrAlreadyRunning):
		writeError(w, http.StatusConflict, err)
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)
	default:
		w.Header().Set("Location", "/v1/runs/"+started.ID)
		writeJSON(w, http.StatusAccepted, started)
	}
}

func (h *handler) listRuns(w http.ResponseWriter, r *http.Request) {
	f := run.Filter{Pipeline: r.URL.Query().Get("pipeline")}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit < 1 {
			writeError(w, http.StatusBadRequest, errors.New("limit must be a positive integer"))
			return
		}
		f.Limit = limit
	}

	runs, err := h.deps.Runs.List(r.Context(), f)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
}

func (h *handler) getRun(w http.ResponseWriter, r *http.Request) {
	got, err := h.deps.Runs.Get(r.Context(), r.PathValue("id"))
	switch {
	case errors.Is(err, run.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case err != nil:
		writeError(w, http.StatusInternalServerError, err)
	default:
		writeJSON(w, http.StatusOK, got)
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// requireToken enforces "Authorization: Bearer <token>" when token is
// set. The comparison is constant-time so response latency does not
// leak how much of a guessed token was right.
func requireToken(token string, next http.Handler) http.Handler {
	if token == "" {
		return next
	}
	want := []byte(token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(got), want) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="aurora-exporter"`)
			writeError(w, http.StatusUnauthorized, errors.New("missing or invalid bearer token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func accessLog(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		level := slog.LevelInfo
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			level = slog.LevelDebug // probes would otherwise flood the log
		}
		log.Log(r.Context(), level, "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"durationMs", time.Since(start).Milliseconds(),
		)
	})
}

func recoverer(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler { //nolint:errorlint // sentinel compared by identity, per net/http docs
					panic(v)
				}
				log.Error("http handler panic", "path", r.URL.Path, "panic", v)
				writeError(w, http.StatusInternalServerError, errors.New("internal error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
