package jsonata

import (
	"os"
	"strings"
	"testing"
)

func TestUpdateCompatibilityMatrix(t *testing.T) {
	if os.Getenv("UPDATE_MATRIX") != "1" {
		t.Skip("Set UPDATE_MATRIX=1 to update README.md")
	}

	matrix, err := GenerateCompatibilityMatrix(testData)
	if err != nil {
		t.Fatalf("Failed to generate compatibility matrix: %v", err)
	}

	readmeData, err := os.ReadFile("../README.md")
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}

	readme := string(readmeData)

	startToken := "| Feature Group | Passed | Failed | Unsupported |"
	startIndex := strings.Index(readme, startToken)
	if startIndex == -1 {
		t.Fatalf("Could not find start of compatibility matrix in README.md")
	}

	endIndex := strings.Index(readme[startIndex:], "\n\n")
	if endIndex == -1 {
		endIndex = len(readme)
	} else {
		endIndex += startIndex
	}

	newReadme := readme[:startIndex] + strings.TrimSpace(matrix) + readme[endIndex:]

	err = os.WriteFile("../README.md", []byte(newReadme), 0644)
	if err != nil {
		t.Fatalf("Failed to write README.md: %v", err)
	}
}
