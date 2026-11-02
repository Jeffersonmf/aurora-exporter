package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/goleak"

	"github.com/Jeffersonmf/aurora-exporter/internal/httpapi"
	"github.com/Jeffersonmf/aurora-exporter/internal/run"
	"github.com/Jeffersonmf/aurora-exporter/internal/service"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

type catalog struct{}

func (catalog) Pipelines() []service.Info {
	return []service.Info{{Name: "brands", Enabled: true}}
}

func (catalog) Pipeline(name string) (service.Info, bool) {
	if name == "brands" {
		return service.Info{Name: "brands"}, true
	}
	return service.Info{}, false
}

type triggerer struct{ err error }

func (t triggerer) Trigger(name string, trig run.Trigger) (run.Run, error) {
	if t.err != nil {
		return run.Run{}, t.err
	}
	return run.Run{ID: "r1", Pipeline: name, Trigger: trig, Status: run.StatusRunning}, nil
}

type readiness struct{ err error }

func (r readiness) Ready(context.Context) error { return r.err }

func newServer(t *testing.T, mod func(*httpapi.Deps)) *httptest.Server {
	t.Helper()
	store := run.NewMemoryStore(0)
	_ = store.Save(context.Background(), run.Run{ID: "r0", Pipeline: "brands", Status: run.StatusSucceeded})
	d := httpapi.Deps{Catalog: catalog{}, Triggerer: triggerer{}, Runs: store, Readiness: readiness{}, Version: "1.2.3"}
	if mod != nil {
		mod(&d)
	}
	srv := httptest.NewServer(httpapi.NewHandler(d))
	t.Cleanup(srv.Close)
	return srv
}

func do(t *testing.T, srv *httptest.Server, method, path, token string) (int, map[string]any, http.Header) {
	t.Helper()
	req, _ := http.NewRequestWithContext(context.Background(), method, srv.URL+path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body, resp.Header
}

func TestAPI_Endpoints(t *testing.T) {
	t.Parallel()
	srv := newServer(t, nil)

	cases := []struct {
		method, path string
		status       int
		contains     string
	}{
		{"GET", "/healthz", 200, "ok"},
		{"GET", "/readyz", 200, "ready"},
		{"GET", "/version", 200, "1.2.3"},
		{"GET", "/v1/pipelines", 200, "brands"},
		{"GET", "/v1/pipelines/brands", 200, "brands"},
		{"GET", "/v1/pipelines/nope", 404, "not found"},
		{"POST", "/v1/pipelines/brands/runs", 202, "running"},
		{"GET", "/v1/runs?pipeline=brands&limit=5", 200, "r0"},
		{"GET", "/v1/runs?limit=zero", 400, "limit"},
		{"GET", "/v1/runs/r0", 200, "succeeded"},
		{"GET", "/v1/runs/missing", 404, "not found"},
		{"DELETE", "/v1/pipelines/brands", 405, ""},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			t.Parallel()
			status, body, _ := do(t, srv, tc.method, tc.path, "")
			if status != tc.status {
				t.Fatalf("status = %d, want %d (body %v)", status, tc.status, body)
			}
			if tc.contains != "" && !strings.Contains(fmt.Sprint(body), tc.contains) {
				t.Fatalf("body %v does not contain %q", body, tc.contains)
			}
		})
	}
}

func TestAPI_TriggerSetsLocation(t *testing.T) {
	t.Parallel()
	_, _, header := do(t, newServer(t, nil), "POST", "/v1/pipelines/brands/runs", "")
	if got := header.Get("Location"); got != "/v1/runs/r1" {
		t.Fatalf("Location = %q", got)
	}
}

func TestAPI_TriggerErrorsMapToStatus(t *testing.T) {
	t.Parallel()
	cases := map[error]int{
		service.ErrAlreadyRunning: http.StatusConflict,
		service.ErrNotFound:       http.StatusNotFound,
		errors.New("boom"):        http.StatusInternalServerError,
	}
	for err, want := range cases {
		srv := newServer(t, func(d *httpapi.Deps) { d.Triggerer = triggerer{err: fmt.Errorf("wrapped: %w", err)} })
		if status, _, _ := do(t, srv, "POST", "/v1/pipelines/brands/runs", ""); status != want {
			t.Errorf("%v: status = %d, want %d", err, status, want)
		}
	}
}

func TestAPI_ReadinessFailure(t *testing.T) {
	t.Parallel()
	srv := newServer(t, func(d *httpapi.Deps) { d.Readiness = readiness{err: errors.New("redshift unreachable")} })
	status, body, _ := do(t, srv, "GET", "/readyz", "")
	if status != http.StatusServiceUnavailable || !strings.Contains(fmt.Sprint(body), "redshift unreachable") {
		t.Fatalf("status = %d body = %v", status, body)
	}
}

func TestAPI_BearerTokenProtectsV1Only(t *testing.T) {
	t.Parallel()
	srv := newServer(t, func(d *httpapi.Deps) { d.AuthToken = "s3cret" })

	if status, _, h := do(t, srv, "GET", "/v1/pipelines", ""); status != http.StatusUnauthorized || h.Get("WWW-Authenticate") == "" {
		t.Fatalf("no token: status = %d", status)
	}
	if status, _, _ := do(t, srv, "GET", "/v1/pipelines", "wrong"); status != http.StatusUnauthorized {
		t.Fatalf("wrong token: status = %d", status)
	}
	if status, _, _ := do(t, srv, "GET", "/v1/pipelines", "s3cret"); status != http.StatusOK {
		t.Fatalf("right token: status = %d", status)
	}
	if status, _, _ := do(t, srv, "GET", "/healthz", ""); status != http.StatusOK {
		t.Fatalf("probes must stay open: status = %d", status)
	}
}

type panicCatalog struct{ catalog }

func (panicCatalog) Pipelines() []service.Info { panic("boom") }

func TestAPI_RecoversFromPanics(t *testing.T) {
	t.Parallel()
	srv := newServer(t, func(d *httpapi.Deps) { d.Catalog = panicCatalog{} })
	if status, _, _ := do(t, srv, "GET", "/v1/pipelines", ""); status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
}
