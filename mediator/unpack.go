package mediator

import (
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/callback"
)

func (m *Mediator) UnpackMessage(body string) (didcomm.Message, error) {
	// var message didcomm.Message = didcomm.Message{}
	options := didcomm.UnpackOptions{
		ExpectDecryptByAllKeys:  true,
		UnwrapReWrappingForward: true,
	}
	msgCh := make(chan didcomm.Message, 1)
	errCh := make(chan callback.UnpackErrorPair, 1)
	unpackCB := callback.NewUnpackResultCallback(msgCh, errCh)

	bodyString := string(body)
	dc := m.Messages
	go dc.Unpack(bodyString, options, unpackCB)
	select {
	case e := <-errCh:
		m.Logger.Error("Error unpacking message:", "msg", e.Msg)
		return didcomm.Message{}, e.Err
	case message := <-msgCh:
		return message, nil
	}
}
