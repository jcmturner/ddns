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
	intServerErrorMsg     = "error processing request"
)

type DDNSUpdate struct {
	Domain     string
	ZoneID     string
	ZoneIDPtr  *string `json:"-"`
	Record     string
	FQDN       string
	RecordType string
	RRType     route53.RRType `json:"-"`
	NewValue   string
}

type DDNSError struct {
	ErrorMessage string
}

func (e DDNSError) Error() string {
	return e.ErrorMessage
}

var Log = log.New(os.Stderr, "DDNS: ", log.Lshortfile)
var Debug *log.Logger

func init() {
	// Initialise Debug logger based on lambda env variable
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		Debug = log.New(os.Stderr, "DDNS Debug: ", log.Lshortfile)
	} else {
		Debug = log.New(ioutil.Discard, "", log.Lshortfile)

	}
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if _, ok := req.QueryStringParameters["type"]; !ok {
		Log.Printf("%d - no 'type' query param", http.StatusBadRequest)
		return badRequest(missingQueryParamsMsg)
	}
	if _, ok := req.QueryStringParameters["value"]; !ok {
		Log.Printf("%d - no 'value' query param", http.StatusBadRequest)
		return badRequest(missingQueryParamsMsg)
	}
	rrtype, ok := validate(req.QueryStringParameters["type"], req.QueryStringParameters["value"])
	if !ok {
		return badRequest(invalidQueryParamsMsg)
	}

	awsCl := new(awsclient.AWSClient)
	r53Cl, err := awsCl.R53()
	if err != nil {
		err := fmt.Errorf("error getting AWS client: %v", err)
		return intServError(err)
	}

	domain := req.PathParameters["domain"]
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	record := req.PathParameters["record"]
	fqdn := fmt.Sprintf("%s.%s", record, domain)
	zoneID, err := r53.ZoneID(r53Cl, req.PathParameters["domain"])
	if err != nil {
		err = fmt.Errorf("error getting ZoneID: %v", err)
		Log.Print(err.Error())
		return badRequest(invalidDomainMsg)
	}

	d := DDNSUpdate{
		Domain:     domain,
		ZoneID:     aws.StringValue(zoneID),
		ZoneIDPtr:  zoneID,
		Record:     record,
		FQDN:       fqdn,
		RecordType: string(rrtype),
		RRType:     rrtype,
		NewValue:   req.QueryStringParameters["value"],
	}
	Debug.Printf("update requested: %+v", d)

	err = r53.UpdateRecord(r53Cl, d.ZoneIDPtr, d.FQDN, d.NewValue, d.RRType)
	if err != nil {
		err = fmt.Errorf("error updating record: %v", err)
		Log.Print(err.Error())
		return intServError(err)
	}

	return success(d)
}

func badRequest(msg string) (events.APIGatewayProxyResponse, error) {
	e := DDNSError{ErrorMessage: msg}
	var s string
	b, err := json.Marshal(e)
	if err != nil {
		err = fmt.Errorf("could not marshal error: %v", err)
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusBadRequest}, err
	} else {
		s = string(b)
	}
	return events.APIGatewayProxyResponse{Body: s, StatusCode: http.StatusBadRequest}, nil
}

func intServError(intErr error) (events.APIGatewayProxyResponse, error) {
	e := DDNSError{ErrorMessage: intServerErrorMsg}
	var s string
	b, err := json.Marshal(e)
	if err != nil {
		err = fmt.Errorf("could not marshal error: %v", err)
		return events.APIGatewayProxyResponse{Body: intServerErrorMsg, StatusCode: http.StatusInternalServerError}, intErr
	} else {
		s = string(b)
	}
	return events.APIGatewayProxyResponse{Body: s, StatusCode: http.StatusInternalServerError}, intErr
}

func success(d DDNSUpdate) (events.APIGatewayProxyResponse, error) {
	Log.Printf("successful update: %+v", d)
	b, err := json.Marshal(d)
	if err != nil {
		err = fmt.Errorf("update succeeded, error marshaling response: %v", err)
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusPartialContent}, err
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
