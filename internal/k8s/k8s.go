package k8s

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type ResourceValue struct {
	Raw  string
	MiB  int64
	CPU  string
}

type Workload struct {
	Kind       string
	Name       string
	Containers []Container
}

type Container struct {
	Name    string
	Image   string
	Env     map[string]string
	Limits  ResourceValue
	Probes  ProbeInfo
	Mounts  []VolumeMount
}

type ProbeInfo struct {
	HasReadiness bool
	HasLiveness  bool
}

type VolumeMount struct {
	Name string
	Path string
}

type ServiceInfo struct {
	Name string
	Type string
}

func ParseWorkload(path string) (*Workload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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

	kind := ""
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "kind" {
			kind = root.Content[i+1].Value
		}
	}

	if kind != "Deployment" && kind != "DaemonSet" && kind != "StatefulSet" {
		return nil, nil
	}

	w := &Workload{Kind: kind}

	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "metadata" {
			w.Name = getNestedString(root.Content[i+1], "name")
		}
		if root.Content[i].Value == "spec" {
			fillContainers(&w.Containers, root.Content[i+1], kind)
		}
	}

	return w, nil
}

func fillContainers(containers *[]Container, spec *yaml.Node, kind string) {
	if spec.Kind != yaml.MappingNode {
		return
	}

	var podSpec *yaml.Node
	if kind == "Deployment" || kind == "StatefulSet" {
		podSpec = findValue(findValue(spec, "template"), "spec")
	} else {
		podSpec = findValue(spec, "template")
		if podSpec != nil {
			podSpec = findValue(podSpec, "spec")
		}
	}

	if podSpec == nil {
		return
	}

	containerList := findValue(podSpec, "containers")
	if containerList == nil || containerList.Kind != yaml.SequenceNode {
		return
	}

	for _, item := range containerList.Content {
		if item.Kind != yaml.MappingNode {
			continue
		}

		c := Container{
			Name: getNestedString(item, "name"),
			Image: getNestedString(item, "image"),
			Env: map[string]string{},
		}

		resources := findValue(item, "resources")
		if resources != nil {
			limits := findValue(resources, "limits")
			if limits != nil {
				mem := getNestedString(limits, "memory")
				c.Limits = ResourceValue{Raw: mem}
				if mi, ok := parseMemoryMiB(mem); ok {
					c.Limits.MiB = mi
				}
				c.Limits.CPU = getNestedString(limits, "cpu")
			}
		}

		envList := findValue(item, "env")
		if envList != nil && envList.Kind == yaml.SequenceNode {
			for _, env := range envList.Content {
				name := getNestedString(env, "name")
				value := getNestedString(env, "value")
				if name != "" {
					c.Env[name] = value
				}
			}
		}

		readinessProbe := findValue(item, "readinessProbe")
		c.Probes.HasReadiness = readinessProbe != nil

		livenessProbe := findValue(item, "livenessProbe")
		c.Probes.HasLiveness = livenessProbe != nil

		vmList := findValue(item, "volumeMounts")
		if vmList != nil && vmList.Kind == yaml.SequenceNode {
			for _, vm := range vmList.Content {
				c.Mounts = append(c.Mounts, VolumeMount{
					Name: getNestedString(vm, "name"),
					Path: getNestedString(vm, "mountPath"),
				})
			}
		}

		*containers = append(*containers, c)
	}
}

func ParseService(path string) (*ServiceInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
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

	kind := ""
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "kind" {
			kind = root.Content[i+1].Value
		}
	}

	if kind != "Service" {
		return nil, nil
	}

	svc := &ServiceInfo{}
	for i := 0; i < len(root.Content)-1; i += 2 {
		if root.Content[i].Value == "metadata" {
			svc.Name = getNestedString(root.Content[i+1], "name")
		}
		if root.Content[i].Value == "spec" {
			svc.Type = getNestedString(root.Content[i+1], "type")
		}
	}

	if svc.Type == "" {
		svc.Type = "ClusterIP"
	}

	return svc, nil
}

func findValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func getNestedString(node *yaml.Node, key string) string {
	if node == nil || node.Kind != yaml.MappingNode {
		return ""
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1].Value
		}
	}
	return ""
}

func parseMemoryMiB(s string) (int64, bool) {
	if s == "" {
		return 0, false
	}

	s = strings.TrimSpace(s)
	var value float64
	var unit string

	fmt.Sscanf(s, "%f%s", &value, &unit)
	if value <= 0 {
		return 0, false
	}

	unit = strings.TrimSpace(unit)
	switch strings.ToUpper(unit) {
	case "MI":
		return int64(value), true
	case "GI":
		return int64(value * 1024), true
	case "M":
		return int64(value * 1000 * 1000 / (1024 * 1024)), true
	case "G":
		return int64(value * 1000 * 1000 * 1000 / (1024 * 1024)), true
	case "K":
		return int64(value * 1000 / (1024 * 1024)), true
	case "KI":
		return int64(value / 1024), true
	default:
		return 0, false
	}
}
