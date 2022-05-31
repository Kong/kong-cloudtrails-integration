FROM golang:1.18.2-alpine AS build_base 

WORKDIR /tmp/kong-cloudtrails-integration
COPY go.mod . 
COPY go.sum . 
COPY aws-sdk-go@v0.0.0-unpublished ./aws-sdk-go@v0.0.0-unpublished

# RUN go mod download 

COPY . . 

RUN CGO_ENABLED=0 go test -v 

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./out/kong-cloudtrails-integration .  

FROM public.ecr.aws/lambda/go:1

COPY --from=build_base /tmp/kong-cloudtrails-integration/out/kong-cloudtrails-integration ${LAMBDA_TASK_ROOT}

CMD [ "kong-cloudtrails-integration" ]
