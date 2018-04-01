package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var user = os.Getenv("USER")
var passwd = os.Getenv("PASSWORD")

func main() {
	lambda.Start(handleAuth)
}

func handleAuth(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	// Configure API GW to use the Authorization header
	u, p, err := parseBasicHeaderValue(event.AuthorizationToken)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, fmt.Errorf("Unauthorized: %v", err)
	}
	if u != user || p != passwd {
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
