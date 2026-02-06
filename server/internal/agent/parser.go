package agent

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	backtickPath = regexp.MustCompile("`([^`]+)`")
	quotePath    = regexp.MustCompile("\"([^\"]+)\"|'([^']+)'")
	filePathLike = regexp.MustCompile(`(?:^|\s)([\w./-]+\.[\w]+)`)
	checklist    = regexp.MustCompile(`^[\s✓*\-+]+([\w./-]+\.[\w]+)`)
)

func ExtractFilePaths(text string) []string {
	candidates := []string{}
	for _, match := range backtickPath.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			candidates = append(candidates, match[1])
		}
	}
	for _, match := range quotePath.FindAllStringSubmatch(text, -1) {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				candidates = append(candidates, match[i])
			}
		}
	}
	for _, match := range checklist.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			candidates = append(candidates, match[1])
		}
	}
	for _, match := range filePathLike.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			candidates = append(candidates, match[1])
		}
	}

	seen := map[string]struct{}{}
	paths := []string{}
	for _, raw := range candidates {
		path := strings.TrimSpace(raw)
		path = strings.Trim(path, ",;:")
		if path == "" {
			continue
		}
		if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
			continue
		}
		path = filepath.Clean(path)
		if path == "." || path == ".." {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths
}
