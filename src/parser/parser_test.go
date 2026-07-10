package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveTrailingComma(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"[1,2,3,\n]", "[1,2,3]"},
		{"[1,2,3]", "[1,2,3]"},
		{"[]", "[]"},
		{"[1,2,3,  ]", "[1,2,3]"},
	}

	for _, c := range cases {
		got := removeTrailingComma(c.input)
		if got != c.expected {
			t.Errorf("removeTrailingComma(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

func TestParseSTCMFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.stcm")
	content := `[{"groupname":"g1","variablename":"v1","variabledata":[{"x":1000.0,"y":1.0},{"x":2000.0,"y":2.0}]}]`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	data, err := ParseSTCMFile(path)
	if err != nil {
		t.Fatalf("ParseSTCMFile failed: %v", err)
	}

	group, ok := data["g1"]
	if !ok {
		t.Fatalf("expected group g1")
	}

	variable, ok := group["v1"]
	if !ok {
		t.Fatalf("expected variable v1")
	}

	if len(variable.X) != 2 || variable.X[0] != 1000.0 || variable.X[1] != 2000.0 {
		t.Errorf("unexpected X values: %v", variable.X)
	}

	if len(variable.Y) != 2 || variable.Y[0] != 1.0 || variable.Y[1] != 2.0 {
		t.Errorf("unexpected Y values: %v", variable.Y)
	}
}
