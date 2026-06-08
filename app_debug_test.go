package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestBuildBugReportURL_ContainsVersion(t *testing.T) {
	raw := buildBugReportURL("v1.0.0-rc.1", "windows", "amd64", "")
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	if got := u.Query().Get("version"); got != "v1.0.0-rc.1" {
		t.Errorf("version param = %q, want %q", got, "v1.0.0-rc.1")
	}
}

func TestBuildBugReportURL_TemplateParam(t *testing.T) {
	raw := buildBugReportURL("dev", "linux", "amd64", "")
	u, _ := url.Parse(raw)
	if got := u.Query().Get("template"); got != "bug_report.yml" {
		t.Errorf("template param = %q, want %q", got, "bug_report.yml")
	}
}

func TestBuildBugReportURL_ContainsOperatingSystem(t *testing.T) {
	raw := buildBugReportURL("dev", "windows", "amd64", "")
	u, _ := url.Parse(raw)
	if got := u.Query().Get("operating-system"); got != "windows/amd64" {
		t.Errorf("operating-system param = %q, want %q", got, "windows/amd64")
	}
}

func TestBuildBugReportURL_LogsIncluded(t *testing.T) {
	tail := "some log output"
	raw := buildBugReportURL("dev", "linux", "amd64", tail)
	u, _ := url.Parse(raw)
	if got := u.Query().Get("logs"); got != tail {
		t.Errorf("logs param = %q, want %q", got, tail)
	}
}

func TestBuildBugReportURL_EmptyLogsTailOmitted(t *testing.T) {
	raw := buildBugReportURL("dev", "linux", "amd64", "")
	u, _ := url.Parse(raw)
	if _, ok := u.Query()["logs"]; ok {
		t.Error("logs param should be absent when log tail is empty")
	}
}

func TestBuildBugReportURL_LogsCappedAt4000(t *testing.T) {
	// Simulate a large log file tail already capped by the caller.
	// buildBugReportURL itself does not cap; verify the URL is still valid
	// when a large log string is passed.
	largeTail := strings.Repeat("x", 4000)
	raw := buildBugReportURL("dev", "darwin", "arm64", largeTail)
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("URL with large logs is invalid: %v", err)
	}
	if got := u.Query().Get("logs"); len(got) != 4000 {
		t.Errorf("logs param length = %d, want 4000", len(got))
	}
}

func TestBuildBugReportURL_BaseURL(t *testing.T) {
	raw := buildBugReportURL("dev", "linux", "amd64", "")
	if !strings.HasPrefix(raw, bugReportRepo) {
		t.Errorf("URL does not start with repo base %q: %s", bugReportRepo, raw)
	}
}
