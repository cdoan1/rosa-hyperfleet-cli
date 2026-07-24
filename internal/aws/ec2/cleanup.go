package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	smithy "github.com/aws/smithy-go"
)

// CleanVPCForDeletion removes orphaned VPC endpoints and non-default security
// groups from a VPC so that its CloudFormation stack can be deleted cleanly.
//
// This is a temporary workaround for OCPBUGS-74960 where HyperShift's CPO
// fails to clean up VPC endpoint resources during HCP deletion.
func CleanVPCForDeletion(ctx context.Context, cfg aws.Config, vpcID string) error {
	client := ec2.NewFromConfig(cfg)

	if err := deleteOrphanedVPCEndpoints(ctx, client, vpcID); err != nil {
		return fmt.Errorf("cleaning VPC endpoints: %w", err)
	}

	if err := deleteNonDefaultSecurityGroups(ctx, client, vpcID); err != nil {
		return fmt.Errorf("cleaning security groups: %w", err)
	}

	return nil
}

func deleteNonDefaultSecurityGroups(ctx context.Context, client *ec2.Client, vpcID string) error {
	var securityGroups []types.SecurityGroup

	paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("describing security groups: %w", err)
		}
		securityGroups = append(securityGroups, page.SecurityGroups...)
	}

	nonDefault := make([]types.SecurityGroup, 0, len(securityGroups))
	for _, sg := range securityGroups {
		if aws.ToString(sg.GroupName) != "default" {
			nonDefault = append(nonDefault, sg)
		}
	}
	if len(nonDefault) == 0 {
		return nil
	}

	// Revoke ingress/egress rules that reference other SGs in this set so
	// cross-group dependencies don't block deletion.
	sgIDs := make(map[string]bool, len(nonDefault))
	for _, sg := range nonDefault {
		sgIDs[aws.ToString(sg.GroupId)] = true
	}
	for _, sg := range nonDefault {
		sgID := aws.ToString(sg.GroupId)
		for _, rule := range sg.IpPermissions {
			for _, pair := range rule.UserIdGroupPairs {
				if sgIDs[aws.ToString(pair.GroupId)] && aws.ToString(pair.GroupId) != sgID {
					log.Printf("  revoking ingress rule referencing %s from SG %s", aws.ToString(pair.GroupId), sgID)
					if _, err := client.RevokeSecurityGroupIngress(ctx, &ec2.RevokeSecurityGroupIngressInput{
						GroupId: aws.String(sgID),
						IpPermissions: []types.IpPermission{{
							IpProtocol:       rule.IpProtocol,
							FromPort:         rule.FromPort,
							ToPort:           rule.ToPort,
							UserIdGroupPairs: []types.UserIdGroupPair{pair},
						}},
					}); err != nil {
						log.Printf("  warning: failed to revoke ingress rule on SG %s: %v", sgID, err)
					}
				}
			}
		}
		for _, rule := range sg.IpPermissionsEgress {
			for _, pair := range rule.UserIdGroupPairs {
				if sgIDs[aws.ToString(pair.GroupId)] && aws.ToString(pair.GroupId) != sgID {
					log.Printf("  revoking egress rule referencing %s from SG %s", aws.ToString(pair.GroupId), sgID)
					if _, err := client.RevokeSecurityGroupEgress(ctx, &ec2.RevokeSecurityGroupEgressInput{
						GroupId: aws.String(sgID),
						IpPermissions: []types.IpPermission{{
							IpProtocol:       rule.IpProtocol,
							FromPort:         rule.FromPort,
							ToPort:           rule.ToPort,
							UserIdGroupPairs: []types.UserIdGroupPair{pair},
						}},
					}); err != nil {
						log.Printf("  warning: failed to revoke egress rule on SG %s: %v", sgID, err)
					}
				}
			}
		}
	}

	var cleanupErr error
	for _, sg := range nonDefault {
		sgID := aws.ToString(sg.GroupId)
		log.Printf("  deleting security group %s (%s)", sgID, aws.ToString(sg.GroupName))
		_, err := client.DeleteSecurityGroup(ctx, &ec2.DeleteSecurityGroupInput{
			GroupId: aws.String(sgID),
		})
		if err != nil {
			log.Printf("  warning: failed to delete SG %s: %v", sgID, err)
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("delete SG %s: %w", sgID, err))
		}
	}

	return cleanupErr
}

func deleteOrphanedVPCEndpoints(ctx context.Context, client *ec2.Client, vpcID string) error {
	var allEndpoints []types.VpcEndpoint

	paginator := ec2.NewDescribeVpcEndpointsPaginator(client, &ec2.DescribeVpcEndpointsInput{
		Filters: []types.Filter{
			{Name: aws.String("vpc-id"), Values: []string{vpcID}},
		},
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("describing VPC endpoints: %w", err)
		}
		allEndpoints = append(allEndpoints, page.VpcEndpoints...)
	}

	var endpointIDs []string
	for _, ep := range allEndpoints {
		endpointIDs = append(endpointIDs, aws.ToString(ep.VpcEndpointId))
	}
	if len(endpointIDs) == 0 {
		return nil
	}

	log.Printf("  deleting %d VPC endpoint(s): %v", len(endpointIDs), endpointIDs)
	deleteOut, err := client.DeleteVpcEndpoints(ctx, &ec2.DeleteVpcEndpointsInput{
		VpcEndpointIds: endpointIDs,
	})
	if err != nil {
		return fmt.Errorf("deleting VPC endpoints: %w", err)
	}
	if deleteOut != nil && len(deleteOut.Unsuccessful) > 0 {
		var failErr error
		for _, item := range deleteOut.Unsuccessful {
			failErr = errors.Join(failErr, fmt.Errorf("endpoint %s: %s",
				aws.ToString(item.ResourceId), aws.ToString(item.Error.Message)))
		}
		return fmt.Errorf("some VPC endpoints failed to delete: %w", failErr)
	}

	log.Printf("  waiting for VPC endpoints to be fully deleted...")
	for attempt := 0; attempt < 60; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
		remaining, err := client.DescribeVpcEndpoints(ctx, &ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: endpointIDs,
		})
		if err != nil {
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "InvalidVpcEndpointId.NotFound" {
				log.Printf("  VPC endpoints deleted")
				return nil
			}
			return fmt.Errorf("waiting for VPC endpoint deletion: %w", err)
		}
		stillActive := 0
		for _, ep := range remaining.VpcEndpoints {
			if ep.State != types.StateDeleted {
				stillActive++
			}
		}
		if stillActive == 0 {
			log.Printf("  VPC endpoints deleted")
			return nil
		}
		log.Printf("  still waiting for %d VPC endpoint(s)...", stillActive)
	}

	return fmt.Errorf("timed out waiting for VPC endpoints to be deleted")
}
