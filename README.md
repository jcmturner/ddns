# DDNS

Implements an equivalent of DynDNS using AWS API Gateway and Lambda to update DNS managed in AWS Route53

## Building & Deploying

Prerequisites
- Go https://golang.org/dl/
- Hashicorp terraform https://www.terraform.io/
- AWS Account which is authoritative for the zone(s) that contain the DNS records to be updated

There are two Lambda functions to build and then apply terraform to deploy the functions and create the API gateway
configuration

Run the following commands. You will be prompted for your AWS access keys and secret keys, and a username and password
that will be used to secure requests to the API.
```bash
export GOOS=linux && \
export GOARCH=amd64 && \
cd $GOPATH/src/github.com/jcmturner/ddns/deploy && \
go build -i -o $GOPATH/src/github.com/jcmturner/ddns/deploy/main $GOPATH/src/github.com/jcmturner/ddns/authorizer/basicauth.go && \
zip ./auth.zip main && \
go build -i -o $GOPATH/src/github.com/jcmturner/ddns/deploy/main $GOPATH/src/github.com/jcmturner/ddns/update/update.go && \
zip ./ddns.zip main && \
terraform apply

```