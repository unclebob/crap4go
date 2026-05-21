package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/unclebob/crap4go/internal/coverage"
)

func TestSortedEntriesParallelMatchesSerial(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(`package sample

func Alpha(x int) int {
	if x > 0 {
		return x
	}
	return 0
}
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.go"), []byte(`package sample

func Beta() int {
	return 1
}
`), 0o644); err != nil {
		t.Fatal(err)
	}

	files := []string{filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")}
	profile := map[string][]coverage.Segment{
		files[0]: {{StartLine: 3, EndLine: 8, Statements: 1, Count: 1}},
		files[1]: {{StartLine: 3, EndLine: 5, Statements: 1, Count: 1}},
	}

	serial, err := sortedEntriesSerial(files, profile)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := sortedEntriesParallel(files, profile, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(parallel) != len(serial) {
		t.Fatalf("expected %d entries, got %d", len(serial), len(parallel))
	}
	for i := range serial {
		if parallel[i].Name != serial[i].Name || parallel[i].Complexity != serial[i].Complexity {
			t.Fatalf("parallel result differs at %d: %#v != %#v", i, parallel[i], serial[i])
		}
	}
}

func TestCoverageCommandDefaultsToGoTestAllPackages(t *testing.T) {
	cmd := coverageCommand("")
	want := []string{"go", "test", "./...", "-coverprofile=" + coverageProfile}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Fatalf("coverage command = %#v, want %#v", cmd.Args, want)
	}
}

func TestCoverageCommandAppendsCoverageFlagToCustomCommand(t *testing.T) {
	cmd := coverageCommand("go test -tags appunit ./internal/app")
	want := "go test -tags appunit ./internal/app -coverprofile=" + coverageProfile
	if !strings.HasSuffix(cmd.Path, "sh") || len(cmd.Args) != 3 || cmd.Args[2] != want {
		t.Fatalf("coverage command = path %q args %#v", cmd.Path, cmd.Args)
	}
}

func TestCoverageCommandReplacesCoveragePlaceholder(t *testing.T) {
	cmd := coverageCommand("go test ./internal/app {coverprofile} -run TestThing")
	want := "go test ./internal/app -coverprofile=" + coverageProfile + " -run TestThing"
	if !strings.HasSuffix(cmd.Path, "sh") || len(cmd.Args) != 3 || cmd.Args[2] != want {
		t.Fatalf("coverage command = path %q args %#v", cmd.Path, cmd.Args)
	}
}
