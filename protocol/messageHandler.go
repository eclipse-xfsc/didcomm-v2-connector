package protocol

import (
	"errors"
	"strings"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	intErr "github.com/eclipse-xfsc/didcomm-v2-connector/internal/errors"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/pkg/constants"
)

func HandleMessage(bodyString string, mediator *mediator.Mediator, bearer string) (packMsg string, err error) {

	messageExpired := false
	messageWrongCreationTime := false

	// unpack message
	msg, err := mediator.UnpackMessage(bodyString)
	if err != nil {
		config.Logger.Error("Error unpacking message", "err", err)
		internal_error := PR_MESSAGE_NOT_UNPACKABLE
		internal_error.To = &[]string{""}
		internal_error.From = &mediator.Did
		timeNow := uint64(time.Now().UTC().Unix())
		internal_error.CreatedTime = &timeNow
		pr, err := mediator.PackPlainMessage(internal_error)
		if err != nil {
			return "", err
		}
		return pr, intErr.ErrUnpackingMessage
	}

	// check optional field created_time and expiration_time
	if msg.CreatedTime != nil {
		now := time.Now().Unix()
		if *msg.CreatedTime > uint64(now) {
			config.Logger.Warn("Received message with creation time in the future", "creationTime", *msg.CreatedTime, "nowTime", now)
			messageWrongCreationTime = true
		}
	}
	// check optional field expiration_time
	if msg.ExpiresTime != nil {
		now := time.Now().Unix()
		if *msg.ExpiresTime < uint64(now) {
			config.Logger.Warn("Received expired message", "expireTime", *msg.ExpiresTime, "nowTime", now)
			messageExpired = true
		}
	}

	// check if did is blocked
	isBlocked, err := mediator.Database.IsBlocked(*msg.From)
	if err != nil {
		errMsg := "unable to check if DID is blocked"
		config.Logger.Error(errMsg, "err", err)
		return "", errors.New(errMsg)
	}

	var responseMsg didcomm.Message = didcomm.Message{}

	if isBlocked {
		responseMsg = PR_DID_BLOCKED
		config.Logger.Info("DID is blocked", "did", *msg.From)
	} else if messageExpired {
		responseMsg = PR_EXPIRED_MESSAGE
	} else if messageWrongCreationTime {
		responseMsg = PR_MESSAGE_WRONG_CREATION_TIME
	} else if strings.HasPrefix(msg.Type, constants.PIURI_COORDINATE_MEDIATION) {

		coordinateMediation := NewCoordinateMediation(mediator)
		responseMsg, err = coordinateMediation.Handle(msg, bearer)
		if err != nil {
			errMsg := "unable to handle coordinate mediation"
			config.Logger.Error(errMsg, "err", err)
			return "", errors.New(errMsg)
		}
	} else if strings.HasPrefix(msg.Type, PIURI_TRUST_PING) {
		trustPing := NewTrustPing(mediator)
		responseMsg, err = trustPing.Handle(msg)
		if err != nil {
			switch {
			case errors.Is(err, intErr.ErrNoPingResponseRequested):
				return "", nil
			default:
				errMsg := "unable to handle trust ping"
				config.Logger.Error(errMsg, "err", err)
				return "", errors.New(errMsg)
			}
		}
	} else if strings.HasPrefix(msg.Type, PIURI_ROUTING) {
		routing := NewRouting(mediator)
		responseMsg, err = routing.Handle(msg)
		if err != nil {
			errMsg := "unable to handle routing"
			config.Logger.Error(errMsg, "err", err)
			return "", errors.New(errMsg)
		}
		if responseMsg.Type == "" {
			return "", nil
		}

	} else if strings.HasPrefix(msg.Type, PIURI_MESSAGEPICKUP) {
		messagePickup := NewMessagePickup(mediator)
		responseMsg, err = messagePickup.Handle(msg)
		if err != nil {
			errMsg := "unable to handle message pickup"
			config.Logger.Error(errMsg, "err", err)
			return "", errors.New(errMsg)
		}

	} else {
		config.Logger.Warn("Message type not handled yet.")
		responseMsg = PR_UNKNOWN_MESSAGE_TYPE
	}

	// pack response
	packMsg, err = packMessage(mediator.Did, *msg.From, responseMsg, mediator)
	if err != nil {
		internal_error := PR_INTERNAL_SERVER_ERROR
		pr, err := packMessage(mediator.Did, *msg.From, internal_error, mediator)
		if err != nil {
			return "", err
		}
		return pr, err
	}

	return
}

func packMessage(from string, to string, responseMsg didcomm.Message, mediator *mediator.Mediator) (packedMsg string, err error) {
	responseMsg.To = &[]string{to}
	responseMsg.From = &from
	timeNow := uint64(time.Now().UTC().Unix())
	responseMsg.CreatedTime = &timeNow

	if config.CurrentConfiguration.DidComm.IsMessageEncrypted {
		packedMsg, err = mediator.PackEncryptedMessage(responseMsg, to, from)
		if err != nil {
			errMsg := "unable to pack encrypted message"
			config.Logger.Error(errMsg, "err", err)
			return "", errors.New(errMsg)
		}
	} else {
		packedMsg, err = mediator.PackPlainMessage(responseMsg)
		if err != nil {
			errMsg := "unable to pack plain message"
			config.Logger.Error(errMsg, "err", err)
			return "", errors.New(errMsg)
		}
	}
	return
}
