package nodepool

import (
	"io"
	"testing"
)

func TestNodePoolCreate_ClusterIDRequired(t *testing.T) {
	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-np"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing --cluster-id, got nil")
	}
	want := "--cluster-id is required"
	if err.Error() != want {
		t.Errorf("Execute() error = %q, want %q", err.Error(), want)
	}
}

func TestNodePoolCreate_ReplicasMustBePositive(t *testing.T) {
	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"my-np", "--cluster-id", "abc123", "--replicas", "0"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for --replicas 0, got nil")
	}
	want := "--replicas must be at least 1"
	if err.Error() != want {
		t.Errorf("Execute() error = %q, want %q", err.Error(), want)
	}
}

func TestNodePoolCreate_RequiresName(t *testing.T) {
	cmd := newCreateCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--cluster-id", "abc123"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing node pool name, got nil")
	}
}

func TestNodePoolList_ClusterIDRequired(t *testing.T) {
	cmd := newListCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing --cluster-id, got nil")
	}
	want := "--cluster-id is required"
	if err.Error() != want {
		t.Errorf("Execute() error = %q, want %q", err.Error(), want)
	}
}

func TestNodePoolList_LimitBounds(t *testing.T) {
	tests := []struct {
		name  string
		limit string
		want  string
	}{
		{"zero", "0", "--limit must be between 1 and 100"},
		{"over max", "101", "--limit must be between 1 and 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newListCommand()
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)
			cmd.SetArgs([]string{"--cluster-id", "abc123", "--limit", tt.limit})

			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute() expected error, got nil")
			}
			if err.Error() != tt.want {
				t.Errorf("Execute() error = %q, want %q", err.Error(), tt.want)
			}
		})
	}
}

func TestNodePoolDelete_RequiresNodePoolID(t *testing.T) {
	cmd := newDeleteCommand()
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() expected error for missing nodepool ID, got nil")
	}
}
