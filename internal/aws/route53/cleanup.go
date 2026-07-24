package route53

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

// CleanHostedZoneForDeletion removes all non-default records (everything except
// SOA and NS) from a Route53 hosted zone so that CloudFormation can delete it.
//
// This is a temporary workaround for OCPBUGS-74960 where HyperShift's CPO
// fails to clean up DNS records (api.*.hypershift.local, *.apps.*.hypershift.local)
// during HCP deletion.
func CleanHostedZoneForDeletion(ctx context.Context, cfg aws.Config, zoneID string) error {
	client := route53.NewFromConfig(cfg)

	var records []types.ResourceRecordSet
	paginator := route53.NewListResourceRecordSetsPaginator(client, &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing records in zone %s: %w", zoneID, err)
		}
		records = append(records, page.ResourceRecordSets...)
	}

	var changes []types.Change
	for _, record := range records {
		if record.Type == types.RRTypeSoa || record.Type == types.RRTypeNs {
			continue
		}
		changes = append(changes, types.Change{
			Action:            types.ChangeActionDelete,
			ResourceRecordSet: &record,
		})
	}

	if len(changes) == 0 {
		log.Printf("  no orphaned records found in zone %s", zoneID)
		return nil
	}

	log.Printf("  will delete %d non-default record(s) from zone %s", len(changes), zoneID)

	var cleanupErr error
	const batchSize = 1000
	for i := 0; i < len(changes); i += batchSize {
		end := i + batchSize
		if end > len(changes) {
			end = len(changes)
		}
		batch := changes[i:end]

		out, err := client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(zoneID),
			ChangeBatch: &types.ChangeBatch{
				Changes: batch,
			},
		})
		if err != nil {
			log.Printf("  warning: failed to delete %d records from zone %s: %v", len(batch), zoneID, err)
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("delete records from zone %s: %w", zoneID, err))
			continue
		}
		log.Printf("  deleted %d records from zone %s", len(batch), zoneID)

		if out.ChangeInfo != nil && out.ChangeInfo.Id != nil {
			waiter := route53.NewResourceRecordSetsChangedWaiter(client)
			if waitErr := waiter.Wait(ctx, &route53.GetChangeInput{
				Id: out.ChangeInfo.Id,
			}, 2*time.Minute); waitErr != nil {
				log.Printf("  warning: timed out waiting for change %s to reach INSYNC: %v", aws.ToString(out.ChangeInfo.Id), waitErr)
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("waiting for change INSYNC: %w", waitErr))
			}
		}
	}

	return cleanupErr
}
