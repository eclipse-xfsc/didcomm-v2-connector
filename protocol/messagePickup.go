package protocol

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
)

type MessagePickup struct {
	mediator *mediator.Mediator
}

type statusBody struct {
	RecipientDid  string `json:"recipient_did"`
	MesssageCount int    `json:"message_count"`
}

func NewMessagePickup(mediator *mediator.Mediator) *MessagePickup {
	return &MessagePickup{
		mediator: mediator,
	}
}

// https://didcomm.org/messagepickup/3.0/

const PIURI_MESSAGEPICKUP = "https://didcomm.org/messagepickup/3.0/"
const PIURI_MESSAGEPICKUP_STATUS_REQUEST = "https://didcomm.org/messagepickup/3.0/status-request"
const PIURI_MESSAGEPICKUP_DELIVERY_REQUEST = "https://didcomm.org/messagepickup/3.0/delivery-request"
const PIURI_MESSAGEPICKUP_MESSAGES_RECEIVED = "https://didcomm.org/messagepickup/3.0/messages-received"
const PIURI_MESSAGEPICKUP_LIVE_DELIVERY_CHANGE = "https://didcomm.org/messagepickup/3.0/live-delivery-change"

func (mp *MessagePickup) Handle(message didcomm.Message) (response didcomm.Message, err error) {

	// Extra headers return "string" with quotation marks if it is a string
	// and 99 if it is an int without quotation marks
	return_route := message.ExtraHeaders["return_route"]
	_, err = strconv.Atoi(return_route)
	if err != nil {
		return_route = strings.Trim(return_route, "\"")
	} else {
		return PR_RETURN_ROUTE_ALL_MISSING, errors.New("return_route must be all")
	}
	if return_route != "all" {
		return PR_RETURN_ROUTE_ALL_MISSING, errors.New("return_route must be all")
	}
	switch message.Type {
	case PIURI_MESSAGEPICKUP_STATUS_REQUEST:
		response, err = mp.handleStatusRequest(message)
	case PIURI_MESSAGEPICKUP_DELIVERY_REQUEST:
		response, err = mp.handleDeliveryRequest(message)
	case PIURI_MESSAGEPICKUP_MESSAGES_RECEIVED:
		response, err = mp.handleMessagesReceived(message)
	case PIURI_MESSAGEPICKUP_LIVE_DELIVERY_CHANGE:
		response, err = mp.handleLiveDeliveryChange(message)
	default:
		err = intErr.ErrUnknownMessageType
	}

	return response, err
}

func (mp *MessagePickup) handleStatusRequest(message didcomm.Message) (response didcomm.Message, err error) {
	type requestBody struct {
		RecipientDid string `json:"recipient_did"`
	}

	var body requestBody
	body, err = extractBody[requestBody](message)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	recipientDid := body.RecipientDid
	remoteDid := *message.From

	match, err := mp.mediator.Database.RecipientAndRemoteDidBelongTogether(recipientDid, remoteDid)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}
	if !match {
		return PR_RECIPIENT_REMOTE_DID_MISMATCH, errors.New("recipient did and remote did do not belong together")
	}

	count, err := mp.mediator.Database.GetMessagesCountForRecipient(recipientDid)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	rb := statusBody{
		RecipientDid:  recipientDid,
		MesssageCount: count,
	}

	responseBodyJson, err := json.Marshal(rb)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   message.Id,
		Type: "https://didcomm.org/messagepickup/3.0/status",
		Body: string(responseBodyJson),
	}

	return response, nil
}

func (mp *MessagePickup) handleDeliveryRequest(message didcomm.Message) (response didcomm.Message, err error) {
	type requestBody struct {
		RecipientDid string `json:"recipient_did"`
		Limit        int    `json:"limit"`
	}

	type responseBody struct {
		RecipientDid string `json:"recipient_did"`
	}

	var body requestBody
	body, err = extractBody[requestBody](message)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	recipientDid := body.RecipientDid
	remoteDid := *message.From
	match, err := mp.mediator.Database.RecipientAndRemoteDidBelongTogether(recipientDid, remoteDid)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}
	if !match {
		return PR_RECIPIENT_REMOTE_DID_MISMATCH, errors.New("recipient did and remote did do not belong together")
	}

	attachments, err := mp.mediator.Database.GetMessagesForRecipient(recipientDid, body.Limit)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	rb := responseBody{
		RecipientDid: recipientDid,
	}

	responseBodyJson, err := json.Marshal(rb)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:          message.Id,
		Type:        "https://didcomm.org/messagepickup/3.0/delivery",
		Body:        string(responseBodyJson),
		Attachments: &attachments,
	}

	return response, nil
}

func (mp *MessagePickup) handleMessagesReceived(message didcomm.Message) (response didcomm.Message, err error) {
	type requestBody struct {
		MessageIdList []string `json:"message_id_list"`
	}

	type deletedBody struct {
		DeleteCount int `json:"delete_count"`
	}

	var body requestBody
	body, err = extractBody[requestBody](message)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	for _, id := range body.MessageIdList {
		match, err := mp.mediator.Database.RemoteDidBelongsToMessage(*message.From, id)
		if err != nil {
			return PR_INTERNAL_SERVER_ERROR, err
		}

		if !match {
			return PR_REMOTE_DID_MESSAGE_MISMATCH, errors.New("remote did does not belong to message")
		}
	}
	count, err := mp.mediator.Database.DeleteMessagesByIds(body.MessageIdList)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	rb := deletedBody{
		DeleteCount: count,
	}

	responseBodyJson, err := json.Marshal(rb)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   message.Id,
		Type: "https://didcomm.org/messagepickup/3.0/status",
		Body: string(responseBodyJson),
	}

	return response, nil
}

func (mp *MessagePickup) handleLiveDeliveryChange(message didcomm.Message) (response didcomm.Message, err error) {
	type responseBody struct {
		Code    string `json:"code"`
		Comment string `json:"comment"`
	}

	rb := responseBody{
		Code:    "e.m.live-mode-not-supported",
		Comment: "Live mode is not supported",
	}

	responseBodyJson, err := json.Marshal(rb)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   message.Id,
		Type: "https://didcomm.org/report-problem/2.0/problem-report",
		Body: string(responseBodyJson),
	}

	return response, nil
}
