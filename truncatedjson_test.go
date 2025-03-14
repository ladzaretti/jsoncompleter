package truncatedjson_test

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ladzaretti/truncatedjson"
)

//go:embed testdata/json
var fsys embed.FS

type testFile struct {
	path    string
	content string
}

func readTestFiles(t *testing.T, fsys embed.FS, dir string) []testFile {
	t.Helper()

	files, err := fsys.ReadDir(dir)
	if err != nil {
		t.Fatalf("directory %q: %v", dir, err)
	}

	filesContent := make([]testFile, 0, len(files))
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		p := filepath.Join(dir, f.Name())
		content, err := fsys.ReadFile(p)
		if err != nil {
			t.Fatalf("read file %q: %v", p, err)
		}

		filesContent = append(filesContent, testFile{path: p, content: string(content)})
	}

	return filesContent
}

func TestComplete(t *testing.T) {
	testFiles := readTestFiles(t, fsys, "testdata/json")

	for _, tt := range testFiles {
		t.Run(tt.path, func(t *testing.T) {
			for i := len(tt.content); i > 0; i-- {
				truncated := tt.content[:i]
				got := truncatedjson.Complete(truncated)

				var j any
				if err := json.Unmarshal([]byte(got), &j); err != nil {
					t.Errorf("Reconstruct(%q) produced invalid JSON: %v (output: %q)", truncated, err, got)
				}
			}
		})
	}
}
