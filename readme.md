# Kong CloudTrails Integration

**WARNING**: this is an early phase POC, under active development and
             **EXPERIMENTAL**. This is not a fully supported product. For any bugs,
             issues, requests, please open an issue. Any issues
             will be fixed on a best-effort basis.

## Reference Architecture

![Kong Cloudtrails Reference Architecture](/assets/img/referenceArchitecture.png)

Kong Gateway supports Audit Logs as an Enterprise Feature. When enabled on the Global Control Plane, every api request called to the Kong admin API and DAO object created, updated, or deleted in the Kong database with relevant data including RBAC, Workspace, and a TTL for the entry. These audit log entries, in turn, are retrievable via endpoints on the Kong admin api: /audit/requests, and /audit/objects.

For the Kong integration with the CloudTrail Lake,an AWS Lambda function in combination with AWS ElastiCache-Redis are deployed into existing VPC where the Kong Global Control Plane resides. The lambda function will strictly call the /audit/requests endpoint, process and remove duplicate entries by evaluating existing keys in Redis before transforming and submitting the audit log entries to cloudtrails. Each request_id retrieved from Kong comes with a defined TTL. All new entries are validated against Redis, and similiary any new entries are submitted to Redis. Finally, AWS CloudWatch is used to schedule the lambda function so that it will process audit logs on an hourly schedule.

## Getting Started

### Requirements

There are 3 major steps to have kong audit logs publish to CloudTrail Lake with this project:

1. Have created an AWS Channel to CloudTrail and can provide the Channel ARN.

2. Enable and configure audit logs on the Kong Global Control Plane.

3. Terraform - to deploy the AWS Infrastructure. This terraform has been validated on Terraform v1.2.3 and aws provider 4.19.0.

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

A terraform script with the belwos version requirements is provided that will create the required AWS infrastructure in an existing vpc where the Kong Global Control Plane resides.

#### Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= v1.2.3 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | ~> 4.19.0 |

 **Providers**

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | ~> 4.19.0 |

#### Prequisites

The `prerequisites` for the terraform script are:

1. Provide an existing VPC, and Kong Gateway running in the existing vpc.

2. At least 2 subnets ids in the VPC - to support elasticache.

3. Have all env variables to connect to Kong Gateway and CloudTrail event data store: Channel ARN, Kong Gateway URL, certificates, and Admin Token (if required).

The following resources will be created in the VPC:

* Redis ElastiCache - 2 Node (primary and replica), Multi-AZ Enabled, Auto-failure Enabled, Cluster mode off and hosted in the provided VPC.

* Security Group for the lambda function to access resources in the vpc.

* Role and Policy  for the lambda function.

* Lambda Function - with the role/policies, security group, image, the environment variables configured and hosted in the provided VPC.

* CloudWatch to schedule the lambda function on an 1 hr job.

* CloudTrails Event Store - created in the appropriate region.

#### Terraform Variables

| Var                         | Description                                                                                                                                    | Type         | Default                                    |
|-----------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|--------------|--------------------------------------------|
| resource_name               | Name to provide Lambda Function and related resources                                                                                          | string       | "kong-cloudtrails-integration"             |
| resource_tags               | Tags to set for all resources                                                                                                                  | map(string)  | {project = "kong-cloudtrail-integrations"} |
| aws_region                  | AWS Region where Existing Resources will be deployed                                                                                           | string       |                                            |
| existing_vpc                | Name of Existing VPC to deploy resources into                                                                                                  | string       |                                            |
| existing_subnet_ids         | List of Existing Subnet Ids to deploy resources into                                                                                           | list(string) |                                            |
| security_group              | Name of the Security Group that will be created in the VPC to support the lamba function                                                       | string       |                                            |
| lambda_env                  | Environment Variable to Assign the Lamba Function  - More details [Lambda Environment Variables](#lambda-environment-variables) section below. | object       |                                            |
| lambda_env.KONG_ADMIN_API   | URL to the Kong Gateway Admin API                                                                                                              | string       |                                            |
| lambda_env.KONG_SUPERADMIN  | Define if a superadmin token is needed to call Kong Admin API                                                                | bool         |                                            |
| lambda_env.KONG_ADMIN_TOKEN | The Kong super user admin token                                                                                                           | string       |                                            |
| lambda_env.KONG_ROOT_CA     | Required if a Custom CA is configured on the Kong Admin API. If not required null can be used                                     | string       |                                            |
| lambda_env.REDIS_DB         | Recommend define as 0                                                                                                                          | string       |                                            |
| image                       | URL to Kong CloudTrails Image, kong/cloudtrails-integration is available in dockerhub.                                                                                                                 | string       |                             |


**Note:** More details on the Lambda Env Configuration in [Lambda Environment Variables](#lambda-environment-variables) section below.

#### Example terraform.tvars file

```terraform
aws_region          = "us-east-1"
existing_vpc        = "vpc-uzcrqlyml0mdejmduvy"
existing_subnet_ids = ["subnet-zmuavkc6xnatd7cd1bm", "subnet-7n4ae9ua3uhjw5dhgzx", "subnet-kan0csgez5lwh5ancl0"]
security_group      = "kong-ct-sg"
lambda_env = {
  KONG_ADMIN_API   = "http://ec2-5-531-26-7.compute-1.amazonaws.com:8001"
  KONG_SUPERADMIN  = true
  KONG_ADMIN_TOKEN = "test"
  KONG_ROOT_CA     = ""
  REDIS_DB         = 0
  CHANNEL_ARN     = "arn:aws:cloudtrail:us-east-1:123456789651:channel/07441ab6-c4a1-4c8a-943d-a2f0c50c8a76"
}

image         = "kong/cloudtrails-integration:1.0.0"
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

**Step 3** - Create the tvars file like the sample above and then the command below to create the execution plan:

```console
terraform plan -out=plan.out -var-file 'my-vars.tvars'
```

**Step 4** - Execute the actions defined in the plan:

```console
terraform apply "plan.out"
```

(Note: Terraform Destroy - AWS Lambda functions created in a VPC will have ENI's create and attached to the Security Group. It will take time to detach and destroy the ENIs)

### Understanding the AWS Lambda Configuration

#### Lambda Environment Variables

The list of environment variables that can be configured on the lamba function include:

| Var              | Description                                           | Type    | Required                    | Default               | Example                                                                              | Configurable on Terraform Script                                                                                               |
|------------------|-------------------------------------------------------|---------|-----------------------------|-----------------------|--------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| KONG_ADMIN_API   | Url to kong control plane, includes protocol and port | string  | yes                         | http://localhost:8001 | https://my-gateway.com:8001                                                          | yes                                                                                                                            |
| KONG_SUPERADMIN  | true/false - is super admin enabled on the admin api  | boolean | yes                         | false                 | true                                                                                 | yes                                                                                                                            |
| KONG_ADMIN_TOKEN | Kong Admin Token                                      | string  | yes - if superadmin enabled |                       | test                                                                                 | yes                                                                                                                            |
| KONG_ROOT_CA     | Root CA that matches the certificate configured on the Kong Admin API      | string  | no                          |                       | disabled                                                                             | yes |
| REDIS_DB         | Redis Database Index                                  | string  | no                          | 0                     | 0                                                                                    | yes                                                                                                                            |
| REDIS_HOST       | url:port to redis                                     | string  | yes                         | localhost:6379        | redis-url:6379                                                                       | no - populated by terraform when elasticache is created                                                                |
| CHANNEL_ARN      | channel_arn provided by AWS to ingest events          | string  | yes                         |                       | arn:aws:cloudtrail:us-east1:01234567890:channel/EXAMPLE8-0558-4f7e-a06a-43969EXAMPLE | yes                                                                                                                            |

## Overview of Event Data Audit Entry Published to AWS CloudTrails

Here is quick review of how a Kong Audit Request Entry is Mapped to a CloudTrail EventData Entry.

**Sample Response from /audit/requests on Kong Admin API**

```json
{
  "client_ip": "172.31.76.246",
  "signature": null,
  "removed_from_payload": null,
  "status": 200,
  "ttl": 2591940,
  "rbac_user_id": null,
  "path": "/event-hooks/?size=100",
  "payload": null,
  "request_timestamp": 1663013700,
  "workspace": "cca82c73-2365-441b-8860-9e074d93b205",
  "method": "GET",
  "request_id": "8XsOaNYwXgn0VmgA8qYYpVteu14p9GIK"
}
```

| Event Data Field              | /audit/request Attribute | Description                                                                                                                                                                      | Example                                                                                                                        |
|-------------------------------|--------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| version                       |                          |                                                                                                                                                                                  |                                                                                                                                |
| useridentity                  | rbac_user_id             | if rbac_user_is null, useridentity.principalId set to "anonymous" and details appended                                                                                           | useridentity={type=, principalid=anonymous, details={RBAC=Anonymous User on Kong Gateway: Please Enable RBAC on Kong Gateway}} |
| eventsource                   | N/A                      | static variable to identify the source as Kong                                                                                                                                   | kong-gateway                                                                                                                   |
| eventName                     | combine "method"+"path"  | query parameters are stripped from "path". Per AWS standards, "method"+"path" concatenated to represent event                                                                    | GET/event-hooks                                                                                                                |
| event-time                    | request_timestamp        | request_timestamp translated to UTC format.                                                                                                                                      | 2022-09-13 01:09:14.000                                                                                                        |
| uid                           | request_id               | 1-to-1 mapping with request_id in Kong Audit Log Entry.                                                                                                                          | 0rZdMtxyIzgje15RAdkL4YFIF8Z9sNQK                                                                                               |
| requestParameters             |                          | query parameters stripped from "path".                                                                                                                                           | {queryParameters=size=100}                                                                                                     |
| responseParameters            | N/A                      | Not apart of the Kong Audit Log description.                                                                                                                                     | null                                                                                                                           |
| errorcode                     | N/A                      | Not apart of the Kong Audit Log description.                                                                                                                                     | null                                                                                                                           |
| errormessage                  | N/A                      | Not apart of the Kong Audit Log description.                                                                                                                                     | null                                                                                                                           |
| sourceipaddress               | client_ip                | 1-to-1 mapping with client_ip in Kong Audit Log Entry.                                                                                                                           | 172.31.76.246                                                                                                                  |
| additionaleventdata           |                          | Store any additional information and/or same information in original raw format from Kong Audit Log for tracking purposes.                                                       |                                                                                                                                |
| additionaleventdata.workspace | workspace                | Workspace affected in Kong Admin Portal.                                                                                                                                         | cca82c73-2365-441b-8860-9e074d93b205                                                                                           |
| additionaleventdata.hostname  |                          | Hostname defined on Kong Gateway, retrieved from info endpoint(/) on Kong Admin API                                                                                              | http://my-kong-gateway.com:8001                                                                                                |
| additionaleventdata.method    | method                   | 1-1 mapping with method on Kong Audit Log Entry.                                                                                                                                 | GET                                                                                                                            |
| additionaleventdata.signature | signature                | 1-to-1 mapping with signature on Kong Audit Log Entry.                                                                                                                           |                                                                                                                                |
| additionaleventdata.ttl       | ttl                      | 1-to-1 mapping with ttl on Kong Audit Log Entry. Each Audit Log has a ttl in the Kong Gateway Database. This ttl is also used in the backend Redis database to avoid duplicates. | 2591940                                                                                                                        |
| additionaleventdata.status    | status                   | 1-to-1 mapping with status on Kong Audit Log Entry.                                                                                                                              | 200                                                                                                                            |

## Local Development

For information on local development, please navigate to [Developer Walkthrough](DEVELOP.md)

## References

[Kong Gateway - Admin API Audit Log][audit_log]

<!---links-->
[audit_log]:https://docs.konghq.com/gateway/latest/admin-api/audit-log/#main