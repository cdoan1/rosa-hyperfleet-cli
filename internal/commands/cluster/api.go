package cluster

import (
	"context"
	"fmt"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	v1alpha1 "github.com/openshift-online/rosa-hyperfleet-api/api/v1alpha1/public"
	hyperfleet "github.com/openshift-online/rosa-hyperfleet-api/clientset"
	"github.com/openshift-online/rosa-hyperfleet-api/clientset/platform"
	hfrest "github.com/openshift-online/rosa-hyperfleet-api/clientset/rest"
	pkgconfig "github.com/openshift-online/rosa-regional-platform-cli/internal/config"
)

func fetchAPIURL(ctx context.Context, baseURL, clusterID string, creds awssdk.Credentials, region string) (string, error) {
	// Load AWS config for clientset
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config: %w", err)
	}

	accountID, err := pkgconfig.GetAccountID()
	if err != nil {
		return "", fmt.Errorf("failed to get account ID: %w", err)
	}

	// Create clientset
	cs, err := hyperfleet.NewForConfig(&hfrest.Config{
		Host:      baseURL,
		AccountID: accountID,
		AWSConfig: awsCfg,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create clientset: %w", err)
	}

	// Get cluster to access status
	cluster, err := cs.HyperfleetV1alpha1().Clusters().Get(ctx, clusterID, platform.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to fetch cluster: %w", err)
	}

	if cluster.Status.ControlPlaneEndpoint.Host != "" {
		return fmt.Sprintf("https://%s:%d", cluster.Status.ControlPlaneEndpoint.Host, cluster.Status.ControlPlaneEndpoint.Port), nil
	}

	return "", nil
}

func fetchClusterByName(ctx context.Context, baseURL, name string, creds awssdk.Credentials, region string) (*v1alpha1.Cluster, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	accountID, err := pkgconfig.GetAccountID()
	if err != nil {
		return nil, fmt.Errorf("failed to get account ID: %w", err)
	}

	cs, err := hyperfleet.NewForConfig(&hfrest.Config{
		Host:      baseURL,
		AccountID: accountID,
		AWSConfig: awsCfg,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	const pageSize = 100
	for offset := int64(0); ; offset += pageSize {
		listOpts := platform.ListOptions{
			Limit:  pageSize,
			Offset: offset,
		}

		clusterList, err := cs.HyperfleetV1alpha1().Clusters().List(ctx, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list clusters: %w", err)
		}

		for i := range clusterList.Items {
			c := &clusterList.Items[i]
			if c.Name == name || string(c.UID) == name {
				return c, nil
			}
		}

		if len(clusterList.Items) < int(pageSize) {
			break
		}
	}
	return nil, fmt.Errorf("cluster %q not found", name)
}
