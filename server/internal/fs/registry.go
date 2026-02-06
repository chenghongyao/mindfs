package fs

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ManagedDir struct {
	ID        string    `json:"id"`
	RootPath  string    `json:"root_path"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Registry struct {
	mu    sync.Mutex
	path  string
	dirs  map[string]ManagedDir
	order []string
}

func NewRegistry(path string) *Registry {
	return &Registry{path: path, dirs: make(map[string]ManagedDir)}
}

func NewDefaultRegistry() (*Registry, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(configDir, "mindfs", "registry.json")
	return NewRegistry(path), nil
}

func (r *Registry) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	payload, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var stored struct {
		Dirs  []ManagedDir `json:"dirs"`
		Order []string     `json:"order"`
	}
	if err := json.Unmarshal(payload, &stored); err != nil {
		return err
	}
	r.dirs = make(map[string]ManagedDir)
	r.order = nil
	seen := make(map[string]struct{})
	for _, dir := range stored.Dirs {
		name := dir.Name
		if name == "" {
			name = filepath.Base(dir.RootPath)
		}
		if name == "" || name == "." || name == string(filepath.Separator) {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		dir.Name = name
		dir.ID = name
		r.dirs[name] = dir
		r.order = append(r.order, name)
	}
	return nil
}

func (r *Registry) Save() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveLocked()
}

func (r *Registry) saveLocked() error {
	if r.path == "" {
		return errors.New("registry path required")
	}
	if err := os.MkdirAll(filepath.Dir(r.path), 0o755); err != nil {
		return err
	}
	dirs := make([]ManagedDir, 0, len(r.dirs))
	for _, id := range r.order {
		if dir, ok := r.dirs[id]; ok {
			dirs = append(dirs, dir)
		}
	}
	payload, err := json.MarshalIndent(map[string]any{
		"dirs":  dirs,
		"order": r.order,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, payload, 0o644)
}

func (r *Registry) List() []ManagedDir {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]ManagedDir, 0, len(r.order))
	for _, id := range r.order {
		if dir, ok := r.dirs[id]; ok {
			result = append(result, dir)
		}
	}
	return result
}

func (r *Registry) Get(id string) (ManagedDir, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	dir, ok := r.dirs[id]
	return dir, ok
}

func (r *Registry) Upsert(root string) (ManagedDir, error) {
	if root == "" {
		return ManagedDir{}, errors.New("root required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	name := filepath.Base(root)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ManagedDir{}, errors.New("invalid directory name")
	}
	dir, ok := r.dirs[name]
	if !ok {
		dir = ManagedDir{
			ID:        name,
			RootPath:  root,
			Name:      name,
			CreatedAt: now,
		}
		r.order = append(r.order, name)
	}
	dir.UpdatedAt = now
	r.dirs[name] = dir
	return dir, r.saveLocked()
}

func (r *Registry) Add(root string) (ManagedDir, error) {
	if root == "" {
		return ManagedDir{}, errors.New("root required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	name := filepath.Base(root)
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ManagedDir{}, errors.New("invalid directory name")
	}
	for _, existing := range r.dirs {
		if existing.Name == name && existing.RootPath != root {
			return ManagedDir{}, errors.New("managed directory name already exists")
		}
	}
	now := time.Now().UTC()
	if _, ok := r.dirs[name]; ok {
		dir := r.dirs[name]
		dir.UpdatedAt = now
		r.dirs[name] = dir
		return dir, r.saveLocked()
	}
	dir := ManagedDir{
		ID:        name,
		RootPath:  root,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.dirs[name] = dir
	r.order = append(r.order, name)
	return dir, r.saveLocked()
}
