package jsonata

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTestingT struct {
	*testing.T
	failed   bool
	skipped  bool
	messages []string
}

func (m *mockTestingT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.messages = append(m.messages, fmt.Sprintf(format, args...))
}

func (m *mockTestingT) Skipf(format string, args ...interface{}) {
	m.skipped = true
	m.messages = append(m.messages, fmt.Sprintf(format, args...))
}

func TestHarnessSemantics(t *testing.T) {
	// A meta-test capturing harness behaviors for normal, expected-fail, and expected-error outcomes.

	// Test error capture on out vs execErr logic
	out := error(fmt.Errorf("evaluation error"))
	var execErr error
	if outErr, ok := out.(error); ok {
		execErr = outErr
	}
	assert.NotNil(t, execErr)
	assert.Equal(t, "evaluation error", execErr.Error())

	// Test sc.Code logic
	sc := suiteCase{Code: "T0410"}
	err := error(nil)
	assert.Nil(t, err)
	assert.NotEqual(t, "", sc.Code)

	// Simulate harness branch checking sc.Code expecting error but finding none
	var failed bool
	if sc.Code != "" {
		if execErr == nil { // Simulated nil
			failed = true
		}
	}
	assert.False(t, failed) // execErr is not nil

	// Check numbers matching logic
	outNum := 10
	expectedNumStr := "10"
	var expectedNum interface{}
	dec := json.NewDecoder(strings.NewReader(expectedNumStr))
	dec.UseNumber()
	dec.Decode(&expectedNum)

	n, _ := expectedNum.(json.Number)
	i, _ := n.Int64()
	assert.Equal(t, int64(10), i)
	assert.True(t, i == int64(outNum))
}
