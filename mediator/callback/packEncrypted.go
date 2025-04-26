package callback

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

type PackEncryptedErrorPair struct {
	Err *didcomm.ErrorKind
	Msg string
}

type PackEncryptedSuccessPair struct {
	Result   string
	Metadata didcomm.PackEncryptedMetadata
}

type PackEncryptedResultCallback struct {
	sucCh chan<- PackEncryptedSuccessPair
	errCh chan<- PackEncryptedErrorPair
}

func NewPackEncryptedResultCallback(sucCh chan<- PackEncryptedSuccessPair, errCh chan<- PackEncryptedErrorPair) *PackEncryptedResultCallback {
	return &PackEncryptedResultCallback{
		sucCh: sucCh,
		errCh: errCh,
	}
}

func (m *PackEncryptedResultCallback) Success(result string, metadata didcomm.PackEncryptedMetadata) {
	m.sucCh <- PackEncryptedSuccessPair{result, metadata}
	close(m.sucCh)
	close(m.errCh)
}

func (m *PackEncryptedResultCallback) Error(err *didcomm.ErrorKind, msg string) {
	m.errCh <- PackEncryptedErrorPair{err, msg}
	close(m.errCh)
	close(m.sucCh)
}
