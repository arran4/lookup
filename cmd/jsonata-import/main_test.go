package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVerifyModeReturnsNoErrorOnUnchanged(t *testing.T) {
	err := run(true)
	if err != nil {
		t.Fatalf("verify mode failed: %v", err)
	}
}

func TestVerifyModeDetectsDrift(t *testing.T) {
	// Need an absolute/relative path from the test running dir (which is cmd/jsonata-import)
	originalFile := filepath.Join("..", "..", "jsonata", "testdata", "test-suite", "datasets", "dataset1.json")

	originalContent, err := os.ReadFile(originalFile)
	if err != nil {
		t.Skip("skipping test as dataset1.json does not exist")
	}

	corrupted := append([]byte(nil), originalContent...)
	corrupted = append(corrupted, 'x')

	if err := os.WriteFile(originalFile, corrupted, 0644); err != nil {
		t.Fatalf("failed to write corrupted file: %v", err)
	}

	err = run(true)

		if err := os.WriteFile(originalFile, originalContent, 0644); err != nil {
			t.Fatalf("failed to restore file: %v", err)
		}

	if err == nil {
		t.Fatalf("expected verify mode to fail on drift, but it passed")
	}
}
