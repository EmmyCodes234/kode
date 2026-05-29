package telemetry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	DefaultPostHogKey  = "phc_yuWiGPM25eYqXhGecMNxhZ9TkdPukx6RqPhgMHJCqqLj"
	DefaultPostHogHost = "https://us.posthog.com"
)

type PostHogClient struct {
	APIKey      string
	Host        string
	DistinctID  string
	Enabled     bool
	ProjectRoot string
}

type postHogEvent struct {
	APIKey     string                 `json:"api_key"`
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  string                 `json:"timestamp"`
}

func NewPostHogClient(projectRoot string) *PostHogClient {
	apiKey := os.Getenv("KODE_POSTHOG_API_KEY")
	host := os.Getenv("KODE_POSTHOG_HOST")
	enabled := os.Getenv("KODE_NO_TELEMETRY") == ""

	if apiKey == "" {
		apiKey = DefaultPostHogKey
	}
	if host == "" {
		host = DefaultPostHogHost
	}

	// Load from .kode/kode.json directly to prevent import cycles
	configPath := filepath.Join(projectRoot, ".kode", "kode.json")
	if data, err := os.ReadFile(configPath); err == nil {
		var raw map[string]interface{}
		if err := json.Unmarshal(data, &raw); err == nil {
			if tel, ok := raw["telemetry"].(map[string]interface{}); ok {
				if active, ok := tel["enabled"].(bool); ok && !active {
					enabled = false
				}
				if key, ok := tel["posthog_api_key"].(string); ok && key != "" {
					apiKey = key
				}
				if h, ok := tel["posthog_host"].(string); ok && h != "" {
					host = h
				}
			}
		}
	}

	return &PostHogClient{
		APIKey:      apiKey,
		Host:        host,
		DistinctID:  getAnonymousID(),
		Enabled:     enabled,
		ProjectRoot: projectRoot,
	}
}

func getAnonymousID() string {
	name, err := os.Hostname()
	if err != nil {
		name = "unknown-host"
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "unknown-home"
	}
	// Combine and hash
	h := sha256.New()
	h.Write([]byte(name + ":" + home + ":" + runtime.GOARCH + ":" + runtime.GOOS))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Track logs an event asynchronously in the background
func (ph *PostHogClient) Track(event string, properties map[string]interface{}) {
	if !ph.Enabled {
		return
	}

	// Run in background goroutine to prevent blocking execution
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if properties == nil {
			properties = make(map[string]interface{})
		}
		properties["distinct_id"] = ph.DistinctID
		properties["$lib"] = "kode-go"
		properties["$os"] = runtime.GOOS
		properties["$arch"] = runtime.GOARCH

		payload := postHogEvent{
			APIKey:     ph.APIKey,
			Event:      event,
			Properties: properties,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return
		}

		url := fmt.Sprintf("%s/capture/", ph.Host)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
		if err != nil {
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
	}()
}
