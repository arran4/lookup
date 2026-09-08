package jsonata

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/arran4/lookup"
)

func TestHarnessSemantics(t *testing.T) {
	tests := []struct {
		name              string
		testID            string
		execErr           error
		scCode            string
		expectPass        bool
		isUnsupported     bool
		unsupportedReason string
		isUndefined       bool
		setupErr          error
		want              harnessOutcome
	}{
		{
			name:       "Setup error fails unconditionally",
			testID:     "test/setup",
			setupErr:   fmt.Errorf("failed to read dataset"),
			expectPass: false,
			want:       harnessOutcome{Failed: true, Message: "Setup failed: failed to read dataset"},
		},
		{
			name:       "Normal expected failure runCase error",
			testID:     "test/1",
			execErr:    fmt.Errorf("evaluation error"),
			expectPass: false,
			want:       harnessOutcome{Skipped: true, Message: "Expected failure (runCase error): evaluation error"},
		},
		{
			name:       "New unexpected error",
			testID:     "test/2",
			execErr:    fmt.Errorf("evaluation error"),
			expectPass: true,
			want:       harnessOutcome{Failed: true, Message: "runCase failed: evaluation error"},
		},
		{
			name:       "Expected error code matching an error (success execution path)",
			testID:     "test/3",
			execErr:    fmt.Errorf("Argument 1 of function"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Message: "pass-execution-error"},
		},
		{
			name:       "Expected error code missing incidental matching",
			testID:     "test/3_incidental",
			execErr:    fmt.Errorf("Some unrelated failure referencing T0410 but not strictly"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Failed: true, Message: "Expected error T0410 but got different error: Some unrelated failure referencing T0410 but not strictly"},
		},
		{
			name:       "Expected error code exact matching fallback",
			testID:     "test/3_exact",
			execErr:    fmt.Errorf("[T0410] General evaluation failure"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Message: "pass-execution-error"},
		},
		{
			name:       "Expected error code matching an error (expected failure unexpected pass)",
			testID:     "test/3b",
			execErr:    fmt.Errorf("Argument 1 of function"),
			scCode:     "T0410",
			expectPass: false,
			want:       harnessOutcome{Failed: true, Message: "Unexpected pass! Test test/3b is marked as expected failure but it produced expected error T0410"},
		},
		{
			name:       "WRONG expected error (wrong string)",
			testID:     "test/wrong",
			execErr:    fmt.Errorf("some random error"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Failed: true, Message: "Expected error T0410 but got different error: some random error"},
		},
		{
			name:       "WRONG expected error matching incidental prefix",
			testID:     "test/wrong_incidental",
			execErr:    fmt.Errorf("XT0410: unrelated error"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Failed: true, Message: "Expected error T0410 but got different error: XT0410: unrelated error"},
		},
		{
			name:       "Expected failure WRONG expected error matching incidental prefix",
			testID:     "test/wrong_fail_incidental",
			execErr:    fmt.Errorf("XT0410: unrelated error"),
			scCode:     "T0410",
			expectPass: false,
			want:       harnessOutcome{Skipped: true, Message: "Expected failure (wrong error): Expected T0410 but got: XT0410: unrelated error"},
		},
		{
			name:       "Exact bounded match works",
			testID:     "test/exact_match",
			execErr:    fmt.Errorf("T0410: exact match error"),
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Message: "pass-execution-error"},
		},
		{
			name:       "WRONG expected error on an expected failure",
			testID:     "test/wrong_fail",
			execErr:    fmt.Errorf("some random error"),
			scCode:     "T0410",
			expectPass: false,
			want:       harnessOutcome{Skipped: true, Message: "Expected failure (wrong error): Expected T0410 but got: some random error"},
		},
		{
			name:        "Missing-paths with undefined evaluation",
			testID:      "missing-paths/case000",
			execErr:     lookup.NewInvalidor("path", lookup.ErrNoSuchPath),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{},
		},
		{
			name:        "Missing-paths with undefined evaluation legacy fallback",
			testID:      "missing-paths/case001",
			execErr:     lookup.NewInvalidor("path", fmt.Errorf("element not found at simple path")),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{},
		},
		{
			name:        "Real evaluator failure disguised as undefined should FAIL",
			testID:      "missing-paths/case998",
			execErr:     lookup.NewInvalidor("path", fmt.Errorf("evaluation failure")),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{Failed: true, Message: "runCase failed: evaluation failure"},
		},
		{
			name:        "Parsing failure with undefined result",
			testID:      "missing-paths/case999",
			execErr:     fmt.Errorf("parse failed: unexpected token"),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{Failed: true, Message: "runCase failed: parse failed: unexpected token"},
		},
		{
			name:        "Plain error mimicking element not found text",
			testID:      "missing-paths/case998",
			execErr:     fmt.Errorf("some random element not found at simple path error"),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{Failed: true, Message: "runCase failed: some random element not found at simple path error"},
		},
		{
			name:        "Plain error mimicking invalid path",
			testID:      "missing-paths/case997",
			execErr:     fmt.Errorf("some invalid path error"),
			expectPass:  true,
			isUndefined: true,
			want:        harnessOutcome{Failed: true, Message: "runCase failed: some invalid path error"},
		},
		{
			name:       "Malformed dataset fixture load failure",
			testID:     "test/setup_malformed",
			setupErr:   fmt.Errorf("failed to unmarshal dataset: invalid character"),
			expectPass: false,
			want:       harnessOutcome{Failed: true, Message: "Setup failed: failed to unmarshal dataset: invalid character"},
		},
		{
			name:       "Expected error code but got nil error",
			testID:     "test/4",
			scCode:     "T0410",
			expectPass: true,
			want:       harnessOutcome{Failed: true, Message: "Expected error T0410 but got nil"},
		},
		{
			name:              "Missing fixture (unsupported)",
			testID:            "comments/case003",
			execErr:           fmt.Errorf("some unsupported error"),
			expectPass:        true,
			isUnsupported:     true,
			unsupportedReason: "Function definition not implemented",
			want:              harnessOutcome{Skipped: true, Message: "Unsupported test mechanism: Function definition not implemented (err: some unsupported error)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateHarnessOutcome(tt.testID, tt.execErr, tt.scCode, tt.expectPass, tt.isUnsupported, tt.unsupportedReason, tt.isUndefined, tt.setupErr)
			if got != tt.want {
				t.Errorf("evaluateHarnessOutcome() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunCaseDatasetErrors(t *testing.T) {
	tests := []struct {
		name    string
		dataset string
		wantErr string
	}{
		{
			name:    "missing dataset",
			dataset: "nonexistent",
			wantErr: "failed to read",
		},
		{
			name:    "malformed dataset",
			dataset: "malformed_test_dataset",
			wantErr: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runCase(suiteCase{Dataset: tt.dataset}, "1+1")
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tt.dataset)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error to contain %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestJSONNumberSanity(t *testing.T) {
	expectedNum, err := parseJSON("10")
	if err != nil {
		t.Fatalf("parseJSON failed: %v", err)
	}

	n, ok := expectedNum.(json.Number)
	if !ok {
		t.Fatalf("expected JSON number type, got %T", expectedNum)
	}

	i, err := n.Int64()
	if err != nil {
		t.Fatalf("failed to convert JSON number to int64: %v", err)
	}

	if i != int64(10) {
		t.Errorf("expected int64 10, got %d", i)
	}
}
