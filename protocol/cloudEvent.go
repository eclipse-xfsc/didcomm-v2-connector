package protocol

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"text/template"

	"github.com/cloudevents/sdk-go/v2/event"
	cloudeventprovider "github.com/eclipse-xfsc/cloud-event-provider"
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/database"
	"github.com/eclipse-xfsc/didcomm-v2-connector/pkg/messaging"
	"github.com/google/uuid"
)

func SendMessage(message map[string]interface{}, mediatee *database.Mediatee) error {

	switch config.CurrentConfiguration.CloudForwarding.Protocol {
	case config.HTTP:
		return sendCloudEvent(message, mediatee, mediatee.Topic)
	case config.NATS:
		return sendCloudEvent(message, mediatee, mediatee.Topic)
	case config.HYBRID:
		// implement hybrid mode if cloud event provider supports it
		return errors.New("hybrid mode not supported: message will not be sent")
	default:
		return errors.New("unknown cloud forwarding mode")
	}
}

func ReceiveMessage(mediator *mediator.Mediator) {

	config.Logger.Info("Start messaging", "context")

	topic := config.CurrentConfiguration.CloudForwarding.Nats.Topic

	client, err := cloudeventprovider.New(cloudeventprovider.Config{
		Protocol: cloudeventprovider.ProtocolTypeNats,
		Settings: cloudeventprovider.NatsConfig{
			Url:        config.CurrentConfiguration.CloudForwarding.Nats.Url,
			QueueGroup: config.CurrentConfiguration.CloudForwarding.Nats.QueueGroup,
		},
	}, cloudeventprovider.Sub, topic)
	if err != nil {
		config.Logger.Error("unable to connect to cloud event provider", "msg", err)
	}

	defer client.Close()
	routing := NewRouting(mediator)
	// Use a WaitGroup to wait for a message to arrive
	wg := sync.WaitGroup{}
	wg.Add(1)

	config.Logger.Info("Receiving cloud events", "topic", topic)

	err = client.Sub(func(event event.Event) {

		config.Logger.Info("Received cloud event", "context", event.Context)
		config.Logger.Info("Data", "context", string(event.DataEncoded))

		var incomingMessage json.RawMessage
		err := json.Unmarshal(event.DataEncoded, &incomingMessage)
		if err != nil {
			config.Logger.Error("error while unmarshalling received nats message")
			return
		}

		var content messaging.ConnectorMessage

		err = json.Unmarshal(incomingMessage, &content)

		if err != nil {
			config.Logger.Error("error while unmarshalling received nats message")
			return
		}

		attachment := didcomm.Attachment{
			Data: didcomm.AttachmentDataBase64{
				Value: didcomm.Base64AttachmentData{
					Base64: base64.StdEncoding.EncodeToString(incomingMessage),
				},
			},
		}

		var body = make(map[string]interface{})

		body["next"] = content.Did

		bodyJson, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}

		message := didcomm.Message{
			Id:          uuid.NewString(),
			Type:        PIURI_ROUTING_FORWARD,
			To:          &[]string{mediator.Did},
			Attachments: &[]didcomm.Attachment{attachment},
			Body:        string(bodyJson),
		}

		routing.handleForward(message, false)
	})
	if err != nil {
		config.Logger.Error("Error in subscription of cloud event", "msg", err)
	}

	// Wait for a message to come in
	wg.Wait()
}

func sendCloudEvent(message any, mediatee *database.Mediatee, topic string) (err error) {
	if topic == "" {
		topic = "default-http"
	}
	client, err := cloudeventprovider.New(cloudeventprovider.Config{
		Protocol: cloudeventprovider.ProtocolTypeNats,
		Settings: cloudeventprovider.NatsConfig{
			Url:        config.CurrentConfiguration.CloudForwarding.Nats.Url,
			QueueGroup: config.CurrentConfiguration.CloudForwarding.Nats.QueueGroup,
		},
	}, cloudeventprovider.Pub, topic)
	if err != nil {
		config.Logger.Error("Can not create cloudevent client", "msg", err)
		return
	}
	defer client.Close()

	sourceUrl, err := url.JoinPath(config.CurrentConfiguration.CloudForwarding.Nats.Url)

	config.Logger.Info(fmt.Sprintf("message to send as cloud event: %s", message))

	mediatee.Properties["routingKey"] = mediatee.RoutingKey
	mediatee.Properties["remoteDid"] = mediatee.RemoteDid

	jsonData, err := json.Marshal(mediatee.Properties)

	if err != nil {
		config.Logger.Error("cant marshal properties", "msg", err)
		return
	}

	tmpl, err := template.New("template").Parse(string(jsonData))
	var result bytes.Buffer
	err = tmpl.Execute(&result, message)
	if err != nil {
		panic(err)
	}

	event, err := cloudeventprovider.NewEvent(sourceUrl, mediatee.EventType, result.Bytes())
	if err != nil {
		config.Logger.Error("failed to create cloud event", "msg", err)
		return
	}

	if err = client.Pub(event); err != nil {
		config.Logger.Error("failed to send cloud event", "msg", err)
		return
	}

	config.Logger.Info("published cloud event", "topic", topic)

	return
}
