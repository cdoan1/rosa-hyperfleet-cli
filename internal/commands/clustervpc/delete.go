package clustervpc

import (
	"context"
	"fmt"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/clustervpc"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	clusterName string
	region      string
	noWait      bool
}

func newDeleteCommand() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete CLUSTER_NAME",
		Short: "Delete cluster VPC resources",
		Long: `Delete VPC networking resources for a hosted cluster.

This command deletes the CloudFormation stack containing all VPC resources.

Example:
  rosactl cluster-vpc delete my-cluster --region us-east-1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internalaws.RequireRegion(); err != nil {
				return err
			}
			opts.clusterName = args[0]
			opts.region = internalaws.Region()
			return runDelete(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.noWait, "no-wait", false, "Return immediately without waiting for stack deletion to complete")

	return cmd
}

func runDelete(ctx context.Context, opts *deleteOptions) error {
	fmt.Printf("🗑️  Deleting cluster VPC resources for: %s\n", opts.clusterName)
	fmt.Printf("   Region: %s\n", opts.region)
	fmt.Println()

	// Load AWS config
	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create service request
	req := &clustervpc.DeleteVPCRequest{
		ClusterName: opts.clusterName,
		NoWait:      opts.noWait,
		AWSConfig:   cfg,
	}

	fmt.Printf("☁️  Deleting CloudFormation stack: rosa-%s-vpc\n", opts.clusterName)
	if !opts.noWait {
		fmt.Println("   This may take several minutes...")
	}
	fmt.Println()

	// Call service layer
	err = clustervpc.DeleteVPC(ctx, req)
	if err != nil {
		return err
	}

	if opts.noWait {
		fmt.Println("✅ Stack deletion submitted!")
		fmt.Println()
		fmt.Printf("💡 Stack is being deleted asynchronously. Check status with:\n")
		fmt.Printf("   rosactl cluster-vpc describe %s --region %s\n", opts.clusterName, opts.region)
	} else {
		fmt.Println("✅ Cluster VPC resources deleted successfully!")
	}

	return nil
}
