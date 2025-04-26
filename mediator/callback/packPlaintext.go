package callback

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

type PackErrorPair struct {
	Err *didcomm.ErrorKind
	Msg string
}

type PackResultCallback struct {
	msgCh chan<- string
	errCh chan<- PackErrorPair
}

func NewPackResultCallback(msgCh chan<- string, errCh chan<- PackErrorPair) *PackResultCallback {
	return &PackResultCallback{
		msgCh: msgCh,
		errCh: errCh,
	}
}

func (m *PackResultCallback) Success(result string) {
	m.msgCh <- result
	close(m.msgCh)
	close(m.errCh)
}

func (m *PackResultCallback) Error(err *didcomm.ErrorKind, msg string) {
	m.errCh <- PackErrorPair{err, msg}
	close(m.errCh)
	close(m.msgCh)
}
