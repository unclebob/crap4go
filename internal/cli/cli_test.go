package cli

import "testing"

func TestParseHelp(t *testing.T) {
	options := ParseArgs([]string{"--help"})
	if options.Action != "help" {
		t.Fatalf("expected help, got %q", options.Action)
	}
	if options.Message == "" {
		t.Fatal("expected help message")
	}
}

func TestParseFilters(t *testing.T) {
	options := ParseArgs([]string{"internal/foo", "cmd"})
	if options.Action != "analyze" {
		t.Fatalf("expected analyze, got %q", options.Action)
	}
	if len(options.ModuleFilters) != 2 || options.ModuleFilters[0] != "internal/foo" {
		t.Fatalf("unexpected filters: %#v", options.ModuleFilters)
	}
}
