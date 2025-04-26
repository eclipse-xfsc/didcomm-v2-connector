package mediator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
)

// type DidResolver interface {
// 	Resolve(did string, cb *OnDidResolverResult) ErrorCode
// }

type DidDocumentJSON struct {
	DidDocument               `json:"didDocument"`
	DidDocumentMetadataJSON   `json:"didDocumentMetadata"`
	DidResolutionMetadataJSON `json:"didResolutionMetadata"`
}

// DidDocument represents the nested structure inside the main JSON.
type DidDocument struct {
	Context              []string             `json:"@context"`
	ID                   string               `json:"id"`
	VerificationMethod   []VerificationMethod `json:"verificationMethod"`
	KeyAgreement         []string             `json:"keyAgreement"`
	Authentication       []string             `json:"authentication"`
	AssertionMethod      []string             `json:"assertionMethod"`
	CapabilityInvocation []string             `json:"capabilityInvocation"`
	CapabilityDelegation []string             `json:"capabilityDelegation"`
	Service              []Service            `json:"service"`
}

// VerificationMethod represents a nested structure inside the DidDocument.
type VerificationMethod struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Controller         string `json:"controller"`
	PublicKeyMultibase string `json:"publicKeyMultibase"`
}

// Service represents the nested structure inside the DidDocument for the "service" field.
type Service struct {
	// RoutingKeys     []string `json:"routingKeys"`
	// Accept          []string `json:"accept"`
	Type            string          `json:"type"`
	ID              string          `json:"id"`
	ServiceEndpoint ServiceEndpoint `json:"serviceEndpoint"`
}

type ServiceEndpoint struct {
	Uri         string   `json:"uri"`
	Accept      []string `json:"accept"`
	RoutingKeys []string `json:"routingKeys"`
}

// DidDocumentMetadata represents the nested structure inside the main JSON for "didDocumentMetadata".
type DidDocumentMetadataJSON struct{}

// DidResolutionMetadata represents the nested structure inside the main JSON for "didResolutionMetadata".
type DidResolutionMetadataJSON struct {
	ContentType string `json:"contentType"`
}

type DidResolver interface {
	Resolve(did string, cb *didcomm.OnDidResolverResult) didcomm.ErrorCode
	ResolveDid(did string) (*didcomm.DidDoc, error)
	ResolveDidAsJson(did string) (*DidDocumentJSON, error)
}

type UniverseDidResolver struct {
}

func NewDidResolver() *UniverseDidResolver {
	return &UniverseDidResolver{}
}

// Note: Only the Service Endpoint encoding of the official did:peer method is supported.
// https://identity.foundation/peer-did-method-spec/#generating-a-didpeer2
func (u *UniverseDidResolver) Resolve(did string, cb *didcomm.OnDidResolverResult) didcomm.ErrorCode {
	didDoc, err := u.ResolveDid(did)
	if err != nil {
		errorKind := didcomm.NewErrorKindSecretNotFound()
		err = cb.Error(errorKind, err.Error())
		if err != nil {
			config.Logger.Error("Resolve did callback error not working", err)
		}
		return didcomm.ErrorCodeError
	} else {
		err = cb.Success(didDoc)
		if err != nil {
			config.Logger.Error("Resolve did callback success not working", err)
		}
		return didcomm.ErrorCodeSuccess
	}
}

func (u *UniverseDidResolver) ResolveDid(did string) (*didcomm.DidDoc, error) {
	queryUrl, err := url.JoinPath(config.CurrentConfiguration.DidComm.ResolverUrl, "/1.0/identifiers/", did)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(queryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dict map[string]interface{}

	err = json.Unmarshal(body, &dict)

	if err != nil {
		config.Logger.Error("Unmarshaling has failed:", "err", err)
		return nil, err
	}

	_, isDriver := dict["didDocument"]

	if isDriver {
		var didDocJSON DidDocumentJSON
		err = json.Unmarshal(body, &didDocJSON)
		if err != nil {

			uT, ok := err.(*json.UnmarshalTypeError)

			if ok {
				if uT.Field != "didDocument.@context" {
					config.Logger.Error("Unmarshaling has failed:", "err", err)
					return nil, err
				}
			}
		}
		didDoc := u.didDocJsonToDidDoc(didDocJSON)
		return didDoc, nil
	}

	var didDoc didcomm.DidDoc

	err = json.Unmarshal(body, &didDoc)

	return &didDoc, err
}

func (u *UniverseDidResolver) ResolveDidAsJson(did string) (*DidDocumentJSON, error) {
	queryUrl, err := url.JoinPath(config.CurrentConfiguration.DidComm.ResolverUrl, "/1.0/identifiers/", did)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(queryUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var dict map[string]interface{}

	err = json.Unmarshal(body, &dict)

	if err != nil {
		config.Logger.Error("Unmarshaling has failed:", "err", err)
		return nil, err
	}

	_, isDriver := dict["didDocument"]
	var didDocJSON DidDocumentJSON
	if isDriver {
		err = json.Unmarshal(body, &didDocJSON)
		uT, ok := err.(*json.UnmarshalTypeError)

		if ok {
			if uT.Field != "didDocument.@context" {
				config.Logger.Error("Unmarshaling has failed:", "err", err)
				return nil, err
			}
		}
	} else {
		var didDoc DidDocument

		err = json.Unmarshal(body, &didDoc)

		didDocJSON = DidDocumentJSON{
			DidDocument: didDoc,
		}
	}

	return &didDocJSON, err
}

func (u *UniverseDidResolver) didDocJsonToDidDoc(ddJson DidDocumentJSON) *didcomm.DidDoc {
	// Convert VerificationMethods
	var VerificationMethods []didcomm.VerificationMethod = make([]didcomm.VerificationMethod, len(ddJson.DidDocument.VerificationMethod))
	for i, vm := range ddJson.DidDocument.VerificationMethod {
		VerificationMethodType, err := getVerificationType(vm.Type)
		if err != nil {
			config.Logger.Error("didDocJsonToDidDoc:", "err", err)
		}
		VerificationMethods[i] = didcomm.VerificationMethod{
			Id:         vm.ID,
			Type:       VerificationMethodType,
			Controller: vm.Controller,
			VerificationMaterial: didcomm.VerificationMaterialMultibase{
				PublicKeyMultibase: vm.PublicKeyMultibase,
			},
		}
	}

	// Convert Services
	var Services []didcomm.Service = make([]didcomm.Service, len(ddJson.DidDocument.Service))
	for i, s := range ddJson.DidDocument.Service {
		if s.Type != "DIDCommMessaging" {
			// MUST be DIDCommMessaging see https://identity.foundation/didcomm-messaging/spec/#service-endpoint
			continue
		}
		Services[i] = didcomm.Service{
			Id: s.ID,
			ServiceEndpoint: didcomm.ServiceKindDidCommMessaging{
				Value: didcomm.DidCommMessagingService{
					Uri:         s.ServiceEndpoint.Uri,
					Accept:      &s.ServiceEndpoint.Accept,
					RoutingKeys: s.ServiceEndpoint.RoutingKeys,
				},
			},
		}
	}

	didDoc := didcomm.DidDoc{
		Id:                 ddJson.DidDocument.ID,
		Authentication:     ddJson.DidDocument.Authentication,
		KeyAgreement:       ddJson.DidDocument.KeyAgreement,
		VerificationMethod: VerificationMethods,
		Service:            Services,
	}

	return &didDoc
}

func getVerificationType(stype string) (didcomm.VerificationMethodType, error) {
	// VerificationMethodTypeJsonWebKey2020                    VerificationMethodType = 1
	// VerificationMethodTypeX25519KeyAgreementKey2019         VerificationMethodType = 2
	// VerificationMethodTypeEd25519VerificationKey2018        VerificationMethodType = 3
	// VerificationMethodTypeEcdsaSecp256k1VerificationKey2019 VerificationMethodType = 4
	// VerificationMethodTypeX25519KeyAgreementKey2020         VerificationMethodType = 5
	// VerificationMethodTypeEd25519VerificationKey2020        VerificationMethodType = 6
	// VerificationMethodTypeOther                             VerificationMethodType = 7

	switch stype {
	case "JsonWebKey2020":
		return didcomm.VerificationMethodTypeJsonWebKey2020, nil
	case "X25519KeyAgreementKey2019":
		return didcomm.VerificationMethodTypeX25519KeyAgreementKey2019, nil
	case "Ed25519VerificationKey2018":
		return didcomm.VerificationMethodTypeEd25519VerificationKey2018, nil
	case "EcdsaSecp256k1VerificationKey2019":
		return didcomm.VerificationMethodTypeEcdsaSecp256k1VerificationKey2019, nil
	case "X25519KeyAgreementKey2020":
		return didcomm.VerificationMethodTypeX25519KeyAgreementKey2020, nil
	case "Ed25519VerificationKey2020":
		return didcomm.VerificationMethodTypeEd25519VerificationKey2020, nil
	default:
		return didcomm.VerificationMethodTypeOther, fmt.Errorf("unknown VerificationMethodType: %s", stype)
	}

}
