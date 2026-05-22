package model

import (
	"testing"
)

func TestParseComponentID(t *testing.T) {
	tests := []struct {
		input    string
		wantType string
		wantName string
		wantErr  bool
	}{
		{input: "otlp", wantType: "otlp", wantName: ""},
		{input: "otlp/datadog", wantType: "otlp", wantName: "datadog"},
		{input: "foo/bar", wantType: "foo", wantName: "bar"},
		{input: "prometheus", wantType: "prometheus", wantName: ""},
		{input: "batch", wantType: "batch", wantName: ""},
		{input: "debug/1", wantType: "debug", wantName: "1"},
		{input: "count_connector/metrics", wantType: "count_connector", wantName: "metrics"},

		{input: "", wantErr: true},
		{input: "/x", wantErr: true},
		{input: "x/", wantErr: true},
		{input: "x/y/z", wantErr: true},
		{input: "//", wantErr: true},
		{input: "/", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseComponentID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseComponentID(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseComponentID(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("ParseComponentID(%q) type = %q, want %q", tt.input, got.Type, tt.wantType)
			}
			if got.Name != tt.wantName {
				t.Errorf("ParseComponentID(%q) name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}
