package fs

import (
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const defaultMaxReadBytes int64 = 64 * 1024

type ReadResult struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Encoding  string `json:"encoding"`
	Truncated bool   `json:"truncated"`
	Size      int64  `json:"size"`
	Ext       string `json:"ext"`
	Mime      string `json:"mime"`
	Root      string `json:"root,omitempty"`
}

// ResolvePath ensures the requested path stays within the root.
func ResolvePath(root, target string) (string, error) {
	if root == "" {
		return "", errors.New("root required")
	}
	if target == "" {
		return "", errors.New("path required")
	}
	root = filepath.Clean(root)
	if filepath.IsAbs(target) {
		clean := filepath.Clean(target)
		rel, err := filepath.Rel(root, clean)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", errors.New("path outside root")
		}
		return clean, nil
	}
	clean := filepath.Clean(filepath.Join(root, target))
	rel, err := filepath.Rel(root, clean)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", errors.New("path outside root")
	}
	return clean, nil
}

// ReadFile reads a file with a max byte limit and returns a preview if truncated.
func ReadFile(root, target string, maxBytes int64) (ReadResult, error) {
	if maxBytes <= 0 {
		maxBytes = defaultMaxReadBytes
	}
	resolved, err := ResolvePath(root, target)
	if err != nil {
		return ReadResult{}, err
	}
	relPath, err := filepath.Rel(root, resolved)
	if err != nil {
		return ReadResult{}, err
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return ReadResult{}, err
	}
	if info.IsDir() {
		return ReadResult{}, errors.New("path is a directory")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return ReadResult{}, err
	}
	defer file.Close()

	buf := make([]byte, maxBytes)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return ReadResult{}, err
	}
	buf = buf[:n]
	truncated := info.Size() > int64(n)
	encoding := "utf-8"
	content := string(buf)
	if !utf8.Valid(buf) {
		encoding = "binary"
		content = ""
	}
	ext := filepath.Ext(resolved)
	mimeType := mime.TypeByExtension(ext)
	return ReadResult{
		Path:      relPath,
		Name:      filepath.Base(resolved),
		Content:   content,
		Encoding:  encoding,
		Truncated: truncated,
		Size:      info.Size(),
		Ext:       ext,
		Mime:      mimeType,
		Root:      "",
	}, nil
}
