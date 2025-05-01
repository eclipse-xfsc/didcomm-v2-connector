package main

import (
	"github.com/eclipse-xfsc/didcomm-v2-connector/internal/config"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (app *application) NewRouter() *gin.Engine {

	if config.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(sloggin.New(config.Logger))

	// Connections (Mediatees)
	adminGroup := router.Group("admin")
	connectionsGroup := adminGroup.Group("connections")
	connectionsGroup.GET("", app.GetConnections)
	connectionsGroup.GET(":did", app.GetConnection)
	connectionsGroup.PUT(":did", app.UpdateConnection)
	connectionsGroup.DELETE(":did", app.DeleteConnection)
	// Block Connections (Mediatees)
	connectionsGroup.POST("block/:did", app.BlockConnection)
	connectionsGroup.POST("unblock/:did", app.UnblockConnection)
	connectionsGroup.GET("isblocked/:did", app.IsBlocked)
	connectionsGroup.POST("accept", app.AcceptConnection)

	adminGroup.POST("invitation", app.InvitationMessage)

	// messages
	messagesGroup := router.Group("message")
	messagesGroup.POST("receive", app.ReceiveMessage)

	// healthcheck
	router.GET("health", app.HealthCheck)

	// swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
