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

	// Test expected error code missing incidental matching
	// e.g. An unrelated error string simply contains "T0410" without the strict matching format
	outcome = evaluateHarnessOutcome("test/3_incidental", nil, fmt.Errorf("Some unrelated failure referencing T0410 but not strictly"), "T0410", true, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/3_incidental to fail due to mismatch")
	assert.Equal(t, "Expected error T0410 but got different error: Some unrelated failure referencing T0410 but not strictly", outcome.Message)

	// Test expected error code exact matching fallback "[T0410]"
	outcome = evaluateHarnessOutcome("test/3_exact", nil, fmt.Errorf("[T0410] General evaluation failure"), "T0410", true, false, "", false, nil)
	assert.False(t, outcome.Failed, "Expected test/3_exact to not fail")
	assert.False(t, outcome.Skipped, "Expected test/3_exact to not skip")
	assert.Equal(t, "pass-execution-error", outcome.Message)

	// Test expected error code matching an error (expected failure unexpected pass)
	outcome = evaluateHarnessOutcome("test/3b", nil, fmt.Errorf("Argument 1 of function"), "T0410", false, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/3b to fail")
	assert.True(t, strings.Contains(outcome.Message, "Unexpected pass!"), "Expected message to contain Unexpected pass! Message: %s", outcome.Message)

	// Test WRONG expected error (wrong string)
	outcome = evaluateHarnessOutcome("test/wrong", nil, fmt.Errorf("some random error"), "T0410", true, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/wrong to fail")
	assert.Equal(t, "Expected error T0410 but got different error: some random error", outcome.Message)

	// Test WRONG expected error matching incidental prefix
	outcome = evaluateHarnessOutcome("test/wrong_incidental", nil, fmt.Errorf("XT0410: unrelated error"), "T0410", true, false, "", false, nil)
	assert.True(t, outcome.Failed, "Expected test/wrong_incidental to fail")
	assert.Equal(t, "Expected error T0410 but got different error: XT0410: unrelated error", outcome.Message)

	// Test expected failure WRONG expected error matching incidental prefix
	outcome = evaluateHarnessOutcome("test/wrong_fail_incidental", nil, fmt.Errorf("XT0410: unrelated error"), "T0410", false, false, "", false, nil)
	assert.True(t, outcome.Skipped, "Expected test/wrong_fail_incidental to skip")
	assert.Equal(t, "Expected failure (wrong error): Expected T0410 but got: XT0410: unrelated error", outcome.Message)

	// Test exact bounded match works
	outcome = evaluateHarnessOutcome("test/exact_match", nil, fmt.Errorf("T0410: exact match error"), "T0410", true, false, "", false, nil)
	assert.False(t, outcome.Failed, "Expected test/exact_match to not fail")
	assert.False(t, outcome.Skipped, "Expected test/exact_match to not skip")

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

	// Test plain error mimicking element not found text
	plainErr := fmt.Errorf("some random element not found at simple path error")
	outcome = evaluateHarnessOutcome("missing-paths/case998", nil, plainErr, "", true, false, "", true, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: some random element not found at simple path error", outcome.Message)

	// Test plain error mimicking invalid path
	plainErr2 := fmt.Errorf("some invalid path error")
	outcome = evaluateHarnessOutcome("missing-paths/case997", nil, plainErr2, "", true, false, "", true, nil)
	assert.True(t, outcome.Failed)
	assert.Equal(t, "runCase failed: some invalid path error", outcome.Message)

	// Test malformed dataset fixture load failure
	outcome = evaluateHarnessOutcome("test/setup_malformed", nil, nil, "", false, false, "", false, fmt.Errorf("failed to unmarshal dataset: invalid character"))
	assert.True(t, outcome.Failed)
	assert.Equal(t, "Setup failed: failed to unmarshal dataset: invalid character", outcome.Message)

	// Test actual runCase integration coverage for missing dataset
	_, err := runCase(suiteCase{Dataset: "nonexistent"}, "1+1")
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to read"))

	// Test actual runCase integration coverage for malformed dataset (invalid JSON)
	_, err = runCase(suiteCase{Dataset: "malformed_test_dataset"}, "1+1")
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to unmarshal"))

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
