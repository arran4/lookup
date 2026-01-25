package jsonata

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/tools/txtar"
)

type txtarCase struct {
	Name     string
	Input    string
	Expr     string
	Expected string
}

type suiteCase struct {
	ExprFile  string                 `json:"exprFile"`
	Dataset   string                 `json:"dataset"`
	Data      interface{}            `json:"data"`
	Bindings  map[string]interface{} `json:"bindings"`
	Undefined bool                   `json:"undefinedResult"`
}

func parseTxtar(data []byte) ([]txtarCase, error) {
	archive := txtar.Parse(data)
	var cases []txtarCase

	// We expect a structure where we have pairs or triplets of files.
	// However, the previous structure had a JSON file describing the case,
	// pointing to an expr file and a dataset.
	//
	// Let's adopt a convention for the txtar format:
	// Each test case is a set of files prefixed with the case name.
	// e.g. case001.json (config), case001.JSONATA (expr), case001_expected.json (result)
	//
	// Alternatively, we can just iterate the files and group them.

	// Let's use a map to group files by their base name (without extension)
	files := make(map[string]map[string]string)

	for _, f := range archive.Files {
		// name example: case001.json
		ext := ""
		base := f.Name
		if idx := strings.LastIndex(f.Name, "."); idx != -1 {
			ext = f.Name[idx:]
			base = f.Name[:idx]
		}

		// Handle _expected suffix special case
		if strings.HasSuffix(base, "_expected") {
			base = strings.TrimSuffix(base, "_expected")
			ext = "_expected" + ext
		}

		if _, ok := files[base]; !ok {
			files[base] = make(map[string]string)
		}
		files[base][ext] = string(f.Data)
	}

	for name, fileMap := range files {
		c := txtarCase{Name: name}

		// Logic to reconstruct the suiteCase struct from the files
		// The original JSON file (e.g. case001.json) contained:
		// { "dataset": "dataset5__INPUT", "bindings": {}, "exprFile": "case001.JSONATA" }

		configJSON, ok := fileMap[".json"]
		if !ok {
			// If there is no config json, maybe it's a different structure or incomplete case
			continue
		}

		var sc suiteCase
		if err := json.Unmarshal([]byte(configJSON), &sc); err != nil {
			return nil, fmt.Errorf("failed to parse config for %s: %v", name, err)
		}

		// The expression might be in the fileMap if it was included in the txtar
		// Or it might be referenced. Ideally in the txtar approach we include everything.
		if expr, ok := fileMap[".JSONATA"]; ok {
			c.Expr = strings.TrimSpace(expr)
		}

		if expected, ok := fileMap["_expected.json"]; ok {
			c.Expected = expected
		}

		// We still need the dataset name to load the dataset
		// We'll store the raw config input to let the runner handle dataset loading?
		// Or we can just store the dataset name.
		// Let's store the config string as Input for now, but actually we need the dataset name.
		// Wait, 'Input' in txtarCase usually refers to the data being processed.
		// Here the data is in a separate dataset file common to many tests.
		// Let's abuse Input to store the config JSON for now, or add fields.

		// Let's refine txtarCase
		c.Input = configJSON // The suiteCase JSON
		cases = append(cases, c)
	}

	return cases, nil
}
