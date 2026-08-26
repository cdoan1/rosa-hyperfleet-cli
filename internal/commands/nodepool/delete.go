package nodepool

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	"github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/spf13/cobra"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type deleteOptions struct {
	clusterID string
}

func newDeleteCommand() *cobra.Command {
	opts := &deleteOptions{}

	cmd := &cobra.Command{
		Use:   "delete NODEPOOL_ID",
		Short: "Delete a node pool",
		Long: `Delete a node pool from a ROSA hosted cluster.

Examples:
  rosactl nodepool delete <nodepool-id> --cluster-id <cluster-id> --region us-east-1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.clusterID == "" {
				return fmt.Errorf("--cluster-id is required")
			}
			return runDelete(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.clusterID, "cluster-id", "", "Cluster ID (required)")

	return cmd
}

func runDelete(ctx context.Context, nodepoolID string, opts *deleteOptions) error {
	baseURL, err := config.GetPlatformAPIURL()
	if err != nil {
		return err
	}

	cfg, err := aws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	region := cfg.Region
	if region == "" {
		return aws.ErrRegionRequired
	}

	// Load AWS config for clientset
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	accountID, err := config.GetAccountID()
	if err != nil {
		return fmt.Errorf("failed to get account ID: %w", err)
	}

	// Create clientset
	cs, err := hyperfleet.NewForConfig(&hfrest.Config{
		Host:      baseURL,
		AccountID: accountID,
		AWSConfig: awsCfg,
	})
	if err != nil {
		return fmt.Errorf("failed to create clientset: %w", err)
	}

	// Delete nodepool via clientset (namespace = cluster-<uuid> format)
	namespace := "cluster-" + opts.clusterID
	if err := cs.HyperfleetV1alpha1().NodePools(namespace).Delete(ctx, nodepoolID, platform.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("nodepool %q not found", nodepoolID)
		}
		return fmt.Errorf("failed to delete nodepool: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✓ NodePool %s deletion initiated\n", nodepoolID)
	return nil
}
