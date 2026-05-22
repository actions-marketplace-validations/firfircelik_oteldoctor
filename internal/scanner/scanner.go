package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Scanner struct {
	Recursive bool
}

func New() *Scanner {
	return &Scanner{Recursive: true}
}

func (s *Scanner) Scan(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if path == root {
				return nil
			}
			if !s.Recursive {
				return filepath.SkipDir
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

var collectorKeys = []string{"receivers", "exporters", "service", "processors"}

func IsCollectorConfig(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return false, nil
	}

	if len(node.Content) == 0 {
		return false, nil
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return false, nil
	}

	found := 0
	for i := 0; i < len(root.Content)-1; i += 2 {
		key := root.Content[i].Value
		for _, ck := range collectorKeys {
			if key == ck {
				found++
				break
			}
		}
	}

	return found >= 2, nil
}

func FilterByGlob(files []string, patterns []string, exclude bool) []string {
	if len(patterns) == 0 {
		return files
	}

	var result []string
	for _, f := range files {
		matched := false
		for _, pat := range patterns {
			m, err := filepath.Match(pat, filepath.Base(f))
			if err != nil {
				continue
			}
			if m {
				matched = true
				break
			}
		}
		if exclude {
			if !matched {
				result = append(result, f)
			}
		} else {
			if matched {
				result = append(result, f)
			}
		}
	}
	return result
}

func FilterCollectorConfigs(files []string) ([]string, error) {
	var result []string
	for _, f := range files {
		ok, err := IsCollectorConfig(f)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, f)
		}
	}
	return result, nil
}
