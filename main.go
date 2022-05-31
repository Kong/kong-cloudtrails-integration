package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	log "github.com/sirupsen/logrus"

	"github.com/Kong/kong-cloudtrails-integration/internal/service"
)

type ScheduleEvent struct {
	Schedule bool `json:"schedule"`
}
type Response struct {
	Message string `json:"message"`
}

func HandleLambdaEvent(ctx context.Context, event ScheduleEvent) (Response, error) {
	lc, _ := lambdacontext.FromContext(ctx)
	arn := lc.InvokedFunctionArn

	c := service.InitController()
	c.SetAwsARN(arn)

	service.SetController(c)

	err := service.HandleLogs()

	response := Response{}
	if err != nil {
		log.Error("Error occured", err.Error())
		response.Message = err.Error()
		return response, err
	}
	response.Message = "success"

	return response, nil
}

func main() {

	lambda.Start(HandleLambdaEvent)
}
