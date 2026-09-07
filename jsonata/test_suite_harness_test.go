package jsonata

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

func TestHarnessSemantics(t *testing.T) {
	// A meta-test capturing harness behaviors for normal, expected-fail, and expected-error outcomes.

	// Test setup error fails unconditionally even on expected-fail
	outcome := evaluateHarnessOutcome("test/setup", nil, nil, "", false, false, "", false, fmt.Errorf("failed to read dataset"))
	assert.True(t, outcome.Failed)
	assert.Equal(t, "Setup failed: failed to read dataset", outcome.Message)

	// Test normal expected failure runCase error
	outcome = evaluateHarnessOutcome("test/1", nil, fmt.Errorf("evaluation error"), "", false, false, "", false, nil)
	assert.True(t, outcome.Skipped)
	assert.Equal(t, "Expected failure (runCase error): evaluation error", outcome.Message)

	// Test new unexpected error
	outcome = evaluateHarnessOutcome("test/2", nil, fmt.Errorf("evaluation error"), "", true, false, "", false, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: evaluation error", outcome.Message)

	// Test expected error code matching an error (success execution path)
	outcome = evaluateHarnessOutcome("test/3", nil, fmt.Errorf("some error"), "T0410", true, false, "", false, nil)
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)
	assert.Equal(t, "pass-execution-error", outcome.Message)

	// Test expected error code matching an error (expected failure unexpected pass)
	outcome = evaluateHarnessOutcome("test/3b", nil, fmt.Errorf("some error"), "T0410", false, false, "", false, nil)
	assert.True(t, outcome.Failed)
	assert.True(t, strings.Contains(outcome.Message, "Unexpected pass!"))

	// Test missing-paths with undefined evaluation
	invalidPathErr := lookup.NewInvalidor("path", fmt.Errorf("invalid path"))
	outcome = evaluateHarnessOutcome("missing-paths/case000", nil, invalidPathErr, "", true, false, "", true, nil)
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)

	// Test parsing failure with undefined result
	parseErr := fmt.Errorf("parse failed: unexpected token")
	outcome = evaluateHarnessOutcome("missing-paths/case999", nil, parseErr, "", true, false, "", true, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: parse failed: unexpected token", outcome.Message)

	// Test expected error code but got nil error
	outcome = evaluateHarnessOutcome("test/4", nil, nil, "T0410", true, false, "", false, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "Expected error T0410 but got nil", outcome.Message)

	// Test missing fixture (unsupported)
	outcome = evaluateHarnessOutcome("comments/case003", nil, fmt.Errorf("some unsupported error"), "", true, true, "Function definition not implemented", false, nil)
	assert.True(t, outcome.Skipped)
	assert.Equal(t, "Unsupported test mechanism: Function definition not implemented (err: some unsupported error)", outcome.Message)

	// Test undefined evaluation
	// Should ignore missing path error
	invalidPathErr2 := lookup.NewInvalidor("path", fmt.Errorf("invalid path"))
	outcome = evaluateHarnessOutcome("test/5", nil, invalidPathErr2, "", true, false, "", true, nil)
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)

	// Should NOT ignore parse error
	outcome = evaluateHarnessOutcome("test/6", nil, fmt.Errorf("parse error"), "", true, false, "", true, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: parse error", outcome.Message)

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
