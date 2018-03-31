package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	echo := fmt.Sprintf(`AWS Gateway Request Logger: HTTP Method:%s |Path:%s |Body:%s |Headers:%+v |QueryStringParams:%+v |PathParams:%+v`,
		req.HTTPMethod,
		req.Path,
		req.Body,
		req.Headers,
		req.QueryStringParameters,
		req.PathParameters)

	return events.APIGatewayProxyResponse{Body: "Echo: " + echo, StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(handleRequest)
}
