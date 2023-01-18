# Developer Walkthrough

High Level this guide walks through:

* Setting up Kong with audit logs enabled.

* Setting up AWS Infrastructure: ElastiCache, ECR Private Repository, VPC requirements for the Lambda Function.

* Building and running the AWS Lamdba code locally.

* Pushing and configuring the Lambda Function in AWS.

## Setup Infrastructure on AWS - Kong + Kong DB (Docker), AWS ElastiCache, CloudTrails Event Store

### Setup Kong - Docker Install on AWS EC2

Create the EC2 AMI instance.

ssh into the box and install docker:

```shell
sudo amazon-linux-extras install docker
sudo systemctl enable docker
sudo systemctl start docker
sudo usermod -a -G docker ec2-user
```

kill the session and ssh back into the ec2-user to install kong:

```console
docker network create kong-net
```

```console
 docker run -d --name kong-database \
  --network=kong-net \
  -p 5432:5432 \
  -e "POSTGRES_USER=kong" \
  -e "POSTGRES_DB=kong" \
  -e "POSTGRES_PASSWORD=kongpass" \
  postgres:9.6
```

```console
docker run --rm --network=kong-net \
  -e "KONG_DATABASE=postgres" \
  -e "KONG_PG_HOST=kong-database" \
  -e "KONG_PG_PASSWORD=kongpass" \
  -e "KONG_PASSWORD=test" \
 kong/kong-gateway:2.8.1.1-alpine kong migrations bootstrap
```

Export your license as an ENV, example below:

```console
export KONG_LICENSE_DATA='{"license":{"payload":{"admin_seats":"1","customer":"Example Company, Inc","dataplanes":"1","license_creation_date":"2017-07-20","license_expiration_date":"2017-07-20","license_key":"00141000017ODj3AAG_a1V41000004wT0OEAU","product_subscription":"Konnect Enterprise","support_plan":"None"},"signature":"6985968131533a967fcc721244a979948b1066967f1e9cd65dbd8eeabe060fc32d894a2945f5e4a03c1cd2198c74e058ac63d28b045c2f1fcec95877bd790e1b","version":"1"}}'
```

Start the gateway with audit logs enabled:

```console
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

```console
docker exec -it kong-gateway /bin/bash 
export KONG_AUDIT_LOG_IGNORE_METHODS=GET,OPTIONS
export KONG_AUDIT_LOG_IGNORE_PATHS=/status,/audit/requests,/endpoints/
kong reload
```

To make the audit log configuration permanent, included these are envs when first spin up the gateway container or appropriately update the kong conf.

# Build and Test AWS Lambda Function Locally

## Build the docker image

From the parent directory execute the following command:

```console
docker build -t local-kong-cloudtrails-integration .
```

## Run the docker image

```console
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

Host in your preferred registry: AWS ECR, etc, and update the terraform vars to use the appropriate image and spin up the infrastructure.