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

func TestParseTestCommand(t *testing.T) {
	options := ParseArgs([]string{"--test-command", "go test -tags appunit ./internal/app", "internal/app"})
	if options.Action != "analyze" {
		t.Fatalf("expected analyze, got %q", options.Action)
	}
	if options.TestCommand != "go test -tags appunit ./internal/app" {
		t.Fatalf("unexpected test command: %q", options.TestCommand)
	}
	if len(options.ModuleFilters) != 1 || options.ModuleFilters[0] != "internal/app" {
		t.Fatalf("unexpected filters: %#v", options.ModuleFilters)
	}
}

func TestParseMaxWorkersRequiresPositiveInteger(t *testing.T) {
	options := ParseArgs([]string{"--max-workers", "0"})
	if options.Action != "error" {
		t.Fatalf("expected error, got %q", options.Action)
	}
	missing := ParseArgs([]string{"--max-workers"})
	if missing.Action != "error" {
		t.Fatalf("expected error for missing worker value, got %q", missing.Action)
	}
}

func TestParseTestCommandRequiresCommand(t *testing.T) {
	options := ParseArgs([]string{"--test-command"})
	if options.Action != "error" {
		t.Fatalf("expected error, got %q", options.Action)
	}
	empty := ParseArgs([]string{"--test-command", ""})
	if empty.Action != "error" {
		t.Fatalf("expected error for empty test command, got %q", empty.Action)
	}
}
