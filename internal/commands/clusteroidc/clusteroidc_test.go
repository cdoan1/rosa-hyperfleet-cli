package clusteroidc

import (
	"io"
	"testing"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
)

func TestClusterOIDCCreate_RequiresEitherFlag(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error when neither --oidc-issuer-url nor --oidc-config-id provided, got nil")
	}
}

func TestClusterOIDCCreate_MutuallyExclusiveFlags(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster", "--oidc-issuer-url", "https://example.com", "--oidc-config-id", "abc123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error when both --oidc-issuer-url and --oidc-config-id provided, got nil")
	}
}

func TestClusterOIDCCreate_RegionRequired(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster", "--oidc-issuer-url", "https://oidc.example.com/my-cluster"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestClusterOIDCCreate_RequiresClusterName(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--oidc-issuer-url", "https://oidc.example.com/my-cluster"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing cluster name, got nil")
	}
}

func TestClusterOIDCDelete_RegionRequired(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "")

	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestClusterOIDCDelete_RequiresClusterName(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing cluster name, got nil")
	}
}

func TestClusterOIDCList_RegionRequired(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "")

	cmd := newListCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}
