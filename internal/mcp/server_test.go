package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

func TestServerInitialization(t *testing.T) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	s := NewServer("test-mcp", nil, in, out)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServerHandleRequest(t *testing.T) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	s := NewServer("test-mcp", nil, in, out)

	req := Request{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "unknown_method",
	}
	s.handleRequest(context.Background(), req)

	// Out buffer should contain a jsonrpc error
	if out.Len() == 0 {
		t.Fatal("Expected error response for unknown method")
	}

	out.Reset()
	req = Request{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "initialize",
	}
	s.handleRequest(context.Background(), req)
	
	var resp Response
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if id, ok := resp.ID.(float64); !ok || id != 2 {
		t.Errorf("Expected id 2, got %v", resp.ID)
	}
}
