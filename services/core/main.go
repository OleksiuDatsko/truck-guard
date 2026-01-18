package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/handlers"
	"github.com/truckguard/core/src/api/middleware"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func main() {
	repository.InitDB(os.Getenv("DATABASE_URL"))
	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	r := gin.Default()

	r.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/")
	{
		api.GET("/cameras/by-id/:camera_id", handlers.HandleGetConfigByCameraID)
		api.GET("/scales/by-id/:scale_id", handlers.HandleGetConfigByScaleID)
		configs := api.Group("/configs", middleware.RequireCorePermission("manage:configs"))
		{
			configs.GET("/presets", middleware.RequireCorePermission("read:presets"), handlers.HandleListPresets)
			configs.GET("/presets/:id", middleware.RequireCorePermission("read:presets"), handlers.HandleGetPreset)
			configs.POST("/presets", middleware.RequireCorePermission("create:presets"), handlers.HandleCreatePreset)
			configs.PUT("/presets/:id", middleware.RequireCorePermission("update:presets"), handlers.HandleUpdatePreset)
			configs.DELETE("/presets/:id", middleware.RequireCorePermission("delete:presets"), handlers.HandleDeletePreset)

			configs.GET("/cameras", middleware.RequireCorePermission("read:cameras"), handlers.HandleGetCameras)
			configs.GET("/cameras/:id", middleware.RequireCorePermission("read:cameras"), handlers.HandleGetConfigByID)
			configs.POST("/cameras",
				middleware.RequireCorePermission("create:cameras"),
				middleware.RequireCorePermission("create:keys"),
				handlers.HandleCreateCamera,
			)
			configs.PUT("/cameras/:id", middleware.RequireCorePermission("update:cameras"), handlers.HandleUpdateCamera)
			configs.DELETE("/cameras/:id", middleware.RequireCorePermission("delete:cameras"), handlers.HandleDeleteCamera)

			configs.GET("/scales", middleware.RequireCorePermission("read:scales"), handlers.HandleGetScales)
			configs.POST("/scales",
				middleware.RequireCorePermission("create:scales"),
				middleware.RequireCorePermission("create:keys"),
				handlers.HandleCreateScale,
			)
			configs.PUT("/scales/:id", middleware.RequireCorePermission("update:scales"), handlers.HandleUpdateScale)
			configs.DELETE("/scales/:id", middleware.RequireCorePermission("delete:scales"), handlers.HandleDeleteScale)

			configs.GET("/gates", middleware.RequireCorePermission("read:gates"), handlers.HandleGetGates)
			configs.GET("/gates/:id", middleware.RequireCorePermission("read:gates"), handlers.HandleGetGateByID)
			configs.POST("/gates", middleware.RequireCorePermission("create:gates"), handlers.HandleCreateGate)
			configs.PUT("/gates/:id", middleware.RequireCorePermission("update:gates"), handlers.HandleUpdateGate)
			configs.DELETE("/gates/:id", middleware.RequireCorePermission("delete:gates"), handlers.HandleDeleteGate)

			configs.GET("/settings", middleware.RequireCorePermission("read:settings"), handlers.HandleListSettings)
			configs.POST("/settings", middleware.RequireCorePermission("update:settings"), handlers.HandleUpdateSetting)

			configs.GET("/excluded-plates", middleware.RequireCorePermission("read:excluded_plates"), handlers.HandleListExcludedPlates)
			configs.POST("/excluded-plates", middleware.RequireCorePermission("create:excluded_plates"), handlers.HandleCreateExcludedPlate)
			configs.DELETE("/excluded-plates/:id", middleware.RequireCorePermission("delete:excluded_plates"), handlers.HandleDeleteExcludedPlate)

			configs.GET("/flows", middleware.RequireCorePermission("read:flows"), handlers.HandleListFlows)
			configs.GET("/flows/:id", middleware.RequireCorePermission("read:flows"), handlers.HandleGetFlow)
			configs.POST("/flows", middleware.RequireCorePermission("create:flows"), handlers.HandleCreateFlow)
			configs.PUT("/flows/:id", middleware.RequireCorePermission("update:flows"), handlers.HandleUpdateFlow)
			configs.DELETE("/flows/:id", middleware.RequireCorePermission("delete:flows"), handlers.HandleDeleteFlow)

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

			events.GET("/gate", middleware.RequireCorePermission("read:events"), handlers.HandleGetGateEvents)
			events.GET("/gate/:id", middleware.RequireCorePermission("read:events"), handlers.HandleGetGateEventByID)
		}

		permits := api.Group("/permits")
		{
			permits.GET("/", middleware.RequireCorePermission("read:permits"), handlers.HandleGetPermits)
			permits.GET("/:id", middleware.RequireCorePermission("read:permits"), handlers.HandleGetPermitByID)
		}
	
		users := api.Group("/users")
		{
			users.GET("/", middleware.RequireCorePermission("read:users"), handlers.HandleListUsers)
			users.GET("/:id", middleware.RequireCorePermission("read:users"), handlers.HandleGetUser)
			users.POST("/", middleware.RequireCorePermission("create:users"), handlers.HandleCreateUser)
			users.PUT("/:id", middleware.RequireCorePermission("update:users"), handlers.HandleUpdateUser)
			users.DELETE("/:id", middleware.RequireCorePermission("delete:users"), handlers.HandleDeleteUser)
		}
	}

	repository.DB.FirstOrCreate(&models.SystemSetting{Key: "match_window_seconds", Value: "120"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
