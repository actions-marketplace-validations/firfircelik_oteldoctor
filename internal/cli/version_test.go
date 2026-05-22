package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	prevVersion := version
	version = "1.2.3-test"
	defer func() { version = prevVersion }()

	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "1.2.3-test") {
		t.Errorf("expected version '1.2.3-test' in output, got: %s", out)
	}
}
