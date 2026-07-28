package clustervpc

import (
	"context"
	"fmt"
	"strings"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/aws/cloudformation"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster VPC stacks",
		Long: `List all cluster VPC CloudFormation stacks in the current AWS account.

Example:
  rosactl cluster-vpc list --region us-east-1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internalaws.RequireRegion(); err != nil {
				return err
			}
			return runList(cmd.Context())
		},
	}

	return cmd
}

func runList(ctx context.Context) error {
	// Load AWS config
	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create CloudFormation client
	cfnClient := cloudformation.NewClient(cfg)

	// List stacks with prefix
	stacks, err := cfnClient.ListStacks(ctx, "rosa-")
	if err != nil {
		return fmt.Errorf("failed to list stacks: %w", err)
	}

	// Filter VPC stacks
	var vpcStacks []cloudformation.StackInfo
	for _, stack := range stacks {
		if strings.HasSuffix(stack.StackName, "-vpc") {
			vpcStacks = append(vpcStacks, stack)
		}
	}

	if len(vpcStacks) == 0 {
		fmt.Println("No cluster VPC stacks found.")
		return nil
	}

	// Print table
	fmt.Printf("%-30s %-20s %-30s\n", "CLUSTER NAME", "STATUS", "CREATED")
	fmt.Println(strings.Repeat("-", 82))
	for _, stack := range vpcStacks {
		clusterName := strings.TrimPrefix(stack.StackName, "rosa-")
		clusterName = strings.TrimSuffix(clusterName, "-vpc")
		createdTime := ""
		if stack.CreationTime != nil {
			createdTime = stack.CreationTime.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-30s %-20s %-30s\n", clusterName, stack.Status, createdTime)
	}

	return nil
}
