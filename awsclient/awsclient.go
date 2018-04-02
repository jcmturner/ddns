package awsclient

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/route53iface"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
)

type Iface interface {
	R53() (route53iface.Route53API, error)
	SSM() (ssmiface.SSMAPI, error)
}

type AWSClient struct{}

// R53 returns a configured AWS Route53 client
func (a *AWSClient) R53() (route53iface.Route53API, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return new(route53.Route53), err
	}
	return route53.New(cfg), nil
}

// SSM returns a configured AWS SSM client
func (a *AWSClient) SSM() (ssmiface.SSMAPI, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return new(ssm.SSM), err
	}
	return ssm.New(cfg), nil
}
