package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func stubHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("static"))
	})
}

func testConfig() Config {
	return Config{
		Port:         0,
		DevServerURL: "http://localhost:5173",
		ProjectDir:   ".",
		Mode:         "claude",
		RendererHTML: []byte("x"),
		RendererFS:   stubHandler(),
	}
}

func TestHealth(t *testing.T) {
	s := New(testConfig())
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRenderer(t *testing.T) {
	cfg := testConfig()
	cfg.RendererHTML = []byte("<html>hi</html>")
	s := New(cfg)
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hi") {
		t.Error("expected renderer HTML")
	}
}

func TestConfig(t *testing.T) {
	cfg := testConfig()
	cfg.DevServerURL = "http://localhost:3000"
	s := New(cfg)
	req := httptest.NewRequest("GET", "/api/config", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "localhost:3000") {
		t.Error("expected devServerURL in config response")
	}
}

func TestReloadMethodCheck(t *testing.T) {
	cfg := testConfig()
	cfg.DevServerURL = ""
	s := New(cfg)
	req := httptest.NewRequest("GET", "/api/reload", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 405 {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestReloadPost(t *testing.T) {
	cfg := testConfig()
	cfg.DevServerURL = ""
	s := New(cfg)
	req := httptest.NewRequest("POST", "/api/reload", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200 for POST, got %d", w.Code)
	}
}

func TestConsoleEmpty(t *testing.T) {
	cfg := testConfig()
	cfg.DevServerURL = ""
	s := New(cfg)
	req := httptest.NewRequest("GET", "/api/console", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "[]") {
		t.Error("expected empty array")
	}
}

func TestStaticDelegation(t *testing.T) {
	cfg := testConfig()
	cfg.DevServerURL = ""
	s := New(cfg)
	req := httptest.NewRequest("GET", "/_vibeview/style.css", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	if w.Body.String() != "static" {
		t.Errorf("expected 'static', got '%s'", w.Body.String())
	}
}

func TestInjectHandlerSkipsNonHTML(t *testing.T) {
	h := &injectHandler{
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/css")
			w.Write([]byte("body{color:red}"))
		}),
	}
	req := httptest.NewRequest("GET", "/style.css", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "body{color:red}") {
		t.Error("non-HTML should pass through unchanged")
	}
	if strings.Contains(w.Body.String(), "vibeview-error") {
		t.Error("CSS should not be injected with error script")
	}
}

func TestInjectHandlerInjectsHTML(t *testing.T) {
	h := &injectHandler{
		handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(200)
			w.Write([]byte("<html><head></head><body><h1>Hi</h1></body></html>"))
		}),
	}
	req := httptest.NewRequest("GET", "/page.html", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "vibeview-error") {
		t.Error("HTML response should be injected with error script")
	}
	if !strings.Contains(w.Body.String(), "<h1>Hi</h1>") {
		t.Error("original content should be preserved")
	}
}

func TestDiffNoPrevious(t *testing.T) {
	cfg := testConfig()
	s := New(cfg)
	req := httptest.NewRequest("GET", "/api/diff", nil)
	w := httptest.NewRecorder()
	s.mux.ServeHTTP(w, req)
	// Without a browser connected, diff returns unchanged/no-prev message
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
