package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	JSONRPCVersion = "2.0"
	ServerName     = "vibeview"
	ServerVersion  = "0.1.0"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type Server struct {
	reader    *bufio.Reader
	writer    io.Writer
	serverURL string
	client    *http.Client
}

func New(serverURL string) *Server {
	return &Server{
		reader:    bufio.NewReader(os.Stdin),
		writer:    os.Stdout,
		serverURL: serverURL,
		client:    &http.Client{},
	}
}

func (s *Server) Run() error {
	scanner := bufio.NewScanner(s.reader)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		resp := s.handle(req)
		data, _ := json.Marshal(resp)
		fmt.Fprintln(s.writer, string(data))
	}
	return scanner.Err()
}

func (s *Server) handle(req Request) Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return Response{
			JSONRPC: JSONRPCVersion,
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: "method not found"},
		}
	}
}

func (s *Server) handleInitialize(req Request) Response {
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    ServerName,
				"version": ServerVersion,
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		},
	}
}

func (s *Server) handleToolsList(req Request) Response {
	tools := []Tool{
		{
			Name:        "preview_reload",
			Description: "Reload the preview iframe to reflect latest code changes",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_console",
			Description: "Read recent browser console messages from the preview iframe",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_screenshot",
			Description: "View the current preview. Open http://localhost:51820 in a browser to see the visual preview. Use preview_console and preview_reload for state changes.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
	}
	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  map[string]interface{}{"tools": tools},
	}
}

func (s *Server) handleToolsCall(req Request) Response {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	json.Unmarshal(req.Params, &params)

	var result interface{}
	var err *RPCError

	switch params.Name {
	case "preview_reload":
		resp, httpErr := s.client.Post(s.serverURL+"/api/reload", "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			resp.Body.Close()
			result = map[string]string{"status": "reloaded"}
		}
	case "preview_console":
		resp, httpErr := s.client.Get(s.serverURL + "/api/console")
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			var msgs []interface{}
			json.Unmarshal(data, &msgs)
			result = map[string]interface{}{"messages": msgs}
		}
	case "preview_screenshot":
		result = map[string]string{
			"message": "Preview available at http://localhost:51820. Open in browser to view the rendered output.",
		}
	default:
		err = &RPCError{Code: -32601, Message: "unknown tool: " + params.Name}
	}

	return Response{
		JSONRPC: JSONRPCVersion,
		ID:      req.ID,
		Result:  result,
		Error:   err,
	}
}
