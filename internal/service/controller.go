package service

import (
	"github.com/Kong/kong-cloudtrails-integration/internal/awsClient"
	"github.com/Kong/kong-cloudtrails-integration/internal/database"
	"github.com/Kong/kong-cloudtrails-integration/internal/kongClient"
	"github.com/Kong/kong-cloudtrails-integration/internal/utils"
)

type KConfig struct {
	awsArn           string
	KONG_ADMIN_API   string
	KONG_SUPERADMIN  string
	KONG_ADMIN_TOKEN string
	REDIS_DB         int
	REDIS_HOST       string
	REDIS_USERNAME   string
	REDIS_PASSWORD   string
	AWS_REGION       string
}

type Controller struct {
	KC      *kongClient.KongRestClient
	AC      *awsClient.AWSClient
	KR      *database.Database
	Kconfig KConfig
}

func InitController() *Controller {

	kAdminApi := utils.GetEnv("KONG_ADMIN_API", "http://localhost:8001")
	kSuperAdmin := utils.GetEnv("KONG_SUPERADMIN", "false")
	kongAdminToken := utils.GetEnv("KONG_ADMIN_TOKEN", "")

	awsR := utils.GetEnv("REGION", "us-east-1")

	rHost := utils.GetEnv("REDIS_HOST", "localhost:6379")
	rUser := utils.GetEnv("REDIS_USERNAME", "")
	rPass := utils.GetEnv("REDIS_PASSWORD", "")
	rDb := utils.GetEnvInt("REDIS_DB", 0)

	return &Controller{
		KC: kongClient.NewRestClient(kAdminApi, kSuperAdmin, kongAdminToken),
		AC: awsClient.New(awsR),
		KR: database.New(rHost, rUser, rPass, rDb),
		Kconfig: KConfig{
			KONG_ADMIN_API:   kAdminApi,
			KONG_SUPERADMIN:  kSuperAdmin,
			KONG_ADMIN_TOKEN: kongAdminToken,
			REDIS_HOST:       rHost,
			REDIS_USERNAME:   rUser,
			REDIS_PASSWORD:   rPass,
			REDIS_DB:         rDb,
			AWS_REGION:       awsR,
		},
	}
}

func (c *Controller) SetAwsARN(arn string) {
	c.Kconfig.awsArn = arn
}

func (c *Controller) GetAWsARN() string {
	return c.Kconfig.awsArn
}
