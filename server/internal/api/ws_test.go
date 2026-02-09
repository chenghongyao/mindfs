package api

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ctxbuilder "mindfs/server/internal/context"
	"mindfs/server/internal/session"
	"mindfs/server/internal/skills"
)

func TestParseClientContext(t *testing.T) {
	payload := map[string]any{
		"context": map[string]any{
			"current_root": "ignored-by-payload",
			"current_path": "docs/readme.md",
			"selection": map[string]any{
				"file_path": "docs/readme.md",
				"start":     1,
				"end":       3,
				"text":      "abc",
			},
		},
	}

	got := parseClientContext(payload, "mindfs")
	if got.CurrentRoot != "ignored-by-payload" {
		t.Fatalf("unexpected current root: %q", got.CurrentRoot)
	}
	if got.CurrentPath != "docs/readme.md" {
		t.Fatalf("unexpected current path: %q", got.CurrentPath)
	}
	if got.Selection == nil || got.Selection.Text != "abc" {
		t.Fatalf("unexpected selection: %#v", got.Selection)
	}

	got = parseClientContext(map[string]any{}, "fallback-root")
	if got.CurrentRoot != "fallback-root" {
		t.Fatalf("expected fallback root, got %q", got.CurrentRoot)
	}
}

func TestBuildContinuationPromptOnlyKeepsSelection(t *testing.T) {
	h := &WSHandler{}
	prompt := h.buildContinuationPrompt("继续处理", ctxbuilder.ClientContext{
		CurrentRoot: "mindfs",
		CurrentPath: "docs/readme.md",
		Selection: &ctxbuilder.Selection{
			FilePath: "docs/readme.md",
			Start:    0,
			End:      4,
			Text:     "TODO",
		},
	})

	if !strings.Contains(prompt, "继续处理") {
		t.Fatalf("prompt should contain message, got: %q", prompt)
	}
	if !strings.Contains(prompt, "选中内容: TODO") {
		t.Fatalf("prompt should contain selection, got: %q", prompt)
	}
	if strings.Contains(prompt, "当前路径:") {
		t.Fatalf("continuation prompt should not include current path, got: %q", prompt)
	}
}

func TestBuildInitialPromptIncludesServerContext(t *testing.T) {
	tempDir := t.TempDir()
	root := filepath.Join(tempDir, "mindfs")
	managedDir := filepath.Join(root, ".mindfs")
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		t.Fatalf("mkdir managed dir: %v", err)
	}
	if err := skills.SaveDirConfig(managedDir, skills.DirConfig{
		UserDescription: "股票分析目录",
	}); err != nil {
		t.Fatalf("save dir config: %v", err)
	}

	svc := &SessionService{Stores: session.NewStoreManager()}
	store, err := svc.Stores.Get(managedDir)
	if err != nil {
		t.Fatalf("get store: %v", err)
	}
	manager := session.NewManager(store)
	sessionItem, err := manager.Create(context.Background(), session.CreateInput{
		Type:  session.TypeChat,
		Agent: "codex",
		Name:  "context-test",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	h := &WSHandler{Sessions: svc}
	prompt := h.buildInitialPrompt(
		sessionItem,
		&ResolvedRoot{Path: root, ManagedDir: managedDir},
		"帮我总结当前文件",
		ctxbuilder.ClientContext{
			CurrentRoot: "mindfs",
			CurrentPath: "reports/daily.md",
		},
	)

	if !strings.Contains(prompt, "系统上下文:") {
		t.Fatalf("initial prompt should include system section, got: %q", prompt)
	}
	if !strings.Contains(prompt, "目录描述: 股票分析目录") {
		t.Fatalf("initial prompt should include user description, got: %q", prompt)
	}
	if !strings.Contains(prompt, "当前路径: reports/daily.md") {
		t.Fatalf("initial prompt should include current path, got: %q", prompt)
	}
}
