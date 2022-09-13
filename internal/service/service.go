package service

import (
	log "github.com/sirupsen/logrus"

	"github.com/Kong/kong-cloudtrails-integration/model"
)

var c *Controller

func SetController(cont *Controller) {
	c = cont
}

func HandleLogs() error {
	offset := ""
	counter := 0

	for {
		al, ri, err := c.KC.CallAuditLog(offset)
		if err != nil || al == nil {
			log.Fatalf("Error calling Kong /audit/requests endpoint: %s", err.Error())
		}

		kongInfo, err := c.KC.GetKongInfo()
		kongInfo.KongHostname = c.Kconfig.KONG_ADMIN_API

		if err != nil {
			log.Fatalf("Error calling Kong info(/) endpoint: %s", err.Error())
		}
		al.KongData = kongInfo

		al = removeDups(al, ri)
		err = c.AC.PutAuditLogs(al, c.GetFuncArn(), c.GetChannelArn())
		if err != nil {
			log.Infof("Error Publishing Logs to Cloudtrails: %s duplicates still recorded in ElastiCache", err.Error())
		}

		recordNewDUps(al)
		if al.Offset == "" {
			break
		}
		log.Info("iteration: ", counter)
		counter++
		offset = al.Offset
	}

	return nil
}

func removeDups(al *model.AuditLogs, requestIds []string) *model.AuditLogs {
	dups := c.KR.RequestIdsExist(requestIds...)

	for i, v := range dups {
		if v == 1 {
			delete(al.Logs, requestIds[i])
		}
	}
	return al
}

func recordNewDUps(al *model.AuditLogs) {
	c.KR.SetRequestIds(al)
}
