package protocol

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
)

func extractBody[T any](message didcomm.Message) (b T, err error) {
	r := strings.NewReader(message.Body)
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&b)
	if err != nil {
		fmt.Println("Failed to unmarshal message body JSON")
		return b, err
	}
	return b, nil
}
