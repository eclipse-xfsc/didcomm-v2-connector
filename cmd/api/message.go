package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	connectionmanager "github.com/eclipse-xfsc/didcomm-v2-connector/mediator/connectionManager"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/database"
	"github.com/eclipse-xfsc/didcomm-v2-connector/protocol"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

// @Summary		Receives a DIDComm message
// @Schemes
// @Description	Receives a DIDComm message
// @Tags			Message
// @Accept			json
// @Produce		json
// @Param			message	body		didcomm.Message	true	"Message"
// @Success		200	"OK"
// @Failure		400	"Bad Request"
// @Failure		500	"Internal Server Error"
// @Router			/message/receive  [post]
func (app *application) ReceiveMessage(context *gin.Context) {
	bearer := context.Request.Header.Get("Authorization")
	// get body of request
	bodyBytes, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.String(http.StatusBadRequest, "Error reading request body")
		return
	}
	bodyString := string(bodyBytes)

	// handle message
	packMsg, err := protocol.HandleMessage(bodyString, app.mediator, bearer)
	if err != nil {
		if errors.Is(err, intErr.ErrUnpackingMessage) {
			context.Status(http.StatusBadRequest)
			return
		} else {
			context.Status(http.StatusInternalServerError)
			return
		}
	}

	// answer request
	if packMsg != "" {
		var jsonMap map[string]interface{}
		err = json.Unmarshal([]byte(packMsg), &jsonMap)
		if err != nil {
			context.String(http.StatusInternalServerError, "Error marshaling message")
			return
		}

		context.JSON(http.StatusOK, jsonMap)
	} else {
		context.String(http.StatusOK, "")
	}
}

// @Summary		Create a connection invitation which is used for requesting the mediate
// @Schemes
// @Description	Create a connection invitation
// @Tags			Administration
// @Accept			json
// @Produce		json
// @Param			message	body		didcomm.Message	true	"Message"
// @Success		200	"OK"
// @Failure		400	"Bad Request"
// @Failure		500	"Internal Server Error"
// @Router			/admin/invitiation [post]
func (app *application) InvitationMessage(context *gin.Context) {

	m := app.mediator

	var mediateeBase database.MediateeBase
	err := context.ShouldBindJSON(&mediateeBase)

	if err != nil {
		_ = app.SendPr(context, protocol.PR_INVALID_REQUEST, err)
		context.Status(http.StatusBadRequest)
		return
	}

	mediateeBase.RemoteDid = uuid.NewString() //temporary

	isBlocked, err := app.mediator.Database.IsBlocked(mediateeBase.RemoteDid)
	if err != nil {
		_ = app.SendPr(context, protocol.PR_INTERNAL_SERVER_ERROR, err)
		context.Status(http.StatusInternalServerError)
		return
	}
	if isBlocked {
		_ = app.SendPr(context, protocol.PR_DID_BLOCKED, err)
		context.Status(http.StatusLocked)
		return
	}

	err = app.mediator.ConnectionManager.StoreConnection(mediateeBase.Protocol, mediateeBase.RemoteDid, mediateeBase.Topic, mediateeBase.Properties, mediateeBase.EventType, []string{}, mediateeBase.Group)
	if err != nil {
		if err == connectionmanager.ERROR_PROTOCOL_NOT_SUPPORTED {
			_ = app.SendPr(context, protocol.PR_PROTOCOL_NOT_SUPPORTED, err)
			context.Status(http.StatusBadRequest)
		} else if err == connectionmanager.ERROR_INTERNAL {
			_ = app.SendPr(context, protocol.PR_INTERNAL_SERVER_ERROR, err)
			context.Status(http.StatusInternalServerError)
		} else if err == connectionmanager.ERROR_CONNECTION_ALREADY_EXISTS {
			_ = app.SendPr(context, protocol.PR_ALREADY_CONNECTED, err)
			context.Status(http.StatusInternalServerError)
		} else {
			_ = app.SendPr(context, protocol.PR_INTERNAL_SERVER_ERROR, err)
			context.Status(http.StatusInternalServerError)
		}
		return

	}

	oob := protocol.NewOutOfBand(m)

	payload := jwt.MapClaims{
		"exp":          time.Now().Add(time.Minute * time.Duration(config.CurrentConfiguration.TokenExpiration)).Unix(),
		"invitationId": mediateeBase.RemoteDid,
	}

	token, err := mediator.GenerateSignedToken(m.Did, payload, m.SecretsResolver, m.DidResolver)

	msg, err := oob.Handle(config.CurrentConfiguration.Label, token)
	if err != nil {
		context.String(http.StatusInternalServerError, "Error handling out of band")
	} else {

		msg64 := base64.RawURLEncoding.EncodeToString([]byte(msg))
		oob := fmt.Sprintf("%s?_oob=%s", config.CurrentConfiguration.Url, msg64)
		context.String(http.StatusOK, oob)
	}
}
