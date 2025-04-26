package connectionmanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/database"
	"github.com/eclipse-xfsc/didcomm-v2-connector/pkg/constants"
)

type ConnectionManager struct {
	database database.Adapter
}

func NewConnectionManager(database database.Adapter) *ConnectionManager {
	return &ConnectionManager{
		database: database,
	}
}

var ERROR_PROTOCOL_NOT_SUPPORTED = errors.New("protocol not supported")
var ERROR_INTERNAL = errors.New("internal error")
var ERROR_CONNECTION_ALREADY_EXISTS = errors.New("connection already exists")

func (c *ConnectionManager) StoreConnection(protocol string, remoteDid string, topic string, properties map[string]string, eventType string, recipients []string, group string) (err error) {
	switch config.CurrentConfiguration.CloudForwarding.Protocol {
	case config.HTTP:
		if protocol != config.HTTP {
			return ERROR_PROTOCOL_NOT_SUPPORTED
		}
	case config.NATS:
		if protocol != config.NATS {
			return ERROR_PROTOCOL_NOT_SUPPORTED
		}
	case "hybrid":
		if protocol != config.HTTP && protocol != config.NATS {
			return ERROR_PROTOCOL_NOT_SUPPORTED
		}
	default:
		return ERROR_PROTOCOL_NOT_SUPPORTED
	}
	isMediated, err := c.database.IsMediated(remoteDid)
	if err != nil {
		return ERROR_INTERNAL
	}
	if isMediated {
		return ERROR_CONNECTION_ALREADY_EXISTS
	}
	err = c.database.AddMediatee(database.Mediatee{RemoteDid: remoteDid, Protocol: protocol, Topic: topic, EventType: eventType, Properties: properties, RecipientDids: recipients, Group: group})
	if err != nil {
		return ERROR_INTERNAL
	}
	return nil
}

func (c *ConnectionManager) Connect(host string, mediatorPeerDid string, peerdid string, bearer string) (string, error) {

	var mediatonRequest = make(map[string]interface{})
	mediatonRequest["id"] = uuid.NewString()
	mediatonRequest["type"] = "https://didcomm.org/coordinate-mediation/3.0/mediate-request"
	mediatonRequest["body"] = make(map[string]interface{})
	mediatonRequest["from"] = peerdid
	mediatonRequest["to"] = []string{mediatorPeerDid}
	mediatonRequest["created_time"] = time.Now().Unix()
	mediatonRequest["expired_time"] = time.Now().Add(time.Hour).Unix()
	mediatonRequest["attachments"] = []string{}

	res, err := GetHttpResult(mediatonRequest, host, bearer)

	if err != nil {
		return "", err
	}
	res_type, ok := res["type"]
	if !ok {
		return "", errors.New("expected `type` field to be present in response")
	}
	body, ok := res["body"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("expected `body` field to be present in response of type %s", res_type)
	}

	if res_type == constants.PIURI_COORDINATE_MEDIATION_RESPOSE_GRANT {
		routing_did, ok := body["routing_did"].([]interface{})

		if ok {
			if len(routing_did) > 0 {
				return routing_did[0].(string), nil
			}
		}
		return "", errors.New("expected `routing_did` field to be present in response and be not empty")

	} else if res_type == constants.PIURI_COORDINATE_MEDIATION_RESPOSE_DENY {
		return "", errors.New("mediation request was denied")
	} else {
		return "", fmt.Errorf("unexpected response type %s", res_type)
	}
}

func GetHttpResult(input interface{}, url string, bearer string) (map[string]interface{}, error) {

	body, err := json.Marshal(input)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to marshal input"))
	}

	r, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	r.Header.Add("Content-Type", "application/json")

	if bearer != "" {
		r.Header.Add("Authorization", "Bearer "+bearer)
	}

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(r)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var post map[string]interface{}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed to read response (status: %s) body", res.Status))
	}
	if res.StatusCode >= http.StatusMultipleChoices {
		return nil, errors.New(fmt.Sprintf("error response (status: %s) body %s", res.Status, string(data)))
	}
	derr := json.Unmarshal(data, &post)
	if derr != nil {
		return nil, errors.Join(derr, fmt.Errorf("failed to unmarshal response (status: %s) body %s", res.Status, string(data)))
	}

	return post, nil
}
