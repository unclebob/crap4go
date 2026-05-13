package cli

import (
	"runtime"
	"strconv"
)

type Options struct {
	Action        string
	Message       string
	ModuleFilters []string
	MaxWorkers    int
}

const HelpMessage = `Usage: crap4go [module-filter ...]

Runs go test coverage, computes CRAP scores, and prints a report sorted worst first.

Options:
  -h, --help         Print this help message and exit.
  --max-workers N    Analyze source files in parallel using N workers.
                     Defaults to half the number of logical CPUs.

Arguments:
  module-filter    Optional source path fragment. When present, only matching
                   .go source files are analyzed.`

func ParseArgs(args []string) Options {
	options := Options{Action: "analyze", MaxWorkers: defaultMaxWorkers()}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-h" || arg == "--help" {
			return Options{Action: "help", Message: HelpMessage}
		}
		if arg == "--max-workers" {
			i++
			if i >= len(args) {
				return Options{Action: "error", Message: "--max-workers requires a positive integer"}
			}
			workers, err := strconv.Atoi(args[i])
			if err != nil || workers < 1 {
				return Options{Action: "error", Message: "--max-workers requires a positive integer"}
			}
			options.MaxWorkers = workers
			continue
		}
		options.ModuleFilters = append(options.ModuleFilters, arg)
	}
	return options
}

func defaultMaxWorkers() int {
	workers := runtime.NumCPU() / 2
	if workers < 1 {
		return 1
	}
	return workers
}
