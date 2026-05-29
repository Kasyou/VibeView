package server

import (
	"context"
	"fmt"
	"net/http"
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
	cfg     Config
	mux     *http.ServeMux
	wsUp    websocket.Upgrader
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
	console []ConsoleMsg
	httpSrv *http.Server
}

func New(cfg Config) *Server {
	s := &Server{
		cfg:     cfg,
		mux:     http.NewServeMux(),
		clients: make(map[*websocket.Conn]bool),
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
	s.mux.Handle("/_app/", http.StripPrefix("/_app", http.FileServer(http.Dir(s.cfg.ProjectDir))))
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
				s.mu.Lock()
				s.console = append(s.console, ConsoleMsg{
					Level:   str(data["level"]),
					Message: str(data["message"]),
					File:    str(data["file"]),
					Line:    intval(data["line"]),
				})
				s.mu.Unlock()
			}
		}
	}()
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
