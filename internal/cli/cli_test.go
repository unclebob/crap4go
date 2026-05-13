package cli

import "testing"

func TestParseDefaultsMaxWorkersToHalfAvailableCPUs(t *testing.T) {
	options := ParseArgs(nil)
	if options.MaxWorkers != defaultMaxWorkers() {
		t.Fatalf("expected default workers %d, got %d", defaultMaxWorkers(), options.MaxWorkers)
	}
	if options.MaxWorkers < 1 {
		t.Fatalf("expected at least one worker, got %d", options.MaxWorkers)
	}
}

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

func TestParseMaxWorkers(t *testing.T) {
	options := ParseArgs([]string{"--max-workers", "4", "internal/foo"})
	if options.Action != "analyze" {
		t.Fatalf("expected analyze, got %q", options.Action)
	}
	if options.MaxWorkers != 4 {
		t.Fatalf("expected four workers, got %d", options.MaxWorkers)
	}
	if len(options.ModuleFilters) != 1 || options.ModuleFilters[0] != "internal/foo" {
		t.Fatalf("unexpected filters: %#v", options.ModuleFilters)
	}
}

func TestParseMaxWorkersRequiresPositiveInteger(t *testing.T) {
	options := ParseArgs([]string{"--max-workers", "0"})
	if options.Action != "error" {
		t.Fatalf("expected error, got %q", options.Action)
	}
}
