package protocol

import (
	"encoding/json"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

	"github.com/google/uuid"
)

// https://identity.foundation/didcomm-messaging/spec/#problem-codes
type problemReportCode struct {
	sorter     string
	scope      string
	descriptor []string
}

func (p *problemReportCode) generate() string {
	var descriptor string
	for _, d := range p.descriptor {
		descriptor += "." + d
	}
	return p.sorter + "." + p.scope + descriptor
}

type ProblemReport = didcomm.Message

func NewProblemReport(sorter string, scope string, descriptor []string, comment string) ProblemReport {

	type responseBody struct {
		Code    string `json:"code"`
		Comment string `json:"comment"`
	}

	p := &problemReportCode{
		sorter:     sorter,
		scope:      scope,
		descriptor: descriptor,
	}
	problemReportCode := p.generate()
	rb := responseBody{
		Code:    problemReportCode,
		Comment: comment,
	}

	responseBodyJson, err := json.Marshal(rb)
	if err != nil {
		return didcomm.Message{}
	}

	return didcomm.Message{
		Id:   uuid.New().String(),
		Type: "https://didcomm.org/report-problem/2.0/problem-report",
		Body: string(responseBodyJson),
	}
}

const (
	PR_SORTER_ERROR   = "e"
	PR_SORTER_WARNING = "w"
)

const (
	PR_SCOPE_PROTOCOL = "p"
	PR_SCOPE_MESSAGE  = "m"
)

const (
	PR_DESCRIPTOR_TRUST            = "trust"
	PR_DESCRIPTOR_TRUST_CRYPOT     = "trust.crypot"
	PR_DESCRIPTOR_XFER             = "xfer"
	PR_DESCRIPTOR_DID              = "did"
	PR_DESCRIPTOR_MESSAGE          = "msg"
	PR_DESCRIPTOR_INTERNAL_ERROR   = "me"
	PR_DESCRIPTOR_RESOURCE         = "me.res"
	PR_DESCRIPTOR_REQUIREMENT      = "req"
	PR_DESCRIPTOR_REQUIREMENT_TIME = "req.time"
	PR_DESCRIPTOR_LEGAL            = "legal"
)
