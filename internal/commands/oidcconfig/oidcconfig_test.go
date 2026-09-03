package oidcconfig

import (
	"io"
	"testing"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
)

func TestOidcConfigCreate_RegionRequired(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--type", "managed"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestOidcConfigCreate_InvalidType(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--type", "invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for invalid type, got nil")
	}
}

func TestOidcConfigDescribe_RequiresConfigID(t *testing.T) {
	cmd := newDescribeCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing config ID, got nil")
	}
}

func TestOidcConfigDelete_RequiresConfigID(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing config ID, got nil")
	}
}
