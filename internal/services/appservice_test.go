package services

import (
	"strings"
	"testing"
	"time"
)

func TestGetApplicationInfo(t *testing.T) {
	svc := NewAppService()
	info := svc.GetApplicationInfo()
	if info.Name == "" {
		t.Error("expected non-empty Name")
	}
	if info.Version == "" {
		t.Error("expected non-empty Version")
	}
	if info.Description == "" {
		t.Error("expected non-empty Description")
	}
}

func TestGetApplicationVersion(t *testing.T) {
	svc := NewAppService()
	v := svc.GetApplicationVersion()
	if v == "" {
		t.Error("expected non-empty version")
	}
	if v != appVersion {
		t.Errorf("expected %q, got %q", appVersion, v)
	}
}

func TestGetRuntimeInformation(t *testing.T) {
	svc := NewAppService()
	info := svc.GetRuntimeInformation()
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("expected GoVersion to start with 'go', got %q", info.GoVersion)
	}
	if info.OS == "" {
		t.Error("expected non-empty OS")
	}
	if info.Arch == "" {
		t.Error("expected non-empty Arch")
	}
	if info.StartedAt == "" {
		t.Error("expected non-empty StartedAt")
	}
}

func TestGetFormattedUptime(t *testing.T) {
	svc := &AppService{startedAt: time.Now().Add(-90 * time.Second)}
	uptime := svc.GetFormattedUptime()
	if uptime == "" {
		t.Error("expected non-empty uptime")
	}
	// 90s = 1m 30s
	if !strings.Contains(uptime, "m") {
		t.Errorf("expected minutes in uptime %q", uptime)
	}
}
