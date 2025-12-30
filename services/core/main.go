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

	r.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/v1")
	{
		api.GET("/cameras/:id", handlers.HandleGetConfigByID)
		api.GET("/cameras/by-id/:camera_id", handlers.HandleGetConfigByCameraID)
		configs := api.Group("/configs", middleware.RequireCorePermission("manage:configs"))
		{
			configs.GET("/presets", middleware.RequireCorePermission("read:presets"), handlers.HandleListPresets)
			configs.GET("/presets/:id", middleware.RequireCorePermission("read:presets"), handlers.HandleGetPreset)
			configs.POST("/presets", middleware.RequireCorePermission("create:presets"), handlers.HandleCreatePreset)
			configs.PUT("/presets/:id", middleware.RequireCorePermission("update:presets"), handlers.HandleUpdatePreset)
			configs.DELETE("/presets/:id", middleware.RequireCorePermission("delete:presets"), handlers.HandleDeletePreset)

			configs.GET("/cameras", middleware.RequireCorePermission("read:cameras"), handlers.HandleGetCameras)
			configs.POST("/cameras",
				middleware.RequireCorePermission("create:cameras"),
				middleware.RequireCorePermission("create:keys"),
				handlers.HandleCreateCamera,
			)
			configs.PUT("/cameras/:id", middleware.RequireCorePermission("update:cameras"), handlers.HandleUpdateCamera)
			configs.DELETE("/cameras/:id", middleware.RequireCorePermission("delete:cameras"), handlers.HandleDeleteCamera)

		}

		events := api.Group("/events")
		{
			events.POST("/plate",
				middleware.RequireCorePermission("create:events"),
				middleware.SystemEventLogger("plate"),
				handlers.HandlePlateEvent,
			)
			events.POST("/weight",
				middleware.RequireCorePermission("create:events"),
				middleware.SystemEventLogger("weight"),
				handlers.HandleWeightEvent,
			)
			events.PUT("/plate/:id",
				middleware.RequireCorePermission("update:events"),
				handlers.HandleUpdatePlateEvent,
			)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
