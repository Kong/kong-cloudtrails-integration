terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.19.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}


resource "aws_elasticache_subnet_group" "el_cache_subnet_group" {
  name       = "${var.resource_name}-cache-subnet"
  subnet_ids = var.existing_subnet_ids
  tags       = var.resource_tags
}

resource "aws_elasticache_replication_group" "rg" {
  replication_group_id       = var.resource_name
  description                = "Redis - Kong Cloudtrails Audit Log Infra"
  node_type                  = "cache.t3.small"
  engine                     = "redis"
  engine_version             = "6.2"
  parameter_group_name       = "default.redis6.x"
  port                       = 6379
  num_cache_clusters         = 2
  automatic_failover_enabled = true
  multi_az_enabled           = true
  at_rest_encryption_enabled = true
  subnet_group_name          = aws_elasticache_subnet_group.el_cache_subnet_group.name
  lifecycle {
    ignore_changes = [number_cache_clusters]
  }
  tags       = var.resource_tags
  depends_on = [aws_elasticache_subnet_group.el_cache_subnet_group]
}

resource "aws_iam_role" "iam_role" {
  name = "${var.resource_name}-lambda"
  tags = var.resource_tags

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      }
    ]
  })

  managed_policy_arns = ["arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"]
  inline_policy {
    name = "cloudtrails-putAuditEvents-role"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          "Effect" : "Allow",
          "Action" : "cloudtrail-data:PutAuditEvents",
          "Resource" : "${var.channel_arn}"
        }
      ]
    })
  }
}

resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.resource_name}"
  retention_in_days = 14
}

data "aws_vpc" "selected_vpc" {
  id = var.existing_vpc
}

resource "aws_security_group" "allow_lambda" {
  name        = "allow lambda"
  description = "Allow traffic to Redis, CloudTrails, and Kong Admin API"
  vpc_id      = var.existing_vpc


  egress {
    description = "TCP to Redis"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected_vpc.cidr_block]
  }
  egress {
    description = "HTTPS to CloudTrails-PutAudit"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "HTTP to Kong Admin API"
    from_port   = 8001
    to_port     = 8001
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected_vpc.cidr_block]
  }

  egress {
    description = "HTTPS to Kong Admin API"
    from_port   = 8444
    to_port     = 8444
    protocol    = "tcp"
    cidr_blocks = [data.aws_vpc.selected_vpc.cidr_block]
  }
  timeouts {
    delete = "40m"
  }
  tags = var.resource_tags
}

resource "aws_lambda_function" "lambda" {
  function_name = var.resource_name
  role          = aws_iam_role.iam_role.arn
  package_type  = "Image"
  image_uri     = var.image
  timeout       = 60
  vpc_config {
    subnet_ids         = var.existing_subnet_ids
    security_group_ids = ["${aws_security_group.allow_lambda.id}"]
  }
  environment {
    variables = merge(var.lambda_env, { REDIS_HOST = "${aws_elasticache_replication_group.rg.primary_endpoint_address}:6379" })
  }
  tags = var.resource_tags
  depends_on = [
    aws_elasticache_replication_group.rg,
    aws_cloudwatch_log_group.lambda_logs,
    aws_iam_role.iam_role,
    aws_security_group.allow_lambda
  ]
}

resource "aws_cloudwatch_event_rule" "event_rule" {
  name                = "${var.resource_name}-rule"
  description         = "Fires Kong CloudTrails Lambda Every 1 hr"
  schedule_expression = "rate(1 hour)"
  tags                = var.resource_tags
}

resource "aws_cloudwatch_event_target" "event_target" {
  rule      = aws_cloudwatch_event_rule.event_rule.name
  target_id = "${var.resource_name}-lambda"
  arn       = aws_lambda_function.lambda.arn
}

resource "aws_lambda_permission" "cloudwatch_perm" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.event_rule.arn
}



