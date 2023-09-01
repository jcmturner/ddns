package awsclient

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

const (
	MockParamValue   = "secretStringValue"
	MockHostedZoneID = "/hostedzone/Z119WBBTVP5WFX"
	MockeZoneName    = "test.com."
)

type MockR53 struct{}

func (cl *MockR53) ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error) {
	t := time.Now().UTC()
	return &route53.ChangeResourceRecordSetsOutput{
		ChangeInfo: &route53.ChangeInfo{
			Comment:     params.ChangeBatch.Comment,
			Id:          aws.String("MockRequestID"),
			Status:      route53.ChangeStatusInsync,
			SubmittedAt: &t,
		},
	}, nil
}

func (cl *MockR53) ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
	return &route53.ListHostedZonesOutput{
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
				Name:                   aws.String(MockeZoneName),
				ResourceRecordSetCount: aws.Int64(int64(1)),
			},
		},
		IsTruncated: aws.Bool(false),
		Marker:      params.Marker,
		MaxItems:    params.MaxItems,
	}, nil
}

type MockSSM struct{}

func (cl *MockSSM) GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error) {
	return &ssm.GetParameterOutput{
		Parameter: &ssm.Parameter{
			Name:    params.Name,
			Type:    ssm.ParameterTypeSecureString,
			Value:   aws.String(MockParamValue),
			Version: aws.Int64(int64(1)),
		},
	}, nil
}
