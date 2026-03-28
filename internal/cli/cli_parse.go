package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

type statusRequest struct {
	root            string
	explicitRoot    bool
	historyMode     string
	historyModeSet  bool
	historyLimit    int
	historyLimitSet bool
	verbose         bool
}

type changeNewRequest struct {
	root           string
	explicitRoot   bool
	title          string
	changeType     string
	size           string
	sizeProvided   bool
	description    string
	mode           string
	modeProvided   bool
	contextBundles []string
	bundleProvided bool
}

type changeShapeRequest struct {
	root         string
	explicitRoot bool
	changeID     string
	design       string
	verification string
	tasks        []string
	references   []string
}

type changeCloseRequest struct {
	root               string
	explicitRoot       bool
	changeID           string
	verificationStatus string
	closedAt           time.Time
	supersededBy       []string
}

type changeReallocateRequest struct {
	root         string
	explicitRoot bool
	changeID     string
}

func parseStatusArgs(args []string) (statusRequest, error) {
	if len(args) == 1 && isHelpToken(args[0]) {
		return statusRequest{root: args[0], explicitRoot: true}, nil
	}
	if len(args) > 1 && isHelpToken(args[0]) {
		return statusRequest{}, fmt.Errorf("help does not accept additional arguments")
	}
	request := statusRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		case "--history":
			next, err := assignStatusHistoryModeFlag(args, flag, &request.historyMode)
			if err == nil {
				request.historyModeSet = true
			}
			return next, err
		case "--history-limit":
			next, err := assignStatusHistoryLimitFlag(args, flag, &request.historyLimit)
			if err == nil {
				request.historyLimitSet = true
			}
			return next, err
		case "--verbose":
			if err := requireNoValue("--verbose", flag.hasValue); err != nil {
				return flag.next, err
			}
			request.verbose = true
			return flag.next, nil
		default:
			return flag.next, fmt.Errorf("unknown status flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return statusRequest{}, err
	}
	if strings.TrimSpace(request.historyMode) == "" {
		request.historyMode = statusHistoryModeRecent
	}
	if request.historyLimit == 0 {
		request.historyLimit = defaultStatusHistoryLimit
	}
	baseRequest, err := finalizeOptionalPath(request.root, request.explicitRoot, positionals)
	if err != nil {
		return statusRequest{}, err
	}
	request.root = baseRequest.root
	request.explicitRoot = baseRequest.explicitRoot
	return request, nil
}

func assignStatusHistoryModeFlag(args []string, flag parsedFlag, target *string) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	mode := strings.ToLower(strings.TrimSpace(value))
	switch mode {
	case statusHistoryModeRecent, statusHistoryModeAll, statusHistoryModeNone:
		*target = mode
		return next, nil
	default:
		return flag.next, fmt.Errorf("--history must be one of recent, all, or none")
	}
}

func assignStatusHistoryLimitFlag(args []string, flag parsedFlag, target *int) (int, error) {
	value, next, err := flag.requireValue(args)
	if err != nil {
		return flag.next, err
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || parsed < 1 {
		return flag.next, fmt.Errorf("--history-limit must be a positive integer")
	}
	*target = parsed
	return next, nil
}

func parseChangeNewArgs(args []string) (changeNewRequest, error) {
	request := changeNewRequest{root: "."}
	err := consumeArgs(args, changeNewFlagHandler(args, &request), func(arg string) error {
		return fmt.Errorf("unexpected positional argument %q", arg)
	})
	if err != nil {
		return changeNewRequest{}, err
	}
	return finalizeChangeNewRequest(request)
}

func changeNewFlagHandler(args []string, request *changeNewRequest) func(parsedFlag) (int, error) {
	return func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--title":
			return assignStringFlag(args, flag, &request.title)
		case "--type":
			return assignStringFlag(args, flag, &request.changeType)
		case "--size":
			return assignStringFlagWithProvided(args, flag, &request.size, &request.sizeProvided)
		case "--description":
			return assignStringFlag(args, flag, &request.description)
		case "--shape":
			return assignStringFlagWithProvided(args, flag, &request.mode, &request.modeProvided)
		case "--bundle":
			return appendStringFlagWithProvided(args, flag, &request.contextBundles, &request.bundleProvided)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change new flag %q", flag.raw)
		}
	}
}

func assignStringFlagWithProvided(args []string, flag parsedFlag, target *string, provided *bool) (int, error) {
	next, err := assignStringFlag(args, flag, target)
	if err == nil {
		*provided = true
	}
	return next, err
}

func appendStringFlagWithProvided(args []string, flag parsedFlag, target *[]string, provided *bool) (int, error) {
	next, err := appendStringFlag(args, flag, target)
	if err == nil {
		*provided = true
	}
	return next, err
}

func finalizeChangeNewRequest(request changeNewRequest) (changeNewRequest, error) {
	if strings.TrimSpace(request.title) == "" {
		return changeNewRequest{}, fmt.Errorf("--title is required")
	}
	if strings.TrimSpace(request.changeType) == "" {
		return changeNewRequest{}, fmt.Errorf("--type is required")
	}
	if request.mode != "" && request.mode != string(contracts.ChangeModeMinimum) && request.mode != string(contracts.ChangeModeFull) {
		return changeNewRequest{}, fmt.Errorf("--shape must be %q or %q", contracts.ChangeModeMinimum, contracts.ChangeModeFull)
	}
	return request, nil
}

func parseChangeShapeArgs(args []string) (changeShapeRequest, error) {
	request := changeShapeRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--design":
			return assignStringFlag(args, flag, &request.design)
		case "--verification":
			return assignStringFlag(args, flag, &request.verification)
		case "--task":
			return appendStringFlag(args, flag, &request.tasks)
		case "--reference":
			return appendStringFlag(args, flag, &request.references)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change shape flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeShapeRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change shape requires exactly one change ID")
	if err != nil {
		return changeShapeRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}

func parseChangeCloseArgs(args []string) (changeCloseRequest, error) {
	request := changeCloseRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		switch flag.name {
		case "--verification-status":
			return assignStringFlag(args, flag, &request.verificationStatus)
		case "--superseded-by":
			return appendStringFlag(args, flag, &request.supersededBy)
		case "--closed-at":
			return assignClosedAtFlag(args, flag, &request.closedAt)
		case "--path":
			return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
		default:
			return flag.next, fmt.Errorf("unknown change close flag %q", flag.raw)
		}
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeCloseRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change close requires exactly one change ID")
	if err != nil {
		return changeCloseRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}

func parseChangeReallocateArgs(args []string) (changeReallocateRequest, error) {
	request := changeReallocateRequest{root: "."}
	positionals := make([]string, 0, 1)
	err := consumeArgs(args, func(flag parsedFlag) (int, error) {
		if flag.name != "--path" {
			return flag.next, fmt.Errorf("unknown change reallocate flag %q", flag.raw)
		}
		return assignRootFlag(args, flag, &request.root, &request.explicitRoot)
	}, func(arg string) error {
		positionals = append(positionals, arg)
		return nil
	})
	if err != nil {
		return changeReallocateRequest{}, err
	}
	changeID, err := requireExactPositional(positionals, "change reallocate requires exactly one change ID")
	if err != nil {
		return changeReallocateRequest{}, err
	}
	request.changeID = changeID
	return request, nil
}
