package r53

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/jcmturner/ddns/awsclient"
)

// UpdateRecord upserts a route53 record
func UpdateRecord(ctx context.Context, cl awsclient.R53Client, zoneID *string, fqdn, value string, rectype types.RRType) error {
	in := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeActionUpsert,
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(fqdn),
						Type: rectype,
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(value),
							},
						},
						TTL: aws.Int64(60),
					},
				},
			},
			Comment: aws.String("DDNS update."),
		},
		HostedZoneId: zoneID,
	}
	_, err := cl.ChangeResourceRecordSets(ctx, in)
	return err
}

// ZoneID returns the zone ID given the zone name
func ZoneID(ctx context.Context, cl awsclient.R53Client, zone string) (*string, error) {
	if !strings.HasSuffix(zone, ".") {
		zone = zone + "."
	}
	in := new(route53.ListHostedZonesInput)
	more := true
	for more {
		out, err := cl.ListHostedZones(ctx, in)
		if err != nil {
			return nil, err
		}
		for _, z := range out.HostedZones {
			if aws.ToString(z.Name) == zone {
				return z.Id, nil
			}
		}
		more = out.IsTruncated
		in.Marker = out.NextMarker
	}
	return nil, fmt.Errorf("zone %s not found", zone)
}
