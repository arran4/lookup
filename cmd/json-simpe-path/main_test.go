package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const exampleJSON = `{"name":"foo","spec":{"replicas":3},"metadata":{"name":"prod-service"}}`

func TestExamples(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "doc.json")
	if err := os.WriteFile(fname, []byte(exampleJSON), 0644); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		args  []string
		stdin string
		want  string
	}{
		{"field", []string{"-f", fname, ".spec.replicas"}, "", "3"},
		{"raw", []string{"-raw", ".spec.replicas"}, exampleJSON, "3"},
		{"grep", []string{"-f", fname, "-grep", "^prod", "-raw", ".metadata.name"}, "", "prod-service"},
		{"count", []string{"-f", fname, "-count", ".metadata.name"}, "", "1"},
	}

	for _, c := range cases {
		var in io.Reader = bytes.NewBufferString(c.stdin)
		var out bytes.Buffer
		err := run(c.args, in, &out, io.Discard)
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := strings.TrimSpace(out.String())
		if got != c.want {
			t.Errorf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}
