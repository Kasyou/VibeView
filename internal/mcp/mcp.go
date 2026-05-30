package mcp

import (
	"bufio"
	"bytes"
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
			Description: "Capture a screenshot of the current preview page. Returns a base64 PNG image of what the user sees in the browser. Requires a browser to be connected to the preview server.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_inspect",
			Description: "Query an element's position, size, styles, and text content from the preview page. Provide a CSS selector to target a specific element (e.g. 'button', '.card', '#app').",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"selector": {Type: "string", Description: "CSS selector to query (e.g. 'h1', '.button', '#app')"},
				},
				Required: []string{"selector"},
			},
		},
		{
			Name:        "preview_diff",
			Description: "Compare the current preview screenshot with the previous one. Returns whether visual changes were detected, along with both before/after screenshots if changed. Useful for verifying that code changes produced the expected visual result.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_show",
			Description: "Push visual content to the Claude whiteboard. Send markdown text (headings, lists, code, tables) to be rendered as styled cards in the browser. Use this to visualize your reasoning, show conclusions, draw comparison tables, or present architecture overviews.",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"content": {Type: "string", Description: "Markdown content to render on the whiteboard"},
					"title":   {Type: "string", Description: "Optional card title shown above the content"},
				},
				Required: []string{"content"},
			},
		},
		{
			Name:        "preview_clear",
			Description: "Clear all cards from the Claude whiteboard. Use when starting a new topic or at the user's request.",
			InputSchema: InputSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "preview_history",
			Description: "Query card history from the Claude whiteboard. Returns paginated cards with #seq, time, title and content. Use offset and limit to browse older cards (e.g. offset=0 limit=10 for the latest 10, offset=10 limit=10 for the next batch).",
			InputSchema: InputSchema{
				Type: "object",
				Properties: map[string]Property{
					"offset": {Type: "integer", Description: "Number of cards to skip (default: 0)"},
					"limit":  {Type: "integer", Description: "Max cards to return (default: 30)"},
				},
			},
		},
		{
			Name:        "preview_stop",
			Description: "Stop the VibeView preview server. Use this when the user is done previewing to free up resources.",
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
		// Request screenshot from the browser via HTTP API
		resp, httpErr := s.client.Post(s.serverURL+"/api/screenshot", "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			raw, _ := io.ReadAll(resp.Body)
			var data map[string]interface{}
			json.Unmarshal(raw, &data)
			if img, ok := data["image"]; ok && img != "" {
				result = map[string]interface{}{
					"image":    img,
					"mimeType": "image/png",
				}
			} else {
				result = map[string]string{
					"message": "Screenshot not available. Open http://localhost:51820 in browser to see the preview. Ensure a browser tab is connected.",
				}
			}
		}

	case "preview_inspect":
		selector := "body"
		if sel, ok := params.Arguments["selector"]; ok {
			if s, ok2 := sel.(string); ok2 && s != "" {
				selector = s
			}
		}
		resp, httpErr := s.client.Post(s.serverURL+"/api/inspect?selector="+selector, "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			raw, _ := io.ReadAll(resp.Body)
			var data map[string]interface{}
			json.Unmarshal(raw, &data)
			result = data
		}

	case "preview_diff":
		resp, httpErr := s.client.Get(s.serverURL + "/api/diff")
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			raw, _ := io.ReadAll(resp.Body)
			var data map[string]interface{}
			json.Unmarshal(raw, &data)
			result = data
		}

	case "preview_show":
		content := ""
		title := ""
		if c, ok := params.Arguments["content"]; ok {
			content = fmt.Sprintf("%v", c)
		}
		if t, ok := params.Arguments["title"]; ok {
			title = fmt.Sprintf("%v", t)
		}
		if content == "" {
			err = &RPCError{Code: -32602, Message: "content required"}
		} else {
			body, _ := json.Marshal(map[string]string{"content": content, "title": title})
			resp, httpErr := s.client.Post(s.serverURL+"/api/show", "application/json", bytes.NewReader(body))
			if httpErr != nil {
				err = &RPCError{Code: -32000, Message: httpErr.Error()}
			} else {
				resp.Body.Close()
				result = map[string]string{"status": "shown"}
			}
		}

	case "preview_clear":
		resp, httpErr := s.client.Post(s.serverURL+"/api/clear", "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			resp.Body.Close()
			result = map[string]string{"status": "cleared"}
		}

	case "preview_history":
		offset := 0
		limit := 30
		if o, ok := params.Arguments["offset"]; ok {
			if f, ok2 := o.(float64); ok2 { offset = int(f) }
		}
		if l, ok := params.Arguments["limit"]; ok {
			if f, ok2 := l.(float64); ok2 { limit = int(f) }
		}
		resp, httpErr := s.client.Get(fmt.Sprintf("%s/api/cards?offset=%d&limit=%d", s.serverURL, offset, limit))
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			defer resp.Body.Close()
			raw, _ := io.ReadAll(resp.Body)
			var data map[string]interface{}
			json.Unmarshal(raw, &data)
			result = data
		}

	case "preview_stop":
		resp, httpErr := s.client.Post(s.serverURL+"/api/shutdown", "application/json", nil)
		if httpErr != nil {
			err = &RPCError{Code: -32000, Message: httpErr.Error()}
		} else {
			resp.Body.Close()
			result = map[string]string{"status": "stopped"}
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
