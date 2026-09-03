package oidcconfig

import (
	"github.com/spf13/cobra"
)

// NewOidcConfigCommand creates the oidc-config command
func NewOidcConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oidc-config",
		Short: "Manage OIDC configurations",
		Long: `Manage OIDC configurations for ROSA hosted clusters.

OIDC configurations pre-allocate OIDC issuer URLs and CloudFront distributions,
enabling the OIDC-first cluster creation flow that eliminates the 10-15 minute
IAM eventual consistency delay.

Typical OIDC-first workflow:
  1. rosactl oidc-config create --type managed
  2. rosactl cluster-oidc create my-cluster --oidc-config-id <id> --region us-east-1
  3. rosactl cluster-iam create my-cluster --oidc-issuer-url <url> --region us-east-1
  4. rosactl cluster create my-cluster --oidc-config-id <id> --region us-east-1

For more information, see: docs/guides/OIDC-FIRST-FLOW.md`,
	}

	cmd.AddCommand(newCreateCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newDescribeCommand())
	cmd.AddCommand(newDeleteCommand())

	return cmd
}
