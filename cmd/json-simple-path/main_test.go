package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/lookup/cmd/internal/cli"
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
		err := cli.Run("json-simple-path", c.args, in, &out, io.Discard, "JSON")
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		got := strings.TrimSpace(out.String())
		if got != c.want {
			t.Errorf("%s: want %q got %q", c.name, c.want, got)
		}
	}
}

func TestStrictExamples(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "doc.json")
	if err := os.WriteFile(fname, []byte(exampleJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Should fail with strict mode due to missing path
	var out bytes.Buffer
	err := cli.Run("json-simple-path", []string{"-strict", "-f", fname, ".spec.missing"}, nil, &out, io.Discard, "JSON")
	if err == nil {
		t.Fatalf("expected error in strict mode for missing path, got nil")
	}

	// Should fail with strict mode due to syntax error
	err = cli.Run("json-simple-path", []string{"-strict", "-f", fname, ".spec[0"}, nil, &out, io.Discard, "JSON")
	if err == nil {
		t.Fatalf("expected error in strict mode for malformed syntax, got nil")
	}
}

func TestStrictMultipleQueries(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "doc.json")
	if err := os.WriteFile(fname, []byte(`{"A": 1, "B": 2}`), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := cli.Run("json-simple-path", []string{"-strict", "-f", fname, "-e", "A", "-e", "C"}, nil, &out, io.Discard, "JSON")
	if err == nil {
		t.Fatalf("expected error for missing second query in strict mode")
	}
}

func TestStrictMultipleDocuments(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "doc.json")
	if err := os.WriteFile(fname, []byte(`{"A": 1}
{"B": 2}`), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := cli.Run("json-simple-path", []string{"-strict", "-f", fname, "A"}, nil, &out, io.Discard, "JSON")
	if err == nil {
		t.Fatalf("expected error for missing path in second document in strict mode")
	}
}

func TestStrictConflictingFlags(t *testing.T) {
	tmp := t.TempDir()
	fname := filepath.Join(tmp, "doc.json")
	if err := os.WriteFile(fname, []byte(`{"A": 1}`), 0644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := cli.Run("json-simple-path", []string{"-strict", "-f", fname, "-json", "-yaml", "A"}, nil, &out, io.Discard, "JSON")
	if err == nil || !strings.Contains(err.Error(), "conflicting output flags") {
		t.Fatalf("expected conflicting flags error, got: %v", err)
	}
}
