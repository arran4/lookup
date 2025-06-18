package jsonata

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

type suiteCase struct {
	ExprFile  string                 `json:"exprFile"`
	Dataset   string                 `json:"dataset"`
	Data      interface{}            `json:"data"`
	Bindings  map[string]interface{} `json:"bindings"`
	Undefined bool                   `json:"undefinedResult"`
}

func loadDataset(t *testing.T, name string) interface{} {
	path := filepath.Join("testdata", "test-suite", "datasets", name+".json")
	return loadJSONFile(t, filepath.Join("jsonata", path))
}

func loadJSONFile(t *testing.T, path string) interface{} {
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", path, err)
	}
	return v
}

func runCase(t *testing.T, casePath string, c suiteCase) interface{} {
	var data interface{}
	if c.Data != nil {
		data = c.Data
	} else if c.Dataset != "" {
		data = loadDataset(t, c.Dataset)
	}
	exprFile := filepath.Join(filepath.Dir(casePath), c.ExprFile)
	exprBytes, err := ioutil.ReadFile(filepath.Clean(exprFile))
	if err != nil {
		t.Fatalf("failed to read expression file: %v", err)
	}
	expr := strings.TrimSpace(string(exprBytes))
	ast, err := Parse(expr)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	q := Compile(ast)
	root := lookup.Reflect(data)
	return q.Run(lookup.NewScope(root, root)).Raw()
}

func loadExpected(t *testing.T, caseFile string) interface{} {
	expectedPath := strings.TrimSuffix(caseFile, ".json") + "_expected.json"
	return loadJSONFile(t, expectedPath)
}

// TestJSONataFieldsGroup loads test cases from the upstream JSONata
// "fields" group and evaluates each expression against the provided
// dataset. Expected results are stored in companion _expected.json files.
func TestJSONataFieldsGroup(t *testing.T) {
	skipIf(t, testFeatureFieldsGroup, "fields group")
	files, err := filepath.Glob("jsonata/testdata/test-suite/groups/fields/case*.json")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			bytes, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatalf("failed to read case: %v", err)
			}
			var c suiteCase
			if err := json.Unmarshal(bytes, &c); err != nil {
				t.Fatalf("invalid case: %v", err)
			}
			out := runCase(t, f, c)
			expected := loadExpected(t, f)
			assert.Equal(t, expected, out)
		})
	}
}

// TestJSONataArrayConstructorGroup loads cases from the array-constructor group
// of the upstream JSONata test suite. Expressions are stored in companion
// .JSONATA files next to each case JSON.
func TestJSONataArrayConstructorGroup(t *testing.T) {
	skipIf(t, testFeatureArrayConstructorGroup, "array constructor group")
	files, err := filepath.Glob("jsonata/testdata/test-suite/groups/array-constructor/case*.json")
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			bytes, err := ioutil.ReadFile(f)
			if err != nil {
				t.Fatalf("failed to read case: %v", err)
			}
			var c suiteCase
			if err := json.Unmarshal(bytes, &c); err != nil {
				t.Fatalf("invalid case: %v", err)
			}
			out := runCase(t, f, c)
			expected := loadExpected(t, f)
			assert.Equal(t, expected, out)
		})
	}
}
