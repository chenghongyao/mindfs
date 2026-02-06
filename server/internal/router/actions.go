package router

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"mindfs/server/internal/fs"
)

const DefaultViewVersion = "v1"

// ActionService wires default action handlers.
type ActionService struct {
	Root       string
	ManagedDir string
	Router     *Router
	Registry   *fs.Registry
}

func (s *ActionService) RegisterDefaults() error {
	if s.Router == nil {
		return errors.New("router required")
	}
	return s.Router.Register("open", DefaultViewVersion, s.openHandler)
}

func (s *ActionService) openHandler(ctx context.Context, req ActionRequest) (ActionResponse, error) {
	if req.Version != DefaultViewVersion {
		return ActionResponse{Status: "error", Handled: false, Error: map[string]any{"code": "version_mismatch", "message": "version mismatch"}}, nil
	}
	root := s.Root
	rootName := ""
	if req.Root != "" && s.Registry != nil {
		dir, ok := s.Registry.Get(req.Root)
		if !ok {
			return ActionResponse{Status: "error", Handled: true, Error: map[string]any{"code": "root_not_found", "message": "root not found"}}, nil
		}
		root = dir.RootPath
		rootName = dir.Name
	} else {
		rootName = filepath.Base(root)
	}
	start := time.Now()
	result, err := fs.ReadFile(root, req.Path, 128*1024)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return ActionResponse{Status: "error", Handled: true, Error: map[string]any{"code": "read_failed", "message": err.Error()}, Effects: []any{map[string]any{"type": "latency", "payload": latency}}}, nil
	}
	result.Root = rootName
	managedDir := s.ManagedDir
	if root != s.Root {
		managedDir = filepath.Join(root, ".mindfs")
	}
	state, _ := fs.LoadState(managedDir)
	data := map[string]any{
		"file":    result,
		"state":   state,
		"latency": latency,
	}
	return ActionResponse{Status: "ok", Handled: true, Data: data}, nil
}
