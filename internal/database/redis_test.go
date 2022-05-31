package database

import (
	"context"
	"testing"
	"time"

	"github.com/Kong/kong-cloudtrails-integration/model"
	"github.com/go-redis/redismock/v8"
	"github.com/stretchr/testify/assert"
)

func TestRequestIdsExist(t *testing.T) {
	redis, mock := redismock.NewClientMock()

	db := Database{
		client: redis,
		ctx:    context.TODO(),
	}

	keys := []string{"abcdf", "abcd235", "abscknod"}
	expected := []int{1, 0, 1}

	mock.ExpectExists("abcdf").SetVal(1)
	mock.ExpectExists("abcd235").SetVal(0)
	mock.ExpectExists("abscknod").SetVal(1)

	results := db.RequestIdsExist(keys...)

	assert.Equal(t, results, expected)

	err := mock.ExpectationsWereMet()
	if err != nil {
		assert.Fail(t, err.Error())
	}
}

func TestSetRequestIds(t *testing.T) {
	redis, mock := redismock.NewClientMock()

	db := Database{
		client: redis,
		ctx:    context.TODO(),
	}

	al := &model.AuditLogs{
		Logs: map[string]model.AuditRequest{
			"0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp": model.AuditRequest{
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

	mock.ExpectTxPipeline()
	ttl := time.Duration(2504744) * time.Second
	mock.ExpectSet("0bAf1b0nfOZsMSEByhhgn8FHdra7DjRp", true, ttl).SetVal("DONE")
	mock.ExpectTxPipelineExec()

	db.SetRequestIds(al)
	err := mock.ExpectationsWereMet()
	if err != nil {
		assert.Fail(t, err.Error())
	}
}
