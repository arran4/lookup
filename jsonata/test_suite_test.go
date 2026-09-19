package jsonata

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
)

//go:embed testdata
var testData embed.FS

func loadDataset(name string) (interface{}, error) {
	filename := path.Join("testdata", "test-suite", "datasets", name+".json")
	data, err := fs.ReadFile(testData, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", filename, err)
	}
	return v, nil
}

func parseJSON(data string) (interface{}, error) {
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json: %w", err)
	}
	return v, nil
}

func runCase(c suiteCase, expr string, undefinedInput bool) (interface{}, error) {
	var data interface{}
	var err error
	if undefinedInput {
		data = Undefined{}
	} else if c.Data != nil {
		data = c.Data
	} else if c.Dataset != "" {
		data, err = loadDataset(c.Dataset)
		if err != nil {
			return nil, err
		}
	} else {
		// fallback for cases with data: null or dataset: null (which unmarshals to interface{}(nil))
		// but since we checked undefinedInput explicitly, this really means explicit null.
		data = nil
	}

	ast, err := Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse failed: %w", err)
	}
	q := Compile(ast)
	root := lookup.Reflect(data)
	ctx := &evaluator.Context{
		Functions: GetStandardFunctions(),
	}
	res := q.Run(lookup.NewScopeWithContext(nil, root, ctx))
	if res == nil {
		return nil, nil
	}
	// Return the Pathor directly so the caller can check if it's an error/Invalidor
	return res, nil
}

func TestGroups(t *testing.T) {
	entries, err := testData.ReadDir("testdata/test-suite/groups")
	if err != nil {
		t.Fatalf("failed to list groups: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txtar") {
			continue
		}
		groupName := strings.TrimSuffix(entry.Name(), ".txtar")
		t.Run(groupName, func(t *testing.T) {
			runTxtarGroup(t, path.Join("testdata/test-suite/groups", entry.Name()), groupName)
		})
	}
}

func runTxtarGroup(t *testing.T, filename string, groupName string) {
	data, err := fs.ReadFile(testData, filename)
	if err != nil {
		t.Fatalf("failed to read txtar file %s: %v", filename, err)
	}

	cases, err := parseTxtar(data)
	if err != nil {
		t.Fatalf("failed to parse txtar file %s: %v", filename, err)
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			testID := groupName + "/" + c.Name
			if reason, ok := unsupportedTests[testID]; ok {
				t.Skipf("Unsupported: %s", reason)
				return
			}

			expectPass := !expectedFailures[testID]

			var sc suiteCase
			if err := json.Unmarshal([]byte(c.Input), &sc); err != nil {
				t.Fatalf("invalid suite case config: %v", err)
			}

			var rawMap map[string]json.RawMessage
			if err := json.Unmarshal([]byte(c.Input), &rawMap); err != nil {
				t.Fatalf("failed to unmarshal raw map: %v", err)
			}

			_, hasData := rawMap["data"]
			datasetMsg, hasDataset := rawMap["dataset"]

			undefinedInput := false
			if !hasData {
				if !hasDataset {
					undefinedInput = true
				} else if string(datasetMsg) == "null" || string(datasetMsg) == "\"\"" {
					// The upstream test suite uses "dataset": "" to mean undefined input in some contexts.
					undefinedInput = true
				}
			}

			// Capture panic to treat as failure instead of crash
			defer func() {
				if r := recover(); r != nil {
					if expectPass {
						t.Fatalf("panic: %v", r)
					} else {
						t.Skipf("Expected failure (panic): %v", r)
					}
				}
			}()

			res, err := runCase(sc, c.Expr, undefinedInput)
			var out interface{}

			// Setup errors, e.g. failing to read dataset
			var setupErr error
			if err != nil && (strings.Contains(err.Error(), "failed to read") || strings.Contains(err.Error(), "failed to unmarshal")) {
				setupErr = err
			}

			// Some tests are meant to fail parsing or execution. In those cases sc.Code is set.
			// Alternatively if out is an error/Invalidor it should be treated as an error.
			var execErr error
			if err != nil && setupErr == nil {
				execErr = err
			} else if resErr, ok := res.(error); ok {
				execErr = resErr
			}

			// We delegate to a pure outcome evaluation function to make logic testable
			isUnsupported := false
			unsupportedReason := ""
			if reason, ok := unsupportedTests[testID]; ok {
				isUnsupported = true
				unsupportedReason = reason
			}

			outcome := evaluateHarnessOutcome(testID, execErr, sc.Code, expectPass, isUnsupported, unsupportedReason, sc.Undefined, setupErr)

			if outcome.Failed {
				t.Fatalf("%s", outcome.Message)
			}
			if outcome.Skipped {
				t.Skipf("%s", outcome.Message)
			}
			if outcome.Message != "" {
				// This handles return paths logically where it was just an execution level failure/skip
				return
			}

			var expected interface{}
			if c.Expected == "null" && sc.Undefined {
				expected = nil
			} else {
				var err error
				expected, err = parseJSON(c.Expected)
				if err != nil {
					t.Fatalf("failed to parse expected json: %v", err)
				}
			}

			// Apply final materialization boundary
			var actualUndefined bool
			out, actualUndefined = materializeHarnessValue(res) // Pass 'res' which is the raw interface/pathor

			match := false
			if sc.Undefined {
				match = actualUndefined
			} else if actualUndefined {
				match = false
			} else {
				match = jsonataValuesEqual(expected, out)
			}

			if expectPass {
				if !match {
					expectedStr := "null"
					if sc.Undefined {
						expectedStr = "undefined"
					} else if expected != nil {
						expectedStr = fmt.Sprintf("%v (%T)", expected, expected)
					}
					gotStr := "null"
					if actualUndefined {
						gotStr = "undefined"
					} else if out != nil {
						gotStr = fmt.Sprintf("%v (%T)", out, out)
					}
					t.Fatalf("\nTEST_ID: %s\nEXPR: %s\nEXPECTED_UNDEF: %v\nEXPECTED: %s\nFINAL_VALUE: %s\n---", testID, c.Expr, sc.Undefined, expectedStr, gotStr)
				}
			} else {
				if match {
					t.Fatalf("Unexpected pass! Test %s is marked as expected failure but it passed. Remove it from expectedFailures.", testID)
				} else {
					t.Skipf("Expected failure. Expected: %v, Got: %v", expected, out)
				}
			}
		})
	}
}

type harnessOutcome struct {
	Failed  bool
	Skipped bool
	Message string
}

func evaluateHarnessOutcome(testID string, execErr error, scCode string, expectPass bool, isUnsupported bool, unsupportedReason string, isUndefined bool, setupErr error) harnessOutcome {
	if setupErr != nil {
		return harnessOutcome{Failed: true, Message: fmt.Sprintf("Setup failed: %v", setupErr)}
	}

	if isUnsupported {
		return harnessOutcome{Skipped: true, Message: fmt.Sprintf("Unsupported test mechanism: %v (err: %v)", unsupportedReason, execErr)}
	}

	if scCode != "" {
		if execErr == nil {
			if expectPass {
				return harnessOutcome{Failed: true, Message: fmt.Sprintf("Expected error %s but got nil", scCode)}
			} else {
				return harnessOutcome{Skipped: true, Message: fmt.Sprintf("Expected failure: Expected error %s but got nil", scCode)}
			}
		}

		// The test expects an error code.
		// Since we don't have all jsonata specific error codes built, we provide explicit
		// support mappings or check if the exact error code is directly matched in our error strings.
		matchedError := false
		errStr := execErr.Error()

		// If JSONata error codes aren't cleanly mapping to Go errors, specify precise fallbacks here.
		if scCode == "T0410" && strings.Contains(errStr, "Argument 1 of function") {
			matchedError = true
		} else if scCode == "S0201" && strings.Contains(errStr, "syntax error") {
			matchedError = true
		} else if scCode == "S0106" && strings.Contains(errStr, "unclosed comment") {
			matchedError = true
		} else if strings.Contains(errStr, fmt.Sprintf("[%s]", scCode)) || strings.Contains(errStr, fmt.Sprintf(" %s:", scCode)) || strings.HasPrefix(errStr, fmt.Sprintf("%s:", scCode)) {
			// General fallback for exactly structured JSONata error strings if they existed.
			// Explicitly bounds tokens with prefixes/spaces to avoid incidental substrings like "XT0410"
			matchedError = true
		}

		if !matchedError {
			if expectPass {
				return harnessOutcome{Failed: true, Message: fmt.Sprintf("Expected error %s but got different error: %v", scCode, execErr)}
			} else {
				return harnessOutcome{Skipped: true, Message: fmt.Sprintf("Expected failure (wrong error): Expected %s but got: %v", scCode, execErr)}
			}
		}

		if !expectPass {
			// If we expected the error and got the matching error, the test actually *passed*.
			// If it's on the expectedFailures list, it's an unexpected pass!
			return harnessOutcome{Failed: true, Message: fmt.Sprintf("Unexpected pass! Test %s is marked as expected failure but it produced expected error %s", testID, scCode)}
		}
		return harnessOutcome{Message: "pass-execution-error"} // Return a message to stop further assertions on outcome
	}

	if execErr != nil {
		// Only ignore true MissingPath or explicitly undefined evaluation outcomes.
		if isUndefined {
			var invalidor *lookup.Invalidor
			if errors.As(execErr, &invalidor) {
				unwrapped := invalidor.Unwrap()
				if errors.Is(unwrapped, lookup.ErrNoSuchPath) || strings.Contains(unwrapped.Error(), "element not found at simple path") {
					return harnessOutcome{} // Valid undefined path result, pass for assertion phase
				}
			}
		}

		if expectPass {
			return harnessOutcome{Failed: true, Message: fmt.Sprintf("runCase failed: %v", execErr)}
		} else {
			return harnessOutcome{Skipped: true, Message: fmt.Sprintf("Expected failure (runCase error): %v", execErr)}
		}
	}

	return harnessOutcome{}
}

func materializeHarnessValue(v interface{}) (value interface{}, undefined bool) {
	// First check invalidors mapping to missing elements BEFORE raw extraction
	if inv, ok := v.(*lookup.Invalidor); ok {
		if IsUndefinedError(inv) {
			return nil, true
		}
	}

	if p, ok := v.(lookup.Pathor); ok {
		v = p.Raw()
	}

	v = Materialize(v)
	if _, ok := v.(Undefined); ok {
		return nil, true
	}

	return v, false
}

func jsonataValuesEqual(expected, actual interface{}) bool {
	if reflect.DeepEqual(expected, actual) {
		return true
	}

	// JSON numbers are semantically numbers regardless of the Go numeric
	// representation chosen by the parser/evaluator.
	if e, ok := jsonataNumericValue(expected); ok {
		a, ok := jsonataNumericValue(actual)
		return ok && e.Cmp(a) == 0
	}

	ev := reflect.ValueOf(expected)
	av := reflect.ValueOf(actual)

	if ev.IsValid() && av.IsValid() &&
		(ev.Kind() == reflect.Slice || ev.Kind() == reflect.Array) &&
		(av.Kind() == reflect.Slice || av.Kind() == reflect.Array) {
		if ev.Len() != av.Len() {
			return false
		}
		for i := 0; i < ev.Len(); i++ {
			if !jsonataValuesEqual(ev.Index(i).Interface(), av.Index(i).Interface()) {
				return false
			}
		}
		return true
	}

	em, expectedIsMap := expected.(map[string]interface{})
	am, actualIsMap := actual.(map[string]interface{})
	if expectedIsMap || actualIsMap {
		if !expectedIsMap || !actualIsMap || len(em) != len(am) {
			return false
		}

		for key, expectedValue := range em {
			actualValue, ok := am[key]
			if !ok || !jsonataValuesEqual(expectedValue, actualValue) {
				return false
			}
		}
		return true
	}

	return false
}
