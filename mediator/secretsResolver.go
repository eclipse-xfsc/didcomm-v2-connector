package mediator

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

// Is needed to have the store functionallity
type SecretsResolver interface {
	GetSecret(secretid string, cb *didcomm.OnGetSecretResult) didcomm.ErrorCode
	FindSecrets(secretids []string, cb *didcomm.OnFindSecretsResult) didcomm.ErrorCode
	StoreSecret(secret didcomm.Secret)
}
