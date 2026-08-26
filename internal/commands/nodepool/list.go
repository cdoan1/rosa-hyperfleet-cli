package nodepool

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	"github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/spf13/cobra"
)

type listOptions struct {
	clusterID string
	limit     int
	offset    int
	output    string
}

func newListCommand() *cobra.Command {
	opts := &listOptions{
		limit:  50,
		offset: 0,
	}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List node pools for a cluster",
		Long: `List node pools for a ROSA hosted cluster.

Examples:
  rosactl nodepool list --cluster-id <id>
  rosactl nodepool list --cluster-id <id> --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.clusterID == "" {
				return fmt.Errorf("--cluster-id is required")
			}
			if opts.limit < 1 || opts.limit > 100 {
				return fmt.Errorf("--limit must be between 1 and 100")
			}
			if opts.offset < 0 {
				return fmt.Errorf("--offset must be non-negative")
			}
			return runList(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.clusterID, "cluster-id", "", "Cluster ID (required)")
	cmd.Flags().IntVar(&opts.limit, "limit", opts.limit, "Maximum number of nodepools to return (1-100)")
	cmd.Flags().IntVar(&opts.offset, "offset", opts.offset, "Number of nodepools to skip")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "table", "Output format: table or json")

	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
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

	// List nodepools via clientset (namespace = cluster-<uuid> format)
	listOpts := platform.ListOptions{
		Limit:  int64(opts.limit),
		Offset: int64(opts.offset),
	}

	namespace := "cluster-" + opts.clusterID
	nodepoolList, err := cs.HyperfleetV1alpha1().NodePools(namespace).List(ctx, listOpts)
	if err != nil {
		return fmt.Errorf("failed to list nodepools: %w", err)
	}

	if opts.output == "json" {
		prettyJSON, err := json.MarshalIndent(nodepoolList.Items, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(prettyJSON))
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(w, "ID\tNAME\tREPLICAS\tINSTANCE_TYPE\tPHASE"); err != nil {
		return err
	}

	for _, np := range nodepoolList.Items {
		replicas := "-"
		instanceType := "-"
		phase := string(np.Status.Phase)

		if np.Spec.NodePool.Replicas != nil {
			replicas = fmt.Sprintf("%d", *np.Spec.NodePool.Replicas)
		}
		if np.Spec.NodePool.Platform.AWS != nil && np.Spec.NodePool.Platform.AWS.InstanceType != "" {
			instanceType = np.Spec.NodePool.Platform.AWS.InstanceType
		}
		if phase == "" {
			phase = "-"
		}

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			string(np.UID), np.Name, replicas, instanceType, phase); err != nil {
			return err
		}
	}

	return w.Flush()
}
