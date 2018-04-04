package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jcmturner/ddns/awsclient"
	"github.com/stretchr/testify/assert"
)

const (
	TestUserName            = "testuser"
	AuthHeaderValue         = "Basic dGVzdHVzZXI6c2VjcmV0U3RyaW5nVmFsdWU="
	AuthHeaderInvalidUser   = "Basic aW52YWxpZDpzZWNyZXRTdHJpbmdWYWx1ZQ=="
	AuthHeaderInvalidPasswd = "Basic dGVzdHVzZXI6aW52YWxpZA=="
)

func TestGetPasswd(t *testing.T) {
	awsCl := new(awsclient.MockClient)
	ssmCl, _ := awsCl.SSM()
	passwd, err := getPasswd(ssmCl, "/store/location")
	if err != nil {
		t.Errorf("error getting password: %v", err)
	}
	if passwd != awsclient.MockParamValue {
		t.Errorf("incorrect value of password")
	}
}

func TestParseBasicHeaderValue(t *testing.T) {
	u, p, err := parseBasicHeaderValue(AuthHeaderValue)
	if err != nil {
		t.Errorf("failed to parse auth header")
	}
	if u != TestUserName {
		t.Errorf("username is not correct")
	}
	if p != awsclient.MockParamValue {
		t.Errorf("password is not correct")
	}
}

func TestHandleAuth(t *testing.T) {
	user = TestUserName
	passwd = awsclient.MockParamValue

	var ctx context.Context
	e := events.APIGatewayCustomAuthorizerRequest{
		Type:               "TOKEN",
		AuthorizationToken: AuthHeaderValue,
		MethodArn:          "test:method:arn",
	}
	resp, err := handleAuth(ctx, e)
	if err != nil {
		t.Fatalf("error is not nil for valid token: %v", err)
	}
	assert.Equal(t, TestUserName, resp.PrincipalID, "Principal in response not correct")
	assert.Equal(t, "2012-10-17", resp.PolicyDocument.Version, "Poilcy document version not as expected")
	assert.True(t, len(resp.PolicyDocument.Statement) > 0, "No statements in policy")

	e.AuthorizationToken = AuthHeaderInvalidPasswd
	_, err = handleAuth(ctx, e)
	assert.NotNil(t, err, "Error should not be nil for invalid password")
	e.AuthorizationToken = AuthHeaderInvalidUser
	_, err = handleAuth(ctx, e)
	assert.NotNil(t, err, "Error should not be nil for invalid user")
}
