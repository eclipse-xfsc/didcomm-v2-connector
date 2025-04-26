package messaging

import "encoding/json"

type ConnectorMessage struct {
	Did     string          `json:"did"`
	Payload json.RawMessage `json:"payload"`
}

type InvitationNotify struct {
	InvitationId string `json:"invitationId"`
	Did          string `json:"did"`
}
