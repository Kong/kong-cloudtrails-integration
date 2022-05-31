# Kong CloudTrails Integration

**WARNING**: this is an early phase POC, under active development and
             **EXPERIMENTAL**. This is not a fully supported product. For any bugs,
             issues, requests, please open an issue. Any issues
             will be fixed on a best-effort basis.

## Reference Architecture

![Kong Cloudtrails Reference Architecture](/assets/img/referenceArchitecture.png)

Kong Gateway - supports Audit Logs as an Enterprise Feature. When enabled on the Global Control Plane, every api request called to the Kong admin api and dao object created, updated, or deleted in the Kong database with relevant data including RBAC, Workspace, and a TTL for the entry. These audit log entries in turn are retrievable via endpoints on the Kong admin api: /audit/requests, and /audit/objects.

For the Kong Cloudtrails Integration Project, in order to publish Kong Audit Logs to Cloudtrails an AWS Lambda function in combination with AWS ElastiCache-Redis are deployed into existing VPC where the Kong Global Control Plane resides. The lambda function is written in GO and packaged as an image available on dockerhub (link).

The AWS Lambda function will strictly call the /audit/requests endpoint, process and remove duplicate entries by evaluating existing keys in Redis before transforming and submitting the audit log entries to cloudtrails. The duplicate logic is supported with Redis. Each request_id retrieved from Kong comes with a defined TTL. Using the Expire feature with Redis, the keys will expire with the TTL provided by Kong. All new entries are validated against Redis, and similiary any new entries are submitted to Redis. Finally, AWS CloudWatch is used to schedule the lambda function so that it will process audit logs on a cron schedule.

## Getting Started

### Requirements

There are 2 major steps to have kong audit logs publish to cloudtrails with this project:

1. Enable and configure audit logs on the Kong Global Control Plane.

2. Terraform - to deploy the AWS Infrastructure. This terraform has been validated on Terraform v1.2.3 and aws provider 4.19.0.

### How to Enable Audit Logging on the Kong Global Control Plane

**Step 1** - Audit logging is disabled by default. Enable audit logging via the kong.conf:

```shell
audit_log = on
```

or via environment variables:

```shell
export KONG_AUDIT_LOG=on
```

The reload or restart kong gateway for the gateway to detect the new config changes: `kong reload` or `kong restart`.

**Step 2** - Optional, configuration of audit/request entries generated:

There are 2 configurations available: ignore certain rest api methods, and ignore paths. Again these should be added to the kong.conf or as an environment variable:

```shell
audit_log_ignore_methods = GET,OPTIONS
audit_log_ignore_paths =/status,/audit/requests,/endpoints/
```

More details can be found on [Kong Gateway - Admin API Audit Log][audit_log]

### Deploy AWS Infrastructure - Terraform

A terraform script with `aws version "~> 4.19.0"` is provided that will create the required AWS infrastructure in an existing vpc where the Kong Global Control Plane resides.

The `prerequisite` for the terraform script is to `have an existing VPC where the Kong Control Plane runs`. The terraform script will create the following resources in the existing vpc:

* Redis ElastiCache - Cluster mode disabled, deployed in the vpc

* Security Group for the Lambda Function to access resources in the vpc

* Role and Policy  for AWS Lambda Function

* AWS Lambda Function - with the role/policies, security group, the ENVs configured in the vpc

* CloudWatch to schedule the lambda function on an 1 hr cron job

* CloudTrails Event Store - in the AWS Region

#### Terraform Variables

| Var                         | Description                                                                              | Type         | Default                                    |
|-----------------------------|------------------------------------------------------------------------------------------|--------------|--------------------------------------------|
| resource_name               | Name to provide Lambda Function and related resources                                    | string       | "kong-cloudtrails-integration"             |
| resource_tags               | Tags to set for all resources                                                            | map(string)  | {project = "kong-cloudtrail-integrations"} |
| existing_vpc                | Name of Existing VPC to deploy resources into                                            | string       |                                            |
| existing_subnet_ids         | List of Existing Subnet Ids to deploy resources into                                     | list(string) |                                            |
| security_group              | Name of the Security Group that will be created in the VPC to support the lamba function | string       |                                            |
| cloudtrail_retention        | Cloud Trail Event Data Store Retention Policy                                            | number       |                                            |
| lambda_env                  | Environment Variable to Assign the Lamba Function                                        | object       |                                            |
| lambda_env.KONG_ADMIN_API   |                                                                                          | string       |                                            |
| lambda_env.KONG_SUPERADMIN  |                                                                                          | string       |                                            |
| lambda_env.KONG_ADMIN_TOKEN |                                                                                          | string       |                                            |  
| lambda_env.REDIS_DB         |                                                                                          | string       |                                            |
| image                       | URL to Kong CloudTrails Image                                                            | string       | docker-hub-url                             |

**Note:** More details on the Lambda Env Configuration in [Lambda Environment Variables](#lambda-environment-variables) section below.

#### Example terraform.tvars file

```terraform
existing_vpc        = "vpc-uzcrqlyml0mdejmduvy"
existing_subnet_ids = ["subnet-zmuavkc6xnatd7cd1bm", "subnet-7n4ae9ua3uhjw5dhgzx", "subnet-kan0csgez5lwh5ancl0"]
security_group      = "kong-ct-sg"
lambda_env = {
  KONG_ADMIN_API   = "http://ec2-5-531-26-7.compute-1.amazonaws.com:8001"
  KONG_SUPERADMIN  = true
  KONG_ADMIN_TOKEN = "test"
  REDIS_DB         = 0
}

image         = "698461937376.dkr.ecr.us-east-1.amazonaws.com/my-repo/kong-cloudtrails-integration:latest"
resource_name = "kong-ct-integration"
```

#### Deployment

**Step 1** - Export AWS variables:

```console
export AWS_ACCESS_KEY_ID=
export AWS_SECRET_ACCESS_KEY=
```

**Step 2** - Navigate to `terraform/` in this repo and initialize the working terraform directory:

```console
terraform init
```

**Step 3** - Create the tvars file like the sample above and then create the infrastructure:

```console
terraform apply -var-file 'my-vars.tvars'
```

(Note: Terraform Destroy - AWS Lambda functions created in a VPC will have ENI's create and attached to the Security Group. It will take time to detach and destroy the ENIs)

### Understanding the AWS Lambda Configuration

### Lambda Environment Variables

The list of environment variables that can be configured on the lamba function include:

| Var              | Description                                           | Type    | Required                    | Default               | Example                     | Configurable on Terraform Script                                |
|------------------|-------------------------------------------------------|---------|-----------------------------|-----------------------|-----------------------------|-----------------------------------------------------------------|
| KONG_ADMIN_API   | Url to kong control plane, includes protocol and port | string  | yes                         | http://localhost:8001 | https://my-gateway.com:8001 | yes                                                             |
| KONG_SUPERADMIN  | true/false - is super admin enabled on the admin api  | boolean | yes                         | false                 | true                        | yes                                                             |
| KONG_ADMIN_TOKEN | Super admin token                                     | string  | yes - if superadmin enabled |                       | test                        | yes                                                             |
| REDIS_DB         | Redis Database Index                                  | string  | no                          | 0                     | 0                           | yes                                                             |
| REDIS_HOST       | url:port to redis                                     | string  | yes                         | localhost:6379        | redis-url:6379              | no - This is populated by terraform when elasticache is created |

## Local Development

For information on local development, please navigate to [Developer Walkthrough](DEVELOP.md)

## References

[Kong Gateway - Admin API Audit Log][audit_log]

<!---links-->
[audit_log]:https://docs.konghq.com/gateway/latest/admin-api/audit-log/#main