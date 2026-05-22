package extractor

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

var configMapKeys = []string{
	"collector.yaml",
	"collector.yml",
	"relay",
	"otel-collector-config",
	"config.yaml",
}

type EmbeddedConfig struct {
	Path          string
	SourceFile    string
	ConfigMapName string
	DataKey       string
}

func IsConfigMap(path string) (bool, error) {
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

	var apiVersion, kind string
	for i := 0; i < len(root.Content)-1; i += 2 {
		switch root.Content[i].Value {
		case "apiVersion":
			apiVersion = root.Content[i+1].Value
		case "kind":
			kind = root.Content[i+1].Value
		}
	}

	if apiVersion != "v1" || kind != "ConfigMap" {
		return false, nil
	}

	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value != "data" {
			continue
		}
		dataNode := root.Content[i+1]
		if dataNode.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j < len(dataNode.Content)-1; j += 2 {
			key := dataNode.Content[j].Value
			for _, ck := range configMapKeys {
				if key == ck {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func Extract(path string) ([]EmbeddedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	if len(node.Content) == 0 {
		return nil, nil
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil
	}

	configMapName := ""
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value != "metadata" {
			continue
		}
		metaNode := root.Content[i+1]
		if metaNode.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j < len(metaNode.Content)-1; j += 2 {
			if metaNode.Content[j].Value == "name" {
				configMapName = metaNode.Content[j+1].Value
			}
		}
	}

	var result []EmbeddedConfig

	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value != "data" {
			continue
		}
		dataNode := root.Content[i+1]
		if dataNode.Kind != yaml.MappingNode {
			continue
		}
		for j := 0; j < len(dataNode.Content)-1; j += 2 {
			key := dataNode.Content[j].Value
			value := dataNode.Content[j+1].Value

			matched := false
			for _, ck := range configMapKeys {
				if key == ck {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}

			tmpFile, err := os.CreateTemp("", "oteldoctor-extracted-*.yaml")
			if err != nil {
				for _, r := range result {
					os.Remove(r.Path)
				}
				return nil, fmt.Errorf("creating temp file: %w", err)
			}

			if _, err := tmpFile.WriteString(value); err != nil {
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				return nil, fmt.Errorf("writing temp file: %w", err)
			}
			tmpFile.Close()

			result = append(result, EmbeddedConfig{
				Path:          tmpFile.Name(),
				SourceFile:    path,
				ConfigMapName: configMapName,
				DataKey:       key,
			})
		}
	}

	return result, nil
}

func Cleanup(configs []EmbeddedConfig) {
	for _, c := range configs {
		os.Remove(c.Path)
	}
}

func sourcePath(file, cmName, key string) string {
	name := filepath.Base(file)
	if cmName != "" {
		return fmt.Sprintf("%s/%s::%s", name, cmName, key)
	}
	return fmt.Sprintf("%s::%s", name, key)
}

func SourcePathDisplay(sourceFile, cmName, key string) string {
	return sourcePath(sourceFile, cmName, key)
}
