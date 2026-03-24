package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

type adapterRequest struct {
	root         string
	explicitRoot bool
	tool         string
}

var adapterToolPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

func parseAdapterSyncArgs(args []string) (adapterRequest, error) {
	request := adapterRequest{root: "."}
	positionals := make([]string, 0, 2)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown adapter sync flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return adapterRequest{}, err
	}
	if len(positionals) == 0 || len(positionals) > 2 {
		return adapterRequest{}, fmt.Errorf("adapter sync requires <tool> and optional [path]")
	}
	if err := assignAdapterTool(&request, positionals[0]); err != nil {
		return adapterRequest{}, err
	}
	if len(positionals) == 2 {
		if request.explicitRoot {
			return adapterRequest{}, fmt.Errorf("cannot use both --path and a positional path argument")
		}
		request.root = positionals[1]
		request.explicitRoot = true
	}
	return request, nil
}

func assignAdapterTool(request *adapterRequest, value string) error {
	request.tool = strings.TrimSpace(value)
	if request.tool == "" {
		return fmt.Errorf("adapter sync tool must not be empty")
	}
	if request.tool == "." || request.tool == ".." {
		return fmt.Errorf("adapter sync tool %q is invalid", request.tool)
	}
	if strings.Contains(request.tool, "/") || strings.Contains(request.tool, "\\") {
		return fmt.Errorf("adapter sync tool %q must not contain path separators", request.tool)
	}
	if !adapterToolPattern.MatchString(request.tool) {
		return fmt.Errorf("adapter sync tool %q must match %s", request.tool, adapterToolPattern)
	}
	cleaned := filepath.Clean(request.tool)
	if cleaned != request.tool {
		return fmt.Errorf("adapter sync tool %q is invalid", request.tool)
	}
	return nil
}
