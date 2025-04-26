package protocol

// https://identity.foundation/didcomm-messaging/spec/#trust-ping-protocol-20

import (
	"encoding/json"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"

	"github.com/google/uuid"
)

const PIURI_TRUST_PING = "https://didcomm.org/trust-ping/2.0/ping"
const PIURI_TRUST_PING_RESPONSE = "https://didcomm.org/trust-ping/2.0/ping-response"

type TrustPing struct {
	mediator *mediator.Mediator
}

func NewTrustPing(mediator *mediator.Mediator) *TrustPing {
	return &TrustPing{
		mediator: mediator,
	}
}

func (tp *TrustPing) Handle(message didcomm.Message) (response didcomm.Message, err error) {
	type body struct {
		ResponseRequested bool `json:"response_requested"`
	}
	var messageBody body

	err = json.Unmarshal([]byte(message.Body), &messageBody)
	if err != nil {
		return didcomm.Message{}, err
	}

	if messageBody.ResponseRequested {
		response = didcomm.Message{
			Id:   uuid.Must(uuid.NewRandom()).String(),
			Type: PIURI_TRUST_PING_RESPONSE,
			From: &tp.mediator.Did,
			Thid: &message.Id,
			Body: "{}",
		}

		return response, err
	} else {
		return didcomm.Message{}, intErr.ErrNoPingResponseRequested
	}
}
