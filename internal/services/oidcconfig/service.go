package oidcconfig

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

// CreateOidcConfigRequest contains parameters for creating an OIDC config
type CreateOidcConfigRequest struct {
	Type           string // "managed" or "unmanaged"
	Region         string
	PlatformAPIURL string
	AWSConfig      awssdk.Config
}

// CreateOidcConfigResponse contains the created OIDC config
type CreateOidcConfigResponse struct {
	OidcConfig *v1alpha1.OidcConfig
}

// GetOidcConfigRequest contains parameters for getting an OIDC config
type GetOidcConfigRequest struct {
	ID             string
	PlatformAPIURL string
	AWSConfig      awssdk.Config
}

// GetOidcConfigResponse contains the retrieved OIDC config
type GetOidcConfigResponse struct {
	OidcConfig *v1alpha1.OidcConfig
}

// ListOidcConfigsRequest contains parameters for listing OIDC configs
type ListOidcConfigsRequest struct {
	PlatformAPIURL string
	AWSConfig      awssdk.Config
}

// ListOidcConfigsResponse contains the list of OIDC configs
type ListOidcConfigsResponse struct {
	Items []v1alpha1.OidcConfig
}

// DeleteOidcConfigRequest contains parameters for deleting an OIDC config
type DeleteOidcConfigRequest struct {
	ID             string
	PlatformAPIURL string
	AWSConfig      awssdk.Config
}

// CreateOidcConfig creates a new OIDC configuration via the platform API
func CreateOidcConfig(ctx context.Context, req *CreateOidcConfigRequest) (*CreateOidcConfigResponse, error) {
	cs, err := createClientset(ctx, req.PlatformAPIURL)
	if err != nil {
		return nil, err
	}

	oidcConfig := &v1alpha1.OidcConfig{
		Spec: v1alpha1.OidcConfigSpec{
			Type:      req.Type,
			IssuerUrl: "", // Empty for managed configs — platform API will compute it
		},
	}

	created, err := cs.HyperfleetV1alpha1().OidcConfigs().Create(ctx, oidcConfig, platform.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC config: %w", err)
	}

	return &CreateOidcConfigResponse{
		OidcConfig: created,
	}, nil
}

// GetOidcConfig retrieves an OIDC configuration by ID
func GetOidcConfig(ctx context.Context, req *GetOidcConfigRequest) (*GetOidcConfigResponse, error) {
	cs, err := createClientset(ctx, req.PlatformAPIURL)
	if err != nil {
		return nil, err
	}

	oidcConfig, err := cs.HyperfleetV1alpha1().OidcConfigs().Get(ctx, req.ID, platform.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC config: %w", err)
	}

	return &GetOidcConfigResponse{
		OidcConfig: oidcConfig,
	}, nil
}

// ListOidcConfigs lists all OIDC configurations
func ListOidcConfigs(ctx context.Context, req *ListOidcConfigsRequest) (*ListOidcConfigsResponse, error) {
	cs, err := createClientset(ctx, req.PlatformAPIURL)
	if err != nil {
		return nil, err
	}

	var allConfigs []v1alpha1.OidcConfig
	const pageSize = 100

	for offset := int64(0); ; offset += pageSize {
		listOpts := platform.ListOptions{
			Limit:  pageSize,
			Offset: offset,
		}

		list, err := cs.HyperfleetV1alpha1().OidcConfigs().List(ctx, listOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to list OIDC configs: %w", err)
		}

		allConfigs = append(allConfigs, list.Items...)

		if len(list.Items) < int(pageSize) {
			break
		}
	}

	return &ListOidcConfigsResponse{
		Items: allConfigs,
	}, nil
}

// DeleteOidcConfig deletes an OIDC configuration by ID
func DeleteOidcConfig(ctx context.Context, req *DeleteOidcConfigRequest) error {
	cs, err := createClientset(ctx, req.PlatformAPIURL)
	if err != nil {
		return err
	}

	err = cs.HyperfleetV1alpha1().OidcConfigs().Delete(ctx, req.ID, platform.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete OIDC config: %w", err)
	}

	return nil
}

// createClientset creates a new hyperfleet clientset
func createClientset(ctx context.Context, baseURL string) (*hyperfleet.Clientset, error) {
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

	return cs, nil
}
