package awsClient

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/Kong/kong-cloudtrails-integration/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtraildataservice"
	log "github.com/sirupsen/logrus"
)

type AWSClient struct {
	Sess *session.Session
}

func New(region string) *AWSClient {
	ac := AWSClient{
		Sess: session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String(region),
			},
		})),
	}
	return &ac
}

func (ac *AWSClient) PutAuditLogs(al *model.AuditLogs, arn string) error {
	ctd := cloudtraildataservice.New(ac.Sess)

	input := cloudtraildataservice.PutAuditEventsInput{
		AuditEvents: []*cloudtraildataservice.AuditEvent{},
	}

	awsAccountId := createRecipientId(arn)

	for _, v := range al.Logs {

		ae := ac.transformAuditEvent(&v, al.KongData, awsAccountId)

		if ae != nil {
			input.AuditEvents = append(input.AuditEvents, ae)
		}
	}

	log.Info("Number of records sending to CloudTrails: ", len(input.AuditEvents))
	if len(input.AuditEvents) == 0 {
		return nil
	}

	resp, err := ctd.PutAuditEvents(&input)
	if err != nil {
		return err
	}

	if len(resp.Failed) > 0 {
		log.Error(resp.Failed)
		return errors.New("failed requests submitting to cloudtrails")
	}

	return nil
}

func (ac *AWSClient) transformAuditEvent(ar *model.AuditRequest, kongInfo model.KongData, awsAccountId string) *cloudtraildataservice.AuditEvent {
	if ar == nil {
		return nil
	}

	id := ar.Request_id
	timestamp := time.Unix(int64(ar.Request_timestamp), 0).UTC().Format(time.RFC3339)
	url, qp := splitEventNameQueryParameters(ar.Path)
	eventName := createEventName(ar.Method, url)
	rp := createRequestParameters(ar.Payload, qp)
	ui := createUserIdentity(ar.Rbac_user_id)

	eventData := model.OpenAuditEvent{
		EventVersion:      kongInfo.KongVersion,
		EventSource:       "kong-gateway",
		EventTime:         timestamp,
		EventName:         eventName,
		RequestParameters: rp,
		SourceIPAddress:   ar.Client_ip,
		AdditionalEventData: model.AdditionalEventData{
			Method:       ar.Method,
			Status:       ar.Status,
			Signature:    ar.Signature,
			Ttl:          ar.Ttl,
			Workspace:    ar.Workspace,
			KongHostname: kongInfo.KongHostname,
		},
		RecipientAccountId: awsAccountId,
		UUID:               ar.Request_id,
		UserIdentity:       ui,
	}

	ed, err := json.Marshal(&eventData)

	if err != nil {
		log.Error("Error tranforming AuditRequest to AWS PutAuditEvent: ", err.Error())
		return nil
	}

	return &cloudtraildataservice.AuditEvent{
		Id:        aws.String(id),
		EventData: aws.String(string(ed)),
	}
}

func splitEventNameQueryParameters(path string) (eventName string, qp string) {
	i := strings.Index(path, "?")
	eventName = path
	qp = ""

	if i != -1 {
		eventName = path[0:i]
		qp = path[i+1:]
	}
	return eventName, qp
}

func createRecipientId(arn string) string {
	s := strings.Split(arn, ":")

	return s[4]
}

func createUserIdentity(rbacId string) model.UserIdentity {

	pi := rbacId
	details := make(map[string]interface{})

	if rbacId == "" {
		pi = "anonymous"
		details["RBAC"] = "Anonymous User on Kong Gateway: Please Enable RBAC on Kong Gateway"
	}

	return model.UserIdentity{Type: "", PrincipalId: pi, Details: details}
}

func createEventName(kongMethod string, kongPath string) string {
	return kongMethod + kongPath[1:]
}

func createRequestParameters(payload string, queryParams string) map[string]interface{} {

	rp := make(map[string]interface{})
	json.Unmarshal([]byte(payload), &rp)
	if queryParams != "" {
		rp["queryParameters"] = queryParams
	}

	return rp
}
