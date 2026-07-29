package cluster

import (
	"io"
	"testing"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
)

func TestCreateCommand_RegionRequired(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestCreateCommand_DryRunAndPayloadMutuallyExclusive(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster", "--dry-run", "--payload", "cluster.json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
	want := "cannot use both --dry-run and --payload flags"
	if err.Error() != want {
		t.Errorf("Execute() error = %q, want %q", err.Error(), want)
	}
}

func TestCreateCommand_RequiresClusterName(t *testing.T) {
	t.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error, got nil")
	}
}
