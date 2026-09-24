package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/txtar"
)

func upstreamTestArchive(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tr := tar.NewWriter(gz)
	fixtures := map[string]string{
		"test/test-suite/datasets/sample.json": `{"x":9007199254740993}`,
		"test/test-suite/groups/recovery/case000.json": `{"expr":"$","data":9007199254740993,"result":9007199254740993}`,
		"test/test-suite/groups/recovery/case001.json": `[{"expr":"$","data":null,"result":null},{"expr":"$","data":"null","result":"null"}]`,
		"test/test-suite/groups/recovery/case002.json": `4`,
		"test/test-suite/groups/recovery/case002.jsonata": `$`,
		"test/test-suite/groups/recovery/case003.json": `{"expr":"$","dataset":null,"undefinedResult":true}`,
		"test/test-suite/groups/recovery/case004.json": `{"expr":"x","dataset":"sample","bindings":{"p":1},"code":"T0410"}`,
	}
	for name, text := range fixtures {
		path := "jsonata-test-archive/" + name
		data := []byte(text)
		if err := tr.WriteHeader(&tar.Header{Name: path, Mode: 0644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tr.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tr.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func findArchiveFile(t *testing.T, archive *txtar.Archive, name string) string {
	t.Helper()
	for _, file := range archive.Files {
		if file.Name == name {
			return string(file.Data)
		}
	}
	t.Fatalf("generated archive is missing %s", name)
	return ""
}

func TestImporterPreservesInputAndExpandsCases(t *testing.T) {
	out := t.TempDir()
	source := upstreamTestArchive(t)
	if err := importArchive(bytes.NewReader(source), out, false); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(out, "groups", "recovery.txtar"))
	if err != nil {
		t.Fatal(err)
	}
	archive := txtar.Parse(data)
	if got := len(archive.Files); got != 13 {
		t.Fatalf("expected 13 executable case files (5 cases with result files), got %d", got)
	}
	if got := findArchiveFile(t, archive, "case000.json"); !strings.Contains(got, `9007199254740993`) || strings.Contains(got, `9007199254740992`) {
		t.Fatalf("large integer was rounded: %s", got)
	}
	if got := findArchiveFile(t, archive, "case000_expected.json"); strings.TrimSpace(got) != "9007199254740993" {
		t.Fatalf("expected large integer changed: %s", got)
	}
	for _, name := range []string{"case001_0", "case001_1"} {
		if got := findArchiveFile(t, archive, name+".json"); !strings.Contains(got, `"exprFile": "`+name+`.JSONATA"`) {
			t.Fatalf("array case %s cannot load its expression: %s", name, got)
		}
		findArchiveFile(t, archive, name+".JSONATA")
	}
	if got := findArchiveFile(t, archive, "case001_0.json"); !strings.Contains(got, `"data": null`) {
		t.Fatalf("inline null lost: %s", got)
	}
	if got := findArchiveFile(t, archive, "case001_1.json"); !strings.Contains(got, `"data": "null"`) {
		t.Fatalf("string null lost: %s", got)
	}
	if got := findArchiveFile(t, archive, "case002.json"); !strings.Contains(got, `"data": 4`) {
		t.Fatalf("scalar input lost: %s", got)
	}
	if got := findArchiveFile(t, archive, "case003.json"); !strings.Contains(got, `"dataset": null`) || !strings.Contains(got, `"undefinedResult": true`) {
		t.Fatalf("undefined input/result lost: %s", got)
	}
	if got := findArchiveFile(t, archive, "case004.json"); !strings.Contains(got, `"dataset": "sample"`) || !strings.Contains(got, `"bindings"`) || !strings.Contains(got, `"T0410"`) {
		t.Fatalf("dataset/bindings/error code lost: %s", got)
	}
	if got, err := os.ReadFile(filepath.Join(out, "datasets", "sample.json")); err != nil || !strings.Contains(string(got), `9007199254740993`) {
		t.Fatalf("dataset lost or modified: %s, %v", got, err)
	}
	if err := importArchive(bytes.NewReader(source), out, true); err != nil {
		t.Fatalf("verification rejected unchanged generated fixtures: %v", err)
	}
}

func TestImporterVerifyDetectsDriftAndStaleFiles(t *testing.T) {
	out := t.TempDir()
	source := upstreamTestArchive(t)
	if err := importArchive(bytes.NewReader(source), out, false); err != nil {
		t.Fatal(err)
	}
	group := filepath.Join(out, "groups", "recovery.txtar")
	original, err := os.ReadFile(group)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(group, append(bytes.Clone(original), 'x'), 0644); err != nil {
		t.Fatal(err)
	}
	if err := importArchive(bytes.NewReader(source), out, true); err == nil {
		t.Fatal("verification accepted a corrupted group")
	}
	if got, err := os.ReadFile(group); err != nil || !bytes.HasSuffix(got, []byte("x")) {
		t.Fatalf("verify mode modified the corrupted fixture: %v", err)
	}
	if err := os.WriteFile(group, original, 0644); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"groups", "datasets"} {
		stale := filepath.Join(out, kind, "obsolete.fixture")
		if err := os.WriteFile(stale, []byte("stale"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := importArchive(bytes.NewReader(source), out, true); err == nil {
			t.Fatalf("verification accepted stale file in %s", kind)
		}
		if _, err := os.Stat(stale); err != nil {
			t.Fatalf("verification modified %s: %v", stale, err)
		}
		if err := os.Remove(stale); err != nil {
			t.Fatal(err)
		}
	}
}

func TestImporterRejectsMalformedInput(t *testing.T) {
	if err := importArchive(strings.NewReader("not gzip"), t.TempDir(), false); err == nil {
		t.Fatal("malformed input accepted")
	}
}
