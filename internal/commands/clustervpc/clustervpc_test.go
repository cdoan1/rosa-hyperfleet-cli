package clustervpc

import (
	"io"
	"os"
	"testing"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	testutils "github.com/openshift-online/rosa-regional-platform-cli/tests/utils"
)

func TestClusterVPCCreate_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestClusterVPCCreate_RequiresClusterName(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing cluster name, got nil")
	}
}

func TestClusterVPCDelete_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-cluster"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestClusterVPCDelete_RequiresClusterName(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Setenv(internalaws.EnvRegion, "us-east-1")

	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing cluster name, got nil")
	}
}

func TestClusterVPCList_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newListCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}
