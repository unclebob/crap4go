package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/unclebob/crap4go/internal/cli"
	"github.com/unclebob/crap4go/internal/complexity"
	"github.com/unclebob/crap4go/internal/coverage"
	"github.com/unclebob/crap4go/internal/crap"
)

const coverageProfile = "target/coverage/coverage.out"

func main() {
	if err := run(cli.ParseArgs(os.Args[1:])); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(options cli.Options) error {
	if options.Action == "help" {
		fmt.Println(options.Message)
		return nil
	}
	if err := os.RemoveAll(filepath.Dir(coverageProfile)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(coverageProfile), 0o755); err != nil {
		return err
	}
	if err := runCoverage(); err != nil {
		return err
	}
	profile, err := coverage.LoadProfile(coverageProfile)
	if err != nil {
		return err
	}
	entries, err := sortedEntries(options, profile)
	if err != nil {
		return err
	}
	fmt.Print(crap.FormatReport(entries))
	return nil
}

func runCoverage() error {
	cmd := exec.Command("go", "test", "./...", "-coverprofile="+coverageProfile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func sortedEntries(options cli.Options, profile map[string][]coverage.Segment) ([]crap.Entry, error) {
	files, err := findSourceFiles(".")
	if err != nil {
		return nil, err
	}
	files = filterSources(files, options.ModuleFilters)
	var entries []crap.Entry
	for _, file := range files {
		functions, err := complexity.ExtractFunctions(file)
		if err != nil {
			return nil, err
		}
		for _, fn := range functions {
			cov := coverage.CoverageForRange(profile, file, fn.StartLine, fn.EndLine)
			entries = append(entries, crap.Entry{
				Name:       fn.Name,
				Package:    fn.Package,
				Complexity: fn.Complexity,
				Coverage:   cov,
				CRAP:       crap.Score(fn.Complexity, cov),
			})
		}
	}
	return crap.SortByCRAP(entries), nil
}

func findSourceFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "target", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			files = append(files, filepath.ToSlash(path))
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func filterSources(files []string, moduleFilters []string) []string {
	if len(moduleFilters) == 0 {
		return files
	}
	var out []string
	for _, file := range files {
		for _, filter := range moduleFilters {
			if strings.Contains(file, filter) {
				out = append(out, file)
				break
			}
		}
	}
	return out
}
