package oidcconfig

import (
	"context"
	"fmt"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/oidcconfig"
	"github.com/spf13/cobra"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete CONFIG_ID",
		Short: "Delete an OIDC configuration",
		Long: `Delete an OIDC configuration.

This will delete the OIDC configuration from the platform API. If any clusters
are still using this configuration, the deletion will fail.

Example:
  rosactl oidc-config delete 24ab3cd7`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configID := args[0]
			return runDelete(cmd.Context(), configID)
		},
	}

	return cmd
}

func runDelete(ctx context.Context, configID string) error {
	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	platformURL, err := pkgconfig.GetPlatformAPIURL()
	if err != nil {
		return fmt.Errorf("failed to get platform API URL: %w", err)
	}

	fmt.Printf("Deleting OIDC configuration %s...\n", configID)

	req := &oidcconfig.DeleteOidcConfigRequest{
		ID:             configID,
		PlatformAPIURL: platformURL,
		AWSConfig:      cfg,
	}

	err = oidcconfig.DeleteOidcConfig(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("OIDC configuration %s deleted successfully\n", configID)

	return nil
}
