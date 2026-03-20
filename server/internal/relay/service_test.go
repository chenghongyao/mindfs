package relay

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseBindCode(t *testing.T) {
	t.Setenv("MINDFS_RELAY_BASE_URL", "")

	t.Run("url scheme", func(t *testing.T) {
		code, err := ParseBindCode("mindfs://bind?base_url=https%3A%2F%2Frelay.example.com&activation_token=act_123", Credentials{})
		if err != nil {
			t.Fatalf("ParseBindCode() error = %v", err)
		}
		if code.BaseURL != "https://relay.example.com" || code.ActivationToken != "act_123" {
			t.Fatalf("unexpected bind code: %+v", code)
		}
	})

	t.Run("base64 json", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{
			"base_url":         "https://relay.example.com",
			"activation_token": "act_456",
		})
		raw := base64.RawURLEncoding.EncodeToString(payload)
		code, err := ParseBindCode(raw, Credentials{})
		if err != nil {
			t.Fatalf("ParseBindCode() error = %v", err)
		}
		if code.BaseURL != "https://relay.example.com" || code.ActivationToken != "act_456" {
			t.Fatalf("unexpected bind code: %+v", code)
		}
	})
}

func TestCredentialsStoreSaveLoad(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("HOME", configRoot)

	store, err := NewCredentialsStore()
	if err != nil {
		t.Fatalf("NewCredentialsStore() error = %v", err)
	}

	input := Credentials{
		Relay: RelayCredentials{
			DeviceToken: "dev_123",
			NodeID:      "node_123",
			Endpoint:    "wss://relay.example.com/ws/connector",
		},
	}
	if err := store.Save(input); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Relay != input.Relay {
		t.Fatalf("Load() = %+v, want %+v", got.Relay, input.Relay)
	}

	if _, err := os.Stat(store.filePath); err != nil {
		t.Fatalf("credentials file missing: %v", err)
	}
}

func TestServiceActivate(t *testing.T) {
	var activationCalls int
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("HOME", configRoot)

	svc, err := NewService(":7331", "https://relay.example.com|act_abc")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	svc.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://relay.example.com/api/activate" {
				t.Fatalf("unexpected activation URL: %s", req.URL.String())
			}
			activationCalls++
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"device_token":"dev_abc","node_id":"node_abc","endpoint":"wss://relay.example.com/ws/connector"}`)),
			}, nil
		}),
	}

	creds, err := svc.Bind(context.Background(), "https://relay.example.com|act_abc")
	if err != nil {
		t.Fatalf("Bind() error = %v", err)
	}
	if activationCalls != 1 {
		t.Fatalf("activationCalls = %d, want 1", activationCalls)
	}
	if creds.Relay.DeviceToken != "dev_abc" {
		t.Fatalf("device token = %q", creds.Relay.DeviceToken)
	}
}

func TestManagerStartBindsInitialCode(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("HOME", configRoot)

	manager, err := NewManager(":7331", "https://relay.example.com|act_start")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	var activationCalls int
	manager.service.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "http://localhost:7331/health" {
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader("ok")),
					Header:     http.Header{},
				}, nil
			}
			if req.URL.String() != "https://relay.example.com/api/activate" {
				t.Fatalf("unexpected request URL: %s", req.URL.String())
			}
			activationCalls++
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(strings.NewReader(`{"device_token":"dev_start","node_id":"node_start","endpoint":"wss://relay.example.com/ws/connector"}`)),
			}, nil
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if activationCalls != 1 {
		t.Fatalf("activationCalls = %d, want 1", activationCalls)
	}

	creds, err := manager.service.store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if creds.Relay.NodeID != "node_start" {
		t.Fatalf("node id = %q", creds.Relay.NodeID)
	}
}

func TestManagerBindStartsRelayAfterEmptyStart(t *testing.T) {
	configRoot := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configRoot)
	t.Setenv("HOME", configRoot)

	manager, err := NewManager(":7331", "")
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	requests := make(chan string, 8)
	manager.service.client = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			requests <- req.URL.String()
			switch req.URL.String() {
			case "https://relay.example.com/api/activate":
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     http.Header{"Content-Type": []string{"application/json"}},
					Body:       io.NopCloser(strings.NewReader(`{"device_token":"dev_live","node_id":"node_live","endpoint":"wss://relay.example.com/ws/connector"}`)),
				}, nil
			case "http://localhost:7331/health":
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Header:     http.Header{},
					Body:       io.NopCloser(strings.NewReader("ok")),
				}, nil
			default:
				return nil, context.Canceled
			}
		}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if _, err := manager.Bind(context.Background(), "https://relay.example.com|act_live"); err != nil {
		t.Fatalf("Bind() error = %v", err)
	}

	timeout := time.After(2 * time.Second)
	for {
		select {
		case raw := <-requests:
			if raw == "http://localhost:7331/health" {
				return
			}
		case <-timeout:
			t.Fatal("relay did not start after bind")
		}
	}
}

func TestLocalTargetURL(t *testing.T) {
	target, err := localTargetURL("http://127.0.0.1:7331", mustParseURL("/api/file?root=a"))
	if err != nil {
		t.Fatalf("localTargetURL() error = %v", err)
	}
	if target.String() != "http://127.0.0.1:7331/api/file?root=a" {
		t.Fatalf("localTargetURL() = %s", target.String())
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}
