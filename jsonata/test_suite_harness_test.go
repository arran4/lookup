package jsonata

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHarnessSemantics(t *testing.T) {
	// A meta-test capturing harness behaviors for normal, expected-fail, and expected-error outcomes.

	// Test normal expected failure runCase error
	outcome := evaluateHarnessOutcome("test/1", nil, fmt.Errorf("evaluation error"), "", false, false, "")
	assert.True(t, outcome.Skipped)
	assert.Equal(t, "Expected failure (runCase error): evaluation error", outcome.Message)

	// Test new unexpected error
	outcome = evaluateHarnessOutcome("test/2", nil, fmt.Errorf("evaluation error"), "", true, false, "")
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: evaluation error", outcome.Message)

	// Test expected error code matching an error (success execution path)
	outcome = evaluateHarnessOutcome("test/3", nil, fmt.Errorf("some error"), "T0410", true, false, "")
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)
	assert.Equal(t, "pass-execution-error", outcome.Message)

	// Test expected error code but got nil error
	outcome = evaluateHarnessOutcome("test/4", nil, nil, "T0410", true, false, "")
	assert.True(t, outcome.Failed)
	assert.Equal(t, "Expected error T0410 but got nil", outcome.Message)

	// Test missing fixture (unsupported)
	outcome = evaluateHarnessOutcome("comments/case003", nil, fmt.Errorf("some unsupported error"), "", true, true, "Function definition not implemented")
	assert.True(t, outcome.Skipped)
	assert.Equal(t, "Unsupported test mechanism: Function definition not implemented (err: some unsupported error)", outcome.Message)

	// Check numbers matching logic simulation
	outNum := 10
	expectedNumStr := "10"
	var expectedNum interface{}
	dec := json.NewDecoder(strings.NewReader(expectedNumStr))
	dec.UseNumber()
	_ = dec.Decode(&expectedNum)

	n, _ := expectedNum.(json.Number)
	i, _ := n.Int64()
	assert.Equal(t, int64(10), i)
	assert.True(t, i == int64(outNum))
}
