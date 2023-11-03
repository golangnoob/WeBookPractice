package main

import (
	"github.com/gin-gonic/gin"

	"webooktrial/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
