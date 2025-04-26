package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (app *application) HealthCheck(context *gin.Context) {

	context.Status(http.StatusOK)
}
