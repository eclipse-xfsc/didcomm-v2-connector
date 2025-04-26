package protocol_test

import (
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/protocol"

	"github.com/stretchr/testify/assert"
)

type body struct {
	ResponseRequested bool `json:"response_requested"`
}

var tp *protocol.TrustPing

func init() {
	config.CurrentConfiguration.Database.InMemory = true
	med = mediator.NewMediator(slog.Default())
	tp = protocol.NewTrustPing(med)
}

func TestHandle_JsonSyntaxError(t *testing.T) {

	// Sample data
	msg := didcomm.Message{}

	response, err := tp.Handle(msg)

	// Check if the result matches the expected outcome
	expectedResult := didcomm.Message{}
	var expectedErr *json.SyntaxError

	assert.Equal(t, expectedResult, response)
	assert.True(t, errors.As(err, &expectedErr))
}

func TestHandle_ResponseRequesteFalse(t *testing.T) {

	// Sample data
	messageBody := body{
		ResponseRequested: false,
	}

	jsonBody, _ := json.Marshal(messageBody)

	msg := didcomm.Message{
		Id:   "123",
		Type: protocol.PIURI_TRUST_PING,
		Body: string(jsonBody),
	}

	response, err := tp.Handle(msg)

	// Check if the result matches the expected outcome
	expectedResult := didcomm.Message{}
	expectedError := intErr.ErrNoPingResponseRequested

	assert.Equal(t, expectedResult, response)
	assert.Equal(t, expectedError, err)
}

func TestHandle_ResponseRequesteTrue(t *testing.T) {

	// Sample data
	messageBody := body{
		ResponseRequested: true,
	}

	jsonBody, _ := json.Marshal(messageBody)

	msg := didcomm.Message{
		Id:   "123",
		Type: protocol.PIURI_TRUST_PING,
		Body: string(jsonBody),
	}

	response, err := tp.Handle(msg)

	// Check if the result matches the expected outcome
	assert.Equal(t, protocol.PIURI_TRUST_PING_RESPONSE, response.Type)
	assert.Equal(t, msg.Id, *response.Thid)
	assert.Equal(t, nil, err)
}
