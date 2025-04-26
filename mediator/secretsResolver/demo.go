package secretsresolver

import (
	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
)

// GetSecret(secretid string, cb *OnGetSecretResult) ErrorCode
// FindSecrets(secretids []string, cb *OnFindSecretsResult) ErrorCode

// Did does not have meaningful service endpoint
const DID = "did:peer:2.Ez6LSc19SfftNpBDVqcd8NQtef2vinvR3W8s1wVeoYzwy5yiw.Vz6Mkno2XmnAxWb7YbyDJw9hmqWcTuAwQKbtaiw9tjqRjDvMz.SeyJ0IjoiZG0iLCJzIjp7InVyaSI6Imh0dHA6Ly9sb2NhbGhvc3Q6OTA5MC9tZXNzYWdlL3JlY2VpdmUiLCJhIjpbImRpZGNvbW0vdjIiXSwiciI6W119fQ"

type Demo struct {
	secrets map[string]didcomm.Secret
}

func NewDemo() *Demo {

	secrets := make(map[string]didcomm.Secret)

	// store demo secrets
	vSecret := createDemoVerificationSecrets()
	secrets[vSecret.Id] = vSecret
	eSecret := createDemoEncryptionSecrets()
	secrets[eSecret.Id] = eSecret

	return &Demo{
		secrets: secrets,
	}
}

func (d *Demo) GetPlainSecret(secretid string) *didcomm.Secret {
	if secret, ok := d.secrets[secretid]; ok {
		return &secret
	}
	return nil
}

func (d *Demo) GetSecret(secretId string, cb *didcomm.OnGetSecretResult) didcomm.ErrorCode {
	if secret, ok := d.secrets[secretId]; ok {
		err := cb.Success(&secret)
		if err != nil {
			config.Logger.Error("Unable to use success  channel while getting secret", "msg", err)
			return didcomm.ErrorCodeError
		}
		return didcomm.ErrorCodeSuccess
	} else {
		errorKind := didcomm.NewErrorKindSecretNotFound()
		err := cb.Error(errorKind, "Secret not found")
		if err != nil {
			config.Logger.Error("Unable to use error channel while getting secret", "msg", err)
			return didcomm.ErrorCodeError
		}
		return didcomm.ErrorCodeError
	}
}

func (d *Demo) FindSecrets(secretIds []string, cb *didcomm.OnFindSecretsResult) didcomm.ErrorCode {
	var secrets []string
	for _, id := range secretIds {
		if secret, ok := d.secrets[id]; ok {
			secrets = append(secrets, secret.Id)
		}
	}
	err := cb.Success(secrets)
	if err != nil {
		config.Logger.Error("Unable to use success  channel while getting secret", "msg", err)
		return didcomm.ErrorCodeError
	}
	return didcomm.ErrorCodeSuccess
}

func (d *Demo) StoreSecret(secret didcomm.Secret) error {
	d.secrets[secret.Id] = secret
	return nil
}

func createDemoVerificationSecrets() didcomm.Secret {
	// verification secret
	kid := "6Mkno2XmnAxWb7YbyDJw9hmqWcTuAwQKbtaiw9tjqRjDvMz"
	privateKey := "zrv4kkW1Czmu5VCKaHtfuoxFqSLAb4gwk7sjLc89DuzwMe3wZ6VTF5FtoLNRZFAgwBD63waizHTx7ih1yzB2tzAc8bW"
	did := DID
	secret := didcomm.Secret{
		Id:   did + "#" + kid,
		Type: didcomm.SecretTypeEd25519VerificationKey2020,
		SecretMaterial: didcomm.SecretMaterialMultibase{
			PrivateKeyMultibase: privateKey,
		},
	}
	return secret
}

func createDemoEncryptionSecrets() didcomm.Secret {
	kid := "6LSc19SfftNpBDVqcd8NQtef2vinvR3W8s1wVeoYzwy5yiw"
	// privateKey := "z3weggogbWATqzsigwidsXZ1MPFtJwnPgN3LEo5cS6PmzEva"
	privateKey := "z3wehJqXpqKUbrdA3cDz9buSQXxfXYFLuVHUVACmAkNutGme"
	did := DID
	secret := didcomm.Secret{
		Id:   did + "#" + kid,
		Type: didcomm.SecretTypeX25519KeyAgreementKey2020,
		SecretMaterial: didcomm.SecretMaterialMultibase{
			PrivateKeyMultibase: privateKey,
		},
	}
	return secret
}
