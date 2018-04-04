package awsclient

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/route53iface"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
)

const (
	MockParamValue   = "secretStringValue"
	MockHostedZoneID = "/hostedzone/Z119WBBTVP5WFX"
	MockeZoneName    = "test.com."
)

type MockClient struct{}

// R53 returns a configured AWS Route53 client
func (a *MockClient) R53() (route53iface.Route53API, error) {
	return new(R53Mock), nil
}

// SSM returns a configured AWS SSM client
func (a *MockClient) SSM() (ssmiface.SSMAPI, error) {
	return new(SSMMock), nil
}

type R53Mock struct {
	route53iface.Route53API
}

func (r *R53Mock) ListHostedZonesRequest(i *route53.ListHostedZonesInput) route53.ListHostedZonesRequest {
	return route53.ListHostedZonesRequest{
		Request: &aws.Request{
			Data: &route53.ListHostedZonesOutput{
				HostedZones: []route53.HostedZone{
					{
						CallerReference: aws.String("callerref"),
						Config: &route53.HostedZoneConfig{
							Comment:     aws.String("hostedZoneComment"),
							PrivateZone: aws.Bool(false),
						},
						Id: aws.String(MockHostedZoneID),
						LinkedService: &route53.LinkedService{
							Description:      aws.String("linkedDescription"),
							ServicePrincipal: aws.String("linkedServicePrincipal"),
						},
						Name: aws.String(MockeZoneName),
						ResourceRecordSetCount: aws.Int64(int64(1)),
					},
				},
				IsTruncated: aws.Bool(false),
				Marker:      i.Marker,
				MaxItems:    i.MaxItems,
			},
		},
	}
}

func (r *R53Mock) ChangeResourceRecordSetsRequest(i *route53.ChangeResourceRecordSetsInput) route53.ChangeResourceRecordSetsRequest {
	t := time.Now().UTC()
	return route53.ChangeResourceRecordSetsRequest{
		Request: &aws.Request{
			Data: &route53.ChangeResourceRecordSetsOutput{
				ChangeInfo: &route53.ChangeInfo{
					Comment:     i.ChangeBatch.Comment,
					Id:          aws.String("MockRequestID"),
					Status:      route53.ChangeStatusInsync,
					SubmittedAt: &t,
				},
			},
		},
	}
}

type SSMMock struct {
	ssmiface.SSMAPI
}

func (s *SSMMock) GetParameterRequest(i *ssm.GetParameterInput) ssm.GetParameterRequest {
	return ssm.GetParameterRequest{
		Request: &aws.Request{
			Data: &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Name:    i.Name,
					Type:    ssm.ParameterTypeSecureString,
					Value:   aws.String(MockParamValue),
					Version: aws.Int64(int64(1)),
				},
			},
		},
	}
}
