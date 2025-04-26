package callback

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

type UnpackErrorPair struct {
	Err *didcomm.ErrorKind
	Msg string
}

type UnpackResultCallback struct {
	msgCh chan<- didcomm.Message
	errCh chan<- UnpackErrorPair
}

func NewUnpackResultCallback(msgCh chan<- didcomm.Message, errCh chan<- UnpackErrorPair) *UnpackResultCallback {
	return &UnpackResultCallback{
		msgCh: msgCh,
		errCh: errCh,
	}
}

func (m *UnpackResultCallback) Success(result didcomm.Message, metadata didcomm.UnpackMetadata) {
	m.msgCh <- result
	close(m.msgCh)
	close(m.errCh)
}

func (m *UnpackResultCallback) Error(err *didcomm.ErrorKind, msg string) {
	m.errCh <- UnpackErrorPair{err, msg}
	close(m.errCh)
	close(m.msgCh)
}
