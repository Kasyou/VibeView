package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Config struct {
	Port         int
	DevServerURL string
	ProjectDir   string
	RendererHTML []byte
	RendererFS   http.Handler
}

type ConsoleMsg struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	File    string `json:"file"`
	Line    int    `json:"line"`
}

type Server struct {
	cfg        Config
	mux        *http.ServeMux
	wsUp       websocket.Upgrader
	clients    map[*websocket.Conn]bool
	mu         sync.Mutex
	console    []ConsoleMsg
	httpSrv    *http.Server
	screenshot     string                 // latest screenshot base64
	prevScreenshot string                 // previous screenshot for diff comparison
	scrReqs        map[string]chan string  // pending screenshot requests
	inspReqs       map[string]chan string  // pending inspect requests
	scrMu          sync.Mutex
}

func New(cfg Config) *Server {
	s := &Server{
		cfg:     cfg,
		mux:     http.NewServeMux(),
		clients: make(map[*websocket.Conn]bool),
		scrReqs:  make(map[string]chan string),
		inspReqs: make(map[string]chan string),
		wsUp: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/health", s.health)
	s.mux.HandleFunc("/ws", s.ws)
	s.mux.HandleFunc("/api/reload", s.reload)
	s.mux.HandleFunc("/api/config", s.config)
	s.mux.HandleFunc("/api/console", s.getConsole)
	s.mux.Handle("/_vibeview/", http.StripPrefix("/_vibeview", s.cfg.RendererFS))
	s.mux.Handle("/_app/", http.StripPrefix("/_app", &injectHandler{
		handler: http.FileServer(http.Dir(s.cfg.ProjectDir)),
	}))
	s.mux.HandleFunc("/api/screenshot", s.handleScreenshot)
	s.mux.HandleFunc("/api/inspect", s.handleInspect)
	s.mux.HandleFunc("/api/diff", s.handleDiff)
	s.mux.HandleFunc("/", s.renderer)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) config(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"devServerURL":"%s"}`, s.cfg.DevServerURL)
}

func (s *Server) renderer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(s.cfg.RendererHTML)
}

func (s *Server) reload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	s.Broadcast("reload", nil)
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) getConsole(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	s.mu.Lock()
	msgs := s.console
	s.console = nil
	s.mu.Unlock()

	if msgs == nil {
		w.Write([]byte(`[]`))
		return
	}
	w.Write([]byte(`[`))
	for i, m := range msgs {
		if i > 0 {
			w.Write([]byte(`,`))
		}
		fmt.Fprintf(w, `{"level":"%s","message":"%s","file":"%s","line":%d}`,
			escapeJSON(m.Level), escapeJSON(m.Message), escapeJSON(m.File), m.Line)
	}
	w.Write([]byte(`]`))
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	conn, err := s.wsUp.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.clients, conn)
			s.mu.Unlock()
			conn.Close()
		}()
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				return
			}
			if msg["type"] == "console" {
				data, ok := msg["data"].(map[string]interface{})
				if !ok {
					continue
				}
				level := str(data["level"])
				message := str(data["message"])
				file := str(data["file"])
				line := intval(data["line"])

				s.mu.Lock()
				s.console = append(s.console, ConsoleMsg{
					Level: level, Message: message, File: file, Line: line,
				})
				s.mu.Unlock()

				// Echo to terminal
				if file != "" {
					fmt.Fprintf(os.Stderr, "  [%s] %s (%s:%d)\n", level, message, file, line)
				} else {
					fmt.Fprintf(os.Stderr, "  [%s] %s\n", level, message)
				}
			}
			if msg["type"] == "screenshot-data" {
				data, ok := msg["data"].(map[string]interface{})
				if !ok {
					continue
				}
				reqID := str(msg["id"])
				image := str(data["image"])
				s.scrMu.Lock()
				if ch, ok := s.scrReqs[reqID]; ok {
					ch <- image
					delete(s.scrReqs, reqID)
				}
				s.scrMu.Unlock()
				// Store latest screenshot, keep previous for diff
				if image != "" {
					s.mu.Lock()
					s.prevScreenshot = s.screenshot
					s.screenshot = image
					s.mu.Unlock()
				}
			}
			if msg["type"] == "inspect-data" {
				data, ok := msg["data"].(map[string]interface{})
				if !ok {
					continue
				}
				reqID := str(msg["id"])
				// Marshal data back to JSON string for the channel
				jsonBytes, _ := json.Marshal(data)
				s.scrMu.Lock()
				if ch, ok := s.inspReqs[reqID]; ok {
					ch <- string(jsonBytes)
					delete(s.inspReqs, reqID)
				}
				s.scrMu.Unlock()
			}
		}
	}()
}

func (s *Server) handleScreenshot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Return latest cached screenshot
		s.mu.Lock()
		img := s.screenshot
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if img == "" {
			w.Write([]byte(`{"image":"","error":"no screenshot available"}`))
			return
		}
		fmt.Fprintf(w, `{"image":"%s"}`, img)

	case http.MethodPost:
		// Request a new screenshot from the browser
		reqID := fmt.Sprintf("%d", time.Now().UnixNano())
		ch := make(chan string, 1)
		s.scrMu.Lock()
		s.scrReqs[reqID] = ch
		s.scrMu.Unlock()

		s.Broadcast("screenshot-request", map[string]string{"id": reqID})

		// Wait for response (timeout 5s)
		select {
		case img := <-ch:
			w.Header().Set("Content-Type", "application/json")
			if img == "" {
				w.Write([]byte(`{"image":"","error":"screenshot capture failed"}`))
			} else {
				fmt.Fprintf(w, `{"image":"%s"}`, img)
			}
		case <-time.After(5 * time.Second):
			s.scrMu.Lock()
			delete(s.scrReqs, reqID)
			s.scrMu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"image":"","error":"screenshot timeout (no browser connected?)"}`))
		}

	default:
		http.Error(w, "method not allowed", 405)
	}
}

func (s *Server) handleInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}

	selector := r.URL.Query().Get("selector")
	if selector == "" {
		selector = "body"
	}

	reqID := fmt.Sprintf("insp-%d", time.Now().UnixNano())
	ch := make(chan string, 1)
	s.scrMu.Lock()
	s.inspReqs[reqID] = ch
	s.scrMu.Unlock()

	s.Broadcast("inspect-request", map[string]string{
		"id":       reqID,
		"selector": selector,
	})

	select {
	case result := <-ch:
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
	case <-time.After(5 * time.Second):
		s.scrMu.Lock()
		delete(s.inspReqs, reqID)
		s.scrMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"found":false,"error":"timeout"}`))
	}
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Request a fresh screenshot first
	reqID := fmt.Sprintf("diff-%d", time.Now().UnixNano())
	ch := make(chan string, 1)
	s.scrMu.Lock()
	s.scrReqs[reqID] = ch
	s.scrMu.Unlock()

	s.Broadcast("screenshot-request", map[string]string{"id": reqID})

	var currentImg string
	select {
	case img := <-ch:
		currentImg = img
	case <-time.After(5 * time.Second):
		s.scrMu.Lock()
		delete(s.scrReqs, reqID)
		s.scrMu.Unlock()
		w.Write([]byte(`{"changed":false,"error":"screenshot timeout"}`))
		return
	}

	s.mu.Lock()
	prevImg := s.prevScreenshot
	s.mu.Unlock()

	if prevImg == "" {
		w.Write([]byte(`{"changed":false,"message":"no previous screenshot to compare"}`))
		return
	}

	// Compare: check overall length and a sample of bytes
	changed := len(currentImg) != len(prevImg)
	if !changed && len(currentImg) > 200 {
		// Compare middle portion
		mid := len(currentImg) / 2
		changed = currentImg[mid:mid+100] != prevImg[mid:mid+100]
	}

	if changed {
		fmt.Fprintf(w, `{"changed":true,"message":"visual changes detected","before":"%s","after":"%s"}`,
			prevImg, currentImg)
	} else {
		w.Write([]byte(`{"changed":false,"message":"no visual changes detected"}`))
	}
}

func (s *Server) Broadcast(msgType string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for conn := range s.clients {
		conn.WriteJSON(map[string]interface{}{
			"type": msgType,
			"data": data,
		})
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.cfg.Port)
	s.httpSrv = &http.Server{Addr: addr, Handler: s.mux}
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Close() error {
	if s.httpSrv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		return s.httpSrv.Shutdown(ctx)
	}
	return nil
}

func escapeJSON(s string) string {
	// minimal JSON string escaping for console message fields
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b = append(b, '\\', '"')
		case '\\':
			b = append(b, '\\', '\\')
		case '\n':
			b = append(b, '\\', 'n')
		case '\r':
			b = append(b, '\\', 'r')
		case '\t':
			b = append(b, '\\', 't')
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

func str(v interface{}) string {
	s, _ := v.(string)
	return s
}

func intval(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

// injectHandler wraps a file server and injects an error-catching script
// into HTML responses. This allows VibeView to capture JS errors from
// the project being previewed and forward them to the terminal.
type injectHandler struct {
	handler http.Handler
}

var injectScript = []byte(`<script>
(function(){
  var _vibeviewErrors=[];
  window.onerror=function(m,s,l,c,e){
    var err={type:'vibeview-error',message:m,file:s,line:l};
    _vibeviewErrors.push(err);
    try{window.parent.postMessage(err,'*');}catch(_){}
  };
  window.addEventListener('unhandledrejection',function(e){
    var msg=e.reason?e.reason.message||String(e.reason):'Unhandled Promise';
    var err={type:'vibeview-error',message:msg,file:'',line:0};
    try{window.parent.postMessage(err,'*');}catch(_){}
  });
  document.addEventListener('DOMContentLoaded',function(){
    _vibeviewErrors.forEach(function(err){
      try{window.parent.postMessage(err,'*');}catch(_){}
    });
  });
})();
</script>
</body>`)

func (h *injectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only inject into HTML files that aren't directory redirects
	path := r.URL.Path
	if !isHTML(path) || path == "/index.html" {
		h.handler.ServeHTTP(w, r)
		return
	}

	// Capture response in buffer
	buf := &bytes.Buffer{}
	rec := &responseRecorder{buf: buf}
	h.handler.ServeHTTP(rec, r)

	// Don't inject for non-200 or non-HTML responses
	ct := rec.Header().Get("Content-Type")
	if rec.status != 200 || (ct != "" && !hasPrefix(ct, "text/html")) {
		rec.writeTo(w)
		return
	}

	body := buf.Bytes()
	if len(body) == 0 {
		rec.writeTo(w)
		return
	}

	// Inject the error-catching script before </body>
	body = bytes.Replace(body, []byte("</body>"), injectScript, 1)

	// Copy headers from recorded response
	for k, v := range rec.Header() {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteHeader(rec.status)
	w.Write(body)
}

type responseRecorder struct {
	buf    *bytes.Buffer
	header http.Header
	status int
}

func (r *responseRecorder) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = 200
	}
	return r.buf.Write(b)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
}

func (r *responseRecorder) writeTo(w http.ResponseWriter) {
	for k, v := range r.header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	if r.status == 0 {
		r.status = 200
	}
	w.WriteHeader(r.status)
	if r.buf.Len() > 0 {
		io.Copy(w, r.buf)
	}
}

func isHTML(path string) bool {
	return path == "/" ||
		path == "/index.html" ||
		(len(path) > 5 && path[len(path)-5:] == ".html") ||
		(len(path) > 4 && path[len(path)-4:] == ".htm")
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
