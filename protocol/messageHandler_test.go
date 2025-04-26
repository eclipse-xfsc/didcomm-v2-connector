package protocol_test

import (
	"encoding/base64"
	"encoding/json"
	"github.com/google/uuid"
	"strconv"
	"testing"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/protocol"

	"github.com/stretchr/testify/assert"
)

func init() {
	config.CurrentConfiguration.Database.InMemory = true
	//med := mediator.NewMediator(slog.Default())
	//tp := protocol.NewTrustPing(med)
}

func TestHandleMessage(t *testing.T) {
	data := map[string]any{
		"id":   uuid.New().String(),
		"type": "https://didcomm.org/routing/2.0/forward",
		"body": map[string]string{
			"next": "did:example:1",
		},
		"from": "did:example:2",
		"to":   []string{"did:example:3"},
		"attachments": []map[string]any{
			map[string]any{
				"id": uuid.New().String(),
				"data": map[string]string{
					"base64": base64.StdEncoding.EncodeToString([]byte(`"credential_offer": "offer"`)),
				},
			},
		},
	}

	bodyString, _ := json.Marshal(data)

	packedMsg, err := protocol.HandleMessage(string(bodyString), med, "")
	assert.Nil(t, err)
	print(packedMsg)

}

func TestHandleMessage_MessageExpired(t *testing.T) {
	// Sample data
	now := uint64(time.Now().Unix())
	nwoStr := strconv.FormatUint(now, 10)
	bodyString := "" +
		"{" +
		"\"id\": \"123456789abcdefghi\"," +
		"\"type\": \"type\"," +
		"\"body\":\"{}\"," +
		"\"from\": \"did:from\"," +
		"\"to\": [" +
		"\"did:to\"" +
		"]," +
		"\"created_time\": " + nwoStr + "," +
		"\"expires_time\": " + nwoStr + "," +
		"\"attachments\": []" +
		"}"

	packedMsg, err := protocol.HandleMessage(bodyString, med, "")

	// Check if the result matches the expected outcome
	prType := "https://didcomm.org/report-problem/2.0/problem-report"
	prComment := "Message has expired"

	assert.Equal(t, nil, err)
	assert.Contains(t, packedMsg, prType, prComment)
}

func TestHandleMessage_MessageCreatedWrong(t *testing.T) {

	now := time.Now()
	future := uint64(now.Add(time.Hour * 10).Unix())
	futureStr := strconv.FormatUint(future, 10)
	// Sample data
	bodyString := "" +
		"{" +
		"\"id\": \"123456789abcdefghi\"," +
		"\"type\": \"type\"," +
		"\"body\":\"{}\"," +
		"\"from\": \"did:from\"," +
		"\"to\": [" +
		"\"did:to\"" +
		"]," +
		"\"created_time\": " + futureStr + "," +
		"\"expires_time\": " + futureStr + "," +
		"\"attachments\": []" +
		"}"

	packedMsg, err := protocol.HandleMessage(bodyString, med, "")

	// Check if the result matches the expected outcome
	prType := "https://didcomm.org/report-problem/2.0/problem-report"
	prComment := "Message creation time is in the future"

	assert.Equal(t, nil, err)
	assert.Contains(t, packedMsg, prType)
	assert.Contains(t, packedMsg, prComment)
}
