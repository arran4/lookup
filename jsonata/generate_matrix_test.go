package jsonata

import (
	"os"
	"strings"
	"testing"
)

func TestCompatibilityMatrixUpToDate(t *testing.T) {
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

	currentMatrix := strings.TrimSpace(readme[startIndex:endIndex])
	expectedMatrix := strings.TrimSpace(matrix)

	if currentMatrix != expectedMatrix {
		t.Errorf("README.md compatibility matrix is out of date. Please run 'go generate ./jsonata' to update it.\n\nExpected:\n%s\n\nActual:\n%s", expectedMatrix, currentMatrix)
	}
}
