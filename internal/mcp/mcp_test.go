package mcp

import (
	"encoding/json"
	"testing"
)

func TestInitialize(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	resp := s.handle(req)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("wrong protocol version: %v", result["protocolVersion"])
	}
}

func TestToolsList(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 2, Method: "tools/list"}
	resp := s.handle(req)
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("result is not a map")
	}
	tools, ok := result["tools"].([]Tool)
	if !ok {
		t.Fatalf("tools is not []Tool, got %T", result["tools"])
	}
	if len(tools) < 2 {
		t.Errorf("expected at least 2 tools, got %d", len(tools))
	}
}

func TestUnknownMethod(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 3, Method: "nonexistent"}
	resp := s.handle(req)
	if resp.Error == nil {
		t.Error("expected error for unknown method")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("expected -32601, got %d", resp.Error.Code)
	}
}

func TestToolsCallUnknown(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	params, _ := json.Marshal(map[string]interface{}{
		"name":      "nonexistent_tool",
		"arguments": map[string]interface{}{},
	})
	req := Request{JSONRPC: "2.0", ID: 4, Method: "tools/call", Params: params}
	resp := s.handle(req)
	if resp.Error == nil {
		t.Error("expected error for unknown tool")
	}
}

func TestJSONRPCFormatting(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	resp := s.handle(req)

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed Response
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if parsed.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", parsed.JSONRPC)
	}
}

func TestFiveTools(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "tools/list"}
	resp := s.handle(req)
	result, _ := resp.Result.(map[string]interface{})
	tools, _ := result["tools"].([]Tool)

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}

	expected := []string{"preview_reload", "preview_console", "preview_screenshot", "preview_inspect", "preview_diff"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
	if len(tools) != 5 {
		t.Errorf("expected 5 tools, got %d", len(tools))
	}
}

func TestInspectHasSelectorParam(t *testing.T) {
	s := &Server{client: nil, serverURL: "http://localhost:51820"}
	req := Request{JSONRPC: "2.0", ID: 1, Method: "tools/list"}
	resp := s.handle(req)
	result, _ := resp.Result.(map[string]interface{})
	tools, _ := result["tools"].([]Tool)

	for _, tool := range tools {
		if tool.Name == "preview_inspect" {
			if _, ok := tool.InputSchema.Properties["selector"]; !ok {
				t.Error("preview_inspect should have selector property")
			}
			if len(tool.InputSchema.Required) == 0 || tool.InputSchema.Required[0] != "selector" {
				t.Error("preview_inspect should require selector")
			}
			return
		}
	}
	t.Error("preview_inspect tool not found")
}
