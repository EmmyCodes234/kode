package gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (s *Server) proxyToOpenModel(w http.ResponseWriter, r *http.Request, body []byte, model Model) {
	key := s.litePool.Next()
	if key == "" {
		http.Error(w, `{"error":"no Lite pool keys available"}`, http.StatusBadGateway)
		return
	}

	// Parse incoming OpenAI chat request
	var openAIReq map[string]any
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	var upstreamURL string
	var upstreamBody []byte

	switch model.Protocol {
	case ProtocolMessages:
		upstreamURL = "https://api.openmodel.ai/v1/messages"

		// Clone and adapt for Anthropic Messages format
		msg := make(map[string]any)
		for k, v := range openAIReq {
			msg[k] = v
		}
		// Anthropic requires max_tokens
		if _, ok := msg["max_tokens"]; !ok {
			msg["max_tokens"] = 4096
		}
		// Remove unsupported fields
		delete(msg, "stream")
		delete(msg, "temperature")
		delete(msg, "top_p")
		var err error
		upstreamBody, err = json.Marshal(msg)
		if err != nil {
			http.Error(w, `{"error":"marshal error"}`, http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, `{"error":"unsupported protocol"}`, http.StatusBadGateway)
		return
	}

	// Proxy the converted request
	proxyReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstreamURL, bytes.NewReader(upstreamBody))
	if err != nil {
		http.Error(w, `{"error":"proxy error"}`, http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", "Bearer "+key)
	proxyReq.Header.Set("Anthropic-Version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	upstreamResp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"upstream error: %v"}`, err), http.StatusBadGateway)
		return
	}
	defer upstreamResp.Body.Close()

	respBody, err := io.ReadAll(upstreamResp.Body)
	if err != nil {
		http.Error(w, `{"error":"read error"}`, http.StatusInternalServerError)
		return
	}

	// If upstream returned non-200, pass through as-is
	if upstreamResp.StatusCode != http.StatusOK {
		for k, vs := range upstreamResp.Header {
			for _, v := range vs {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(upstreamResp.StatusCode)
		w.Write(respBody)
		return
	}

	// Convert Anthropic response back to OpenAI chat completion format
	switch model.Protocol {
	case ProtocolMessages:
		var anthResp map[string]any
		if err := json.Unmarshal(respBody, &anthResp); err != nil {
			http.Error(w, `{"error":"parse upstream response"}`, http.StatusInternalServerError)
			return
		}
		openAIResp := convertAnthropicToOpenAI(anthResp, openAIReq)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(openAIResp)
	}
}

func convertAnthropicToOpenAI(anthResp map[string]any, origReq map[string]any) map[string]any {
	// Extract content text from Anthropic response
	content := ""
	if contentArr, ok := anthResp["content"].([]any); ok && len(contentArr) > 0 {
		if firstBlock, ok := contentArr[0].(map[string]any); ok {
			if t, ok := firstBlock["text"].(string); ok {
				content = t
			}
		}
	}

	// Extract stop reason
	stopReason := "stop"
	if sr, ok := anthResp["stop_reason"].(string); ok {
		switch sr {
		case "end_turn":
			stopReason = "stop"
		case "max_tokens":
			stopReason = "length"
		case "tool_use":
			stopReason = "tool_calls"
		}
	}

	// Build OpenAI-compatible response
	return map[string]any{
		"id":      anthResp["id"],
		"object":  "chat.completion",
		"created": 1700000000,
		"model":   origReq["model"],
		"choices": []any{
			map[string]any{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": stopReason,
			},
		},
		"usage": map[string]any{
			"prompt_tokens":     anthResp["usage"].(map[string]any)["input_tokens"],
			"completion_tokens": anthResp["usage"].(map[string]any)["output_tokens"],
			"total_tokens":      anthResp["usage"].(map[string]any)["input_tokens"].(float64) + anthResp["usage"].(map[string]any)["output_tokens"].(float64),
		},
	}
}
