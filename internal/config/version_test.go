package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// PackageJSON represents the structure of package.json
type PackageJSON struct {
	Version string `json:"version"`
}

// TestVersionConsistency ensures all version declarations match
func TestVersionConsistency(t *testing.T) {
	// Read version from package.json
	packageJSONPath := filepath.Join("..", "..", "app", "package.json")

	// Check if package.json exists (test might be run from different directories)
	if _, err := os.Stat(packageJSONPath); os.IsNotExist(err) {
		t.Skip("package.json not found, skipping version consistency test")
		return
	}

	packageData, err := os.ReadFile(packageJSONPath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(packageData, &pkg); err != nil {
		t.Fatalf("Failed to parse package.json: %v", err)
	}

	// Compare with backend version
	if pkg.Version != Version {
		t.Errorf("Version mismatch!\n  Frontend (package.json): %s\n  Backend (version.go): %s",
			pkg.Version, Version)
	}

	t.Logf("Version consistency check passed: %s", Version)
}

// TestVersionFormat ensures version follows CalVer format
func TestVersionFormat(t *testing.T) {
	// Test version format: YYYY.MM.PATCH
	if len(Version) < 8 {
		t.Errorf("Version format too short: %s", Version)
		return
	}

	// Basic format validation
	if Version[4] != '.' || Version[7] != '.' {
		t.Errorf("Version format invalid: %s (expected YYYY.MM.PATCH)", Version)
	}

	t.Logf("Version format check passed: %s", Version)
}

// TestBuildInfo ensures build info structure is correct
func TestBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo("test-sha", "2025-07-06T10:30:00Z")

	if buildInfo.Version != Version {
		t.Errorf("BuildInfo version mismatch: got %s, want %s", buildInfo.Version, Version)
	}

	if buildInfo.Service != ServiceName {
		t.Errorf("BuildInfo service mismatch: got %s, want %s", buildInfo.Service, ServiceName)
	}

	if buildInfo.GitSHA != "test-sha" {
		t.Errorf("BuildInfo GitSHA mismatch: got %s, want %s", buildInfo.GitSHA, "test-sha")
	}

	if buildInfo.BuildTime != "2025-07-06T10:30:00Z" {
		t.Errorf("BuildInfo BuildTime mismatch: got %s, want %s", buildInfo.BuildTime, "2025-07-06T10:30:00Z")
	}

	t.Logf("BuildInfo structure check passed")
}
