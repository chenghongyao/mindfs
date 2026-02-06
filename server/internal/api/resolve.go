package api

import (
	"errors"
	"path/filepath"

	"mindfs/server/internal/fs"
)

type resolvedRoot struct {
	Path       string
	Name       string
	ManagedDir string
}

func resolveRoot(rootID, defaultRoot string, registry *fs.Registry) (resolvedRoot, error) {
	root := defaultRoot
	name := filepath.Base(defaultRoot)
	if rootID != "" {
		if registry == nil {
			return resolvedRoot{}, errors.New("registry not configured")
		}
		dir, ok := registry.Get(rootID)
		if !ok {
			return resolvedRoot{}, errors.New("root not found")
		}
		root = dir.RootPath
		name = dir.Name
	}
	if root == "" {
		return resolvedRoot{}, errors.New("root not configured")
	}
	return resolvedRoot{
		Path:       root,
		Name:       name,
		ManagedDir: filepath.Join(root, ".mindfs"),
	}, nil
}
