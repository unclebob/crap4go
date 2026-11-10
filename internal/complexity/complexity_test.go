package complexity

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractFunctions(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := `package sample

func Simple() int {
	return 1
}

func Branchy(x int) int {
	if x > 0 {
		return x
	}
	return 0
}
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	functions, err := ExtractFunctions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) != 2 {
		t.Fatalf("expected two functions, got %#v", functions)
	}
	if functions[0].Name != "Simple" || functions[0].Complexity != 1 {
		t.Fatalf("unexpected simple function: %#v", functions[0])
	}
	if functions[1].Name != "Branchy" || functions[1].Complexity != 2 {
		t.Fatalf("unexpected branchy function: %#v", functions[1])
	}
}

func TestExtractMethodName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := `package sample

type Widget struct{}

func (w *Widget) Run() {
	for i := 0; i < 10; i++ {
	}
}
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	functions, err := ExtractFunctions(path)
	if err != nil {
		t.Fatal(err)
	}
	if functions[0].Name != "Widget.Run" {
		t.Fatalf("unexpected method name: %q", functions[0].Name)
	}
}

func TestExtractGenericAndSelectorMethodNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "generic.go")
	source := `package sample

type Box[T any] struct{}
type Pair[K comparable, V any] struct{}

func (b *Box[T]) Get() {}
func (p Pair[K, V]) Put() {}
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	functions, err := ExtractFunctions(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(functions) != 2 || functions[0].Name != "Box.Get" || functions[1].Name != "Pair.Put" {
		t.Fatalf("unexpected generic method names: %#v", functions)
	}
}

func TestCyclomaticComplexityCountsGoDecisionPoints(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	source := `package sample

func Branchy(x int, ch chan int) int {
	if x > 0 && x < 10 {
		return 1
	}
	switch x {
	case 1, 2:
		return 2
	default:
		return 3
	}
}
`
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	functions, err := ExtractFunctions(path)
	if err != nil {
		t.Fatal(err)
	}
	if functions[0].Complexity != 5 {
		t.Fatalf("expected complexity 5, got %d", functions[0].Complexity)
	}
}
