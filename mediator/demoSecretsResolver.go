package mediator

import (
	"fmt"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
)

// GetSecret(secretid string, cb *OnGetSecretResult) ErrorCode
// FindSecrets(secretids []string, cb *OnFindSecretsResult) ErrorCode

type DemoSecretsResolver struct {
	secrets map[string]didcomm.Secret
}

func NewDemoSecretsResolver() *DemoSecretsResolver {
	secrets := make(map[string]didcomm.Secret)
	return &DemoSecretsResolver{
		secrets: secrets,
	}
}

func (d *DemoSecretsResolver) GetSecret(secretId string, cb *didcomm.OnGetSecretResult) didcomm.ErrorCode {
	if secret, ok := d.secrets[secretId]; ok {
		cb.Success(&secret)
		return didcomm.ErrorCodeSuccess
	} else {
		errorKind := didcomm.NewErrorKindSecretNotFound()
		cb.Error(errorKind, "Secret not found")
		return didcomm.ErrorCodeError
	}
}

func (d *DemoSecretsResolver) FindSecrets(secretIds []string, cb *didcomm.OnFindSecretsResult) didcomm.ErrorCode {
	var secrets []string
	for _, id := range secretIds {
		if secret, ok := d.secrets[id]; ok {
			secrets = append(secrets, secret.Id)
		}
	}
	cb.Success(secrets)
	return didcomm.ErrorCodeSuccess
}

func (d *DemoSecretsResolver) StoreSecret(secret didcomm.Secret) {
	d.secrets[secret.Id] = secret
}

// Demo Print Function
func (d *DemoSecretsResolver) PrintSecrets() {
	// print json like secrets
	for _, secret := range d.secrets {
		fmt.Println("{")
		fmt.Println("  Id:", secret.Id)
		fmt.Println("  Type:", secret.Type)
		fmt.Println("  SecretMaterial:")
		fmt.Println("    PrivateKeyMultibase:", secret.SecretMaterial.(didcomm.SecretMaterialMultibase).PrivateKeyMultibase)
		fmt.Println("}")
	}
}
