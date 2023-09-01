package awsclient

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// R53 is an interface used to enable mock client for testing.
// The subset of methods the AWS route53 client implements that we use are specified here.
type R53Client interface {
	ChangeResourceRecordSets(ctx context.Context, params *route53.ChangeResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ChangeResourceRecordSetsOutput, error)
	ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
}

type SSMClient interface {
	GetParameter(ctx context.Context, params *ssm.GetParameterInput, optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

// R53 returns a configured AWS Route53 client
func R53(ctx context.Context) (R53Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cl := route53.NewFromConfig(cfg)
	if cl == nil {
		return nil, err
	}
	return cl, nil
}

// SSM returns a configured AWS SSM client
func SSM(ctx context.Context) (SSMClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	cl := ssm.NewFromConfig(cfg)
	if cl == nil {
		return nil, err
	}
	return cl, nil
}
