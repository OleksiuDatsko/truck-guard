package main

import (
	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/handlers"
	"github.com/truckguard/core/src/api/middleware"
	"github.com/truckguard/core/src/repository"
	"os"
)

func main() {
	repository.InitDB(os.Getenv("DATABASE_URL"))

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/v1")
	api.POST("/events/plate",
		middleware.RequireCorePermission("create:events"),
		middleware.SystemEventLogger("plate"),
		handlers.HandlePlateEvent,
	)
	api.POST("/events/weight",
		middleware.RequireCorePermission("create:events"),
		middleware.SystemEventLogger("weight"),
		handlers.HandleWeightEvent,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
