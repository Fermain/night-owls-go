package config

import (
	"fmt"
	"runtime"
)

// Version information - using Community CalVer format: YYYY.MM.PATCH
// NOTE: This version should match app/package.json - consider automated sync
const (
	Version     = "2025.07.1"
	ServiceName = "night-owls-go"
)

// BuildInfo holds build-time information
type BuildInfo struct {
	Version   string `json:"version"`
	GitSHA    string `json:"git_sha,omitempty"`
	BuildTime string `json:"build_time,omitempty"`
	GoVersion string `json:"go_version"`
	Service   string `json:"service"`
}

// GetBuildInfo returns the current build information
func GetBuildInfo(gitSHA, buildTime string) BuildInfo {
	return BuildInfo{
		Version:   Version,
		GitSHA:    gitSHA,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
		Service:   ServiceName,
	}
}

// GetVersionString returns a formatted version string
func GetVersionString() string {
	return fmt.Sprintf("%s v%s", ServiceName, Version)
} 