package aws

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegion(t *testing.T) {
	t.Setenv(EnvRegion, "us-west-2")
	if got := Region(); got != "us-west-2" {
		t.Errorf("Region() = %q, want %q", got, "us-west-2")
	}

	t.Setenv(EnvRegion, "")
	if got := Region(); got != "" {
		t.Errorf("Region() = %q, want empty string when unset", got)
	}
}

func TestProfile(t *testing.T) {
	t.Setenv(EnvProfile, "my-profile")
	if got := Profile(); got != "my-profile" {
		t.Errorf("Profile() = %q, want %q", got, "my-profile")
	}

	t.Setenv(EnvProfile, "")
	if got := Profile(); got != "" {
		t.Errorf("Profile() = %q, want empty string when unset", got)
	}
}

func TestRequireRegion(t *testing.T) {
	// Start with EnvRegion empty (os.Getenv returns "" for both unset and "").
	t.Setenv(EnvRegion, "")

	if err := RequireRegion(); err != ErrRegionRequired {
		t.Errorf("RequireRegion() = %v, want ErrRegionRequired", err)
	}

	t.Setenv(EnvRegion, "us-east-1")
	if err := RequireRegion(); err != nil {
		t.Errorf("RequireRegion() = %v, want nil", err)
	}
}

func TestNewConfig_UnknownProfile(t *testing.T) {
	// Point the SDK at an empty config file so no profiles exist.
	emptyConfig, err := os.CreateTemp(t.TempDir(), "aws-config-*")
	if err != nil {
		t.Fatal(err)
	}
	if err := emptyConfig.Close(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AWS_CONFIG_FILE", emptyConfig.Name())
	t.Setenv(EnvProfile, "nonexistent-profile")

	_, err = NewConfig(context.Background())
	if err == nil {
		t.Fatal("NewConfig() expected error for unknown profile, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-profile") {
		t.Errorf("error should mention profile name, got: %v", err)
	}
}

func TestNewConfig_ProfileInCustomConfigFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Isolate from the real ~/.aws/config so only the custom file is seen.
	t.Setenv("HOME", tmpDir)

	// Write a custom config file containing the profile we want to validate.
	configFile := filepath.Join(tmpDir, "aws-config")
	if err := os.WriteFile(configFile, []byte("[profile my-profile]\nregion = eu-central-1\n"), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AWS_CONFIG_FILE", configFile)
	t.Setenv(EnvProfile, "my-profile")
	t.Setenv(EnvRegion, "eu-central-1")

	if _, err := NewConfig(context.Background()); err != nil {
		t.Fatalf("NewConfig() unexpected error for profile in custom AWS_CONFIG_FILE: %v", err)
	}
}

func TestNewConfig_RegionApplied(t *testing.T) {
	t.Setenv(EnvRegion, "eu-west-1")
	t.Setenv(EnvProfile, "") // empty means no profile is selected

	cfg, err := NewConfig(context.Background())
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("cfg.Region = %q, want %q", cfg.Region, "eu-west-1")
	}
}
