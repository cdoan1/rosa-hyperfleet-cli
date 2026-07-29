package config

import (
	"os"
	"path/filepath"
	"testing"

	testutils "github.com/openshift-online/rosa-regional-platform-cli/tests/utils"
)

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "rosactl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", originalHome) }()
	_ = os.Setenv("HOME", tmpDir)

	// Test saving a config
	testURL := "https://api.example.com"
	err = SetPlatformAPIURL(testURL)
	if err != nil {
		t.Fatalf("Failed to set platform API URL: %v", err)
	}

	// Verify the config file was created
	configPath := filepath.Join(tmpDir, configDir, configFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Test loading the config
	url, err := GetPlatformAPIURL()
	if err != nil {
		t.Fatalf("Failed to get platform API URL: %v", err)
	}

	if url != testURL {
		t.Errorf("Expected URL %q, got %q", testURL, url)
	}
}

func TestGetPlatformAPIURL_NotConfigured(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "rosactl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Override the home directory for testing
	originalHome := os.Getenv("HOME")
	defer func() { _ = os.Setenv("HOME", originalHome) }()
	_ = os.Setenv("HOME", tmpDir)

	// Test getting URL when not configured
	_, err = GetPlatformAPIURL()
	if err == nil {
		t.Error("Expected error when platform API URL is not configured")
	}
}

func TestGetRegion(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantRegion string
		wantErr    bool
	}{
		{
			name:       "standard API host",
			url:        "https://api.us-east-1.example.com",
			wantRegion: "us-east-1",
		},
		{
			name:       "execute-api URL",
			url:        "https://abc123.execute-api.ap-southeast-2.amazonaws.com",
			wantRegion: "ap-southeast-2",
		},
		{
			name:       "us-west-2 region",
			url:        "https://platform.us-west-2.internal.example.com",
			wantRegion: "us-west-2",
		},
		{
			name:       "GovCloud region",
			url:        "https://api.us-gov-west-1.example.com",
			wantRegion: "us-gov-west-1",
		},
		{
			name:    "no region in URL",
			url:     "https://api.example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "rosactl-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer func() { _ = os.RemoveAll(tmpDir) }()

			defer testutils.RestoreEnv("HOME", os.Getenv("HOME"))
			_ = os.Setenv("HOME", tmpDir)

			if err := SetPlatformAPIURL(tt.url); err != nil {
				t.Fatalf("SetPlatformAPIURL() error: %v", err)
			}

			region, err := GetRegion()
			if tt.wantErr {
				if err == nil {
					t.Error("GetRegion() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("GetRegion() unexpected error: %v", err)
			}
			if region != tt.wantRegion {
				t.Errorf("GetRegion() = %q, want %q", region, tt.wantRegion)
			}
		})
	}
}

func TestGetRegion_NotConfigured(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rosactl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	defer testutils.RestoreEnv("HOME", os.Getenv("HOME"))
	_ = os.Setenv("HOME", tmpDir)

	_, err = GetRegion()
	if err == nil {
		t.Error("GetRegion() expected error when platform URL not configured")
	}
}
