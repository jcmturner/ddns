package main

import (
	"context"
	"ddns/awsclient"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
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
		Debug = log.New(io.Discard, "", log.Lshortfile)
	}
	if v := flag.Lookup("test.v"); v != nil {
		// We are in a test
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic("cannot load client config: " + err.Error())
	}
	cl := ssm.NewFromConfig(cfg)
	if cl == nil {
		panic("cannot create SSM client: " + err.Error())
	}
	passwd, err = getPasswd(cl, passwd_store)
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

func getPasswd(cl awsclient.SSMClient, store string) (string, error) {
	in := &ssm.GetParameterInput{
		Name:           aws.String(store),
		WithDecryption: aws.Bool(true),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	out, err := cl.GetParamter(ctx, in)
	if err != nil {
		return "", err
	}
	return aws.ToString(out.Parameter.Value), nil
}
