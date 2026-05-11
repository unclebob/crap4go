package coverage

import (
	"strings"
	"testing"
)

func TestParseProfile(t *testing.T) {
	profile, err := ParseProfile(strings.NewReader(`mode: set
github.com/unclebob/example/internal/foo/foo.go:3.10,5.2 2 1
github.com/unclebob/example/internal/foo/foo.go:8.1,9.2 1 0
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(profile) != 1 {
		t.Fatalf("expected one file, got %#v", profile)
	}
	segments := profile["github.com/unclebob/example/internal/foo/foo.go"]
	if len(segments) != 2 {
		t.Fatalf("expected two segments, got %#v", segments)
	}
}

func TestCoverageForRangeUsesStatementWeightedCoverage(t *testing.T) {
	profile := map[string][]Segment{
		"foo.go": {
			{StartLine: 1, EndLine: 2, Statements: 2, Count: 1},
			{StartLine: 3, EndLine: 4, Statements: 1, Count: 0},
		},
	}
	coverage := CoverageForRange(profile, "foo.go", 1, 4)
	if coverage == nil || *coverage != 100.0*2.0/3.0 {
		t.Fatalf("unexpected coverage: %v", coverage)
	}
}

func TestCoverageForRangeMatchesModuleQualifiedFileBySuffix(t *testing.T) {
	profile := map[string][]Segment{
		"github.com/unclebob/example/internal/foo/foo.go": {
			{StartLine: 1, EndLine: 1, Statements: 1, Count: 1},
		},
	}
	coverage := CoverageForRange(profile, "internal/foo/foo.go", 1, 1)
	if coverage == nil || *coverage != 100.0 {
		t.Fatalf("unexpected coverage: %v", coverage)
	}
}

func TestCoverageForRangeIsNilWhenFileMissing(t *testing.T) {
	if CoverageForRange(map[string][]Segment{}, "foo.go", 1, 2) != nil {
		t.Fatal("expected nil coverage")
	}
}
