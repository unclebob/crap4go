package cli

type Options struct {
	Action        string
	Message       string
	ModuleFilters []string
}

const HelpMessage = `Usage: crap4go [module-filter ...]

Runs go test coverage, computes CRAP scores, and prints a report sorted worst first.

Options:
  -h, --help    Print this help message and exit.

Arguments:
  module-filter    Optional source path fragment. When present, only matching
                   .go source files are analyzed.`

func ParseArgs(args []string) Options {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return Options{Action: "help", Message: HelpMessage}
		}
	}
	return Options{Action: "analyze", ModuleFilters: append([]string(nil), args...)}
}
