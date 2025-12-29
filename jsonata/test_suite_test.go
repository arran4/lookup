package jsonata

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"path"
	"strings"
	"testing"

	"github.com/arran4/lookup"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata
var testData embed.FS

type suiteCase struct {
	ExprFile  string                 `json:"exprFile"`
	Dataset   string                 `json:"dataset"`
	Data      interface{}            `json:"data"`
	Bindings  map[string]interface{} `json:"bindings"`
	Undefined bool                   `json:"undefinedResult"`
}

func loadDataset(t *testing.T, name string) interface{} {
	filename := path.Join("testdata", "test-suite", "datasets", name+".json")
	data, err := fs.ReadFile(testData, filename)
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}
	var v interface{}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("failed to unmarshal %s: %v", filename, err)
	}
	return v
}

func parseJSON(t *testing.T, data string) interface{} {
	var v interface{}
	dec := json.NewDecoder(strings.NewReader(data))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}
	return v
}

func runCase(t *testing.T, c suiteCase, expr string) interface{} {
	var data interface{}
	if c.Data != nil {
		data = c.Data
	} else if c.Dataset != "" {
		data = loadDataset(t, c.Dataset)
	}

	ast, err := Parse(expr)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	q := Compile(ast)
	root := lookup.Reflect(data)
	return q.Run(lookup.NewScope(root, root)).Raw()
}

// TestJSONataFieldsGroup loads test cases from the upstream JSONata
// "fields" group and evaluates each expression against the provided
// dataset. Expected results are stored in companion _expected.json files.
func TestJSONataFieldsGroup(t *testing.T) {
	skipIf(t, testFeatureFieldsGroup, "fields group")
	runTxtarGroup(t, "testdata/test-suite/groups/fields.txtar")
}

// TestJSONataArrayConstructorGroup loads cases from the array-constructor group
// of the upstream JSONata test suite. Expressions are stored in companion
// .JSONATA files next to each case JSON.
func TestJSONataArrayConstructorGroup(t *testing.T) {
	skipIf(t, testFeatureArrayConstructorGroup, "array constructor group")
	runTxtarGroup(t, "testdata/test-suite/groups/array-constructor.txtar")
}

func runTxtarGroup(t *testing.T, filename string) {
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
			var sc suiteCase
			if err := json.Unmarshal([]byte(c.Input), &sc); err != nil {
				t.Fatalf("invalid suite case config: %v", err)
			}

			out := runCase(t, sc, c.Expr)
			expected := parseJSON(t, c.Expected)
			assert.Equal(t, expected, out)
		})
	}
}
