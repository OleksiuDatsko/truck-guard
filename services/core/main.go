package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/api/handlers"
	"github.com/truckguard/core/src/api/handlers/data"
	"github.com/truckguard/core/src/api/middleware"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/pkg/telemetry"
	"github.com/truckguard/core/src/repository"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func main() {
	logger := telemetry.NewLogger("truckguard-core")
	slog.SetDefault(logger)
	if err := telemetry.Init("truckguard-core"); err != nil {
		logger.Error("otel init failed", "error", err)
		os.Exit(1)
	}
	defer telemetry.Shutdown(context.Background())

	repository.InitDB(os.Getenv("DATABASE_URL"))
	repository.InitRedis(os.Getenv("REDIS_ADDR"))

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("truckguard-core"))
	r.Use(middleware.Logger())
	r.Use(middleware.MetricsMiddleware())

	// Health check with service_up gauge
	meter := otel.Meter("truckguard-core")
	serviceUpGauge, _ := meter.Int64ObservableGauge("service_up",
		metric.WithDescription("Service health status (1 for up)"),
	)
	_, _ = meter.RegisterCallback(func(ctx context.Context, observer metric.Observer) error {
		observer.ObserveInt64(serviceUpGauge, 1)
		return nil
	}, serviceUpGauge)

	r.Match([]string{"GET", "HEAD"}, "/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/")
	{
		api.GET("/cameras/by-id/:camera_id", handlers.HandleGetConfigByCameraID)
		api.GET("/scales/by-id/:scale_id", handlers.HandleGetConfigByScaleID)
		configs := api.Group("/configs", middleware.RequireCorePermission("manage:configs"))
		{
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

			configs.GET("/settings", middleware.RequireCorePermission("read:settings"), handlers.HandleListSettings)
			configs.POST("/settings", middleware.RequireCorePermission("update:settings"), handlers.HandleUpdateSetting)

			configs.GET("/excluded-plates", middleware.RequireCorePermission("read:excluded_plates"), handlers.HandleListExcludedPlates)
			configs.POST("/excluded-plates", middleware.RequireCorePermission("create:excluded_plates"), handlers.HandleCreateExcludedPlate)
			configs.DELETE("/excluded-plates/:id", middleware.RequireCorePermission("delete:excluded_plates"), handlers.HandleDeleteExcludedPlate)

		}

		dataGroup := api.Group("/data")
		{
			dataGroup.GET("/posts", middleware.RequireCorePermission("read:settings"), data.HandleListPosts)
			dataGroup.POST("/posts", middleware.RequireCorePermission("manage:settings"), data.HandleCreatePost)
			dataGroup.PUT("/posts/:id", middleware.RequireCorePermission("manage:settings"), data.HandleUpdatePost)
			dataGroup.DELETE("/posts/:id", middleware.RequireCorePermission("manage:settings"), data.HandleDeletePost)

			dataGroup.GET("/modes", middleware.RequireCorePermission("read:settings"), data.HandleListModes)
			dataGroup.POST("/modes", middleware.RequireCorePermission("manage:settings"), data.HandleCreateMode)
			dataGroup.PUT("/modes/:id", middleware.RequireCorePermission("manage:settings"), data.HandleUpdateMode)
			dataGroup.DELETE("/modes/:id", middleware.RequireCorePermission("manage:settings"), data.HandleDeleteMode)

			dataGroup.GET("/vehicle-types", middleware.RequireCorePermission("read:settings"), data.HandleListVehicleTypes)
			dataGroup.POST("/vehicle-types", middleware.RequireCorePermission("manage:settings"), data.HandleCreateVehicleType)
			dataGroup.PUT("/vehicle-types/:id", middleware.RequireCorePermission("manage:settings"), data.HandleUpdateVehicleType)
			dataGroup.DELETE("/vehicle-types/:id", middleware.RequireCorePermission("manage:settings"), data.HandleDeleteVehicleType)

			dataGroup.GET("/payment-types", middleware.RequireCorePermission("read:settings"), data.HandleListPaymentTypes)
			dataGroup.POST("/payment-types", middleware.RequireCorePermission("manage:settings"), data.HandleCreatePaymentType)
			dataGroup.PUT("/payment-types/:id", middleware.RequireCorePermission("manage:settings"), data.HandleUpdatePaymentType)
			dataGroup.DELETE("/payment-types/:id", middleware.RequireCorePermission("manage:settings"), data.HandleDeletePaymentType)

			dataGroup.GET("/companies", middleware.RequireCorePermission("read:settings"), data.HandleListCompanies)
			dataGroup.POST("/companies", middleware.RequireCorePermission("manage:settings"), data.HandleCreateCompany)
			dataGroup.PUT("/companies/:id", middleware.RequireCorePermission("manage:settings"), data.HandleUpdateCompany)
			dataGroup.DELETE("/companies/:id", middleware.RequireCorePermission("manage:settings"), data.HandleDeleteCompany)
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

		permits := api.Group("/permits")
		{
			permits.GET("", middleware.RequireCorePermission("read:permits"), handlers.HandleGetPermits)
			permits.GET("/:id", middleware.RequireCorePermission("read:permits"), handlers.HandleGetPermitByID)
			permits.POST("", middleware.RequireCorePermission("create:permits"), handlers.HandleCreatePermit)
			permits.PUT("/:id", middleware.RequireCorePermission("update:permits"), handlers.HandleUpdatePermit)
		}

		users := api.Group("/users")
		{
			users.GET("", middleware.RequireCorePermission("read:users"), handlers.HandleListUsers)
			users.GET("/me", handlers.HandleGetMyProfile)
			users.PUT("/me", handlers.HandleUpdateMyProfile)
			users.GET("/:id", middleware.RequireCorePermission("read:users"), handlers.HandleGetUser)
			users.DELETE("/:id", middleware.RequireCorePermission("delete:users"), handlers.HandleDeleteUser)
			users.GET("/by-auth-id/:authId", handlers.HandleGetUserByAuthID)
			users.POST("/", middleware.RequireCorePermission("create:users"), handlers.HandleCreateUser)
			users.PUT("/:id", middleware.RequireCorePermission("update:users"), handlers.HandleUpdateUser)
		}
	}

	repository.DB.FirstOrCreate(&models.SystemSetting{Key: "match_window_seconds", Value: "120"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
