package awsClient

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Kong/kong-cloudtrails-integration/model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudtraildata"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestTransformAuditEvent(t *testing.T) {

	ar := &model.AuditRequest{
		Client_ip:            "172.18.0.1",
		Request_id:           "0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp",
		Request_timestamp:    1655123134,
		Ttl:                  2504744,
		Signature:            "",
		Rbac_user_id:         "",
		Workspace:            "0c21a8bb-0e63-4cf1-8b98-7038e1f25468",
		Payload:              "",
		Path:                 "/auth?session_logout=true",
		Method:               "OPTIONS",
		Status:               200,
		Removed_from_payload: "",
	}
	kongInfo := model.KongData{
		KongVersion:  "2.8.1.1-enterprise-edition",
		KongHostname: "http://kong-gateway.com",
	}

	awsAccountId := "123456789"
	ac := &AWSClient{}

	expected := &cloudtraildata.AuditEvent{
		Id:        aws.String("0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp"),
		EventData: aws.String("{\"version\":\"2.8.1.1-enterprise-edition\",\"userIdentity\":{\"type\":\"\",\"principalId\":\"anonymous\",\"details\":{\"RBAC\":\"Anonymous User on Kong Gateway: Please Enable RBAC on Kong Gateway\"}},\"eventSource\":\"kong-gateway\",\"eventName\":\"OPTIONSauth\",\"eventTime\":\"2022-06-13T12:25:34Z\",\"UID\":\"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp\",\"requestParameters\":{\"queryParameters\":\"session_logout=true\"},\"sourceIPAddress\":\"172.18.0.1\",\"additionalEventData\":{\"method\":\"OPTIONS\",\"status\":200,\"signature\":\"\",\"ttl\":2504744,\"workspace\":\"0c21a8bb-0e63-4cf1-8b98-7038e1f25468\",\"konghostname\":\"http://kong-gateway.com\"},\"recipientAccountId\":\"123456789\"}"),
	}

	auditEvent := ac.transformAuditEvent(ar, kongInfo, awsAccountId)
	assert.Equal(t, expected, auditEvent)
}

func TestTransformReturnNil(t *testing.T) {
	ac := AWSClient{}
	kongInfo := model.KongData{
		KongVersion:  "2.8.1.1-enterprise-edition",
		KongHostname: "http://kong-gateway.com",
	}
	ae := ac.transformAuditEvent(nil, kongInfo, "12354545")
	assert.Nil(t, ae)
}

func TestGetRecipientId(t *testing.T) {
	functionArns := []string{
		"arn:aws:rds:us-east-2:123456789012:db:my-mysql-instance-1",
		"arn:aws:::123456789012:db:my-mysql-instance-1",
	}

	for _, v := range functionArns {
		output := createRecipientId(v)
		assert.Equal(t, output, "123456789012")
	}
}

func TestSplitEventNameQueryParameters(t *testing.T) {

	eventNames := []string{
		"/services",
		"/audit/requests?request_id=oops",
	}

	en1, qp1 := splitEventNameQueryParameters(eventNames[0])
	assert.Equal(t, en1, "/services")
	assert.Equal(t, qp1, "")

	en2, qp2 := splitEventNameQueryParameters(eventNames[1])
	assert.Equal(t, "/audit/requests", en2)
	assert.Equal(t, "request_id=oops", qp2)
}

func TestCreateEventName(t *testing.T) {
	method := "GET"
	path := "/services"

	expected := "GETservices"
	actual := createEventName(method, path)
	assert.Equal(t, expected, actual)

}

func TestCreateUserIdentity(t *testing.T) {
	rbacIds := []string{"", "123lnkl43"}

	e1 := model.UserIdentity{
		Type:        "",
		PrincipalId: "anonymous",
		Details: map[string]interface{}{
			"RBAC": "Anonymous User on Kong Gateway: Please Enable RBAC on Kong Gateway",
		},
	}

	e2 := model.UserIdentity{
		Type:        "",
		PrincipalId: "123lnkl43",
		Details:     make(map[string]interface{}),
	}
	actual1 := createUserIdentity(rbacIds[0])
	actual2 := createUserIdentity(rbacIds[1])

	assert.Equal(t, e1, actual1)
	assert.Equal(t, e2, actual2)

}

func TestPutAuditLogs(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://cloudtrail-data.us-east-1.amazonaws.com/PutAuditEvents",
		func(req *http.Request) (*http.Response, error) {

			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"Failed":     []*cloudtraildata.ResultErrorEntry{},
				"Successful": []*cloudtraildata.AuditEventResultEntry{},
			})
		},
	)

	ac := AWSClient{
		Sess: session.Must(session.NewSession(aws.NewConfig().WithRegion("us-east-1").WithCredentials(credentials.AnonymousCredentials))),
	}

	httpmock.ActivateNonDefault(ac.Sess.Config.HTTPClient)

	al := model.AuditLogs{
		Logs: map[string]model.AuditRequest{
			"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp": {
				Client_ip:            "172.18.0.1",
				Request_id:           "0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp",
				Request_timestamp:    0,
				Ttl:                  2504744,
				Signature:            "",
				Rbac_user_id:         "",
				Workspace:            "0c21a8bb-0e63-4cf1-8b98-7038e1f25468",
				Payload:              "",
				Path:                 "/auth?session_logout=true",
				Method:               "OPTIONS",
				Status:               200,
				Removed_from_payload: "",
			},
		},
	}
	functionArn := "arn:aws:rds:us-east-2:123456789012:db:my-mysql-instance-1"
	channelArn := "arn:aws:cloudtrail:useast1:12345678910:eventdatastore/EXAMPLEf852-4e8f-8bd1-bcf6cEXAMPLE"
	err := ac.PutAuditLogs(&al, functionArn, channelArn)
	assert.Nil(t, err)

}

func TestPutAuditLogFail(t *testing.T) {

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", "https://cloudtrail-data.us-east-1.amazonaws.com/PutAuditEvents",
		httpmock.NewErrorResponder(errors.New("something borked")),
	)

	ac := AWSClient{
		Sess: session.Must(session.NewSession(aws.NewConfig().WithRegion("us-east-1").WithCredentials(credentials.AnonymousCredentials))),
	}

	httpmock.ActivateNonDefault(ac.Sess.Config.HTTPClient)

	al := model.AuditLogs{
		Logs: map[string]model.AuditRequest{
			"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp": {
				Client_ip:            "172.18.0.1",
				Request_id:           "0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp",
				Request_timestamp:    0,
				Ttl:                  2504744,
				Signature:            "",
				Rbac_user_id:         "",
				Workspace:            "0c21a8bb-0e63-4cf1-8b98-7038e1f25468",
				Payload:              "",
				Path:                 "/auth?session_logout=true",
				Method:               "OPTIONS",
				Status:               200,
				Removed_from_payload: "",
			},
		},
	}
	functionArn := "arn:aws:rds:us-east-2:123456789012:db:my-mysql-instance-1"
	channelArn := "arn:aws:cloudtrail:useast1:12345678910:eventdatastore/EXAMPLEf852-4e8f-8bd1-bcf6cEXAMPLE"
	err := ac.PutAuditLogs(&al, functionArn, channelArn)
	assert.NotNil(t, err)
}

func testPutAuditLogsFail200(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	fOutput := []*cloudtraildata.ResultErrorEntry{
		{
			ErrorCode:    aws.String("403"),
			ErrorMessage: aws.String("request id failure"),
		},
	}

	httpmock.RegisterResponder("POST", "https://cloudtrail-data.us-east-1.amazonaws.com/PutAuditEvents",
		func(req *http.Request) (*http.Response, error) {

			return httpmock.NewJsonResponse(200, map[string]interface{}{
				"Failed":     fOutput,
				"Successful": []*cloudtraildata.AuditEventResultEntry{},
			})
		},
	)

	ac := AWSClient{
		Sess: session.Must(session.NewSession(aws.NewConfig().WithRegion("us-east-1").WithCredentials(credentials.AnonymousCredentials))),
	}

	httpmock.ActivateNonDefault(ac.Sess.Config.HTTPClient)

	al := model.AuditLogs{
		Logs: map[string]model.AuditRequest{
			"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp": {
				Client_ip:            "172.18.0.1",
				Request_id:           "0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp",
				Request_timestamp:    0,
				Ttl:                  2504744,
				Signature:            "",
				Rbac_user_id:         "",
				Workspace:            "0c21a8bb-0e63-4cf1-8b98-7038e1f25468",
				Payload:              "",
				Path:                 "/auth?session_logout=true",
				Method:               "OPTIONS",
				Status:               200,
				Removed_from_payload: "",
			},
		},
	}
	functionArn := "arn:aws:rds:us-east-2:123456789012:db:my-mysql-instance-1"
	channelArn := "arn:aws:cloudtrail:useast1:12345678910:eventdatastore/EXAMPLEf852-4e8f-8bd1-bcf6cEXAMPLE"

	err := ac.PutAuditLogs(&al, functionArn, channelArn)
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "failed requests submitting to cloudtrails")
}
