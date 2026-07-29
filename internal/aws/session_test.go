package aws

import (
	"context"
	"os"
	"strings"
	"testing"

	testutils "github.com/openshift-online/rosa-regional-platform-cli/tests/utils"
)

func TestRegion(t *testing.T) {
	defer testutils.RestoreEnv(EnvRegion, os.Getenv(EnvRegion))

	os.Setenv(EnvRegion, "us-west-2")
	if got := Region(); got != "us-west-2" {
		t.Errorf("Region() = %q, want %q", got, "us-west-2")
	}

	os.Unsetenv(EnvRegion)
	if got := Region(); got != "" {
		t.Errorf("Region() = %q, want empty string when unset", got)
	}
}

func TestProfile(t *testing.T) {
	defer testutils.RestoreEnv(EnvProfile, os.Getenv(EnvProfile))

	os.Setenv(EnvProfile, "my-profile")
	if got := Profile(); got != "my-profile" {
		t.Errorf("Profile() = %q, want %q", got, "my-profile")
	}

	os.Unsetenv(EnvProfile)
	if got := Profile(); got != "" {
		t.Errorf("Profile() = %q, want empty string when unset", got)
	}
}

func TestRequireRegion(t *testing.T) {
	defer testutils.RestoreEnv(EnvRegion, os.Getenv(EnvRegion))

	os.Unsetenv(EnvRegion)
	if err := RequireRegion(); err != ErrRegionRequired {
		t.Errorf("RequireRegion() = %v, want ErrRegionRequired", err)
	}

	os.Setenv(EnvRegion, "us-east-1")
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
	emptyConfig.Close()

	defer testutils.RestoreEnv("AWS_CONFIG_FILE", os.Getenv("AWS_CONFIG_FILE"))
	os.Setenv("AWS_CONFIG_FILE", emptyConfig.Name())

	defer testutils.RestoreEnv(EnvProfile, os.Getenv(EnvProfile))
	os.Setenv(EnvProfile, "nonexistent-profile")

	_, err = NewConfig(context.Background())
	if err == nil {
		t.Fatal("NewConfig() expected error for unknown profile, got nil")
	}
	if !strings.Contains(err.Error(), "nonexistent-profile") {
		t.Errorf("error should mention profile name, got: %v", err)
	}
}

func TestNewConfig_RegionApplied(t *testing.T) {
	defer testutils.RestoreEnv(EnvRegion, os.Getenv(EnvRegion))
	defer testutils.RestoreEnv(EnvProfile, os.Getenv(EnvProfile))

	os.Setenv(EnvRegion, "eu-west-1")
	os.Unsetenv(EnvProfile)

	cfg, err := NewConfig(context.Background())
	if err != nil {
		t.Fatalf("NewConfig() unexpected error: %v", err)
	}
	if cfg.Region != "eu-west-1" {
		t.Errorf("cfg.Region = %q, want %q", cfg.Region, "eu-west-1")
	}
}
