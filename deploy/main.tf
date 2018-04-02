variable "aws_access_key" {}
variable "aws_secret_key" {}
variable "aws_region" {
  default = "eu-west-2"
}
variable "basic_auth_user" {}
variable "basic_auth_user_password" {}
variable "basic_auth_user_password_store" {
  default = "/ddns/password"
}


provider "aws" {
  access_key = "${var.aws_access_key}"
  secret_key = "${var.aws_secret_key}"
  region     = "${var.aws_region}"
}

data "aws_caller_identity" "current" {}

resource "aws_iam_role" "ddns_lambda" {
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
    sid = "LambdaLogging"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${aws_lambda_function.ddns_lambda.function_name}:*",
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
  role       = "${aws_iam_role.ddns_lambda.name}"
  policy_arn = "${aws_iam_policy.ddns_lambda_policy.arn}"
}

resource "aws_lambda_function" "ddns_lambda" {
  filename         = "ddns.zip"
  function_name    = "ddns"
  role             = "${aws_iam_role.ddns_lambda.arn}"
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

resource "aws_iam_role" "api_basic_auth_lambda" {
  name = "api_basic_auth_lambda"

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

data "aws_iam_policy_document" "api_basic_auth_lambda_policy" {
  statement {
    sid = "LambdaLogging"
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = [
      "arn:aws:logs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${aws_lambda_function.api_basic_auth_lambda.function_name}:*",
    ]
  }
  statement {
    sid = "ParamaterStoreAccess"
    actions = [
      "ssm:GetParameter"
    ]
    resources = [
      "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter${var.basic_auth_user_password_store}",
    ]
  }
}

resource "aws_iam_policy" "api_basic_auth_lambda_policy" {
  name   = "api_basic_auth_lambda_policy"
  path   = "/"
  policy = "${data.aws_iam_policy_document.api_basic_auth_lambda_policy.json}"
}

resource "aws_iam_role_policy_attachment" "api_basic_auth_attach" {
  role       = "${aws_iam_role.api_basic_auth_lambda.name}"
  policy_arn = "${aws_iam_policy.api_basic_auth_lambda_policy.arn}"
}

resource "aws_lambda_function" "api_basic_auth_lambda" {
  filename         = "auth.zip"
  function_name    = "api_basic_auth"
  role             = "${aws_iam_role.api_basic_auth_lambda.arn}"
  handler          = "main"
  source_code_hash = "${base64sha256(file("auth.zip"))}"
  runtime          = "go1.x"
  timeout = 30

  environment {
    variables = {
      USER = "${var.basic_auth_user}"
      PASSWORD_STORE = "${var.basic_auth_user_password_store}"
    }
  }
}

resource "aws_ssm_parameter" "secret" {
  name  = "${var.basic_auth_user_password_store}"
  description  = "DDNS password"
  type  = "SecureString"
  overwrite = true
  value = "${var.basic_auth_user_password}"

  tags {
    Applicaion = "ddns"
  }
}

resource "aws_cloudwatch_log_group" "ddns" {
  name = "/aws/lambda/${aws_lambda_function.ddns_lambda.function_name}"
  retention_in_days = 14
  tags {
    Application = "ddns"
  }
}

resource "aws_cloudwatch_log_group" "auth" {
  name = "/aws/lambda/${aws_lambda_function.api_basic_auth_lambda.function_name}"
  retention_in_days = 14
  tags {
    Application = "ddns"
  }
}

resource "aws_api_gateway_rest_api" "ddns" {
  name = "ddns"
  description = "DDNS API"
}

resource "aws_api_gateway_resource" "ddns" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  parent_id = "${aws_api_gateway_rest_api.ddns.root_resource_id}"
  path_part = "ddns"
}

resource "aws_api_gateway_resource" "ddns_domain" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  parent_id = "${aws_api_gateway_resource.ddns.id}"
  path_part = "{domain}"
}

resource "aws_api_gateway_resource" "ddns_domain_record" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  parent_id = "${aws_api_gateway_resource.ddns_domain.id}"
  path_part = "{record+}"
}

resource "aws_api_gateway_method" "ddns_api_method" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  resource_id = "${aws_api_gateway_resource.ddns_domain_record.id}"
  http_method = "GET"
  authorization = "CUSTOM"
  authorizer_id = "${aws_api_gateway_authorizer.basic.id}"
  api_key_required = false
  request_parameters {
    "method.request.querystring.type" = true
    "method.request.querystring.value" = true
  }
}

resource "aws_api_gateway_integration" "ddns_api_method-integration" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  resource_id = "${aws_api_gateway_resource.ddns_domain_record.id}"
  http_method = "${aws_api_gateway_method.ddns_api_method.http_method}"
  type = "AWS_PROXY"
  integration_http_method = "POST"
  uri                     = "arn:aws:apigateway:${var.aws_region}:lambda:path/2015-03-31/functions/${aws_lambda_function.ddns_lambda.arn}/invocations"
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.ddns_lambda.arn}"
  principal     = "apigateway.amazonaws.com"
  # More: http://docs.aws.amazon.com/apigateway/latest/developerguide/api-gateway-control-access-using-iam-policies-to-invoke-api.html
  source_arn = "arn:aws:execute-api:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${aws_api_gateway_rest_api.ddns.id}/*/${aws_api_gateway_method.ddns_api_method.http_method}${aws_api_gateway_resource.ddns_domain_record.path}"
}

resource "aws_api_gateway_deployment" "ddns_v1" {
  rest_api_id = "${aws_api_gateway_rest_api.ddns.id}"
  stage_name = "v1"
  depends_on = [
    "aws_api_gateway_method.ddns_api_method",
    "aws_api_gateway_integration.ddns_api_method-integration"
  ]
}

resource "aws_api_gateway_usage_plan" "ddns_usage" {
  name         = "ddns"
  description  = "ddns"
  api_stages {
    api_id = "${aws_api_gateway_rest_api.ddns.id}"
    stage  = "${aws_api_gateway_deployment.ddns_v1.stage_name}"
  }
}

//resource "aws_api_gateway_api_key" "ddns_client" {
//  name = "ddns_client"
//}

//resource "aws_api_gateway_usage_plan_key" "ddns" {
//  key_id        = "${aws_api_gateway_api_key.ddns_client.id}"
//  key_type      = "API_KEY"
//  usage_plan_id = "${aws_api_gateway_usage_plan.ddns_usage.id}"
//}

output "dev_url" {
  value = "https://${aws_api_gateway_deployment.ddns_v1.rest_api_id}.execute-api.${var.aws_region}.amazonaws.com/${aws_api_gateway_deployment.ddns_v1.stage_name}"
}

resource "aws_api_gateway_authorizer" "basic" {
  name                   = "basic_auth"
  rest_api_id            = "${aws_api_gateway_rest_api.ddns.id}"
  authorizer_uri         = "${aws_lambda_function.api_basic_auth_lambda.invoke_arn}"
  authorizer_credentials = "${aws_iam_role.invocation_role.arn}"
}

resource "aws_iam_role" "invocation_role" {
  name = "api_gateway_auth_invocation"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "apigateway.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "invocation_policy" {
  name = "api_gateway_auth_invocation"
  role = "${aws_iam_role.invocation_role.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "lambda:InvokeFunction",
      "Effect": "Allow",
      "Resource": "${aws_lambda_function.api_basic_auth_lambda.arn}"
    }
  ]
}
EOF
}