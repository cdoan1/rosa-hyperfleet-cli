package oidcconfig

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/services/oidcconfig"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List OIDC configurations",
		Long: `List all OIDC configurations for your account.

Example:
  rosactl oidc-config list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context())
		},
	}

	return cmd
}

func runList(ctx context.Context) error {
	cfg, err := internalaws.NewConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	platformURL, err := pkgconfig.GetPlatformAPIURL()
	if err != nil {
		return fmt.Errorf("failed to get platform API URL: %w", err)
	}

	req := &oidcconfig.ListOidcConfigsRequest{
		PlatformAPIURL: platformURL,
		AWSConfig:      cfg,
	}

	resp, err := oidcconfig.ListOidcConfigs(ctx, req)
	if err != nil {
		return err
	}

	if len(resp.Items) == 0 {
		fmt.Println("No OIDC configurations found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tISSUER URL\tPHASE\tTHUMBPRINT")
	for _, config := range resp.Items {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			config.Name,
			config.Spec.Type,
			config.Spec.IssuerUrl,
			config.Status.Phase,
			config.Status.Thumbprint,
		)
	}
	w.Flush()

	return nil
}
