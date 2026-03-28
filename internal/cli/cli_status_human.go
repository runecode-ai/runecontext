package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/runecode-systems/runecontext/internal/contracts"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiDim    = "\x1b[2m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiBlue   = "\x1b[34m"
	ansiCyan   = "\x1b[36m"
	ansiRed    = "\x1b[31m"
)

type statusRenderOptions struct {
	color   bool
	explain bool
}

func renderHumanStatus(absRoot string, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary, options statusRenderOptions) string {
	if summary == nil {
		return ""
	}
	var builder strings.Builder
	appendStatusHeader(&builder, absRoot, summary, options)
	appendStatusSection(&builder, "In Flight", summary.Active, options)
	appendStatusSection(&builder, "Recently Completed", summary.Closed, options)
	appendStatusSection(&builder, "Replaced", summary.Superseded, options)
	if options.explain {
		appendStatusExplainHuman(&builder, loaded, summary, options)
	}
	return builder.String()
}

func appendStatusHeader(builder *strings.Builder, absRoot string, summary *contracts.ProjectStatusSummary, options statusRenderOptions) {
	builder.WriteString(styleStatusText("RuneContext Status", ansiBold+ansiBlue, options.color))
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Root: %s\n", absRoot))
	builder.WriteString(fmt.Sprintf("Config: %s\n", summary.SelectedConfigPath))
	builder.WriteString(fmt.Sprintf("Version: %s  Assurance: %s\n", summary.RuneContextVersion, summary.AssuranceTier))
	builder.WriteString(renderBundleSummary(summary.BundleIDs))
	builder.WriteString("\n\n")
}

func renderBundleSummary(bundleIDs []string) string {
	if len(bundleIDs) == 0 {
		return "Bundles (0): none"
	}
	return fmt.Sprintf("Bundles (%d): %s", len(bundleIDs), strings.Join(bundleIDs, ", "))
}

func appendStatusSection(builder *strings.Builder, title string, entries []contracts.ChangeStatusEntry, options statusRenderOptions) {
	builder.WriteString(styleStatusText(title, ansiBold, options.color))
	builder.WriteString(fmt.Sprintf(" (%d)\n", len(entries)))
	if len(entries) == 0 {
		builder.WriteString("  (none)\n\n")
		return
	}
	for _, entry := range entries {
		appendStatusEntry(builder, entry, options)
	}
	builder.WriteString("\n")
}

func appendStatusEntry(builder *strings.Builder, entry contracts.ChangeStatusEntry, options statusRenderOptions) {
	builder.WriteString(fmt.Sprintf("- %s [%s %s] %s  %s\n", entry.ID, emptyAsDash(entry.Type), emptyAsDash(entry.Size), entry.Title, renderVerificationBadge(entry.VerificationStatus, options.color)))
	relationshipLines := statusRelationshipLines(entry)
	for i, item := range relationshipLines {
		prefix := "  |-- "
		if i == len(relationshipLines)-1 {
			prefix = "  `-- "
		}
		builder.WriteString(prefix + item + "\n")
	}
}

func statusRelationshipLines(entry contracts.ChangeStatusEntry) []string {
	lines := make([]string, 0, 7)
	appendStatusRelationLine(&lines, "depends on", entry.DependsOn)
	appendStatusRelationLine(&lines, "related", entry.RelatedChanges)
	appendStatusRelationLine(&lines, "supersedes", entry.Supersedes)
	appendStatusRelationLine(&lines, "superseded by", entry.SupersededBy)
	if entry.CreatedAt != "" {
		lines = append(lines, fmt.Sprintf("created: %s", entry.CreatedAt))
	}
	if entry.ClosedAt != "" {
		lines = append(lines, fmt.Sprintf("closed: %s", entry.ClosedAt))
	}
	lines = append(lines, fmt.Sprintf("path: %s", entry.Path))
	return lines
}

func appendStatusRelationLine(lines *[]string, label string, ids []string) {
	if len(ids) == 0 {
		return
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	*lines = append(*lines, fmt.Sprintf("%s: %s", label, strings.Join(sorted, ", ")))
}

func renderVerificationBadge(status string, useColor bool) string {
	label := fmt.Sprintf("[%s]", emptyAsDash(status))
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "passed":
		return styleStatusText(label, ansiGreen, useColor)
	case "pending":
		return styleStatusText(label, ansiYellow, useColor)
	case "failed":
		return styleStatusText(label, ansiRed, useColor)
	case "skipped":
		return styleStatusText(label, ansiCyan, useColor)
	default:
		return styleStatusText(label, ansiDim, useColor)
	}
}

func styleStatusText(value, code string, enabled bool) string {
	if !enabled || value == "" {
		return value
	}
	return code + value + ansiReset
}

func emptyAsDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func appendStatusExplainHuman(builder *strings.Builder, loaded *contracts.LoadedProject, summary *contracts.ProjectStatusSummary, options statusRenderOptions) {
	lines := appendStatusExplainLines(nil, loaded, summary)
	if len(lines) == 0 {
		return
	}
	builder.WriteString(styleStatusText("Explain", ansiBold, options.color))
	builder.WriteString("\n")
	for _, item := range lines {
		builder.WriteString(fmt.Sprintf("- %s: %s\n", item.key, item.value))
	}
	builder.WriteString("\n")
}

func shouldUseStatusColor(w io.Writer) bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if term := strings.TrimSpace(strings.ToLower(os.Getenv("TERM"))); term == "" || term == "dumb" {
		return false
	}
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
