package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadProfileReadsFileAndHandlesMissingFile(t *testing.T) {
	missing, err := LoadProfile(filepath.Join(t.TempDir(), "missing.out"))
	if err != nil || missing != nil {
		t.Fatalf("expected nil profile and nil error for missing file, got %#v, %v", missing, err)
	}

	path := filepath.Join(t.TempDir(), "coverage.out")
	content := "mode: set\ngithub.com/unclebob/example/foo.go:1.1,2.2 1 1\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	profile, err := LoadProfile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(profile["github.com/unclebob/example/foo.go"]) != 1 {
		t.Fatalf("expected one segment, got %#v", profile)
	}
}

func TestParseProfileRejectsMalformedLines(t *testing.T) {
	cases := []string{
		"badline",
		"nocolon 1 1",
		"foo.go:1.1 1 1",
		"foo.go:bad,2.2 1 1",
		"foo.go:x.1,2.2 1 1",
		"foo.go:1.x,2.2 1 1",
		"foo.go:1.1,bad 1 1",
		"foo.go:1.1,2.2 x 1",
		"foo.go:1.1,2.2 1 x",
	}
	for _, tc := range cases {
		if _, err := ParseProfile(strings.NewReader(tc)); err == nil {
			t.Fatalf("expected error for malformed line %q", tc)
		}
	}
}

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
