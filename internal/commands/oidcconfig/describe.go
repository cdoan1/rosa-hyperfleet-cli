package oidcconfig

import (
	"context"
	"fmt"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/oidcconfig"
	"github.com/spf13/cobra"
)

func newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe CONFIG_ID",
		Short: "Describe an OIDC configuration",
		Long: `Show detailed information about an OIDC configuration.

Example:
  rosactl oidc-config describe 24ab3cd7`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configID := args[0]
			return runDescribe(cmd.Context(), configID)
		},
	}

	return cmd
}

func runDescribe(ctx context.Context, configID string) error {
	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	platformURL, err := pkgconfig.GetPlatformAPIURL()
	if err != nil {
		return fmt.Errorf("failed to get platform API URL: %w", err)
	}

	req := &oidcconfig.GetOidcConfigRequest{
		ID:             configID,
		PlatformAPIURL: platformURL,
		AWSConfig:      cfg,
	}

	resp, err := oidcconfig.GetOidcConfig(ctx, req)
	if err != nil {
		return err
	}

	config := resp.OidcConfig

	fmt.Println("OIDC Configuration Details:")
	fmt.Printf("   Config ID:           %s\n", config.Name)
	fmt.Printf("   Type:                %s\n", config.Spec.Type)
	fmt.Printf("   Issuer URL:          %s\n", config.Spec.IssuerUrl)
	if config.Status.Phase != "" {
		fmt.Printf("   Phase:               %s\n", config.Status.Phase)
	}
	if config.Status.Thumbprint != "" {
		fmt.Printf("   Thumbprint:          %s\n", config.Status.Thumbprint)
	}
	if config.Spec.SecretArn != "" {
		fmt.Printf("   Secret ARN:          %s\n", config.Spec.SecretArn)
	}
	if config.Spec.InstallerRoleArn != "" {
		fmt.Printf("   Installer Role ARN:  %s\n", config.Spec.InstallerRoleArn)
	}
	if config.Status.LastUsedTimestamp != nil {
		fmt.Printf("   Last Used:           %s\n", config.Status.LastUsedTimestamp.Time.Format("2006-01-02 15:04:05"))
	}
	fmt.Printf("   Resource Version:    %s\n", config.ResourceVersion)
	fmt.Printf("   Generation:          %d\n", config.Generation)

	if len(config.Status.Conditions) > 0 {
		fmt.Println("\nConditions:")
		for _, cond := range config.Status.Conditions {
			fmt.Printf("   %s:\n", cond.Type)
			fmt.Printf("      Status:  %s\n", cond.Status)
			if cond.Reason != "" {
				fmt.Printf("      Reason:  %s\n", cond.Reason)
			}
			if cond.Message != "" {
				fmt.Printf("      Message: %s\n", cond.Message)
			}
			fmt.Printf("      Last Transition: %s\n", cond.LastTransitionTime.Format("2006-01-02 15:04:05"))
		}
	}

	return nil
}
