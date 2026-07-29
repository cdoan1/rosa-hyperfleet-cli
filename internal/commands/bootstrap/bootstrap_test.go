package bootstrap

import (
	"io"
	"os"
	"testing"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	testutils "github.com/openshift-online/rosa-regional-platform-cli/tests/utils"
)

func TestBootstrapCreate_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	// --image-uri is MarkFlagRequired; provide it so Cobra's required-flag check
	// passes and RunE fires where RequireRegion() is checked.
	cmd.SetArgs([]string{"--image-uri", "111122223333.dkr.ecr.us-east-1.amazonaws.com/myapp:latest"})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestBootstrapDelete_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}

func TestBootstrapStatus_RegionRequired(t *testing.T) {
	defer testutils.RestoreEnv(internalaws.EnvRegion, os.Getenv(internalaws.EnvRegion))
	os.Unsetenv(internalaws.EnvRegion)

	cmd := newStatusCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != internalaws.ErrRegionRequired {
		t.Errorf("Execute() = %v, want ErrRegionRequired", err)
	}
}
