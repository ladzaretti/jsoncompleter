package jsoncompleter_test

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/ladzaretti/jsoncompleter"
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

	completerWithOpts := jsoncompleter.New(
		jsoncompleter.WithTruncationMarker("__CUSTOM__"),
		jsoncompleter.WithMarkTruncation(true),
	)

	completerWithoutOpts := jsoncompleter.New()

	for _, tt := range testFiles {
		withOpts := testSuite{completerWithOpts, tt.content}
		t.Run("with opts: "+tt.path, withOpts.testDynamicTruncation)

		withoutOpts := testSuite{completerWithoutOpts, tt.content}
		t.Run("without opts: "+tt.path, withoutOpts.testDynamicTruncation)
	}
}

type testSuite struct {
	completer   *jsoncompleter.Completer
	jsonContent string
}

func (s *testSuite) testDynamicTruncation(t *testing.T) {
	for i := len(s.jsonContent); i > 0; i-- {
		truncated := s.jsonContent[:i]
		got := s.completer.Complete(truncated)

		var anything any
		if err := json.Unmarshal([]byte(got), &anything); err != nil {
			t.Errorf("Complete(%q) produced invalid JSON: %v (output: %q)", shorten(truncated), err, shorten(got))
		}
	}
}

const maxLen = 200

func shorten(s string) string {
	if len(s) > maxLen {
		return "..." + s[len(s)-maxLen:]
	}

	return s
}
