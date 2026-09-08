package jsonata

// This file is used to generate the JSONata compatibility matrix in the README.

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

type matrixStats struct {
	Passed      int
	Failed      int
	Unsupported int
}

// generateCompatibilityMatrix builds the markdown matrix string mapping
func generateCompatibilityMatrix(fsys embed.FS) (string, error) {
	entries, err := fsys.ReadDir("testdata/test-suite/groups")
	if err != nil {
		return "", fmt.Errorf("failed to list groups: %w", err)
	}

	groups := make(map[string]*matrixStats)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txtar") {
			continue
		}

		groupName := strings.TrimSuffix(entry.Name(), ".txtar")
		groups[groupName] = &matrixStats{}

		data, err := fs.ReadFile(fsys, path.Join("testdata/test-suite/groups", entry.Name()))
		if err != nil {
			return "", fmt.Errorf("failed to read %s: %w", entry.Name(), err)
		}

		cases, err := parseTxtar(data)
		if err != nil {
			return "", fmt.Errorf("failed to parse txtar file %s: %w", entry.Name(), err)
		}

		for _, c := range cases {
			testID := groupName + "/" + c.Name

			if _, ok := unsupportedTests[testID]; ok {
				groups[groupName].Unsupported++
				continue
			}

			if expectedFailures[testID] {
				groups[groupName].Failed++
			} else {
				groups[groupName].Passed++
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("| Feature Group | Passed | Failed | Unsupported |\n")
	sb.WriteString("|---|---|---|---|\n")

	var keys []string
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		s := groups[k]
		fmt.Fprintf(&sb, "| %s | %d | %d | %d |\n", k, s.Passed, s.Failed, s.Unsupported)
	}

	return sb.String(), nil
}
