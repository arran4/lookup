package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/tools/txtar"
)

const UpstreamRepo = "jsonata-js/jsonata"

// UpstreamRev is the commit to which upstream tag v2.0.2 resolves.
const UpstreamRev = "9e4c1cbe97859d04bd32d194eb49c6f485e0ffea"

type TestCase struct {
	JSON   []byte
	Expr   []byte
	Result []byte
}

func main() {
	if len(os.Args) > 2 || (len(os.Args) == 2 && os.Args[1] != "verify") {
		log.Fatal("usage: go run ./cmd/jsonata-import [verify]")
	}
	verify := len(os.Args) == 2
	if err := run(verify); err != nil {
		log.Fatal(err)
	}
}

func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above working directory")
		}
		dir = parent
	}
}

func run(verifyOnly bool) error {
	root, err := repositoryRoot()
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", UpstreamRepo, UpstreamRev)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upstream %s: %s", url, resp.Status)
	}
	outDir := filepath.Join(root, "jsonata", "testdata", "test-suite")
	return importArchive(resp.Body, outDir, verifyOnly)
}

// importArchive accepts a pinned upstream .tar.gz stream. Keeping I/O injected
// makes conversion and verification testable without networking or tracked files.
func importArchive(source io.Reader, outDir string, verifyOnly bool) error {
	gz, err := gzip.NewReader(source)
	if err != nil {
		return fmt.Errorf("open upstream archive: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	groupJSON := make(map[string]map[string][]byte)
	groupExpr := make(map[string]map[string][]byte)
	groupResult := make(map[string]map[string][]byte)
	datasets := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read upstream archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		parts := strings.Split(hdr.Name, "/")
		if len(parts) < 4 || parts[1] != "test" || parts[2] != "test-suite" {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return fmt.Errorf("read %s: %w", hdr.Name, err)
		}
		data = withNewline(data)
		subpath := parts[3:]
		if len(subpath) == 2 && subpath[0] == "datasets" && strings.HasSuffix(subpath[1], ".json") {
			name := strings.TrimSuffix(subpath[1], ".json")
			if name == "" {
				return fmt.Errorf("invalid dataset name: %s", hdr.Name)
			}
			datasets[name] = data
			continue
		}
		if len(subpath) != 3 || subpath[0] != "groups" {
			continue
		}
		group, filename := subpath[1], subpath[2]
		ext := strings.ToLower(filepath.Ext(filename))
		base := filename[:len(filename)-len(ext)]
		if ext != ".json" && ext != ".jsonata" {
			continue
		}
		if group == "" || base == "" {
			return fmt.Errorf("invalid case name: %s", hdr.Name)
		}
		if strings.HasSuffix(base, "_expected") {
			caseName := strings.TrimSuffix(base, "_expected")
			if groupResult[group] == nil {
				groupResult[group] = make(map[string][]byte)
			}
			groupResult[group][caseName] = data
			continue
		}
		if ext == ".json" {
			if groupJSON[group] == nil {
				groupJSON[group] = make(map[string][]byte)
			}
			groupJSON[group][base] = data
		} else if ext == ".jsonata" {
			if groupExpr[group] == nil {
				groupExpr[group] = make(map[string][]byte)
			}
			groupExpr[group][filename] = data
			groupExpr[group][base] = data
		}
	}
	if len(groupJSON) == 0 {
		return fmt.Errorf("upstream archive contained no JSONata suite groups")
	}
	files := make(map[string][]byte)
	for name, data := range datasets {
		files[filepath.Join("datasets", name+".json")] = data
	}
	for group, jsonCases := range groupJSON {
		archive := new(txtar.Archive)
		names := make([]string, 0, len(jsonCases))
		for name := range jsonCases {
			names = append(names, name)
		}
		sort.Strings(names)
		generated := make(map[string]bool)
		for _, name := range names {
			jsonData := jsonCases[name]
			var value interface{}
			dec := json.NewDecoder(bytes.NewReader(jsonData))
			dec.UseNumber()
			if err := dec.Decode(&value); err != nil {
				return fmt.Errorf("decode %s/%s: %w", group, name, err)
			}
			var trailing interface{}
			if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
				return fmt.Errorf("trailing JSON in %s/%s: %v", group, name, err)
			}
			if entries, ok := value.([]interface{}); ok {
				if len(entries) == 0 {
					return fmt.Errorf("empty case array in %s/%s", group, name)
				}
				for i, entry := range entries {
					caseName := name + "_" + strconv.Itoa(i)
					resData := groupResult[group][caseName]
					if len(resData) == 0 {
						resData = groupResult[group][name]
					}
					if err := appendCase(archive, group, caseName, entry, groupExpr[group], resData); err != nil {
						return fmt.Errorf("%s/%s: %w", group, caseName, err)
					}
					generated[caseName] = true
				}
			} else {
				resData := groupResult[group][name]
				if err := appendCase(archive, group, name, value, groupExpr[group], resData); err != nil {
					return fmt.Errorf("%s/%s: %w", group, name, err)
				}
				generated[name] = true
			}
		}
		if len(generated) == 0 {
			return fmt.Errorf("no cases generated for group %s", group)
		}
		files[filepath.Join("groups", group+".txtar")] = txtar.Format(archive)
	}
	if err := synchronize(outDir, files, verifyOnly); err != nil {
		return err
	}
	log.Printf("JSONata suite: %d groups, %d datasets, %d output files", len(groupJSON), len(datasets), len(files))
	return nil
}

func withNewline(data []byte) []byte {
	if len(data) == 0 || data[len(data)-1] == '\n' {
		return data
	}
	return append(data, '\n')
}

func appendCase(archive *txtar.Archive, group string, name string, source interface{}, exprs map[string][]byte, result []byte) error {
	metadata, isObject := source.(map[string]interface{})
	if !isObject {
		// A scalar case is input data, not valid harness metadata by itself.
		metadata = map[string]interface{}{"data": source}
	}
	var expr []byte
	if inline, ok := metadata["expr"]; ok {
		text, ok := inline.(string)
		if !ok {
			return fmt.Errorf("expression must be a string")
		}
		expr = []byte(text)
	} else if ref, ok := metadata["expr-file"]; ok {
		fn, ok := ref.(string)
		if !ok {
			return fmt.Errorf("expr-file must be a string")
		}
		if exprs != nil {
			expr = exprs[fn]
			if len(expr) == 0 {
				expr = exprs[strings.TrimSuffix(fn, ".jsonata")]
			}
		}
		if len(expr) == 0 {
			return fmt.Errorf("referenced expression file %s not found in group %s", fn, group)
		}
	} else if ref, ok := metadata["exprFile"]; ok {
		fn, ok := ref.(string)
		if !ok {
			return fmt.Errorf("exprFile must be a string")
		}
		if exprs != nil {
			expr = exprs[fn]
			if len(expr) == 0 {
				expr = exprs[strings.TrimSuffix(fn, ".jsonata")]
			}
		}
		if len(expr) == 0 {
			return fmt.Errorf("referenced expression file %s not found in group %s", fn, group)
		}
	} else if exprs != nil {
		if data, ok := exprs[name+".jsonata"]; ok {
			expr = data
		} else if data, ok := exprs[name]; ok {
			expr = data
		}
	}
	if len(expr) == 0 {
		return fmt.Errorf("missing expression")
	}

	delete(metadata, "expr")
	delete(metadata, "expr-file")
	delete(metadata, "exprFile")
	metadata["exprFile"] = name + ".JSONATA"

	if len(result) == 0 {
		if inline, ok := metadata["result"]; ok {
			var err error
			result, err = json.Marshal(inline)
			if err != nil {
				return fmt.Errorf("encode expected result: %w", err)
			}
		}
	}
	delete(metadata, "result")

	encoded, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	archive.Files = append(archive.Files, txtar.File{Name: name + ".json", Data: withNewline(encoded)})
	archive.Files = append(archive.Files, txtar.File{Name: name + ".JSONATA", Data: withNewline(expr)})
	if len(bytes.TrimSpace(result)) > 0 {
		archive.Files = append(archive.Files, txtar.File{Name: name + "_expected.json", Data: withNewline(result)})
	}
	return nil
}

func synchronize(outDir string, expected map[string][]byte, verifyOnly bool) error {
	// Inspect the two managed directories rather than rewriting arbitrary files
	// under the project tree. In verify mode not a single file is modified.
	for _, kind := range []string{"groups", "datasets"} {
		dir := filepath.Join(outDir, kind)
		entries, err := os.ReadDir(dir)
		if errors.Is(err, os.ErrNotExist) && !verifyOnly {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("create %s: %w", dir, err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("read %s: %w", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				return fmt.Errorf("unexpected directory in managed suite: %s", entry.Name())
			}
			rel := filepath.Join(kind, entry.Name())
			if rel == filepath.Join("datasets", "malformed_test_dataset.json") {
				continue
			}
			if _, ok := expected[rel]; !ok {
				if verifyOnly {
					return fmt.Errorf("stale fixture: %s", rel)
				}
				if err := os.Remove(filepath.Join(outDir, rel)); err != nil {
					return fmt.Errorf("remove stale fixture %s: %w", rel, err)
				}
			}
		}
	}
	paths := make([]string, 0, len(expected))
	for p := range expected {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, rel := range paths {
		p := filepath.Join(outDir, rel)
		if verifyOnly {
			actual, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("missing fixture %s: %w", rel, err)
			}
			if !bytes.Equal(actual, expected[rel]) {
				return fmt.Errorf("fixture content differs: %s", rel)
			}
		} else if err := os.WriteFile(p, expected[rel], 0644); err != nil {
			return fmt.Errorf("write fixture %s: %w", rel, err)
		}
	}
	return nil
}
