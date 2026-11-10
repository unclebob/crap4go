package coverage

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Segment struct {
	File       string
	StartLine  int
	StartCol   int
	EndLine    int
	EndCol     int
	Statements int
	Count      int
}

func LoadProfile(path string) (map[string][]Segment, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	return ParseProfile(f)
}

func ParseProfile(r io.Reader) (map[string][]Segment, error) {
	out := map[string][]Segment{}
	scanner := bufio.NewScanner(r)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		segment, err := parseProfileLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNumber, err)
		}
		out[segment.File] = append(out[segment.File], segment)
	}
	return out, scanner.Err()
}

func CoverageForRange(profile map[string][]Segment, file string, startLine, endLine int) *float64 {
	segments := segmentsForFile(profile, file)
	if len(segments) == 0 {
		return nil
	}
	total := 0
	covered := 0
	for _, segment := range segments {
		if segment.EndLine < startLine || segment.StartLine > endLine {
			continue
		}
		total += segment.Statements
		if segment.Count > 0 {
			covered += segment.Statements
		}
	}
	if total == 0 {
		zero := 0.0
		return &zero
	}
	pct := 100.0 * float64(covered) / float64(total)
	return &pct
}

func segmentsForFile(profile map[string][]Segment, file string) []Segment {
	normalized := normalizePath(file)
	if segments := profile[normalized]; len(segments) > 0 {
		return segments
	}
	for candidate, segments := range profile {
		if suffixMatch(normalizePath(candidate), normalized) {
			return segments
		}
	}
	return nil
}

func suffixMatch(path, suffix string) bool {
	pathParts := pathSegments(path)
	suffixParts := pathSegments(suffix)
	if len(suffixParts) == 0 || len(suffixParts) > len(pathParts) {
		return false
	}
	offset := len(pathParts) - len(suffixParts)
	for i := range suffixParts {
		if pathParts[offset+i] != suffixParts[i] {
			return false
		}
	}
	return true
}

func normalizePath(path string) string {
	return strings.TrimPrefix(strings.ReplaceAll(path, "\\", "/"), "./")
}

func pathSegments(path string) []string {
	parts := strings.Split(normalizePath(path), "/")
	var out []string
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseProfileLine(line string) (Segment, error) {
	fields := strings.Fields(line)
	if len(fields) != 3 {
		return Segment{}, fmt.Errorf("invalid coverage segment %q", line)
	}
	file, rangePart, found := strings.Cut(fields[0], ":")
	if !found {
		return Segment{}, fmt.Errorf("missing file separator")
	}
	startLine, startCol, endLine, endCol, err := parseRange(rangePart)
	if err != nil {
		return Segment{}, err
	}
	statements, err := strconv.Atoi(fields[1])
	if err != nil {
		return Segment{}, fmt.Errorf("invalid statement count: %w", err)
	}
	count, err := strconv.Atoi(fields[2])
	if err != nil {
		return Segment{}, fmt.Errorf("invalid execution count: %w", err)
	}
	return Segment{
		File: file, StartLine: startLine, StartCol: startCol, EndLine: endLine,
		EndCol: endCol, Statements: statements, Count: count,
	}, nil
}

func parseRange(rangePart string) (int, int, int, int, error) {
	startPart, endPart, found := strings.Cut(rangePart, ",")
	if !found {
		return 0, 0, 0, 0, fmt.Errorf("missing range separator")
	}
	startLine, startCol, err := parsePosition(startPart)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	endLine, endCol, err := parsePosition(endPart)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	return startLine, startCol, endLine, endCol, nil
}

func parsePosition(value string) (int, int, error) {
	linePart, colPart, found := strings.Cut(value, ".")
	if !found {
		return 0, 0, fmt.Errorf("invalid position %q", value)
	}
	line, err := strconv.Atoi(linePart)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid line: %w", err)
	}
	col, err := strconv.Atoi(colPart)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid column: %w", err)
	}
	return line, col, nil
}
