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
	outcome = evaluateHarnessOutcome("test/3", nil, fmt.Errorf("Argument 1 of function"), "T0410", true, false, "", false, nil)
	assert.False(t, outcome.Failed, "Expected test/3 to not fail")
	assert.False(t, outcome.Skipped, "Expected test/3 to not skip")
	assert.Equal(t, "pass-execution-error", outcome.Message)

	// Test expected error code matching an error (expected failure unexpected pass)
	outcome = evaluateHarnessOutcome("test/3b", nil, fmt.Errorf("parse failed"), "T0410", false, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/3b to fail")
	assert.True(t, strings.Contains(outcome.Message, "Unexpected pass!"), "Expected message to contain Unexpected pass! Message: %s", outcome.Message)

	// Test WRONG expected error (wrong string)
	outcome = evaluateHarnessOutcome("test/wrong", nil, fmt.Errorf("some random error"), "T0410", true, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/wrong to fail")
	assert.Equal(t, "Expected error T0410 but got different error: some random error", outcome.Message)

	// Test WRONG expected error on an expected failure (should skip not unexpectedly pass)
	outcome = evaluateHarnessOutcome("test/wrong_fail", nil, fmt.Errorf("some random error"), "T0410", false, false, "", false, nil)
	assert.True(t, outcome.Skipped, "Expected test/wrong_fail to skip")
	assert.Equal(t, "Expected failure (wrong error): Expected T0410 but got: some random error", outcome.Message)

	// Test missing-paths with undefined evaluation
	invalidPathErr := lookup.NewInvalidor("path", lookup.ErrNoSuchPath)
	outcome = evaluateHarnessOutcome("missing-paths/case000", nil, invalidPathErr, "", true, false, "", true, nil)
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)

	invalidPathErr2 := lookup.NewInvalidor("path", fmt.Errorf("element not found at simple path"))
	outcome = evaluateHarnessOutcome("missing-paths/case001", nil, invalidPathErr2, "", true, false, "", true, nil)
	assert.False(t, outcome.Failed)
	assert.False(t, outcome.Skipped)

	// Test real evaluator failure disguised as undefined should FAIL
	realEvalErr := lookup.NewInvalidor("path", fmt.Errorf("evaluation failure"))
	outcome = evaluateHarnessOutcome("missing-paths/case998", nil, realEvalErr, "", true, false, "", true, nil)
	assert.True(t, outcome.Failed, "Expected case998 to fail")
	assert.True(t, strings.Contains(outcome.Message, "runCase failed: evaluation failure"), "Expected string in: %s", outcome.Message)

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
