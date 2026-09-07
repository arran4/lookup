package jsonata

import (
	"os"
	"strings"
	"testing"
)

func TestCompatibilityMatrixUpToDate(t *testing.T) {
	matrix, err := generateCompatibilityMatrix(testData)
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

	// Normalize Windows line endings just in case
	currentMatrix = strings.ReplaceAll(currentMatrix, "\r\n", "\n")
	expectedMatrix = strings.ReplaceAll(expectedMatrix, "\r\n", "\n")

	if currentMatrix != expectedMatrix {
		t.Errorf("README.md compatibility matrix is out of date. Please run 'UPDATE_MATRIX=1 go test ./jsonata -run TestUpdateCompatibilityMatrix' to update it.\n\nExpected:\n%s\n\nActual:\n%s", expectedMatrix, currentMatrix)
	}
}
