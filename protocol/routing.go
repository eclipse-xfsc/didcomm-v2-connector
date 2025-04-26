package protocol

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
)

// https://identity.foundation/didcomm-messaging/spec/#routing-protocol-20

const PIURI_ROUTING = "https://didcomm.org/routing/2.0/"
const PIURI_ROUTING_FORWARD = "https://didcomm.org/routing/2.0/forward"

type Routing struct {
	mediator *mediator.Mediator
}

func NewRouting(mediator *mediator.Mediator) *Routing {
	return &Routing{
		mediator: mediator,
	}
}

func (rt *Routing) Handle(message didcomm.Message) (response didcomm.Message, err error) {

	switch message.Type {
	case PIURI_ROUTING_FORWARD:
		response, err = rt.handleForward(message, true)
	default:
		err = intErr.ErrUnknownMessageType
		response = PR_UNKNOWN_MESSAGE_TYPE
	}
	return
}

func (rt *Routing) handleForward(message didcomm.Message, inbound bool) (pr ProblemReport, err error) {
	t := uint64(time.Now().UTC().Unix())
	if message.ExpiresTime != nil && *message.ExpiresTime < t {
		return PR_EXPIRED_MESSAGE, errors.New("message has expired")
	}
	type requestBody struct {
		Next string `json:"next"`
	}
	var body requestBody
	body, err = extractBody[requestBody](message)

	if err != nil {
		return PR_COULD_NOT_FORWARD_MESSAGE, err
	}

	if len(*message.Attachments) != 1 {
		return PR_COULD_NOT_FORWARD_MESSAGE, errors.New("message must have exactly one attachment")
	}

	attachment := (*message.Attachments)[0]

	isMediated, err := rt.mediator.Database.IsMediated(body.Next)

	if err != nil {
		config.Logger.Error("error checking mediatee", "err", err)
		return PR_COULD_NOT_FORWARD_MESSAGE, err
	}

	if isMediated {
		config.Logger.Debug("Next is registered as mediator, park message in outbox.")
		err = rt.mediator.Database.AddMessage(body.Next, attachment)
		if err != nil {
			config.Logger.Error("could not add message to inbox", "err", err)
			return PR_COULD_NOT_FORWARD_MESSAGE, err
		}
		return PR_COULD_NOT_FORWARD_MESSAGE, err

	} else {

		isRegistered, err := rt.mediator.Database.IsRecipientDidRegistered(body.Next)
		if err != nil {
			return PR_COULD_NOT_FORWARD_MESSAGE, err
		}

		if isRegistered {
			/*here it should be in future a iteration over all recipients, together with a format selection. The forward is not precisly specified in the moment, and should
			    be discussed deeper. E.g. Recipient forward types, recipient forward version etc. pp For now here is just an simple forward to nats, potentially this could be enhanced by web
				socket etc.
			*/
			config.Logger.Info("Recipient is registered")
			if !inbound {
				//In the case of a did here must be later on a decision if a service endpoint should be used or
				//to use the message box. for now is it the messagebox

				mediatee, err := rt.mediator.Database.GetMediateeByRecipientDid(body.Next)
				if err != nil {
					config.Logger.Error("error getting mediatee", "err", err)
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}

				config.Logger.Debug("Message is outgoing, resolve remote did")
				didDoc, err := rt.mediator.DidResolver.ResolveDid(mediatee.RemoteDid)

				if err != nil {
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}

				/*
					The peer did for a device has no endpoint where you try to send to
				*/
				if len(didDoc.Service) == 0 {
					config.Logger.Debug("No direct forwarding possible (no service found in remote did), park message in outbox")
					err = rt.mediator.Database.AddMessage(body.Next, attachment)
					if err != nil {
						config.Logger.Error("could not add message to inbox", "err", err)
						return PR_COULD_NOT_FORWARD_MESSAGE, err
					}
				} else {
					/*
						if service endpoint exists which is didcomm compatible, forward message as it is.
					*/
					config.Logger.Debug("Endpoint found within did, post it to there directly")
					for _, x := range didDoc.Service {
						val, ok := x.ServiceEndpoint.(didcomm.ServiceKindDidCommMessaging)

						if ok {

							message.From = &rt.mediator.Did
							message.To = &[]string{mediatee.RemoteDid}

							return rt.ForwardMessage(message, val.Value.Uri)
						}
					}

					return PR_COULD_NOT_FORWARD_MESSAGE, errors.New("No compatible endpoint found in the list")
				}

			} else {

				/*
					Here should be later an seperation between cloud and attachment forward. E.g. by using different service endpoints within the routing did

					if ... {

						ForwardMessageAttachment...
					} else

				*/
				config.Logger.Debug("Incoming message, handle it as normal mediation and forward b64 attachment")
				messageB64 := attachment.Data.(didcomm.AttachmentDataBase64).Value.Base64
				messageDecoded, err := base64.StdEncoding.DecodeString(messageB64)
				if err != nil {
					config.Logger.Error("Error decoding message", "err", err)
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}

				mediatee, err := rt.mediator.Database.GetMediateeByRecipientDid(body.Next)
				if err != nil {
					config.Logger.Error("error getting mediatee", "err", err)
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}

				var content map[string]interface{}

				err = json.Unmarshal(messageDecoded, &content)

				if err != nil {
					config.Logger.Error("marshalling body failed", "err", err)
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}

				err = SendMessage(content, mediatee)
				if err != nil {
					config.Logger.Error("unable to send message to cloud", "err", err)
					return PR_COULD_NOT_FORWARD_MESSAGE, err
				}
			}
		}
	}
	return didcomm.Message{}, err
}

func (rt *Routing) ForwardMessage(message didcomm.Message, endpoint string) (pr ProblemReport, err error) {
	to := *message.To
	packMsg, err := packMessage(*message.From, to[0], message, rt.mediator)
	if err != nil {
		config.Logger.Error("Problem Report", "err", err)
		return didcomm.Message{}, err
	}
	r := strings.NewReader(packMsg)

	resp, err := http.Post(endpoint, "application/didcomm-plain+json", r)
	if err != nil {
		return PR_COULD_NOT_FORWARD_MESSAGE, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return PR_NEXT_DENIED_MESSAGE, errors.New("forwarded message was not accepted by the next recipient")
	}
	return didcomm.Message{}, nil
}

func (rt *Routing) ForwardAttachmentMessage(message didcomm.Attachment, recipientDid string, endpoint string) (pr ProblemReport, err error) {
	messageDecoded := message.Data.(didcomm.AttachmentDataBase64).Value.Base64
	body, err := base64.StdEncoding.DecodeString(messageDecoded)
	if err != nil {
		return PR_COULD_NOT_FORWARD_MESSAGE, err
	}
	r := bytes.NewReader(body)
	resp, err := http.Post(endpoint, "application/didcomm-plain+json", r)
	if err != nil {
		return PR_COULD_NOT_FORWARD_MESSAGE, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return PR_NEXT_DENIED_MESSAGE, errors.New("forwarded message was not accepted by the next recipient")
	}
	return didcomm.Message{}, nil
}
