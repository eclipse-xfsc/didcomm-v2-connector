package mediator

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	secretsresolver "github.com/eclipse-xfsc/didcomm-v2-connector/mediator/secretsResolver"

	"github.com/golang-jwt/jwt"
	multibase "github.com/multiformats/go-multibase"
)

// https://identity.foundation/peer-did-method-spec/#method-2-multiple-inception-key-without-doc

func NumAlgo2(services []didcomm.Service, secretResolver secretsresolver.Adapter, didResolver DidResolver) (peerDid string, err error) {

	encPub, encPriv, err := generateX25519Base58BTC()
	if err != nil {
		config.Logger.Error("Error generating X25518 keypair:", err)
		return
	}
	encSecret := *createSecretFromKeyPair(encPub, encPriv, didcomm.SecretTypeX25519KeyAgreementKey2020)

	signPub, singPriv, err := generateEd25519Base58BTC()
	if err != nil {
		config.Logger.Error("Error generating Ed25519 keypair:", err)
		return
	}
	signSecret := *createSecretFromKeyPair(signPub, singPriv, didcomm.SecretTypeEd25519VerificationKey2020)
	serviceB64URL := encodeServicesToB64URL(services)
	peerDid = fmt.Sprintf("did:peer:2.E%s.V%s%s", encPub, signPub, serviceB64URL)

	didDoc, err := didResolver.ResolveDid(peerDid)
	if err != nil {
		config.Logger.Error("Error resolving peer DID:", err)
		return
	}

	// store encryption keys in secret resolver
	for _, key := range didDoc.VerificationMethod {
		s := strings.Split(key.Id, "#")
		identifier := s[len(s)-1]
		for _, secret := range []didcomm.Secret{encSecret, signSecret} {

			if identifier == secret.Id {
				secret.Id = key.Id
				err = secretResolver.StoreSecret(secret)
				if err != nil {
					config.Logger.Error("Error storing secret:", err)
					return
				}
			}
		}
	}

	return
}

// Generate Encryption Keypair
func generateX25519Base58BTC() (public string, private string, err error) {
	randEncryption := rand.Reader

	privPrefix := []byte{0x82, 0x26}
	privateEncryption, err := ecdh.X25519().GenerateKey(randEncryption)
	if err != nil {
		return "", "", err
	}
	// append prefix
	privWPrefix := append(privPrefix, privateEncryption.Bytes()...)

	private, err = multibase.Encode(multibase.Base58BTC, privWPrefix)
	if err != nil {
		return "", "", err
	}

	pubPrefix := []byte{0xEC, 0x01}
	publicEncryption := privateEncryption.PublicKey()
	pubWPrefix := append(pubPrefix, publicEncryption.Bytes()...)
	public, err = multibase.Encode(multibase.Base58BTC, pubWPrefix)
	if err != nil {
		return "", "", err
	}

	return public, private, nil
}

// Generate Signing Keypair
func generateEd25519Base58BTC() (public string, private string, err error) {
	randEncryption := rand.Reader
	publicRaw, privateRaw, err := ed25519.GenerateKey(randEncryption)
	if err != nil {
		return "", "", err
	}

	privPrefix := []byte{0x80, 0x26}
	privateRaw = append(privPrefix, privateRaw...)
	private, err = multibase.Encode(multibase.Base58BTC, privateRaw)
	if err != nil {
		return "", "", err
	}

	pubPrefix := []byte{0xED, 0x01}
	publicRaw = append(pubPrefix, publicRaw...)
	public, err = multibase.Encode(multibase.Base58BTC, publicRaw)
	if err != nil {
		return "", "", err
	}

	return public, private, nil
}

type serviceEncoded struct {
	T string `json:"t"`
	S struct {
		URI string   `json:"uri"`
		A   []string `json:"a"`
		R   []string `json:"r"`
	} `json:"s"`
}

func encodeServiceToB64URL(service didcomm.Service) (multibaseService string) {
	serviceKind := service.ServiceEndpoint.(didcomm.ServiceKindDidCommMessaging)
	shortServiceStruct := serviceEncoded{
		T: "dm",
		S: struct {
			URI string   "json:\"uri\""
			A   []string "json:\"a\""
			R   []string "json:\"r\""
		}{
			URI: serviceKind.Value.Uri,
			A:   *serviceKind.Value.Accept,
			R:   serviceKind.Value.RoutingKeys,
		},
	}

	json, err := json.Marshal(shortServiceStruct)
	if err != nil {
		config.Logger.Error("Error marshalling didcomm service:", err)
		return
	}
	m := base64.RawURLEncoding.EncodeToString(json)

	// https://identity.foundation/peer-did-method-spec/#generating-a-didpeer2
	// peer DID method requires the = character to be removed from the base64url encoded string
	m = strings.ReplaceAll(m, "=", "")
	return m
}

func encodeServicesToB64URL(services []didcomm.Service) (service string) {
	serviceString := ""
	for _, s := range services {
		serviceString += fmt.Sprintf(".S%s", encodeServiceToB64URL(s))
	}
	return serviceString
}

func createSecretFromKeyPair(public string, private string, t didcomm.SecretType) *didcomm.Secret {
	// first character is the multibase identifier
	// expected to be z for base58btc
	id := public[1:]
	secret := didcomm.Secret{
		Id:   id,
		Type: t,
		SecretMaterial: didcomm.SecretMaterialMultibase{
			PrivateKeyMultibase: private,
		},
	}
	return &secret
}

func GenerateSignedToken(did string, payload jwt.MapClaims, secretResolver secretsresolver.Adapter, didResolver DidResolver) (string, error) {
	doc, err := didResolver.ResolveDid(did)
	if err != nil {
		config.Logger.Error("Error resolving peer DID:", err)
		return "", err
	}

	if len(doc.Authentication) == 0 {
		return "", errors.New("No auth key found")
	}

	secret := secretResolver.GetPlainSecret(doc.Authentication[0])
	if secret == nil {
		return "", errors.New("No secret found")
	}

	_, by, err := multibase.Decode(secret.SecretMaterial.(didcomm.SecretMaterialMultibase).PrivateKeyMultibase)

	if err != nil {
		return "", err
	}

	privateKey := ed25519.PrivateKey(by[2:])

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)

	return token.SignedString(privateKey)
}

func VerifySignedToken(tokenString, did string, secretResolver secretsresolver.Adapter, didResolver DidResolver) (string, error) {
	doc, err := didResolver.ResolveDid(did)
	if err != nil {
		config.Logger.Error("Error resolving peer DID:", err)
		return "", err
	}

	if len(doc.Authentication) == 0 {
		return "", errors.New("No auth key found")
	}

	secret := secretResolver.GetPlainSecret(doc.Authentication[0])
	if secret == nil {
		return "", errors.New("No secret found")
	}

	_, by, err := multibase.Decode(secret.SecretMaterial.(didcomm.SecretMaterialMultibase).PrivateKeyMultibase)

	if err != nil {
		return "", err
	}

	parts := strings.Split(tokenString, "Bearer")

	if len(parts) > 1 {
		tokenString = parts[1]
	} else {
		tokenString = parts[0]
	}

	tok, err := jwt.Parse(strings.TrimLeft(tokenString, " "), func(token *jwt.Token) (interface{}, error) {
		// Überprüfen Sie, ob der Signierungs-Algorithmus korrekt ist
		if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("Unsupported Algorithm: %v", token.Header["alg"])
		}

		return ed25519.PrivateKey(by[2:]).Public(), nil
	})

	claims, ok := tok.Claims.(jwt.MapClaims)

	if !tok.Valid || !ok {
		return "", errors.New("Token not valid")
	}

	return claims["invitationId"].(string), nil
}
