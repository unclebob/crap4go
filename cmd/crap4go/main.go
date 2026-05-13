package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

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
	if options.Action == "error" {
		return fmt.Errorf("%s\n\n%s", options.Message, cli.HelpMessage)
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
	if options.MaxWorkers <= 1 || len(files) <= 1 {
		return sortedEntriesSerial(files, profile)
	}
	return sortedEntriesParallel(files, profile, options.MaxWorkers)
}

func sortedEntriesSerial(files []string, profile map[string][]coverage.Segment) ([]crap.Entry, error) {
	var entries []crap.Entry
	for _, file := range files {
		fileEntries, err := entriesForFile(file, profile)
		if err != nil {
			return nil, err
		}
		entries = append(entries, fileEntries...)
	}
	return crap.SortByCRAP(entries), nil
}

func sortedEntriesParallel(files []string, profile map[string][]coverage.Segment, maxWorkers int) ([]crap.Entry, error) {
	if maxWorkers > len(files) {
		maxWorkers = len(files)
	}
	type job struct {
		index int
		file  string
	}
	type result struct {
		index   int
		entries []crap.Entry
		err     error
	}

	jobs := make(chan job, len(files))
	results := make(chan result, len(files))
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				entries, err := entriesForFile(job.file, profile)
				results <- result{index: job.index, entries: entries, err: err}
			}
		}()
	}
	for i, file := range files {
		jobs <- job{index: i, file: file}
	}
	close(jobs)
	wg.Wait()
	close(results)

	byFile := make([][]crap.Entry, len(files))
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		byFile[result.index] = result.entries
	}
	var entries []crap.Entry
	for _, fileEntries := range byFile {
		entries = append(entries, fileEntries...)
	}
	return crap.SortByCRAP(entries), nil
}

func entriesForFile(file string, profile map[string][]coverage.Segment) ([]crap.Entry, error) {
	functions, err := complexity.ExtractFunctions(file)
	if err != nil {
		return nil, err
	}
	var entries []crap.Entry
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
	return entries, nil
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
