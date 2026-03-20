package relay

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hashicorp/yamux"
)

const (
	wsFrameData  byte = 1
	wsFrameClose byte = 2
)

type Service struct {
	localAddr string
	localURL  string
	store     *CredentialsStore
	client    *http.Client
}

type activationResponse struct {
	DeviceToken string `json:"device_token"`
	NodeID      string `json:"node_id"`
	Endpoint    string `json:"endpoint"`
}

func NewService(localAddr, _ string) (*Service, error) {
	store, err := NewCredentialsStore()
	if err != nil {
		return nil, err
	}
	return &Service{
		localAddr: localAddr,
		localURL:  addrToURL(localAddr, ""),
		store:     store,
		client:    &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (s *Service) Run(ctx context.Context) error {
	creds, err := s.store.Load()
	if err != nil {
		return err
	}
	if creds.Relay.DeviceToken == "" || creds.Relay.Endpoint == "" {
		return nil
	}
	if err := s.waitForLocalServer(ctx); err != nil {
		return err
	}

	backoff := time.Second
	for {
		err := s.runSession(ctx, creds.Relay)
		if ctx.Err() != nil {
			return nil
		}
		if isPermanentRelayError(err) {
			return err
		}
		log.Printf("[relay] reconnecting after error: %v", err)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (s *Service) Bind(ctx context.Context, bindCode string) (Credentials, error) {
	creds, err := s.store.Load()
	if err != nil {
		return Credentials{}, err
	}
	bind, err := ParseBindCode(strings.TrimSpace(bindCode), creds)
	if err != nil {
		return Credentials{}, err
	}
	activated, err := s.activate(ctx, bind)
	if err != nil {
		return Credentials{}, err
	}
	creds = Credentials{Relay: activated}
	if err := s.store.Save(creds); err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

func (s *Service) activate(ctx context.Context, bind BindCode) (RelayCredentials, error) {
	activateURL, err := buildActivateURL(bind.BaseURL)
	if err != nil {
		return RelayCredentials{}, err
	}
	body, err := json.Marshal(map[string]string{"activation_token": bind.ActivationToken})
	if err != nil {
		return RelayCredentials{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, activateURL, bytes.NewReader(body))
	if err != nil {
		return RelayCredentials{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return RelayCredentials{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return RelayCredentials{}, fmt.Errorf("relay activation failed: %s %s", resp.Status, strings.TrimSpace(string(payload)))
	}

	var out activationResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return RelayCredentials{}, err
	}
	return RelayCredentials{
		DeviceToken: strings.TrimSpace(out.DeviceToken),
		NodeID:      strings.TrimSpace(out.NodeID),
		Endpoint:    strings.TrimSpace(out.Endpoint),
	}, nil
}

func (s *Service) runSession(ctx context.Context, creds RelayCredentials) error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+creds.DeviceToken)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, creds.Endpoint, headers)
	if err != nil {
		return err
	}
	defer conn.Close()

	wsConn := NewWebSocketNetConn(conn)
	muxSession, err := yamux.Client(wsConn, nil)
	if err != nil {
		return err
	}
	defer muxSession.Close()

	errCh := make(chan error, 1)
	go func() {
		for {
			stream, err := muxSession.Accept()
			if err != nil {
				errCh <- err
				return
			}
			go func() {
				if err := s.handleStream(ctx, stream); err != nil {
					log.Printf("[relay] stream failed: %v", err)
				}
			}()
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Service) handleStream(ctx context.Context, stream net.Conn) error {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	req, err := http.ReadRequest(reader)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	if websocket.IsWebSocketUpgrade(req) {
		return s.proxyWebSocket(req, stream)
	}
	return s.proxyHTTP(req, stream)
}

func (s *Service) proxyHTTP(req *http.Request, stream io.Writer) error {
	targetURL, err := localTargetURL(s.localURL, req.URL)
	if err != nil {
		return err
	}
	outbound := req.Clone(req.Context())
	outbound.URL = targetURL
	outbound.RequestURI = ""
	outbound.Host = targetURL.Host

	resp, err := s.client.Do(outbound)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return resp.Write(stream)
}

func (s *Service) proxyWebSocket(req *http.Request, stream io.ReadWriter) error {
	targetURL, err := websocketTargetURL(s.localURL, req.URL)
	if err != nil {
		return err
	}
	headers := cloneHeader(req.Header)
	headers.Del("Connection")
	headers.Del("Upgrade")
	headers.Del("Sec-WebSocket-Key")
	headers.Del("Sec-WebSocket-Version")
	headers.Del("Sec-WebSocket-Extensions")

	dialer := *websocket.DefaultDialer
	if protocol := strings.TrimSpace(req.Header.Get("Sec-WebSocket-Protocol")); protocol != "" {
		dialer.Subprotocols = splitHeaderValues(protocol)
	}
	localConn, resp, err := dialer.DialContext(req.Context(), targetURL, headers)
	if err != nil {
		if resp != nil {
			_ = resp.Write(stream)
		}
		return err
	}
	defer localConn.Close()

	if resp == nil {
		return errors.New("relay websocket upgrade missing response")
	}
	if err := resp.Write(stream); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go bridgeStreamToWebSocket(stream, localConn, errCh)
	go bridgeWebSocketToStream(localConn, stream, errCh)
	err = <-errCh
	_ = writeWSCloseFrame(stream, websocket.CloseNormalClosure, "connector_closed")
	return err
}

func (s *Service) waitForLocalServer(ctx context.Context) error {
	healthURL := strings.TrimSuffix(s.localURL, "/") + "/health"
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			return err
		}
		resp, err := s.client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func addrToURL(addr, path string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return strings.TrimSuffix(addr, "/") + path
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		host = "localhost"
		port = strings.TrimPrefix(addr, ":")
	}
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "7331"
	}
	return fmt.Sprintf("http://%s:%s%s", host, port, path)
}

func localTargetURL(base string, requestURL *url.URL) (*url.URL, error) {
	target, err := url.Parse(strings.TrimSuffix(base, "/") + requestURL.RequestURI())
	if err != nil {
		return nil, err
	}
	target.Fragment = ""
	return target, nil
}

func websocketTargetURL(base string, requestURL *url.URL) (string, error) {
	target, err := localTargetURL(base, requestURL)
	if err != nil {
		return "", err
	}
	switch target.Scheme {
	case "http":
		target.Scheme = "ws"
	case "https":
		target.Scheme = "wss"
	default:
		return "", fmt.Errorf("unsupported websocket target scheme: %s", target.Scheme)
	}
	return target.String(), nil
}

func splitHeaderValues(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func cloneHeader(header http.Header) http.Header {
	clone := make(http.Header, len(header))
	for key, values := range header {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}

func isPermanentRelayError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "401") || strings.Contains(msg, "403")
}
