package fs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// FileCreatedCallback 文件创建回调，包含完整信息
type FileCreatedCallback func(relativePath, sessionKey string, size int64)

type FileWatcher struct {
	rootPath   string
	managedDir string
	sessionKey string
	watcher    *fsnotify.Watcher
	mu         sync.Mutex
	onCreate   FileCreatedCallback
	done       chan struct{}
}

func NewFileWatcher(rootPath, managedDir, sessionKey string, onCreate FileCreatedCallback) (*FileWatcher, error) {
	if rootPath == "" {
		return nil, errors.New("root path required")
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	fw := &FileWatcher{
		rootPath:   rootPath,
		managedDir: managedDir,
		sessionKey: sessionKey,
		watcher:    w,
		onCreate:   onCreate,
		done:       make(chan struct{}),
	}
	if err := fw.addWatchRecursive(rootPath); err != nil {
		_ = w.Close()
		return nil, err
	}
	go fw.run()
	return fw, nil
}

// UpdateSessionKey 更新当前关联的 Session Key
func (w *FileWatcher) UpdateSessionKey(sessionKey string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sessionKey = sessionKey
}

func (w *FileWatcher) Close() {
	w.mu.Lock()
	select {
	case <-w.done:
		w.mu.Unlock()
		return
	default:
		close(w.done)
	}
	w.mu.Unlock()
	_ = w.watcher.Close()
}

func (w *FileWatcher) run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == 0 {
				continue
			}
			if w.shouldIgnore(event.Name) {
				continue
			}
			info, err := os.Stat(event.Name)
			if err != nil {
				continue
			}
			if info.IsDir() {
				_ = w.addWatchRecursive(event.Name)
				continue
			}
			rel, err := filepath.Rel(w.rootPath, event.Name)
			if err != nil {
				continue
			}

			// 获取当前 session key
			w.mu.Lock()
			sessionKey := w.sessionKey
			w.mu.Unlock()

			// 更新 file-meta.json
			if w.managedDir != "" && sessionKey != "" {
				_ = UpdateFileMeta(w.managedDir, rel, sessionKey, "agent")
			}

			// 调用回调
			if w.onCreate != nil {
				w.onCreate(rel, sessionKey, info.Size())
			}
		case _, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
		case <-w.done:
			return
		}
	}
}

func (w *FileWatcher) addWatchRecursive(path string) error {
	return filepath.WalkDir(path, func(entryPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if w.shouldIgnore(entryPath) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			_ = w.watcher.Add(entryPath)
		}
		return nil
	})
}

func (w *FileWatcher) shouldIgnore(path string) bool {
	if w.managedDir != "" && strings.HasPrefix(path, w.managedDir) {
		return true
	}
	return false
}

var _ fs.FS = os.DirFS(".")
