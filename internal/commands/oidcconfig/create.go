package oidcconfig

import (
	"context"
	"fmt"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/oidcconfig"
	"github.com/spf13/cobra"
)

type createOptions struct {
	configType string
	region     string
}

func newCreateCommand() *cobra.Command {
	opts := &createOptions{}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an OIDC configuration",
		Long: `Create an OIDC configuration for ROSA hosted clusters.

This command creates an OIDC configuration via the platform API, which pre-allocates
an OIDC issuer URL and CloudFront distribution. This enables the OIDC-first cluster
creation flow that eliminates the 10-15 minute IAM eventual consistency delay.

The platform API will:
1. Generate a random OIDC config ID (e.g., "24ab3cd7")
2. Provision CloudFront distribution and S3 bucket
3. Construct issuer URL (e.g., "https://d1234.cloudfront.net/24ab3cd7")
4. Fetch TLS thumbprint from CloudFront

Example:
  rosactl oidc-config create --type managed --region us-east-1

Next steps:
  1. Create OIDC provider in your AWS account:
     rosactl cluster-oidc create my-cluster --oidc-config-id <id> --region us-east-1

  2. Create IAM roles:
     rosactl cluster-iam create my-cluster --oidc-issuer-url <url> --region us-east-1

  3. Create cluster:
     rosactl cluster create my-cluster --oidc-config-id <id> --region us-east-1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := internalaws.RequireRegion(); err != nil {
				return err
			}
			opts.region = internalaws.Region()
			return runCreate(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.configType, "type", "managed", "OIDC config type (managed or unmanaged)")

	return cmd
}

func runCreate(ctx context.Context, opts *createOptions) error {
	if opts.configType != "managed" && opts.configType != "unmanaged" {
		return fmt.Errorf("invalid type %q, must be 'managed' or 'unmanaged'", opts.configType)
	}

	if opts.configType == "unmanaged" {
		return fmt.Errorf("unmanaged OIDC configs are not yet supported")
	}

	fmt.Println("Creating OIDC configuration...")
	fmt.Printf("   Type: %s\n", opts.configType)
	fmt.Printf("   Region: %s\n", opts.region)
	fmt.Println()

	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	platformURL, err := pkgconfig.GetPlatformAPIURL()
	if err != nil {
		return fmt.Errorf("failed to get platform API URL: %w", err)
	}

	req := &oidcconfig.CreateOidcConfigRequest{
		Type:           opts.configType,
		Region:         opts.region,
		PlatformAPIURL: platformURL,
		AWSConfig:      cfg,
	}

	resp, err := oidcconfig.CreateOidcConfig(ctx, req)
	if err != nil {
		return err
	}

	config := resp.OidcConfig

	fmt.Println("OIDC configuration created successfully!")
	fmt.Printf("   Config ID:    %s\n", config.Name)
	if config.Spec.IssuerUrl != "" {
		fmt.Printf("   Issuer URL:   %s\n", config.Spec.IssuerUrl)
	}
	if config.Status.Thumbprint != "" {
		fmt.Printf("   Thumbprint:   %s\n", config.Status.Thumbprint)
	}
	if config.Status.Phase != "" {
		fmt.Printf("   Phase:        %s\n", config.Status.Phase)
	}
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Println("  1. Create OIDC provider in your AWS account:")
	fmt.Printf("     rosactl cluster-oidc create my-cluster --oidc-config-id %s --region %s\n", config.Name, opts.region)
	fmt.Println()
	fmt.Println("  2. Create IAM roles:")
	if config.Spec.IssuerUrl != "" {
		fmt.Printf("     rosactl cluster-iam create my-cluster --oidc-issuer-url %s --region %s\n", config.Spec.IssuerUrl, opts.region)
	} else {
		fmt.Printf("     rosactl cluster-iam create my-cluster --oidc-issuer-url <issuer-url> --region %s\n", opts.region)
	}
	fmt.Println()
	fmt.Println("  3. Create cluster:")
	fmt.Printf("     rosactl cluster create my-cluster --oidc-config-id %s --region %s\n", config.Name, opts.region)

	return nil
}
