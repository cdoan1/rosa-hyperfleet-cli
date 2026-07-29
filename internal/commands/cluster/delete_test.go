package cluster

import (
	"io"
	"testing"
)

func TestClusterDelete_RequiresClusterName(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing cluster name, got nil")
	}
}
