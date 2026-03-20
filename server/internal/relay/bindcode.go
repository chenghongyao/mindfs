package relay

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type BindCode struct {
	BaseURL         string
	ActivationToken string
}

func ParseBindCode(raw string, creds Credentials) (BindCode, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return BindCode{}, errors.New("bind code required")
	}

	if strings.HasPrefix(raw, "mindfs://bind?") {
		u, err := url.Parse(raw)
		if err != nil {
			return BindCode{}, err
		}
		return BindCode{
			BaseURL:         strings.TrimSpace(u.Query().Get("base_url")),
			ActivationToken: strings.TrimSpace(u.Query().Get("activation_token")),
		}, nil
	}

	if strings.Contains(raw, "|") {
		parts := strings.SplitN(raw, "|", 2)
		return BindCode{
			BaseURL:         strings.TrimSpace(parts[0]),
			ActivationToken: strings.TrimSpace(parts[1]),
		}, nil
	}

	if decoded, err := base64.RawURLEncoding.DecodeString(raw); err == nil {
		var payload struct {
			BaseURL         string `json:"base_url"`
			ActivationToken string `json:"activation_token"`
		}
		if json.Unmarshal(decoded, &payload) == nil && payload.ActivationToken != "" {
			return BindCode{
				BaseURL:         strings.TrimSpace(payload.BaseURL),
				ActivationToken: strings.TrimSpace(payload.ActivationToken),
			}, nil
		}
	}

	baseURL := endpointBaseURL(creds.Relay.Endpoint)
	if baseURL == "" {
		baseURL = strings.TrimSpace(os.Getenv("MINDFS_RELAY_BASE_URL"))
	}
	if baseURL == "" {
		return BindCode{}, fmt.Errorf("bind code missing relay base URL")
	}
	return BindCode{
		BaseURL:         baseURL,
		ActivationToken: raw,
	}, nil
}

func buildActivateURL(baseURL string) (string, error) {
	baseURL = strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return "", errors.New("relay base URL required")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "http", "https":
		u.Path = strings.TrimSuffix(u.Path, "/") + "/api/activate"
		u.RawQuery = ""
		u.Fragment = ""
		return u.String(), nil
	default:
		return "", fmt.Errorf("unsupported relay base URL scheme: %s", u.Scheme)
	}
}

func endpointBaseURL(endpoint string) string {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return ""
	}
	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	default:
		return ""
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimSuffix(u.String(), "/")
}
