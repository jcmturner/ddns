package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/ssmiface"
	"github.com/jcmturner/ddns/awsclient"
)

var user = os.Getenv("USER")
var passwd_store = os.Getenv("PASSWORD_STORE")
var passwd string

var Log = log.New(os.Stderr, "Auth: ", log.Lshortfile)
var Debug *log.Logger

func init() {
	// Initialise Debug logger based on lambda env variable
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		Debug = log.New(os.Stderr, "Auth Debug: ", log.Lshortfile)
	} else {
		Debug = log.New(ioutil.Discard, "", log.Lshortfile)
	}
	if v := flag.Lookup("test.v"); v != nil {
		// We are in a test
		return
	}
	awsCl := new(awsclient.AWSClient)
	ssmCl, err := awsCl.SSM()
	if err != nil {
		panic("cannot create SSM client: " + err.Error())
	}
	passwd, err = getPasswd(ssmCl, passwd_store)
	if err != nil {
		panic("cannot get password: " + err.Error())
	}
}

func main() {
	lambda.Start(handleAuth)
}

func handleAuth(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// Configure API GW to use the Authorization header
	u, p, err := parseBasicHeaderValue(event.AuthorizationToken)
	if err != nil {
		Debug.Printf("Authentication failed: %v", err)
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized: %v", err)
	}
	if u != user || p != passwd {
		Debug.Printf("Invalid credentials. Provided: %s %s Reference: %s %s", u, p, user, passwd)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: u,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Allow",
					Resource: []string{event.MethodArn},
				},
			},
		},
	}, nil
}

func parseBasicHeaderValue(s string) (username, password string, err error) {
	h := strings.SplitN(s, " ", 2)
	if len(h) != 2 {
		return "", "", errors.New("request does not have a valid authorization header")
	}
	if strings.ToLower(h[0]) != "basic" {
		return "", "", errors.New("request does not have a valid basic auth header")
	}
	b, err := base64.StdEncoding.DecodeString(h[1])
	if err != nil {
		return "", "", errors.New("could not base64 decode auth header")
	}
	v := string(b)
	vc := strings.SplitN(v, ":", 2)
	return vc[0], vc[1], nil
}

func getPasswd(ssmapi ssmiface.SSMAPI, store string) (string, error) {
	in := &ssm.GetParameterInput{
		Name:           aws.String(store),
		WithDecryption: aws.Bool(true),
	}
	req := ssmapi.GetParameterRequest(in)
	out, err := req.Send()
	if err != nil {
		return "", err
	}
	return aws.StringValue(out.Parameter.Value), nil
}
