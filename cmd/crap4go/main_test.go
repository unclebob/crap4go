package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/unclebob/crap4go/internal/cli"
	"github.com/unclebob/crap4go/internal/coverage"
)

func TestFindSourceFilesSkipsIgnoredDirsAndTests(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "target"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "ignored.go"), []byte("package git\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "target", "ignored.go"), []byte("package target\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "foo.go"), []byte("package pkg\nfunc Foo() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "foo_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := findSourceFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || !strings.HasSuffix(files[0], "pkg/foo.go") {
		t.Fatalf("unexpected source files: %#v", files)
	}

	filtered := filterSources(files, []string{"pkg"})
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered file, got %#v", filtered)
	}
	unfiltered := filterSources(files, nil)
	if len(unfiltered) != 1 {
		t.Fatalf("expected 1 unfiltered file, got %#v", unfiltered)
	}
	none := filterSources(files, []string{"nonexistent"})
	if len(none) != 0 {
		t.Fatalf("expected 0 matching files, got %#v", none)
	}
}

func TestRunHelpErrorAndCustomTestCommand(t *testing.T) {
	if err := run(cli.Options{Action: "help", Message: cli.HelpMessage}); err != nil {
		t.Fatalf("expected nil error for help action, got %v", err)
	}
	if err := run(cli.Options{Action: "error", Message: "bad flag"}); err == nil {
		t.Fatal("expected error for error action")
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte("package sample\nfunc Add(a, b int) int { return a + b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	opts := cli.Options{
		Action:      "analyze",
		MaxWorkers:  1,
		TestCommand: "printf 'mode: set\\nsample.go:2.1,2.40 1 1\\n' > " + coverageProfile + " && true {coverprofile}",
	}
	if err := run(opts); err != nil {
		t.Fatalf("expected run to succeed with custom test command, got %v", err)
	}

	optsParallel := cli.Options{
		Action:      "analyze",
		MaxWorkers:  2,
		TestCommand: "printf 'mode: set\\nsample.go:2.1,2.40 1 1\\n' > " + coverageProfile + " && true {coverprofile}",
	}
	if err := os.WriteFile(filepath.Join(dir, "second.go"), []byte("package sample\nfunc Sub(a, b int) int { return a - b }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run(optsParallel); err != nil {
		t.Fatalf("expected parallel run to succeed, got %v", err)
	}
}

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
