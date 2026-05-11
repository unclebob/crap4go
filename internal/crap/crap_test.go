package crap

import (
	"strings"
	"testing"
)

func TestScore(t *testing.T) {
	coverage := 50.0
	score := Score(4, &coverage)
	if score == nil {
		t.Fatal("expected score")
	}
	if *score != 6.0 {
		t.Fatalf("expected 6.0, got %.1f", *score)
	}
}

func TestScoreIsNilForIndeterminateCoverage(t *testing.T) {
	if Score(4, nil) != nil {
		t.Fatal("expected nil score")
	}
}

func TestSortByCRAP(t *testing.T) {
	high := 10.0
	low := 1.0
	entries := []Entry{
		{Name: "low", CRAP: &low},
		{Name: "none"},
		{Name: "high", CRAP: &high},
	}
	sorted := SortByCRAP(entries)
	if sorted[0].Name != "high" || sorted[1].Name != "low" || sorted[2].Name != "none" {
		t.Fatalf("unexpected order: %#v", sorted)
	}
}

func TestFormatReport(t *testing.T) {
	coverage := 100.0
	score := 1.0
	report := FormatReport([]Entry{{
		Name: "Simple", Package: "sample", Complexity: 1, Coverage: &coverage, CRAP: &score,
	}})
	if !strings.Contains(report, "CRAP Report") {
		t.Fatalf("missing title: %s", report)
	}
	if !strings.Contains(report, "Simple") || !strings.Contains(report, "100.0%") {
		t.Fatalf("missing entry data: %s", report)
	}
}
