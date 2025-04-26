package secretsresolver

import "github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"

// Is needed to have the store functionallity
type Adapter interface {
	GetPlainSecret(secretid string) *didcomm.Secret
	GetSecret(secretid string, cb *didcomm.OnGetSecretResult) didcomm.ErrorCode
	FindSecrets(secretids []string, cb *didcomm.OnFindSecretsResult) didcomm.ErrorCode
	StoreSecret(secret didcomm.Secret) error
}
