package awsclient

import (
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/route53iface"
)

type Iface interface {
	R53() (route53iface.Route53API, error)
}

type AWSClient struct{}

// DynamoDB returns a configured AWS DynamoDB client
func (a *AWSClient) R53() (route53iface.Route53API, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return new(route53.Route53), err
	}
	return route53.New(cfg), nil
}
