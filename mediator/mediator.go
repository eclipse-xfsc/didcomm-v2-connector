package mediator

import (
	"fmt"
	"log/slog"
	"net/url"

	"github.com/eclipse-xfsc/didcomm-v2-connector/didcomm"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	connectionManager "github.com/eclipse-xfsc/didcomm-v2-connector/mediator/connectionManager"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator/database"
	secretsresolver "github.com/eclipse-xfsc/didcomm-v2-connector/mediator/secretsResolver"
)

type Mediator struct {
	ConnectionManager *connectionManager.ConnectionManager
	Messages          *didcomm.DidComm
	SecretsResolver   secretsresolver.Adapter
	DidResolver       DidResolver
	Did               string
	Database          database.Adapter
	Logger            *slog.Logger
}

func NewMediator(logger *slog.Logger) *Mediator {

	var m Mediator

	// set database
	if config.CurrentConfiguration.Database.InMemory {
		m.Database = database.NewDemo()
	} else {
		m.Database = database.NewCassandra()
	}

	// create connection manager
	connectionManager := connectionManager.NewConnectionManager(m.Database)
	m.ConnectionManager = connectionManager

	// create DidResolver
	m.DidResolver = NewDidResolver()

	if config.CurrentConfiguration.Database.InMemory {
		m.SecretsResolver = secretsresolver.NewDemo()
	} else {
		m.SecretsResolver = secretsresolver.NewCassandra()
	}

	// create peer did of mediator
	m.createDidIfNeeded()

	dcomm := didcomm.NewDidComm(m.DidResolver, m.SecretsResolver)
	m.Messages = dcomm
	m.Logger = logger
	return &m
}

func (m *Mediator) createDidIfNeeded() {

	// check
	peerDid, err := m.Database.GetMediatorDid()

	if err != nil {
		config.Logger.Error("Unable to get mediator did", "msg", err)
	}
	if peerDid == "" {
		services, err := m.CreateMediatorService()
		if err != nil {
			config.Logger.Error("Unable to create mediator service", "msg", err)
			panic("Mediator can not be used without a service")
		}
		// Create DID for mediator
		peerDid, err = NumAlgo2(services, *&m.SecretsResolver, m.DidResolver)
		if err != nil {
			config.Logger.Error("Unable to create mediator DID", "msg", err)
			panic("Mediator can not be used without a DID")
		}
		err = m.Database.StoreMediatorDid(peerDid)
		if err != nil {
			config.Logger.Error("Unable to store mediator DID", "msg", err)
			panic("Mediator can not be used without ayDID")
		}
	}
	m.Did = peerDid

	config.Logger.Info(fmt.Sprintf("Mediator Peer DID: %s", peerDid))

}

func (m *Mediator) CreateMediatorService() (service []didcomm.Service, err error) {

	s, err := CreateServiceEntry()

	if err != nil {
		config.Logger.Error("can not join url path", err)
		return nil, err
	}

	response := []didcomm.Service{
		s,
	}

	return response, nil
}

func CreateServiceEntry() (didcomm.Service, error) {
	queryUrl, err := url.JoinPath(config.CurrentConfiguration.Url, "/message/receive")

	if err != nil {
		return didcomm.Service{}, err
	}

	serviceKind := didcomm.ServiceKindDidCommMessaging{
		Value: didcomm.DidCommMessagingService{
			Uri:         queryUrl,
			Accept:      &[]string{"didcomm/v2"},
			RoutingKeys: []string{},
		},
	}

	service := didcomm.Service{

		// Service ID
		Id:              "#service-1",
		ServiceEndpoint: serviceKind,
	}

	return service, nil
}
