package main

import (
	"log/slog"
	"net/url"
	goruntime "runtime"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"smb-tools/internal/logger"
)

const bugReportRepo = "https://github.com/tbrittain/smb-tools/issues/new"
const bugReportLogMaxBytes = 4000

// LogFrontendError records a Vue or unhandled-promise error from the frontend
// in the session log file. Fire-and-forget — no return value.
func (a *App) LogFrontendError(message, stack, context string) {
	slog.Error("frontend error", "context", context, "message", message, "stack", stack)
}

// OpenBugReport assembles a pre-filled GitHub bug report URL and opens it in
// the user's default browser. When includeSystemInfo is true, the OS and
// architecture are added as URL parameters.
func (a *App) OpenBugReport(includeSystemInfo bool) error {
	logTail := logger.TailFile(a.logFilePath, bugReportLogMaxBytes)
	issueURL := buildBugReportURL(a.version, goruntime.GOOS, goruntime.GOARCH, logTail, includeSystemInfo)
	runtime.BrowserOpenURL(a.ctx, issueURL)
	return nil
}

// buildBugReportURL is a pure function that assembles the GitHub issue pre-fill
// URL. Extracted for testability (no Wails context required).
func buildBugReportURL(version, goos, goarch, logTail string, includeSystemInfo bool) string {
	params := url.Values{}
	params.Set("template", "bug_report.yml")
	params.Set("version", version)
	if includeSystemInfo {
		params.Set("os", goos+"/"+goarch)
	}
	if logTail != "" {
		params.Set("logs", logTail)
	}
	return bugReportRepo + "?" + params.Encode()
}
