provider "aws" {
  region = "us-east-1"
}

data "aws_region" "current" {}
data "aws_iam_account_alias" "current" {}

resource "aws_iam_role" "ddns_lambda_role" {
  name = "ddns_lambda"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_iam_policy_document" "ddns_lambda_policy" {
  statement {
    sid = "CreateLambdaLogGroup"
    actions = [
      "logs:CreateLogGroup",
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current}:${data.aws_iam_account_alias.current}:*",
    ]
  }

  statement {
    sid = "LambdaLogging"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:${data.aws_region.current}:${data.aws_iam_account_alias.current}:log-group:/aws/lambda/${aws_lambda_function.ddns_lambda.function_name}:*",
    ]
  }

  statement {
    sid = "Route53Perms"
    actions = [
      "route53:ListHostedZones",
      "route53:ChangeResourceRecordSets"
    ]
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "ddns_lambda_policy" {
  name   = "ddns_lambda_policy"
  path   = "/"
  policy = "${data.aws_iam_policy_document.ddns_lambda_policy.json}"
}

resource "aws_iam_role_policy_attachment" "ddns_attach" {
  role       = "${aws_iam_role.ddns_lambda_role.name}"
  policy_arn = "${aws_iam_policy.ddns_lambda_policy.arn}"
}

resource "aws_lambda_function" "ddns_lambda" {
  filename         = "ddns.zip"
  function_name    = "ddns"
  role             = "${aws_iam_role.ddns_lambda_role.arn}"
  handler          = "main"
  source_code_hash = "${base64sha256(file("ddns.zip"))}"
  runtime          = "go1.x"
  timeout = 180

  environment {
    variables = {
      DEBUG = "false"
    }
  }
}

resource "aws_api_gateway_rest_api" "ddns_api" {
  name = "ddns"
  description = "DDNS API"
}

resource "aws_api_gateway_resource" "ddns_api_resource" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  parent_id = "${aws_api_gateway_rest_api.ddns_api.root_resource_id}"
  path_part = "ddns"
}

resource "aws_api_gateway_resource" "ddns_domain_api_resource" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  parent_id = "${aws_api_gateway_resource.ddns_api_resource.id}"
  path_part = "{domain}"
}

resource "aws_api_gateway_resource" "ddns_domain_record_api_resource" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  parent_id = "${aws_api_gateway_resource.ddns_domain_api_resource.id}"
  path_part = "{record+}"
}

resource "aws_api_gateway_method" "ddns_api_method" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  resource_id = "${aws_api_gateway_resource.ddns_domain_record_api_resource.id}"
  http_method = "GET"
  authorization = "NONE"
  api_key_required = true
  request_parameters {
    "method.request.querystring.type" = true
    "method.request.querystring.value" = true
  }
}

resource "aws_api_gateway_integration" "ddns_api_method-integration" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  resource_id = "${aws_api_gateway_resource.ddns_domain_record_api_resource.id}"
  http_method = "${aws_api_gateway_method.ddns_api_method.http_method}"
  type = "AWS_PROXY"
  integration_http_method = "POST"
  uri                     = "arn:aws:apigateway:${data.aws_region.current}:lambda:path/2015-03-31/functions/${aws_lambda_function.ddns_lambda.arn}/invocations"
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.ddns_lambda.arn}"
  principal     = "apigateway.amazonaws.com"
  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${data.aws_region.current}:${data.aws_iam_account_alias.current}:${aws_api_gateway_rest_api.ddns_api.id}/*/${aws_api_gateway_method.ddns_api_method.http_method}${aws_api_gateway_resource.ddns_domain_record_api_resource.path}"
}

resource "aws_api_gateway_api_key" "ddns_client" {
  name = "ddns_client"
}

resource "aws_api_gateway_deployment" "ddns_v1" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
  stage_name = "v1"
}

resource "aws_api_gateway_usage_plan" "ddns_usage" {
  name         = "ddns"
  description  = "ddns"
  api_stages {
    api_id = "${aws_api_gateway_rest_api.ddns_api.id}"
    stage  = "${aws_api_gateway_deployment.ddns_v1.stage_name}"
  }
}

resource "aws_api_gateway_usage_plan_key" "ddns" {
  key_id        = "${aws_api_gateway_api_key.ddns_client.id}"
  key_type      = "API_KEY"
  usage_plan_id = "${aws_api_gateway_usage_plan.ddns_usage.id}"
}

output "dev_url" {
  value = "https://${aws_api_gateway_deployment.ddns_v1.rest_api_id}.execute-api.${data.aws_region.current}.amazonaws.com/${aws_api_gateway_deployment.ddns_v1.stage_name}"
}