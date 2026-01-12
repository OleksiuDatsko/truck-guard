package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/handlers"
	"github.com/truckguard/core/src/api/middleware"
	"github.com/truckguard/core/src/repository"
)

func main() {
	repository.InitDB(os.Getenv("DATABASE_URL"))

	r := gin.Default()

	r.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/")
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
			events.GET("/plate", middleware.RequireCorePermission("read:events"), handlers.HandleGetPlateEvents)
			events.GET("/plate/:id", middleware.RequireCorePermission("read:events"), handlers.HandleGetPlateEventByID)
			events.POST("/plate",
				middleware.RequireCorePermission("create:events"),
				middleware.SystemEventLogger("plate"),
				handlers.HandlePlateEvent,
			)

			events.GET("/weight", middleware.RequireCorePermission("read:events"), handlers.HandleGetWeightEvents)
			events.GET("/weight/:id", middleware.RequireCorePermission("read:events"), handlers.HandleGetWeightEventByID)
			events.POST("/weight",
				middleware.RequireCorePermission("create:events"),
				middleware.SystemEventLogger("weight"),
				handlers.HandleWeightEvent,
			)

			events.PATCH("/plate/:id",
				middleware.RequireCorePermission("update:events"),
				handlers.HandlePatchPlateEvent,
			)

			events.GET("/system", middleware.RequireCorePermission("read:events"), handlers.HandleGetSystemEvents)
			events.GET("/system/:id", middleware.RequireCorePermission("read:events"), handlers.HandleGetSystemEventByID)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
