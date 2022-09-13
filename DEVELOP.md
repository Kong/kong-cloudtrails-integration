# Developer Walkthrough

High Level this guide walks through:

* Setting up Kong with audit logs enabled.

* Setting up AWS Infrastructure: ElastiCache, ECR Private Repository, VPC requirements for the Lambda Function.

* Building and running the AWS Lamdba code locally.

* Pushing and configuring the Lambda Function in AWS.

## Setup Infrastructure on AWS - Kong + Kong DB (Docker), AWS ElastiCache, CloudTrails Event Store

### Setup Kong - Docker Install on AWS EC2

Create the EC2 AMI instance:

```shell
aws ec2 run-instances \
    --image-id  ami-0022f774911c1d690 \
    --count 1 --instance-type t2.micro \
    --key-name df-keypair \
    --security-group-ids sg-0d6e38be721c7f8a8 \
    --subnet-id subnet-09f5e9093a3053cbb \
    --associate-public-ip-address
```

ssh into the box and install docker:

```shell
sudo amazon-linux-extras install docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -a -G docker ec2-user
```

kill the session and ssh back into the ec2-user to install kong:

```shell
docker network create kong-net
```

```shell
 docker run -d --name kong-database \
  --network=kong-net \
  -p 5432:5432 \
  -e "POSTGRES_USER=kong" \
  -e "POSTGRES_DB=kong" \
  -e "POSTGRES_PASSWORD=kongpass" \
  postgres:9.6
```

```shell
docker run --rm --network=kong-net \
  -e "KONG_DATABASE=postgres" \
  -e "KONG_PG_HOST=kong-database" \
  -e "KONG_PG_PASSWORD=kongpass" \
  -e "KONG_PASSWORD=test" \
 kong/kong-gateway:2.8.1.1-alpine kong migrations bootstrap
```

Export your license as an ENV, example below:

```shell
export KONG_LICENSE_DATA='{"license":{"payload":{"admin_seats":"1","customer":"Example Company, Inc","dataplanes":"1","license_creation_date":"2017-07-20","license_expiration_date":"2017-07-20","license_key":"00141000017ODj3AAG_a1V41000004wT0OEAU","product_subscription":"Konnect Enterprise","support_plan":"None"},"signature":"6985968131533a967fcc721244a979948b1066967f1e9cd65dbd8eeabe060fc32d894a2945f5e4a03c1cd2198c74e058ac63d28b045c2f1fcec95877bd790e1b","version":"1"}}'
```

Start the gateway with audit logs enabled:

```shell
docker run -d --name kong-gateway \
  --network=kong-net \
  -e "KONG_DATABASE=postgres" \
  -e "KONG_PG_HOST=kong-database" \
  -e "KONG_PG_USER=kong" \
  -e "KONG_PG_PASSWORD=kongpass" \
  -e "KONG_PROXY_ACCESS_LOG=/dev/stdout" \
  -e "KONG_ADMIN_ACCESS_LOG=/dev/stdout" \
  -e "KONG_PROXY_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_ERROR_LOG=/dev/stderr" \
  -e "KONG_ADMIN_LISTEN=0.0.0.0:8001" \
  -e "KONG_ADMIN_GUI_URL=http://localhost:8002" \
  -e KONG_LICENSE_DATA \
  -e KONG_AUDIT_LOG=on \
  -p 8000:8000 \
  -p 8443:8443 \
  -p 8001:8001 \
  -p 8444:8444 \
  -p 8002:8002 \
  -p 8445:8445 \
  -p 8003:8003 \
  -p 8004:8004 \
  kong/kong-gateway:2.8.1.1-alpine
```

**Optional Audit log Configurations** to ignore unecessary audit logs entries, to test in a temporary fashion:

```shell
docker exec -it kong-gateway /bin/bash 
export KONG_AUDIT_LOG_IGNORE_METHODS=GET,OPTIONS
export KONG_AUDIT_LOG_IGNORE_PATHS=/status,/audit/requests,/endpoints/
kong reload
```

To make the audit log configuration permanent, included these are envs when first spin up the gateway container or appropriately update the kong conf.

### Setup AWS ElastiCache

Optional, you can run redis on the ec2 instance with Kong if you don't want to spin up AWS ElastiCache for testing:

```shell
docker run --name kong-redis -d --network=kong-net -p 6379:6379 redis
```

**AWS Redis ElastiCache**
Create a subnet-group for ElastiCache:

```shell
aws elasticache create-cache-subnet-group \
    --cache-subnet-group-name kong-cloudtrails-integration \
    --cache-subnet-group-description "kong cloudtrails integration" \
    --subnet-ids <subnet-id1> <subnet-id2> <subnet-id3>
```

Create a Single Node Redis ElastiCache:

```shell
aws elasticache create-cache-cluster \
--cache-cluster-id kong-cloudtrails-integration \
--cache-node-type cache.t3.small \
--engine redis \
--engine-version 6.2 \
--num-cache-nodes 1 \
--cache-parameter-group default.redis6.x \
--cache-subnet-group kong-cloudtrails-integration \
--security-group-ids <sg-1>
```

Store the Endpoint on the cluster, this is added to the Lambda ENV variables as REDIS_HOST

# Build and Test AWS Lambda Function Locally

## Build the docker image
docker build -t kong-cloudtrails-integration .

## Run the docker image

```shell
docker run -p 9000:8080 -e AWS_ACCESS_KEY_ID=<access-key> \
  -e AWS_SECRET_ACCESS_KEY=<secret> \
  -e REGION=us-east-1 \
  -e KONG_ADMIN_API=http://ec2-3-228-219-180.compute-1.amazonaws.com:8001 \
  -e KONG_SUPERADMIN=true \
  -e KONG_ADMIN_TOKEN=test \
  -e REDIS_HOST=kong-cloudtrails-integration.klcj8d.0001.use1.cache.amazonaws.com:6379 \
  -e REDIS_DB=0 \
  kong-cloudtrails-integration
```

## Send a CURL request with a payload

curl -XPOST `"http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"ScheduleEvent": true}'`

# Build and Test On AWS Lambda Function on AWS

## Build the Lambda Image

Build the image and host in your preferred registry: AWS ECR, etc.

```shell
docker build -t kong-cloudtrails-integration .
```

### Configure AWS Lambda Function

#### Create the IAM Role

1. Deploy lambda function via the console: [AWS Lambda - Deploy Lambda function as container images](https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-images.html#configuration-images-create).

2. Update the permissions of the IAm role generated for the lambda function:
 
 * Create a new policy, Kong-Cloudtrails. The policy json is located under aws-resources/kong-cloudtrails-policy.json.

 * Attach the policy to the iam role generated for the lambda function

 * Attach the AWS Managed policy, AWSLambdaVPCAccessExecutionRole, to the role.

3. Configure the Lambda function Env Variables via the Functions Console, sample below:

```shell
KONG_ADMIN_API http://ec2-3-239-173-2.compute-1.amazonaws.com:8001
KONG_SUPERADMIN true
KONG_ADMIN_TOKEN test
REDIS_HOST kong-cloudtrails-integration.klcj8d.ng.0001.use1.cache.amazonaws.com:6379
REDIS_DB 0
CHANNEL_ARN arn:aws:cloudtrail:us-east-1:123456789651:channel/07441ab6-c4a1-4c8a-943d-a2f0c50c8a76
```

4. Schedule lamda function to run every hour via CloudWatch: [Tutorial: Schedule AWS Lambda Functions Using CloudWatch Events](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/RunLambdaSchedule.html)

# References

* [Container Image on Amazon EC2 - Installing Docker on AMI 2](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/create-container-image.html)

* [AWS Lambda - Creating Lambda container images](https://docs.aws.amazon.com/lambda/latest/dg/images-create.html)

* [AWS Lambda - Deploy Lambda function as container images](https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-images.html#configuration-images-create)

* [Tutorial: Schedule AWS Lambda Functions Using CloudWatch Events](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/RunLambdaSchedule.html)

* [AWS ElastiCache for Redis - Create Subnet Group](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/SubnetGroups.Creating.html)

* [AWS ElastiCache for Redis - Creating a Cluster](https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/Clusters.Create.html)

* [Kong Gatway - Admin API Audit Log](https://docs.konghq.com/gateway/latest/admin-api/audit-log/)

* [Kong Gateway - Install Kong Gateway On Docker](https://docs.konghq.com/gateway/2.8.x/install-and-run/docker/)