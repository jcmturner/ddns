package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/jcmturner/ddns/awsclient"
	"github.com/jcmturner/ddns/r53"
)

const (
	invalidQueryParamsMsg = "invalid query parameters provided"
	missingQueryParamsMsg = "missing query parameters"
	invalidDomainMsg      = "domain provided is not valid"
	serverErrMsg          = "error processing request"
)

type DDNSError struct {
	ErrorMessage string
}

func (e DDNSError) Error() string {
	return e.ErrorMessage
}

type DDNSResponse struct {
	Domain     string
	ZoneID     string
	Record     string
	RecordType string
	NewValue   string
}

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if _, ok := req.QueryStringParameters["type"]; !ok {
		return failure(missingQueryParamsMsg, http.StatusBadRequest)
	}
	if _, ok := req.QueryStringParameters["value"]; !ok {
		return failure(missingQueryParamsMsg, http.StatusBadRequest)
	}
	rrtype, ok := validate(req.QueryStringParameters["type"], req.QueryStringParameters["value"])
	if !ok {
		return failure(invalidQueryParamsMsg, http.StatusBadRequest)
	}

	awsCl := new(awsclient.AWSClient)
	r53Cl, err := awsCl.R53()
	if err != nil {
		return failure(serverErrMsg, http.StatusInternalServerError)
	}

	zoneID, err := r53.ZoneID(r53Cl, req.PathParameters["domain"])
	if err != nil {
		return failure(invalidDomainMsg, http.StatusBadRequest)
	}

	err = process(awsCl, zoneID, req.PathParameters["record"], req.QueryStringParameters["value"], rrtype)
	if err != nil {
		return failure(serverErrMsg, http.StatusInternalServerError)
	}

	return success(req.PathParameters["domain"], req.PathParameters["record"], req.QueryStringParameters["value"], zoneID, rrtype)
}

func main() {
	lambda.Start(handleRequest)
}

func failure(msg string, HTTPcode int) (events.APIGatewayProxyResponse, error) {
	e := DDNSError{ErrorMessage: msg}
	var s string
	b, err := json.Marshal(e)
	if err != nil {
		s = "could not marshal error: " + err.Error()
	} else {
		s = string(b)
	}
	return events.APIGatewayProxyResponse{Body: s, StatusCode: HTTPcode}, e
}

func success(domain, record, value string, zoneID *string, rrType route53.RRType) (events.APIGatewayProxyResponse, error) {
	r := DDNSResponse{
		Domain:     domain,
		ZoneID:     aws.StringValue(zoneID),
		Record:     record,
		RecordType: string(rrType),
		NewValue:   value,
	}
	b, err := json.Marshal(r)
	if err != nil {
		return failure("update succeeded, error marshaling response", http.StatusPartialContent)
	}
	return events.APIGatewayProxyResponse{Body: string(b), StatusCode: http.StatusOK}, nil
}

func validate(rectype, value string) (route53.RRType, bool) {
	rectype = strings.ToUpper(rectype)
	switch rectype {
	case string(route53.RRTypeA):
		return route53.RRTypeA, validateARecord(value)
	default:
		return route53.RRType(""), false
	}
}

func validateARecord(value string) bool {
	if value == "" {
		return false
	}
	if ip := net.ParseIP(value); ip == nil {
		return false
	}
	return true
}

func process(awsCl awsclient.Iface, zoneID *string, record, value string, rrtype route53.RRType) error {
	r53Cl, err := awsCl.R53()
	if err != nil {
		return err
	}
	return r53.UpdateRecord(r53Cl, zoneID, record, value, rrtype)
}
