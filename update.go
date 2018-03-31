package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
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

var Log = log.New(os.Stderr, "DDNS: ", log.Lshortfile)
var Debug *log.Logger

func init() {
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		Debug = log.New(os.Stderr, "DDNS Debug: ", log.Lshortfile)
	} else {
		Debug = log.New(ioutil.Discard, "", log.Lshortfile)

	}
}

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if _, ok := req.QueryStringParameters["type"]; !ok {
		Log.Printf("%d - no 'type' query param", http.StatusBadRequest)
		return failure(missingQueryParamsMsg, http.StatusBadRequest)
	}
	if _, ok := req.QueryStringParameters["value"]; !ok {
		Log.Printf("%d - no 'value' query param", http.StatusBadRequest)
		return failure(missingQueryParamsMsg, http.StatusBadRequest)
	}
	rrtype, ok := validate(req.QueryStringParameters["type"], req.QueryStringParameters["value"])
	if !ok {
		return failure(invalidQueryParamsMsg, http.StatusBadRequest)
	}

	awsCl := new(awsclient.AWSClient)
	r53Cl, err := awsCl.R53()
	if err != nil {
		Log.Printf("%d - error getting AWS client: %v", http.StatusInternalServerError, err)
		return failure(serverErrMsg, http.StatusInternalServerError)
	}

	domain := req.PathParameters["domain"]
	record := req.PathParameters["record"]
	fqdn := fmt.Sprintf("%s.%s", record, domain)
	zoneID, err := r53.ZoneID(r53Cl, req.PathParameters["domain"])
	if err != nil {
		Log.Printf("error getting ZoneID: %v", err)
		return failure(invalidDomainMsg, http.StatusBadRequest)
	}

	err = r53.UpdateRecord(r53Cl, zoneID, fqdn, req.QueryStringParameters["value"], rrtype)
	if err != nil {
		Log.Printf("error updating record: %v", err)
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
	Log.Printf("successful update: %+v", r)
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
		Debug.Printf("record type: %s", string(route53.RRTypeA))
		return route53.RRTypeA, validateARecord(value)
	default:
		Log.Printf("%d - invalid record type %s", http.StatusBadRequest, rectype)
		return route53.RRType(""), false
	}
}

func validateARecord(value string) bool {
	if value == "" {
		Log.Printf("%d - invalid value is null", http.StatusBadRequest)
		return false
	}
	if ip := net.ParseIP(value); ip == nil {
		Log.Printf("%d - invalid IP address: %s", http.StatusBadRequest, value)
		return false
	}
	return true
}
