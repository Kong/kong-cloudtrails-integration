variable "resource_name" {
  description = "Name to provide Lambda Function and related resources"
  type        = string
  default     = "kong-cloudtrails-integration"
}

variable "resource_tags" {
  description = "Tags to set for all resources"
  type        = map(string)
  default = {
    project = "kong-cloudtrail-integrations"
  }
}

variable "aws_region" {
  description = "Name of the aws region to deploy resoures into"
  type        = string
}

variable "existing_vpc" {
  description = "Name/Id of Existing VPC to deploy resources into"
  type        = string
}

variable "existing_subnet_ids" {
  description = "List of Existing Subnet Ids to deploy resources into"
  type        = list(string)
}

variable "security_group" {
  description = "Name of the Security Group that will be created in the VPC to support the lamba function"
  type        = string
}

variable "channel_arn" {
  description = ""
  type = string
}

variable "lambda_env" {
  description = "Environment Variable to Assign the Lamba Function"
  type = object({
    KONG_ADMIN_API   = string
    KONG_SUPERADMIN  = bool
    KONG_ADMIN_TOKEN = string
    KONG_ROOT_CA     = string
    REDIS_DB         = string
    CHANNEL_ARN      = string
  })
  sensitive = true
}

variable "image" {
  description = "URL to Kong CloudTrails Image"
  type        = string
  default     = "docker-hub-url"
}

