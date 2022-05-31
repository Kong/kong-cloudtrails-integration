package model

type AuditLogs struct {
	Logs     map[string]AuditRequest
	Offset   string
	KongData KongData
}

type AuditRequests struct {
	Data   []AuditRequest
	Offset string `json:"offset"`
}
type AuditRequest struct {
	Client_ip            string `json:"client_ip"`
	Request_id           string `json:"request_id"`
	Request_timestamp    int    `json:"request_timestamp"`
	Ttl                  int    `json:"ttl"`
	Signature            string `json:"signature"`
	Rbac_user_id         string `json:"rbac_user_id"`
	Workspace            string `json:"workspace"`
	Payload              string `json:"payload"`
	Path                 string `json:"path"`
	Method               string `json:"method"`
	Status               int    `json:"status"`
	Removed_from_payload string `json:"removed_from_payload"`
}

type KongData struct {
	KongVersion  string `json:"version"`
	KongHostname string
}

type OpenAuditEvent struct {
	EventVersion        string                 `json:"eventVersion"`
	UserIdentity        UserIdentity           `json:"userIdentity"`
	EventTime           string                 `json:"eventTime"`
	EventSource         string                 `json:"eventSource"`
	EventName           string                 `json:"eventName"`
	RequestParameters   map[string]interface{} `json:"requestParameters,omitempty"`
	ResponseElements    string                 `json:"responseElements,omitempty"`
	SourceIPAddress     string                 `json:"sourceIPAddress,omitempty"`
	UserAgent           string                 `json:"userAgent,omitempty"`
	AdditionalEventData AdditionalEventData    `json:"additionalEventData,omitempty"`
	RecipientAccountId  string                 `json:"recipientAccountId"`
	UUID                string                 `json:"UUID"`
}

type UserIdentity struct {
	Type        string                 `json:"type"`
	PrincipalId string                 `json:"principalId"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

type AdditionalEventData struct {
	Method       string `json:"method"`
	Status       int    `json:"status"`
	Signature    string `json:"signature"`
	Ttl          int    `json:"ttl"`
	Workspace    string `json:"workspace"`
	KongHostname string `json:"konghostname"`
}
