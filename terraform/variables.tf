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

variable "cloudtrail_retention" {
  description = "Cloud Trail Event Data Store Retention Policy"
  type        = number
  default     = 90
}

variable "lambda_env" {
  description = "Environment Variable to Assign the Lamba Function"
  type = object({
    KONG_ADMIN_API   = string
    KONG_SUPERADMIN  = bool
    KONG_ADMIN_TOKEN = string
    REDIS_DB         = string
  })
  sensitive = true
}

variable "image" {
  description = "URL to Kong CloudTrails Image"
  type        = string
  default     = "docker-hub-url"
}

