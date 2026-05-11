package crap

import (
	"fmt"
	"sort"
	"strings"
)

type Entry struct {
	Name       string
	Package    string
	Complexity int
	Coverage   *float64
	CRAP       *float64
}

func Score(complexity int, coveragePct *float64) *float64 {
	if coveragePct == nil {
		return nil
	}
	cc := float64(complexity)
	uncov := 1.0 - (*coveragePct / 100.0)
	score := cc*cc*uncov*uncov*uncov + cc
	return &score
}

func SortByCRAP(entries []Entry) []Entry {
	out := append([]Entry(nil), entries...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].CRAP, out[j].CRAP
		if a == nil && b == nil {
			return out[i].Name < out[j].Name
		}
		if a == nil {
			return false
		}
		if b == nil {
			return true
		}
		return *a > *b
	})
	return out
}

func FormatReport(entries []Entry) string {
	header := fmt.Sprintf("%-30s %-35s %4s %7s %8s", "Function", "Package", "CC", "Cov%", "CRAP")
	sep := strings.Repeat("-", len(header))
	lines := []string{"CRAP Report", "===========", header, sep}
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%-30s %-35s %4d %7s %8s",
			e.Name, e.Package, e.Complexity, formatCoverage(e.Coverage), formatScore(e.CRAP)))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func formatCoverage(coverage *float64) string {
	if coverage == nil {
		return "  N/A "
	}
	return fmt.Sprintf("%5.1f%%", *coverage)
}

func formatScore(score *float64) string {
	if score == nil {
		return "     N/A"
	}
	return fmt.Sprintf("%8.1f", *score)
}
