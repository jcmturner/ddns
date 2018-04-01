package r53

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/route53iface"
)

// UpdateRecord upserts a route53 record
func UpdateRecord(r53 route53iface.Route53API, zoneID *string, fqdn, value string, rectype route53.RRType) error {
	in := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []route53.Change{
				{
					Action: route53.ChangeActionUpsert,
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name: aws.String(fqdn),
						Type: rectype,
						ResourceRecords: []route53.ResourceRecord{
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
	req := r53.ChangeResourceRecordSetsRequest(in)
	_, err := req.Send()
	return err
}

// ZoneID returns the zone ID given the zone name
func ZoneID(r53 route53iface.Route53API, zone string) (*string, error) {
	if !strings.HasSuffix(zone, ".") {
		zone = zone + "."
	}
	in := new(route53.ListHostedZonesInput)
	more := true
	for more {
		req := r53.ListHostedZonesRequest(in)
		out, err := req.Send()
		if err != nil {
			return nil, err
		}
		for _, z := range out.HostedZones {
			if aws.StringValue(z.Name) == zone {
				return z.Id, nil
			}
		}
		more = aws.BoolValue(out.IsTruncated)
		in.Marker = out.NextMarker
	}
	return nil, fmt.Errorf("zone %s not found", zone)
}
