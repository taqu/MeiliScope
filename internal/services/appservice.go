package services

import (
	"fmt"
	"runtime"
	"time"
)

const appVersion = "0.1.0-dev"

// AppVersion returns the application version string for use outside the package.
func AppVersion() string { return appVersion }

// ApplicationInfo contains metadata about the application.
type ApplicationInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
}

// RuntimeInfo contains information about the current runtime environment.
type RuntimeInfo struct {
	GoVersion string `json:"goVersion"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	StartedAt string `json:"startedAt"`
}

// AppService exposes application-level information to the frontend.
type AppService struct {
	startedAt time.Time
}

// NewAppService creates an AppService with the application start time.
func NewAppService() *AppService {
	return &AppService{startedAt: time.Now()}
}

// GetApplicationInfo returns static application metadata.
func (s *AppService) GetApplicationInfo() ApplicationInfo {
	return ApplicationInfo{
		Name:        "MeiliScope",
		Version:     appVersion,
		Description: "Desktop browser for Meilisearch",
	}
}

// GetApplicationVersion returns the application version string.
func (s *AppService) GetApplicationVersion() string {
	return appVersion
}

// GetRuntimeInformation returns Go runtime and environment details.
func (s *AppService) GetRuntimeInformation() RuntimeInfo {
	return RuntimeInfo{
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		StartedAt: s.startedAt.Format(time.RFC3339),
	}
}

// GetFormattedUptime returns a human-readable uptime string.
func (s *AppService) GetFormattedUptime() string {
	d := time.Since(s.startedAt).Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, sec)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, sec)
	}
	return fmt.Sprintf("%ds", sec)
}
