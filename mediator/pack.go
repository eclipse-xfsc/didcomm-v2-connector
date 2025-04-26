package mediator

import (
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/callback"
)

func (m *Mediator) PackEncryptedMessage(message didcomm.Message, to string, from string) (response string, err error) {

	// set PackEncryptedOptions
	pencryptOpt := didcomm.PackEncryptedOptions{
		ProtectSender: false,
		Forward:       true,
		EncAlgAuth:    didcomm.AuthCryptAlgA256cbcHs512Ecdh1puA256kw,
		EncAlgAnon:    didcomm.AnonCryptAlgA256cbcHs512EcdhEsA256kw,
	}

	// prepare callback
	sucCh := make(chan callback.PackEncryptedSuccessPair, 1)
	errCh := make(chan callback.PackEncryptedErrorPair, 1)
	packEncryptCB := callback.NewPackEncryptedResultCallback(sucCh, errCh)

	dc := m.Messages
	// needs to be set, otherwise it will fail when unpacking
	message.Typ = "application/didcomm-plain+json"
	// Signing works as well
	dc.PackEncrypted(message, to, &from, &from, pencryptOpt, packEncryptCB)

	select {
	case err := <-errCh:
		m.Logger.Error("Error packing message:", "msg", err.Msg)
		return "", err.Err
	case suc := <-sucCh:
		return suc.Result, nil
	}
}

func (m *Mediator) PackPlainMessage(message didcomm.Message) (response string, err error) {
	strCh := make(chan string, 1)
	errCh := make(chan callback.PackErrorPair, 1)
	cb := callback.NewPackResultCallback(strCh, errCh)
	dc := m.Messages
	message.Typ = "application/didcomm-plain+json"
	dc.PackPlaintext(message, cb)

	select {
	case e := <-errCh:
		m.Logger.Error("Error packing message:", "msg", e.Msg)
		return "", e.Err
	case m := <-strCh:
		return m, nil
	}
}
