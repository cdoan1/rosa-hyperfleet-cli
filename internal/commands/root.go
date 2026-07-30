package commands

import (
	"fmt"
	"os"

	internalaws "github.com/openshift-online/rosa-regional-platform-cli/internal/aws"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/bootstrap"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/cluster"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/clusteriam"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/clusteroidc"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/clustervpc"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/handler"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/login"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/nodepool"
	"github.com/openshift-online/rosa-regional-platform-cli/internal/commands/version"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
	"github.com/spf13/cobra"
)

var verbose bool

var rootCmd = &cobra.Command{
	Use:   "rosactl",
	Short: "CLI tool for managing AWS resources",
	Long:  "🚀 rosactl is a command-line interface for ROSA Hyperfleet.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if region, _ := cmd.Flags().GetString("region"); region != "" {
			_ = os.Setenv(internalaws.EnvRegion, region)
		} else if region, err := pkgconfig.GetRegion(); err == nil {
			_ = os.Setenv(internalaws.EnvRegion, region)
		}
		if profile, _ := cmd.Flags().GetString("profile"); profile != "" {
			_ = os.Setenv(internalaws.EnvProfile, profile)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().String("region", "", "AWS region (overrides default)")
	rootCmd.PersistentFlags().String("profile", "", "AWS profile (overrides default)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	rootCmd.AddCommand(bootstrap.NewBootstrapCommand())
	rootCmd.AddCommand(cluster.NewClusterCommand())
	rootCmd.AddCommand(clusteriam.NewClusterIAMCommand())
	rootCmd.AddCommand(clusteroidc.NewClusterOIDCCommand())
	rootCmd.AddCommand(clustervpc.NewClusterVPCCommand())
	rootCmd.AddCommand(handler.NewHandlerCommand())
	rootCmd.AddCommand(nodepool.NewNodePoolCommand())
	rootCmd.AddCommand(login.NewLoginCommand())
	rootCmd.AddCommand(version.NewVersionCommand())
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func IsVerbose() bool {
	return verbose
}
