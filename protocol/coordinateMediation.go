package protocol

import (
	"encoding/json"
	"slices"

	"github.com/google/uuid"
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/pkg/constants"
	"github.com/eclipse-xfsc/didcomm-v2-connector/pkg/messaging"
)

// update model
type Update struct {
	RecipientDid string `json:"recipient_did"`
	Action       string `json:"action"`
	Result       string `json:"result"`
}

// https://didcomm.org/coordinate-mediation/3.0/
type CoordinateMediation struct {
	mediator *mediator.Mediator
}

func NewCoordinateMediation(mediator *mediator.Mediator) *CoordinateMediation {
	return &CoordinateMediation{
		mediator: mediator,
	}
}

func (h *CoordinateMediation) Handle(message didcomm.Message, bearer string) (response didcomm.Message, err error) {

	switch message.Type {
	case constants.PIURI_COORDINATE_MEDIATION_REQUEST:
		response, err = h.handleMediationRequest(message, bearer)
	case constants.PIURI_COORDINATE_MEDIATION_UPDATE:
		response, err = h.handleRecipientUpdate(message)
	case constants.PIURI_COORDINATE_MEDIATION_QUERY:
		response, err = h.handleRecipientQuery(message)
	}

	return
}

func (h *CoordinateMediation) handleRecipientQuery(message didcomm.Message) (response didcomm.Message, err error) {

	// recipient query models

	type OutgoingPagination struct {
		Count     int `json:"count"`
		Offset    int `json:"offset"`
		Remaining int `json:"remaining"`
	}

	type IncomingPagination struct {
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	}

	type Query struct {
		Pagination IncomingPagination `json:"paginate"`
	}

	type Did struct {
		RecipientDid string `json:"recipient_did"`
	}

	type DidList struct {
		Dids       []Did              `json:"dids"`
		Pagination OutgoingPagination `json:"pagination"`
	}

	var paginate Query
	incomingDid := *message.From

	errJSON := json.Unmarshal([]byte(message.Body), &paginate)
	if errJSON != nil {
		return PR_INTERNAL_SERVER_ERROR, errJSON
	}

	db := h.mediator.Database
	currentDids, err := db.GetRecipientDids(incomingDid)
	if err != nil {
		config.Logger.Error("Can not get recipient dids from db", "msg", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}
	currentDidsSize := len(currentDids)

	offset, count, remaining := calculatePagination(paginate.Pagination.Limit, paginate.Pagination.Offset, currentDidsSize)

	var didlist = DidList{
		Dids: make([]Did, 0),
		Pagination: OutgoingPagination{
			Count:     count,
			Offset:    offset,
			Remaining: remaining,
		},
	}

	for index := 0 + offset; index < count+offset; index++ {
		didlist.Dids = append(didlist.Dids, Did{
			RecipientDid: currentDids[index],
		})
	}

	bodyJson, err := json.Marshal(didlist)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   uuid.Must(uuid.NewRandom()).String(),
		Type: "https://didcomm.org/coordinate-mediation/3.0/recipient",
		Body: string(bodyJson),
		To:   &[]string{*message.From},
		From: &h.mediator.Did,
	}

	return response, nil
}

func calculatePagination(limit int, incomingOffset int, size int) (offset int, count int, remaining int) {

	if size-incomingOffset-limit < 0 {
		offset = 0
	} else {
		offset = incomingOffset
	}

	if limit >= size {
		count = size
	} else {
		count = limit
	}

	if limit >= size {
		remaining = 0
	} else {
		remaining = size - offset - limit
	}

	return
}

func (h *CoordinateMediation) handleRecipientUpdate(message didcomm.Message) (response didcomm.Message, err error) {
	type IncomingUpdates struct {
		Updates []Update `json:"updates"`
	}

	var incomingUpdates IncomingUpdates
	remoteDid := *message.From

	errJSON := json.Unmarshal([]byte(message.Body), &incomingUpdates)
	if errJSON != nil {
		return PR_INTERNAL_SERVER_ERROR, errJSON
	}

	// get current list of DID from DB
	db := h.mediator.Database
	mediatee, err := db.GetMediatee(remoteDid)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	// update keys
	updatedRecipientDids, recipientDidsToAdd, recipientDidsToDelete, err := update(remoteDid, mediatee.RecipientDids, incomingUpdates.Updates)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	// update keys in db for connection
	for _, key := range recipientDidsToAdd {
		err = db.AddRecipientDid(remoteDid, key)
		if err != nil {
			config.Logger.Error("Unable to add recipient DID", "msg", err)
			// remove not added recipient DID from updatedRecipientDids
			removeUpdateByRecipientDid(&updatedRecipientDids, key)
		}
	}
	// delete keys in db for connection
	for _, key := range recipientDidsToDelete {
		err = db.DeleteRecipientDid(remoteDid, key)
		if err != nil {
			config.Logger.Error("Unable to add recipient DID", "msg", err)
			// remove not deleted recipient DID from updatedRecipientDids
			removeUpdateByRecipientDid(&updatedRecipientDids, key)
		}
	}

	// needed for different json name
	type OutgoingUpdates struct {
		Updates []Update `json:"updated"`
	}

	outgoingUpdates := OutgoingUpdates{
		Updates: updatedRecipientDids,
	}

	bodyJson, err := json.Marshal(outgoingUpdates)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   uuid.Must(uuid.NewRandom()).String(),
		Type: "https://didcomm.org/coordinate-mediation/3.0/recipient-update-response",
		Body: string(bodyJson),
		To:   &[]string{*message.From},
		From: &h.mediator.Did,
	}

	return response, nil
}

func update(did string, dbRecipientDids []string, RecipientDidUpdates []Update) (updatedRecipientDids []Update, recipientDidsToAdd []string, recipientDidsToDelete []string, err error) {

	updatedRecipientDids = []Update{}

	ADD := "add"
	REMOVE := "remove"
	NO_CHANGES := "no_changes"
	SUCCESS := "success"
	CLIENT_ERROR := "client_error"

	recipientDidsToAdd = make([]string, 0)
	recipientDidsToDelete = make([]string, 0)

	for _, update := range RecipientDidUpdates {
		switch update.Action {
		case ADD:
			if slices.Contains(dbRecipientDids, update.RecipientDid) {
				updatedRecipientDids = append(updatedRecipientDids, Update{
					RecipientDid: update.RecipientDid,
					Action:       ADD,
					Result:       NO_CHANGES,
				})
			} else {
				updatedRecipientDids = append(updatedRecipientDids, Update{
					RecipientDid: update.RecipientDid,
					Action:       ADD,
					Result:       SUCCESS,
				})
				// add key to list
				recipientDidsToAdd = append(recipientDidsToAdd, update.RecipientDid)
			}
		case REMOVE:
			if slices.Contains(dbRecipientDids, update.RecipientDid) {
				updatedRecipientDids = append(updatedRecipientDids, Update{
					RecipientDid: update.RecipientDid,
					Action:       REMOVE,
					Result:       SUCCESS,
				})
				// remove key from db list
				recipientDidsToDelete = append(recipientDidsToDelete, update.RecipientDid)
			} else {
				updatedRecipientDids = append(updatedRecipientDids, Update{
					RecipientDid: update.RecipientDid,
					Action:       REMOVE,
					Result:       CLIENT_ERROR,
				})
			}
		default:
			config.Logger.Warn("RecipientUpdate: Unkown update action!")
		}
	}

	return updatedRecipientDids, recipientDidsToAdd, recipientDidsToDelete, nil
}

func (h *CoordinateMediation) handleMediationRequest(message didcomm.Message, bearer string) (response didcomm.Message, err error) {

	id, err := mediator.VerifySignedToken(bearer, h.mediator.Did, h.mediator.SecretsResolver, h.mediator.DidResolver)

	if err != nil {
		config.Logger.Error("Error during verification " + err.Error())
		type body struct {
			Comment string `json:"comment"`
		}
		b := body{
			Comment: "Mediatee cant be registered.",
		}
		bodyJson, err := json.Marshal(b)
		if err != nil {
			config.Logger.Error("Can not marshal string", err)
			return PR_INTERNAL_SERVER_ERROR, err
		}
		response = didcomm.Message{
			Id:   uuid.Must(uuid.NewRandom()).String(),
			Type: constants.PIURI_COORDINATE_MEDIATION_RESPOSE_DENY,
			Body: string(bodyJson),
			From: &h.mediator.Did,
		}
		return response, nil
	}

	ok, err := h.mediator.Database.IsMediated(*message.From)

	if err != nil {
		config.Logger.Error("error during mediatee check", err)
		return PR_INVALID_REQUEST, err
	}

	if ok {
		config.Logger.Error("Connection already exists", err)
		return PR_INVALID_REQUEST, err
	}

	db := h.mediator.Database

	invitation, err := h.mediator.Database.GetMediatee(id)

	if err != nil {
		config.Logger.Error("invitation not found", err)
		return PR_INVALID_REQUEST, err
	}

	service, err := h.mediator.CreateMediatorService()
	if err != nil {
		config.Logger.Error("Mediator service creation failed", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}

	// Grant Message
	routingKey, err := mediator.NumAlgo2(service, *&h.mediator.SecretsResolver, *&h.mediator.DidResolver)
	if err != nil {
		return PR_INTERNAL_SERVER_ERROR, err
	}

	// request model
	type body struct {
		RoutingDid []string `json:"routing_did"`
	}

	b := body{
		RoutingDid: []string{routingKey},
	}

	bodyJson, err := json.Marshal(b)
	if err != nil {
		config.Logger.Error("Can not marshal string", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}
	err = db.SetRoutingKey(*message.From, routingKey)
	if err != nil {
		config.Logger.Error("Can not add meediatee to db", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}

	response = didcomm.Message{
		Id:   uuid.Must(uuid.NewRandom()).String(),
		Type: constants.PIURI_COORDINATE_MEDIATION_RESPOSE_GRANT,
		Body: string(bodyJson),
		// To:   &[]string{*message.From},
		From: &h.mediator.Did,
	}

	err = h.mediator.ConnectionManager.StoreConnection(invitation.Protocol, *message.From, invitation.Topic, invitation.Properties, invitation.EventType, []string{routingKey}, invitation.Group)

	if err != nil {
		config.Logger.Error("error finalizing mediatee", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}

	h.mediator.Database.DeleteMediatee(id)

	key, err := db.GetRoutingKey(*message.From)

	if err != nil {
		config.Logger.Error("Can not get routing key from db", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}

	if key != "" {
		response = didcomm.Message{
			Id:   uuid.Must(uuid.NewRandom()).String(),
			Type: constants.PIURI_COORDINATE_MEDIATION_RESPOSE_DENY,
			Body: "{}",
			From: &h.mediator.Did,
		}
		return response, nil
	}

	mediatee, err := h.mediator.Database.GetMediatee(*message.From)

	if err != nil {
		config.Logger.Error("error getting mediatee", err)
		return PR_INTERNAL_SERVER_ERROR, err
	}

	inv := messaging.InvitationNotify{
		InvitationId: id,
		Did:          mediatee.RoutingKey,
	}

	err = sendCloudEvent(inv, mediatee, config.CurrentConfiguration.CloudForwarding.Nats.Topic+"-invitation")

	return response, err
}

// remove an element from the list based on RecipientDid
func removeUpdateByRecipientDid(updates *[]Update, recipientDid string) {
	for i, update := range *updates {
		if update.RecipientDid == recipientDid {
			// Remove the element at index i
			*updates = append((*updates)[:i], (*updates)[i+1:]...)
			return
		}
	}
}
