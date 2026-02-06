package context

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func LoadViewExamples(dir string) ([]ViewExample, error) {
	if dir == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(configDir, "mindfs", "view-examples")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ViewExample{}, nil
		}
		return nil, err
	}
	items := []ViewExample{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		payload, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}
		var example ViewExample
		if err := json.Unmarshal(payload, &example); err != nil {
			continue
		}
		if example.Prompt == "" && example.Description == "" {
			continue
		}
		items = append(items, example)
	}
	return items, nil
}

func SelectExamples(examples []ViewExample, query string, limit int) []ViewExample {
	if limit <= 0 {
		limit = 3
	}
	if len(examples) == 0 {
		return []ViewExample{}
	}
	query = strings.ToLower(query)
	sort.Slice(examples, func(i, j int) bool {
		return scoreExample(examples[i], query) > scoreExample(examples[j], query)
	})
	if len(examples) > limit {
		examples = examples[:limit]
	}
	return examples
}

func scoreExample(example ViewExample, query string) int {
	score := 0
	if query == "" {
		return score
	}
	if strings.Contains(strings.ToLower(example.Description), query) {
		score += 2
	}
	if strings.Contains(strings.ToLower(example.Prompt), query) {
		score += 1
	}
	return score
}
