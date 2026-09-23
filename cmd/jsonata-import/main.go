package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/txtar"
)

const UpstreamRepo = "jsonata-js/jsonata"
// v2.0.2 tag commit SHA
const UpstreamRev = "2c12574e4c2cf27ab7fa1fbf09d84bf2c8f85f1c"

type TestCase struct {
	JSON    []byte
	Expr    []byte
	Result  []byte
}

func main() {
	verifyOnly := false
	if len(os.Args) > 1 && os.Args[1] == "verify" {
		verifyOnly = true
		log.Println("Running in read-only verification mode")
	}

	err := run(verifyOnly)
	if err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
	if verifyOnly {
		log.Println("Verification passed. No drift detected.")
	}
}

func run(verifyOnly bool) error {
	url := fmt.Sprintf("https://github.com/%s/archive/%s.tar.gz", UpstreamRepo, UpstreamRev)
	log.Printf("Downloading %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	outDir := filepath.Join("jsonata", "testdata", "test-suite")
	groupsDir := filepath.Join(outDir, "groups")
	datasetsDir := filepath.Join(outDir, "datasets")

	tests := make(map[string]map[string]*TestCase)
	datasets := make(map[string][]byte)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		parts := strings.Split(hdr.Name, "/")
		if len(parts) < 3 || parts[1] != "test" || parts[2] != "test-suite" {
			continue
		}

		subpath := strings.Join(parts[3:], "/")

		data, err := io.ReadAll(tr)
		if err != nil {
			return err
		}

		if len(data) > 0 && data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}

		if strings.HasPrefix(subpath, "datasets/") {
			dsName := strings.TrimSuffix(strings.TrimPrefix(subpath, "datasets/"), ".json")
			datasets[dsName] = data
			continue
		}

		if strings.HasPrefix(subpath, "groups/") {
			groupPath := strings.TrimPrefix(subpath, "groups/")
			groupParts := strings.Split(groupPath, "/")
			if len(groupParts) != 2 {
				continue
			}
			groupName := groupParts[0]
			fileName := groupParts[1]

			if tests[groupName] == nil {
				tests[groupName] = make(map[string]*TestCase)
			}

			baseName := fileName
			ext := ""
			if idx := strings.LastIndex(fileName, "."); idx != -1 {
				baseName = fileName[:idx]
				ext = strings.ToLower(fileName[idx:])
			}

			isExpected := strings.HasSuffix(baseName, "_expected")
			caseName := baseName
			if isExpected {
				caseName = strings.TrimSuffix(baseName, "_expected")
			}

			tc := tests[groupName][caseName]
			if tc == nil {
				tc = &TestCase{}
				tests[groupName][caseName] = tc
			}

			if isExpected {
				tc.Result = data
			} else if ext == ".json" {
				tc.JSON = data
			} else if ext == ".jsonata" {
				tc.Expr = data
			}
		}
	}

	log.Printf("Found %d groups", len(tests))
	log.Printf("Found %d datasets", len(datasets))

	if !verifyOnly {
		os.MkdirAll(groupsDir, 0755)
		os.MkdirAll(datasetsDir, 0755)
	}

	for dsName, dsData := range datasets {
		p := filepath.Join(datasetsDir, dsName+".json")
		if verifyOnly {
			existing, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("dataset mismatch (missing or err): %s - %v", dsName, err)
			}
			if !bytes.Equal(existing, dsData) {
				return fmt.Errorf("dataset mismatch (content diff): %s", dsName)
			}
		} else {
			err := os.WriteFile(p, dsData, 0644)
			if err != nil {
				return err
			}
		}
	}

	for groupName, groupTests := range tests {
		archive := new(txtar.Archive)

		caseNames := make([]string, 0, len(groupTests))
		for name := range groupTests {
			caseNames = append(caseNames, name)
		}
		sort.Strings(caseNames)

		for _, caseName := range caseNames {
			tc := groupTests[caseName]
			if tc.JSON != nil {
				var v interface{}
				if err := json.Unmarshal(tc.JSON, &v); err == nil {
					if _, ok := v.([]interface{}); ok {
						continue
					}

					vMap, ok := v.(map[string]interface{})
					if ok {
						if expr, hasExpr := vMap["expr"]; hasExpr {
							if tc.Expr == nil {
								if s, ok := expr.(string); ok {
									tc.Expr = []byte(s)
								}
							}
							if _, ok := vMap["exprFile"]; !ok {
								vMap["exprFile"] = caseName + ".JSONATA"
							}
							delete(vMap, "expr")
						}

						if res, hasRes := vMap["result"]; hasRes {
							if tc.Result == nil {
								if res == nil {
									tc.Result = []byte("null\n")
								} else if s, ok := res.(string); ok && s == "" {
									if b, err := json.Marshal(res); err == nil {
										tc.Result = b
									}
								} else {
									if b, err := json.Marshal(res); err == nil {
										tc.Result = b
									}
								}
							}
							delete(vMap, "result")
						}

						b, _ := json.MarshalIndent(vMap, "", "  ")
						b = append(b, '\n')
						archive.Files = append(archive.Files, txtar.File{
							Name: caseName + ".json",
							Data: b,
						})
					} else {
						archive.Files = append(archive.Files, txtar.File{
							Name: caseName + ".json",
							Data: tc.JSON,
						})
					}
				} else {
					archive.Files = append(archive.Files, txtar.File{
						Name: caseName + ".json",
						Data: tc.JSON,
					})
				}
			}
			if tc.Expr != nil {
				if len(tc.Expr) > 0 && tc.Expr[len(tc.Expr)-1] != '\n' {
					tc.Expr = append(tc.Expr, '\n')
				}
				archive.Files = append(archive.Files, txtar.File{
					Name: caseName + ".JSONATA",
					Data: tc.Expr,
				})
			}
			if tc.Result != nil {
				resultStr := strings.TrimSpace(string(tc.Result))
				if resultStr != "" {
					if len(tc.Result) > 0 && tc.Result[len(tc.Result)-1] != '\n' {
						tc.Result = append(tc.Result, '\n')
					}
					archive.Files = append(archive.Files, txtar.File{
						Name: caseName + "_expected.json",
						Data: tc.Result,
					})
				}
			}
		}

		p := filepath.Join(groupsDir, groupName+".txtar")
		expectedData := txtar.Format(archive)
		if verifyOnly {
			existing, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("group mismatch (missing or err): %s - %v", groupName, err)
			}
			if !bytes.Equal(existing, expectedData) {
				return fmt.Errorf("group mismatch (content diff): %s", groupName)
			}
		} else {
			err := os.WriteFile(p, expectedData, 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
