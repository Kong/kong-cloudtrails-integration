package kongClient

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/Kong/kong-cloudtrails-integration/model"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestCallAuditLogSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://gateway:8001/audit/requests",
		httpmock.NewStringResponder(200, `{"Data":[{"client_ip":"172.18.0.1","request_id":"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp","request_timestamp":"","ttl":2504744,"signature":"","rbac_user_id":"","workspace":"0c21a8bb-0e63-4cf1-8b98-7038e1f25468","payload":"","path":"/auth?session_logout=true","method":"OPTIONS","status":200,"removed_from_payload":""}]}`))

	kc := NewRestClient("http://gateway:8001", "true", "test")
	httpmock.ActivateNonDefault(kc.client.GetClient())

	expected := model.AuditLogs{
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

	auditLogs, ids, err := kc.CallAuditLog("")

	assert.Nil(t, err)
	assert.Equal(t, 1, httpmock.GetTotalCallCount())
	assert.Equal(t, &expected, auditLogs)
	assert.Equal(t, ids, []string{"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp"})

}

func TestCallAuditLogFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://gateway:8001/audit/requests",
		func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("internal server error")
		})

	kc := NewRestClient("http://gateway:8001", "true", "test")
	httpmock.ActivateNonDefault(kc.client.GetClient())

	auditLogs, requestIds, err := kc.CallAuditLog("")

	assert.Equal(t, 1, httpmock.GetTotalCallCount())
	assert.NotNil(t, err)
	assert.Nil(t, auditLogs)
	assert.Nil(t, requestIds)
}

func TestSuperAdminHeader(t *testing.T) {

	kc := NewRestClient("http://gateway:8001", "true", "test")
	expected := "test"
	actual := kc.client.Header.Get("Kong-Admin-Token")
	assert.Equal(t, actual, expected)
}

func TestNoAdminHeader(t *testing.T) {
	kc := NewRestClient("http://gateway:8001", "false", "")
	expected := ""
	actual := kc.client.Header.Get("Kong-Admin-Token")
	assert.Equal(t, actual, expected)
}

func TestOffsetQueryParam(t *testing.T) {
	offsets := []string{"", "naldknasdflknasdf"}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "http://gateway:8001/audit/requests",
		httpmock.NewStringResponder(200, `{"Data":[{"client_ip":"172.18.0.1","request_id":"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp","request_timestamp":"","ttl":2504744,"signature":"","rbac_user_id":"","workspace":"0c21a8bb-0e63-4cf1-8b98-7038e1f25468","payload":"","path":"/auth?session_logout=true","method":"OPTIONS","status":200,"removed_from_payload":""}]}`))

	httpmock.RegisterResponder("GET", `=~^http://gateway:8001/audit/requests(\?offset=\w+)\z`,
		httpmock.NewStringResponder(200, `{"Data":[{"client_ip":"172.18.0.1","request_id":"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp","request_timestamp":"","ttl":2504744,"signature":"","rbac_user_id":"","workspace":"0c21a8bb-0e63-4cf1-8b98-7038e1f25468","payload":"","path":"/auth?session_logout=true","method":"OPTIONS","status":200,"removed_from_payload":""}]}`))

	kc := NewRestClient("http://gateway:8001", "true", "test")
	httpmock.ActivateNonDefault(kc.client.GetClient())

	for _, v := range offsets {
		kc.CallAuditLog(v)
	}
	info := httpmock.GetCallCountInfo()
	assert.Equal(t, 2, httpmock.GetTotalCallCount())
	fmt.Println(info)
}
