package clusteroidc

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/clusteroidc"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/oidcconfig"
	"github.com/spf13/cobra"
)

type createOptions struct {
	clusterName    string
	oidcIssuerURL  string
	oidcThumbprint string
	oidcConfigID   string
	region         string
	noWait         bool
}

func newCreateCommand() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create CLUSTER_NAME",
		Short: "Create cluster OIDC provider",
		Long: `Create an IAM OIDC provider for a hosted cluster.

This command:
1. Fetches the TLS thumbprint from the OIDC issuer URL (unless --oidc-thumbprint is provided)
2. Creates a CloudFormation stack with the IAM OIDC provider (rosa-{cluster-name}-oidc)
3. Updates the IAM roles stack (rosa-{cluster-name}-iam) trust policies with the issuer domain

The IAM roles stack must already exist (created via 'rosactl cluster-iam create').

OIDC-first workflow (recommended):
  rosactl cluster-oidc create my-cluster \
    --oidc-config-id 24ab3cd7 \
    --region us-east-1

Traditional workflow:
  rosactl cluster-oidc create my-cluster \
    --oidc-issuer-url https://d1234.cloudfront.net/my-cluster \
    --region us-east-1`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internalaws.RequireRegion(); err != nil {
				return err
			}
			opts.clusterName = args[0]
			opts.region = internalaws.Region()
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.oidcIssuerURL, "oidc-issuer-url", "", "OIDC issuer URL from the cluster")
	cmd.Flags().StringVar(&opts.oidcThumbprint, "oidc-thumbprint", "", "TLS thumbprint (optional, fetched automatically if omitted)")
	cmd.Flags().StringVar(&opts.oidcConfigID, "oidc-config-id", "", "OIDC config ID (alternative to --oidc-issuer-url)")
	cmd.Flags().BoolVar(&opts.noWait, "no-wait", false, "Return immediately without waiting for stack creation to complete")

	return cmd
}

func runCreate(ctx context.Context, opts *createOptions) error {
	// Validate flags
	if opts.oidcConfigID == "" && opts.oidcIssuerURL == "" {
		return fmt.Errorf("either --oidc-config-id or --oidc-issuer-url must be provided")
	}
	if opts.oidcConfigID != "" && opts.oidcIssuerURL != "" {
		return fmt.Errorf("--oidc-config-id and --oidc-issuer-url are mutually exclusive")
	}

	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Resolve OIDC issuer URL and thumbprint
	var oidcIssuerURL, oidcThumbprint string
	if opts.oidcConfigID != "" {
		// OIDC-first flow: look up config from platform API
		fmt.Println("Creating cluster OIDC provider from OIDC config...")
		fmt.Printf("   Cluster: %s\n", opts.clusterName)
		fmt.Printf("   OIDC Config ID: %s\n", opts.oidcConfigID)
		fmt.Printf("   Region: %s\n", opts.region)
		fmt.Println()

		oidcIssuerURL, oidcThumbprint, err = resolveOIDCConfigDetails(ctx, opts.oidcConfigID, cfg)
		if err != nil {
			return err
		}

		fmt.Printf("Resolved from OIDC config:\n")
		fmt.Printf("   Issuer URL: %s\n", oidcIssuerURL)
		if oidcThumbprint != "" {
			fmt.Printf("   Thumbprint: %s\n", oidcThumbprint)
		}
		fmt.Println()
	} else {
		// Traditional flow: use provided issuer URL
		if !strings.HasPrefix(opts.oidcIssuerURL, "https://") {
			return fmt.Errorf("OIDC issuer URL must start with https://")
		}
		oidcIssuerURL = opts.oidcIssuerURL
		oidcThumbprint = opts.oidcThumbprint

		fmt.Println("Creating cluster OIDC provider...")
		fmt.Printf("   Cluster: %s\n", opts.clusterName)
		fmt.Printf("   OIDC Issuer: %s\n", oidcIssuerURL)
		fmt.Printf("   Region: %s\n", opts.region)
		fmt.Println()
	}

	req := &clusteroidc.CreateOIDCRequest{
		ClusterName:    opts.clusterName,
		OIDCIssuerURL:  oidcIssuerURL,
		OIDCThumbprint: oidcThumbprint,
		NoWait:         opts.noWait,
		AWSConfig:      cfg,
	}

	fmt.Printf("Creating CloudFormation stack: rosa-%s-oidc\n", opts.clusterName)
	if !opts.noWait {
		fmt.Println("   This may take a few minutes...")
	}
	fmt.Println()

	resp, err := clusteroidc.CreateOIDC(ctx, req)
	if err != nil {
		return err
	}

	if opts.noWait {
		fmt.Println("Stack creation submitted!")
		fmt.Printf("   Stack ID: %s\n", resp.StackID)
	} else {
		fmt.Println("Cluster OIDC provider created successfully!")
		fmt.Printf("   Stack ID: %s\n", resp.StackID)
		fmt.Println()

		if len(resp.Outputs) > 0 {
			fmt.Println("Created Resources:")
			for key, value := range resp.Outputs {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	return nil
}

func resolveOIDCConfigDetails(ctx context.Context, configID string, awsConfig aws.Config) (issuerURL, thumbprint string, err error) {
	platformURL, err := pkgconfig.GetPlatformAPIURL()
	if err != nil {
		return "", "", fmt.Errorf("failed to get platform API URL: %w", err)
	}

	req := &oidcconfig.GetOidcConfigRequest{
		ID:             configID,
		PlatformAPIURL: platformURL,
		AWSConfig:      awsConfig,
	}

	resp, err := oidcconfig.GetOidcConfig(ctx, req)
	if err != nil {
		return "", "", fmt.Errorf("failed to get OIDC config: %w", err)
	}

	if resp.OidcConfig.Spec.IssuerUrl == "" {
		return "", "", fmt.Errorf("OIDC config %s has no issuer URL", configID)
	}

	return resp.OidcConfig.Spec.IssuerUrl, resp.OidcConfig.Status.Thumbprint, nil
}
