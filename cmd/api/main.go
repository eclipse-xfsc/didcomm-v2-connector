package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse-xfsc/didcomm-v2-connector/cmd/api/database"
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"
	"github.com/eclipse-xfsc/didcomm-v2-connector/mediator"
	"github.com/eclipse-xfsc/didcomm-v2-connector/protocol"
)

//	@title			DIDComm Connector API
//	@version		1.0
//	@description	The DIDCommConnector can be used as a Mediator and Connection Management Service by parties who want to set up trust with another party. The DIDCommConnector uses DIDComm v2 and provides a message layer and a management component for the following two use cases: - Pairing a cloud solution with a smartphone / app solution - DIDComm v2 based message protocols
//	@description	- Pairing a cloud solution with a smartphone / app solution - DIDComm v2 based message protocols
//	@description	- DIDComm v2 based message protocols
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.change.this/url
//	@contact.email	email@todo.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9090
//	@BasePath	/

//	@securityDefinitions.basic	BasicAuth

//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Description for what is this security definition being used

//	@securitydefinitions.oauth2.application	OAuth2Application
//	@tokenUrl								https://example.com/oauth/token
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.implicit	OAuth2Implicit
//	@authorizationUrl						https://example.com/oauth/authorize
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.password	OAuth2Password
//	@tokenUrl								https://example.com/oauth/token
//	@scope.read								Grants read access
//	@scope.write							Grants write access
//	@scope.admin							Grants read and write access to administrative information

//	@securitydefinitions.oauth2.accessCode	OAuth2AccessCode
//	@tokenUrl								https://example.com/oauth/token
//	@authorizationUrl						https://example.com/oauth/authorize
//	@scope.admin							Grants read and write access to administrative information

type application struct {
	mediator *mediator.Mediator
}

func main() {

	err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	database.NewMigration()
	app := application{
		mediator: mediator.NewMediator(config.Logger),
	}

	if config.IsForwardTypeNats() || config.IsForwardTypeHybrid() {
		// subscribe to nats
		go protocol.ReceiveMessage(app.mediator)
	}

	router := app.NewRouter()
	srv := &http.Server{
		Addr:    ":" + fmt.Sprint(config.CurrentConfiguration.Port),
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			config.Logger.Error("ListenAndServe", "Error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	config.Logger.Info("Server Started")
	<-quit
	config.Logger.Info("Shutting down server in 5 seconds...")
	app.mediator.Database.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error("Server Shutdown:", "msg", err)
	}
	select {
	case <-ctx.Done():
		config.Logger.Info("timeout of 5 seconds.")
	default:
		config.Logger.Info("Server shutdown completed")
	}
	config.Logger.Info("Server exiting")
	config.CurrentConfiguration.LoggerFile.Close()
}
