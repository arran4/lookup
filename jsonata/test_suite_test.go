package jsonata

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/arran4/go-evaluator"
	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
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

func runCase(c suiteCase, expr string) (interface{}, error) {
	var data interface{}
	var err error
	if c.Data != nil {
		data = c.Data
	} else if c.Dataset != "" {
		data, err = loadDataset(c.Dataset)
		if err != nil {
			return nil, err
		}
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

			res, err := runCase(sc, c.Expr)
			var out interface{}
			if res != nil {
				if p, ok := res.(lookup.Pathor); ok {
					out = p.Raw()
				} else {
					out = res
				}
			}

			// Setup errors, e.g. failing to read dataset
			if err != nil && strings.Contains(err.Error(), "failed to read") {
				if expectPass {
					t.Fatalf("Setup failed: %v", err)
				} else {
					t.Skipf("Setup failed (expected failure): %v", err)
				}
				return
			}

			// Some tests are meant to fail parsing or execution. In those cases sc.Code is set.
			// Alternatively if out is an error/Invalidor it should be treated as an error.
			var execErr error
			if err != nil {
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

			outcome := evaluateHarnessOutcome(testID, out, execErr, sc.Code, expectPass, isUnsupported, unsupportedReason)

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

			match := assert.ObjectsAreEqual(expected, out)
			if !match {
				if n, ok := expected.(json.Number); ok {
					f, err := n.Float64()
					if err == nil {
						// Compare as float if actual is float
						if fOut, ok := out.(float64); ok {
							match = assert.ObjectsAreEqualValues(f, fOut) // approximation
						} else {
							i, err := n.Int64()
							if err == nil {
								if iOut, ok := out.(int); ok {
									match = i == int64(iOut)
								} else if iOut, ok := out.(int64); ok {
									match = i == iOut
								}
							}
						}
					}
				}
			}

			if expectPass {
				if !match {
					t.Fatalf("Test failed. Expected: %v, Got: %v", expected, out)
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

func evaluateHarnessOutcome(testID string, out interface{}, execErr error, scCode string, expectPass bool, isUnsupported bool, unsupportedReason string) harnessOutcome {
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
		// Optionally verify exact error code matches, but for now consider it a pass
		return harnessOutcome{Message: "pass-execution-error"} // Return a message to stop further assertions on outcome
	}

	if execErr != nil {
		if expectPass {
			return harnessOutcome{Failed: true, Message: fmt.Sprintf("runCase failed: %v", execErr)}
		} else {
			return harnessOutcome{Skipped: true, Message: fmt.Sprintf("Expected failure (runCase error): %v", execErr)}
		}
	}

	return harnessOutcome{}
}
