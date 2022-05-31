package kongClient

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Kong/kong-cloudtrails-integration/model"

	"github.com/go-resty/resty/v2"
)

type KongRestClient struct {
	client *resty.Client
}

func NewRestClient(baseUrl string, superAdmin string, superAdminToken string) *KongRestClient {

	kc := KongRestClient{
		client: resty.New().SetBaseURL(strings.ToLower(baseUrl)).
			SetHeader("Accept", "application/json").
			SetQueryParam("size", "100"),
	}

	admin, _ := strconv.ParseBool(superAdmin)

	if admin {
		kc.client.SetHeader("Kong-Admin-Token", superAdminToken)
	}

	return &kc
}

func (kc *KongRestClient) CallAuditLog(offset string) (*model.AuditLogs, []string, error) {
	var ar = model.AuditRequests{}
	var requestIds []string

	al := model.AuditLogs{
		Logs: make(map[string]model.AuditRequest),
	}

	if offset != "" {
		kc.client.QueryParam.Set("offset", offset)
	} else {
		kc.client.QueryParam.Del("offset")
	}

	resp, err := kc.client.R().Get("/audit/requests")

	if err != nil {
		log.Error("En error occured calling Kong Admin ApI", err)
		return nil, nil, err
	}

	if !resp.IsSuccess() {
		msg := "Unsuccessful Response from Kong Admin API: " + string(resp.Status()) + "\t" + string(resp.Body())
		return nil, nil, errors.New(msg)

	}

	json.Unmarshal(resp.Body(), &ar)
	requestIds = make([]string, 0, len(ar.Data))
	for _, log := range ar.Data {
		al.Logs[log.Request_id] = log
		requestIds = append(requestIds, log.Request_id)
	}
	al.Offset = ar.Offset

	return &al, requestIds, nil
}

func (kc *KongRestClient) GetKongInfo() (model.KongData, error) {

	var info = model.KongData{}

	resp, err := kc.client.R().Get("/")

	if err != nil {
		return info, err
	}

	if !resp.IsSuccess() {
		msg := "Error from Kong Admin API: " + string(resp.Status()) + "\t" + string(resp.Body())
		return info, errors.New(msg)
	}

	json.Unmarshal(resp.Body(), &info)

	return info, nil
}
