package session

import "sync"

type StoreManager struct {
	mu     sync.Mutex
	stores map[string]*Store
}

func NewStoreManager() *StoreManager {
	return &StoreManager{stores: make(map[string]*Store)}
}

func (m *StoreManager) Get(managedDir string) (*Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if store, ok := m.stores[managedDir]; ok {
		return store, nil
	}
	store, err := NewStore(managedDir)
	if err != nil {
		return nil, err
	}
	m.stores[managedDir] = store
	return store, nil
}

func (m *StoreManager) List() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	paths := make([]string, 0, len(m.stores))
	for key := range m.stores {
		paths = append(paths, key)
	}
	return paths
}
