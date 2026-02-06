package api

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"mindfs/server/internal/agent"
	"mindfs/server/internal/audit"
	"mindfs/server/internal/fs"
	"mindfs/server/internal/router"
)

// HTTPHandler provides REST endpoints for health, tree, file, and action.
type HTTPHandler struct {
	Router   *router.Router
	Root     string
	Views    *router.ViewStoreManager
	Registry *fs.Registry
	Sessions *SessionService
	Prober   *agent.Prober
	Audit    *audit.WriterPool
}

// Routes constructs the chi router with all endpoints.
func (h *HTTPHandler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", h.handleIndex)
	r.Get("/health", h.handleHealth)
	r.Get("/api/tree", h.handleTree)
	r.Get("/api/file", h.handleFile)
	r.Get("/api/view", h.handleView)
	r.Post("/api/view/accept", h.handleViewAccept)
	r.Post("/api/view/revert", h.handleViewRevert)
	r.Get("/api/sessions", h.handleSessions)
	r.Get("/api/sessions/{key}", h.handleSessionGet)
	r.Post("/api/sessions", h.handleSessionCreate)
	r.Post("/api/sessions/{key}/message", h.handleSessionMessage)
	r.Post("/api/skills/{id}/execute", h.handleSkillExecute)
	r.Get("/api/dirs/{id}/config", h.handleDirConfigGet)
	r.Put("/api/dirs/{id}/config", h.handleDirConfigPut)
	r.Get("/api/dirs", h.handleDirs)
	r.Post("/api/dirs", h.handleAddDir)
	r.Post("/api/action", h.handleAction)

	// File metadata API
	r.Get("/api/file/meta", h.handleFileMeta)

	// Directory skills API
	r.Get("/api/dirs/{id}/skills", h.handleDirSkills)

	// Agent status API
	agentHandler := &AgentHandler{Prober: h.Prober}
	r.Mount("/api/agents", agentHandler.Routes())

	// View routes API
	viewHandler := &ViewHandler{Root: h.Root, Registry: h.Registry}
	r.Mount("/api/view", viewHandler.Routes())

	return r
}

// getAuditLogger returns an audit logger for the given root ID
func (h *HTTPHandler) getAuditLogger(rootID string) *audit.Logger {
	if h.Audit == nil {
		return nil
	}
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		return nil
	}
	return audit.NewLogger(h.Audit, rootID, resolved.ManagedDir)
}

func (h *HTTPHandler) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

func (h *HTTPHandler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (h *HTTPHandler) handleTree(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	root := h.Root
	rootID := r.URL.Query().Get("root")
	if rootID != "" && h.Registry != nil {
		if dir, ok := h.Registry.Get(rootID); ok {
			root = dir.RootPath
		} else {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not found"})
			return
		}
	}
	if root == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not configured"})
		return
	}
	dir := r.URL.Query().Get("dir")
	if dir == "" || dir == "." {
		dir = root
	}
	resolved, err := fs.ResolvePath(root, dir)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	entries, err := fs.ListEntries(root, resolved)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"tree": entries})
}

func (h *HTTPHandler) handleFile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	root := h.Root
	rootID := r.URL.Query().Get("root")
	if rootID != "" && h.Registry != nil {
		if dir, ok := h.Registry.Get(rootID); ok {
			root = dir.RootPath
		} else {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not found"})
			return
		}
	}
	if root == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not configured"})
		return
	}
	path := r.URL.Query().Get("path")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "path required"})
		return
	}
	raw := r.URL.Query().Get("raw")
	if raw == "1" {
		resolved, err := fs.ResolvePath(root, path)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		file, err := os.Open(resolved)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err == nil {
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
		}
		ext := filepath.Ext(resolved)
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, file)
		return
	}
	result, err := fs.ReadFile(root, path, 128*1024)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	if rootID != "" && h.Registry != nil {
		if dir, ok := h.Registry.Get(rootID); ok {
			result.Root = dir.Name
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"file": result})
}

func (h *HTTPHandler) handleAction(w http.ResponseWriter, r *http.Request) {
	if h.Router == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	var req struct {
		Action  string         `json:"action"`
		Path    string         `json:"path"`
		Context map[string]any `json:"context"`
		Meta    map[string]any `json:"meta"`
		Version string         `json:"version"`
		Root    string         `json:"root"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
		return
	}
	resp, err := h.Router.Dispatch(r.Context(), router.ActionRequest{
		Action:  req.Action,
		Path:    req.Path,
		Context: req.Context,
		Meta:    req.Meta,
		Version: req.Version,
		Root:    req.Root,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"status":  resp.Status,
		"handled": resp.Handled,
		"data":    resp.Data,
		"view":    resp.View,
		"effects": resp.Effects,
		"error":   resp.Error,
	})
}

func (h *HTTPHandler) handleView(w http.ResponseWriter, r *http.Request) {
	if h.Views == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "view store not configured"})
		return
	}
	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	managedDir := resolved.ManagedDir
	store, err := h.Views.Get(managedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	status, _ := fs.LoadViewStatus(managedDir)
	if view, ts, modTime, ok := store.Get(); ok {
		if info, err := os.Stat(filepath.Join(managedDir, "view.json")); err == nil && info.ModTime().After(modTime) {
			view, err = store.Load()
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"view": view, "updated_at": time.Now().UTC(), "pending": status.Pending, "view_id": status.CurrentID})
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"view": view, "updated_at": ts, "pending": status.Pending, "view_id": status.CurrentID})
		return
	}
	view, err := store.Load()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"view": nil, "pending": status.Pending, "view_id": status.CurrentID})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"view": view, "pending": status.Pending, "view_id": status.CurrentID})
}

func (h *HTTPHandler) handleViewAccept(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	status, err := fs.AcceptView(resolved.ManagedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "pending": status.Pending, "view_id": status.CurrentID})
}

func (h *HTTPHandler) handleViewRevert(w http.ResponseWriter, r *http.Request) {
	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	status, err := fs.RevertView(resolved.ManagedDir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	var view map[string]any
	if h.Views != nil {
		if store, err := h.Views.Get(resolved.ManagedDir); err == nil {
			view, _ = store.Load()
		}
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "pending": status.Pending, "view_id": status.CurrentID, "view": view})
}

func (h *HTTPHandler) handleDirs(w http.ResponseWriter, _ *http.Request) {
	if h.Registry == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "registry not configured"})
		return
	}
	if h.Root == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not configured"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	dirs := h.Registry.List()
	resp := make([]map[string]any, 0, len(dirs))
	for _, dir := range dirs {
		display := dir.Name
		resp = append(resp, map[string]any{
			"id":           dir.ID,
			"root_path":    ".",
			"display_name": display,
			"created_at":   dir.CreatedAt,
			"updated_at":   dir.UpdatedAt,
		})
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"dirs": resp})
}

func (h *HTTPHandler) handleAddDir(w http.ResponseWriter, r *http.Request) {
	if h.Registry == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "registry not configured"})
		return
	}
	if h.Root == "" {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "root not configured"})
		return
	}
	var req struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid json"})
		return
	}
	if req.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "path required"})
		return
	}
	if filepath.IsAbs(req.Path) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "absolute paths not allowed"})
		return
	}
	abs := filepath.Join(h.Root, req.Path)
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "path must be a directory"})
		return
	}
	if _, err := fs.EnsureManagedDir(abs); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	dir, err := h.Registry.Add(abs)
	if err != nil {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	rel, _ := filepath.Rel(h.Root, dir.RootPath)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"id":           dir.ID,
		"root_path":    rel,
		"display_name": dir.Name,
		"created_at":   dir.CreatedAt,
		"updated_at":   dir.UpdatedAt,
	})
}

func (h *HTTPHandler) handleFileMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rootID := r.URL.Query().Get("root")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "path required"})
		return
	}

	meta, err := fs.GetFileMeta(resolved.ManagedDir, path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	if meta == nil {
		_ = json.NewEncoder(w).Encode(map[string]any{"meta": nil})
		return
	}

	_ = json.NewEncoder(w).Encode(map[string]any{
		"meta": map[string]any{
			"source_session": meta.SourceSession,
			"session_name":   meta.SessionName,
			"agent":          meta.Agent,
			"created_at":     meta.CreatedAt,
			"updated_at":     meta.UpdatedAt,
			"created_by":     meta.CreatedBy,
		},
	})
}

func (h *HTTPHandler) handleDirSkills(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rootID := chi.URLParam(r, "id")
	resolved, err := resolveRoot(rootID, h.Root, h.Registry)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}

	// Load skills from .mindfs/skills/ directory
	skillsDir := filepath.Join(resolved.ManagedDir, "skills")
	skills := make([]map[string]any, 0)

	entries, err := os.ReadDir(skillsDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".json" && ext != ".yaml" && ext != ".yml" {
				continue
			}

			skillID := name[:len(name)-len(ext)]
			skillPath := filepath.Join(skillsDir, name)

			// Try to read skill metadata
			data, err := os.ReadFile(skillPath)
			if err != nil {
				continue
			}

			var skillMeta map[string]any
			if err := json.Unmarshal(data, &skillMeta); err != nil {
				continue
			}

			skillName, _ := skillMeta["name"].(string)
			if skillName == "" {
				skillName = skillID
			}
			description, _ := skillMeta["description"].(string)

			skills = append(skills, map[string]any{
				"id":          skillID,
				"name":        skillName,
				"description": description,
				"source":      "directory",
			})
		}
	}

	_ = json.NewEncoder(w).Encode(map[string]any{"skills": skills})
}

const indexHTML = `<!doctype html>
<html lang="zh-CN">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>MindFS</title>
    <style>
      :root {
        color-scheme: light;
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        font-family: "Avenir Next", "Helvetica Neue", Arial, sans-serif;
        background: #f7f7f5;
        color: #222;
      }
      .shell {
        display: grid;
        grid-template-columns: 260px 1fr;
        grid-template-rows: 1fr 64px;
        height: 100vh;
        background: #fff;
      }
      aside {
        grid-row: 1 / span 2;
        border-right: 1px solid #e5e5e5;
        padding: 12px;
        overflow: auto;
        background: linear-gradient(180deg, #faf9f6 0%, #ffffff 100%);
      }
      main {
        padding: 16px;
        overflow: auto;
      }
      footer {
        border-top: 1px solid #e5e5e5;
        padding: 12px 16px;
        display: flex;
        align-items: center;
        gap: 12px;
      }
      .file-button {
        display: block;
        width: 100%;
        text-align: left;
        border: none;
        background: transparent;
        padding: 4px 0;
        cursor: pointer;
        color: #333;
      }
      .card {
        border: 1px solid #e5e5e5;
        border-radius: 8px;
        padding: 10px 12px;
        display: flex;
        justify-content: space-between;
        margin-bottom: 8px;
        background: #fff;
      }
      .status {
        font-size: 12px;
        color: #666;
      }
      input[type="text"] {
        flex: 1;
        padding: 8px 10px;
        border-radius: 6px;
        border: 1px solid #ccc;
      }
      button {
        padding: 8px 12px;
        border-radius: 6px;
        border: 1px solid #ccc;
        background: #f5f5f5;
        cursor: pointer;
      }
    </style>
  </head>
  <body>
    <div class="shell">
      <aside>
        <h3>Files</h3>
        <div id="tree">加载中...</div>
      </aside>
      <main>
        <h2 style="margin-top: 0;">Workspace</h2>
        <div id="list"></div>
      </main>
      <footer>
        <input type="text" placeholder="Ask or type a command..." />
        <button type="button">Run</button>
        <span class="status" id="status">Connected</span>
      </footer>
    </div>
    <script>
      fetch("/api/tree?dir=.")
        .then(function (res) { return res.json(); })
        .then(function (payload) {
          var tree = payload.tree || [];
          var treeEl = document.getElementById("tree");
          var listEl = document.getElementById("list");
          treeEl.innerHTML = "";
          listEl.innerHTML = "";
          tree.forEach(function (entry) {
            var btn = document.createElement("button");
            btn.className = "file-button";
            btn.textContent = (entry.is_dir ? "📁 " : "📄 ") + entry.name;
            treeEl.appendChild(btn);

            var card = document.createElement("div");
            card.className = "card";
            card.innerHTML = "<span>" + entry.name + "</span><span>" + (entry.is_dir ? "Folder" : "File") + "</span>";
            listEl.appendChild(card);
          });
        })
        .catch(function () {
          var treeEl = document.getElementById("tree");
          treeEl.textContent = "无法加载目录树";
        });
    </script>
  </body>
</html>`
