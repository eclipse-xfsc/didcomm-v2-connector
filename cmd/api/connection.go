package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/protocol"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/database"
)

type ConnectionResponse struct {
	database.Mediatee
	DidDoc *mediator.DidDocumentJSON `json:"didDocument,omitempty"`
}

// @Summary	Get connections
// @Schemes
// @Description	Returns a list with the existing connections
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Success		200	{array}	database.Mediatee
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections [get]
func (app *application) GetConnections(context *gin.Context) {
	logTag := "/admin/connections [get]"
	config.Logger.Info(logTag, "Start", true)

	var g *string
	var s *string
	group, ok := context.GetQuery("group")

	if ok {
		g = &group
	}

	search, ok := context.GetQuery("search")

	if ok {
		s = &search
	}

	connections, err := app.mediator.Database.GetMediatees(g)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
	}

	var responseObjects = make([]ConnectionResponse, 0)

	for _, x := range connections {

		// Filter out all Temporary records
		if !strings.Contains(x.RemoteDid, "did:") {
			continue
		}

		if s != nil {

			b, _ := json.Marshal(x)

			if !strings.Contains(string(b), *s) {
				continue
			}
		}

		doc, _ := app.mediator.DidResolver.ResolveDidAsJson(x.RemoteDid)

		response := ConnectionResponse{
			Mediatee: x,
			DidDoc:   doc,
		}

		responseObjects = append(responseObjects, response)
	}

	config.Logger.Info(logTag, "End", true)

	context.JSON(http.StatusOK, responseObjects)

}

// @Summary	Get connection
// @Schemes
// @Description	Returns a connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path		string	true	"DID"
// @Success		200	{object}	database.Mediatee
// @Failure		204	"No object found"
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections/:did [get]
func (app *application) GetConnection(context *gin.Context) {
	logTag := "/admin/connections/{did} [get]"
	did := context.Param("did")
	config.Logger.Info(logTag, "did", did, "Start", true)
	connection, err := app.mediator.Database.GetMediatee(did)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}

	doc, _ := app.mediator.DidResolver.ResolveDidAsJson(did)

	response := ConnectionResponse{
		Mediatee: *connection,
		DidDoc:   doc,
	}
	config.Logger.Info(logTag, "End", true)

	context.IndentedJSON(http.StatusOK, response)
}

// @Summary	Create a new connection
// @Schemes
// @Description	Creates a new connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			connection	body		database.MediateeBase	true	"Connection"
// @Success		201			{object}	string
// @Failure		400			"Bad Request"
// @Failure		423			"Locked"
// @Failure		500			"Internal Server Error"
// @Router			/admin/connections [post]
/*func (app *application) CreateConnection(context *gin.Context) {
	logTag := "/admin/connections [post]"
	config.Logger.Info(logTag, "Start", true)
	context.Header("Content-Type", "application/json")

	var mediateeBase database.MediateeBase
	err := context.ShouldBindJSON(&mediateeBase)

	if err != nil {
		_ = app.SendPr(context, protocol.PR_INVALID_REQUEST, err)
		context.Status(http.StatusBadRequest)
		return
	}

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

	err = app.mediator.ConnectionManager.StoreConnection(mediateeBase.Protocol, mediateeBase.RemoteDid, mediateeBase.Topic, mediateeBase.Properties)
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
	/*oob := protocol.NewOutOfBand(app.mediator)
	invitation, err := oob.Handle(config.CurrentConfiguration.Label)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.String(http.StatusCreated, invitation)
}*/

// @Summary	Deletes a connection
// @Schemes
// @Description	Deletes a connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path string	true	"DID"
// @Success		200			"OK"
// @Failure		500			"Internal Server Error"
// @Router			/admin/connections/{did} [delete]
func (app *application) DeleteConnection(context *gin.Context) {
	logTag := "/admin/connections/{did} [delete]"
	did := context.Param("did")
	config.Logger.Info(logTag, "did", did, "Start", true)
	err := app.mediator.Database.DeleteMediatee(did)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.Status(http.StatusOK)

}

// @Summary	Updates a connection
// @Schemes
// @Description	Updates a connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path string	true	"DID"
// @Success		200			"OK"
// @Failure		500			"Internal Server Error"
// @Router			/admin/connections/{did} [delete]
func (app *application) UpdateConnection(context *gin.Context) {
	logTag := "/admin/connections/{did} [update]"
	did := context.Param("did")
	config.Logger.Info(logTag, "did", did, "Start", true)

	body, err := io.ReadAll(context.Request.Body)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	err = context.Request.Body.Close()

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		config.Logger.Error(logTag, "Error", errors.New("no body"))
		context.Status(http.StatusBadRequest)
		return
	}

	var mediatee database.Mediatee

	err = json.Unmarshal(body, &mediatee)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	if mediatee.RemoteDid != "" && mediatee.RemoteDid != did {
		config.Logger.Error(logTag, "Error", errors.New("did mismatch"))
		context.Status(http.StatusBadRequest)
		return
	}

	mediatee.RemoteDid = did

	err = app.mediator.Database.UpdateMediatee(mediatee)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.Status(http.StatusOK)

}

// @Summary	Blocks connection
// @Schemes
// @Description	Blocks connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path	string	true	"Did to be blocked"
// @Success		200	"OK"
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections/block/{did} [post]
func (app *application) BlockConnection(context *gin.Context) {
	logTag := "/admin/connections/block/{did} [post]"
	did := context.Param("did")
	config.Logger.Info(logTag, "did", did, "Start", true)

	err := app.mediator.Database.BlockMediatee(did)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.Status(http.StatusOK)
}

// @Summary	Unblock existing connection
// @Schemes
// @Description	Blocks existing connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path	string	true	"Did to be unblocked"
// @Success		200	"OK"
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections/unblock/{did} [post]
func (app *application) UnblockConnection(context *gin.Context) {
	logTag := "/admin/connections/unblock/{did} [post]"
	did := context.Param("did")
	config.Logger.Info(logTag, "did", did, "Start", true)
	err := app.mediator.Database.UnblockMediatee(did)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.Status(http.StatusOK)
}

// @Summary	Checks if a remoteDid is blocked
// @Schemes
// @Description	Checks if a remoteDid is blocked
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Param			did	path		string	true	"Did"
// @Success		200	{object}	boolean
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections/isblocked/{did} [get]
func (app *application) IsBlocked(context *gin.Context) {
	logTag := "/admin/connections/isblocked/{did} [get]"
	did := context.Param("did")

	config.Logger.Info(logTag, "did", did, "Start", true)
	ib, err := app.mediator.Database.IsBlocked(did)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.IndentedJSON(http.StatusOK, ib)
}

// @Summary	Accept connection
// @Schemes
// @Description	accept connection
// @Tags			Connections
// @Accept			json
// @Produce		json
// @Success		200	"OK"
// @Failure		500	"Internal Server Error"
// @Router			/admin/connections/unblock/{did} [post]
func (app *application) AcceptConnection(context *gin.Context) {
	logTag := "/admin/connections/accept"
	config.Logger.Info(logTag, "Start", true)

	type Invitation struct {
		Invitation string `json:"invitation"`
		database.MediateeBase
	}

	var inv Invitation

	err := context.ShouldBindJSON(&inv)

	if inv.Invitation == "" ||
		inv.EventType == "" ||
		inv.Properties == nil ||
		inv.Topic == "" ||
		inv.Group == "" {
		_ = app.SendPr(context, protocol.PR_INVALID_REQUEST, err)
		context.Status(http.StatusBadRequest)
		return
	}

	if err != nil {
		_ = app.SendPr(context, protocol.PR_INVALID_REQUEST, err)
		context.Status(http.StatusBadRequest)
		return
	}

	uri, err := url.Parse(inv.Invitation)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	q, err := url.ParseQuery(uri.RawQuery)
	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusInternalServerError)
		return
	}

	oob := q.Get("_oob")

	b, err := base64.RawURLEncoding.DecodeString(oob)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	msg, err := app.mediator.UnpackMessage(string(b))

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}

	var bodyJson map[string]interface{}

	err = json.Unmarshal([]byte(msg.Body), &bodyJson)

	if err != nil {
		config.Logger.Error("error unmarshalling", err)
		context.Status(http.StatusInternalServerError)
		return
	}

	s, err := mediator.CreateServiceEntry()

	if err != nil {
		config.Logger.Error("cant create Service", err)
		context.Status(http.StatusInternalServerError)
		return
	}

	orgPeerDid, err := mediator.NumAlgo2([]didcomm.Service{s}, app.mediator.SecretsResolver, app.mediator.DidResolver)

	if err != nil {
		config.Logger.Error("cant generate peer did", err)
		context.Status(http.StatusInternalServerError)
		return
	}

	// routing key must be extracted from mediaton request
	connectUrl := uri.Scheme + "://" + uri.Host + uri.Path + "/message/receive"
	peerdid, err := app.mediator.ConnectionManager.Connect(connectUrl, *msg.From, orgPeerDid, bodyJson["auth"].(string))

	if err != nil {
		config.Logger.Error("cant connect to the peer", err, "url", connectUrl)
		context.Status(http.StatusInternalServerError)
		return
	}

	err = app.mediator.ConnectionManager.StoreConnection(inv.Protocol, *msg.From, inv.Topic, inv.Properties, inv.EventType, []string{peerdid}, inv.Group)

	if err != nil {
		config.Logger.Error(logTag, "Error", err)
		context.Status(http.StatusBadRequest)
		return
	}
	config.Logger.Info(logTag, "End", true)
	context.JSON(200, peerdid)
}

func (app *application) SendPr(context *gin.Context, pr didcomm.Message, err error) error {
	context.Header("Content-Type", "application/json")
	if err != nil {
		config.Logger.Error("Problem Report", "err", err)
	}
	packMsg, err := app.mediator.PackPlainMessage(pr)
	if err != nil {
		config.Logger.Error("Problem Report", "err", err)
		context.Status(http.StatusBadRequest)
		return err
	}
	context.String(http.StatusBadRequest, packMsg)
	return nil
}
