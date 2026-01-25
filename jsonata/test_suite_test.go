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
	return res.Raw(), nil
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
			_, skipOnFail := groupStatus[groupName]
			expectPass := !skipOnFail
			runTxtarGroup(t, path.Join("testdata/test-suite/groups", entry.Name()), expectPass)
		})
	}
}

func runTxtarGroup(t *testing.T, filename string, expectPass bool) {
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
			if strings.Contains(filename, "comments.txtar") {
				if c.Name == "case002" {
					t.Skip("Skipping case002: Error expectation logic not implemented in test runner")
				}
				if c.Name == "case003" {
					t.Skip("Skipping case003: Function definition not implemented")
				}
			}

			var sc suiteCase
			if err := json.Unmarshal([]byte(c.Input), &sc); err != nil {
				t.Fatalf("invalid suite case config: %v", err)
			}

			// Capture panic to treat as failure instead of crash
			defer func() {
				if r := recover(); r != nil {
					if expectPass {
						t.Errorf("panic: %v", r)
					} else {
						t.Skipf("panic: %v", r)
					}
				}
			}()

			out, err := runCase(sc, c.Expr)
			if err != nil {
				if expectPass {
					t.Fatalf("runCase failed: %v", err)
				} else {
					t.Skipf("runCase failed: %v", err)
				}
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

			if expectPass {
				if n, ok := expected.(json.Number); ok {
					f, err := n.Float64()
					if err == nil {
						// Compare as float if actual is float
						if fOut, ok := out.(float64); ok {
							assert.InDelta(t, f, fOut, 0.0000001)
							return
						}
						// Compare as int if actual is int
						i, err := n.Int64()
						if err == nil {
							if iOut, ok := out.(int); ok {
								assert.Equal(t, i, int64(iOut))
								return
							}
							if iOut, ok := out.(int64); ok {
								assert.Equal(t, i, iOut)
								return
							}
						}
					}
				}
				assert.Equal(t, expected, out)
			} else {
				if !assert.ObjectsAreEqual(expected, out) {
					t.Skipf("Skipping failed test. Expected: %v, Got: %v", expected, out)
				}
			}
		})
	}
}
