package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"mindfs/server/internal/fs"
	"mindfs/server/internal/router"
)

func TestHandleViewRouteIsReachable(t *testing.T) {
	tempDir := t.TempDir()
	root := filepath.Join(tempDir, "mindfs")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("mkdir root: %v", err)
	}
	if _, err := fs.EnsureManagedDir(root); err != nil {
		t.Fatalf("ensure managed dir: %v", err)
	}

	registry := fs.NewRegistry(filepath.Join(tempDir, "registry.json"))
	if _, err := registry.Add(root); err != nil {
		t.Fatalf("add registry dir: %v", err)
	}

	handler := &HTTPHandler{
		Root:     root,
		Registry: registry,
		Views:    router.NewViewStoreManager(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/view?root=mindfs", nil)
	rec := httptest.NewRecorder()
	handler.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := body["view"]; !ok {
		t.Fatalf("expected response to include view field, got: %v", body)
	}
}
