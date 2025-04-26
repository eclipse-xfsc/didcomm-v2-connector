package protocol

import (
	"encoding/json"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"

	"github.com/google/uuid"
)

// https://identity.foundation/didcomm-messaging/spec/#invitation
type OutOfBand struct {
	mediator *mediator.Mediator
}

func NewOutOfBand(mediator *mediator.Mediator) *OutOfBand {
	return &OutOfBand{
		mediator: mediator,
	}
}

func (o *OutOfBand) Handle(label string, bearer string) (response string, err error) {
	type body struct {
		GoalCode string   `json:"goal_code"`
		Goal     string   `json:"goal"`
		Label    string   `json:"label"`
		Accept   []string `json:"accept"`
		Bearer   string   `json:"auth"`
	}

	b := body{
		GoalCode: "request-mediate",
		Goal:     "RequestMediate",
		Label:    label,
		Accept:   []string{"didcomm/v2"},
		Bearer:   bearer,
	}

	bodyJson, err := json.Marshal(b)
	if err != nil {
		panic(err)
	}

	message := didcomm.Message{
		Id:   uuid.Must(uuid.NewRandom()).String(),
		Type: "https://didcomm.org/out-of-band/2.0/invitation",
		Body: string(bodyJson),
		From: &o.mediator.Did,
	}
	packMsg, err := o.mediator.PackPlainMessage(message)
	if err != nil {
		return "", err
	}
	// packMsg64 := b64.StdEncoding.EncodeToString([]byte(packMsg))
	// return packMsg64, nil
	return packMsg, nil
}
