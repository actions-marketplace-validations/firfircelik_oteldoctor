package model

import (
	"fmt"
	"strings"
)

type ComponentID struct {
	Type string
	Name string
}

func ParseComponentID(id string) (ComponentID, error) {
	if id == "" {
		return ComponentID{}, fmt.Errorf("empty component id")
	}

	parts := strings.Split(id, "/")

	switch len(parts) {
	case 1:
		if parts[0] == "" {
			return ComponentID{}, fmt.Errorf("invalid component id %q", id)
		}
		return ComponentID{Type: parts[0]}, nil
	case 2:
		if parts[0] == "" || parts[1] == "" {
			return ComponentID{}, fmt.Errorf("invalid component id %q", id)
		}
		return ComponentID{Type: parts[0], Name: parts[1]}, nil
	default:
		return ComponentID{}, fmt.Errorf("invalid component id %q", id)
	}
}
