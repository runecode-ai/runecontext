package cli

import (
	"fmt"
	"strings"
)

type machineOptions struct {
	jsonOutput     bool
	nonInteractive bool
	dryRun         bool
	explain        bool
	explainWarning string
}

type machineFlagConfig struct {
	allowDryRun   bool
	allowExplain  bool
	explainNotYet bool
}

func parseMachineFlags(args []string, config machineFlagConfig) (machineOptions, []string, error) {
	options := machineOptions{}
	remaining := make([]string, 0, len(args))
	for _, arg := range args {
		name, _, hasValue := strings.Cut(arg, "=")
		switch name {
		case "--json":
			if hasValue {
				return machineOptions{}, nil, fmt.Errorf("--json does not accept a value")
			}
			options.jsonOutput = true
		case "--non-interactive":
			if hasValue {
				return machineOptions{}, nil, fmt.Errorf("--non-interactive does not accept a value")
			}
			options.nonInteractive = true
		case "--dry-run":
			if hasValue {
				return machineOptions{}, nil, fmt.Errorf("--dry-run does not accept a value")
			}
			if !config.allowDryRun {
				return machineOptions{}, nil, fmt.Errorf("--dry-run is only supported for write commands")
			}
			options.dryRun = true
		case "--explain":
			if hasValue {
				return machineOptions{}, nil, fmt.Errorf("--explain does not accept a value")
			}
			if !config.allowExplain {
				return machineOptions{}, nil, fmt.Errorf("--explain is not supported for this command")
			}
			options.explain = true
			if config.explainNotYet {
				options.explainWarning = "--explain is not yet implemented for this command; output will not include explanation details"
			}
		default:
			remaining = append(remaining, arg)
		}
	}
	return options, remaining, nil
}
