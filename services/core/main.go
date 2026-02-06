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
		configs := api.Group("/configs")
		{
			configs.GET("/cameras", handlers.HandleGetCameras)
			configs.GET("/cameras/:id", handlers.HandleGetConfigByID)
			configs.POST("/cameras", handlers.HandleCreateCamera)
			configs.PUT("/cameras/:id", handlers.HandleUpdateCamera)
			configs.DELETE("/cameras/:id", handlers.HandleDeleteCamera)

			configs.GET("/scales", handlers.HandleGetScales)
			configs.POST("/scales", handlers.HandleCreateScale)
			configs.PUT("/scales/:id", handlers.HandleUpdateScale)
			configs.DELETE("/scales/:id", handlers.HandleDeleteScale)

			configs.GET("/settings", handlers.HandleListSettings)
			configs.POST("/settings", handlers.HandleUpdateSetting)

			configs.GET("/excluded-plates", handlers.HandleListExcludedPlates)
			configs.POST("/excluded-plates", handlers.HandleCreateExcludedPlate)
			configs.DELETE("/excluded-plates/:id", handlers.HandleDeleteExcludedPlate)

		}

		dataGroup := api.Group("/data")
		{
			dataGroup.GET("/posts", data.HandleListPosts)
			dataGroup.POST("/posts", data.HandleCreatePost)
			dataGroup.PUT("/posts/:id", data.HandleUpdatePost)
			dataGroup.DELETE("/posts/:id", data.HandleDeletePost)

			dataGroup.GET("/modes", data.HandleListModes)
			dataGroup.POST("/modes", data.HandleCreateMode)
			dataGroup.PUT("/modes/:id", data.HandleUpdateMode)
			dataGroup.DELETE("/modes/:id", data.HandleDeleteMode)

			dataGroup.GET("/vehicle-types", data.HandleListVehicleTypes)
			dataGroup.POST("/vehicle-types", data.HandleCreateVehicleType)
			dataGroup.PUT("/vehicle-types/:id", data.HandleUpdateVehicleType)
			dataGroup.DELETE("/vehicle-types/:id", data.HandleDeleteVehicleType)

			dataGroup.GET("/payment-types", data.HandleListPaymentTypes)
			dataGroup.POST("/payment-types", data.HandleCreatePaymentType)
			dataGroup.PUT("/payment-types/:id", data.HandleUpdatePaymentType)
			dataGroup.DELETE("/payment-types/:id", data.HandleDeletePaymentType)

			dataGroup.GET("/companies", data.HandleListCompanies)
			dataGroup.POST("/companies", data.HandleCreateCompany)
			dataGroup.PUT("/companies/:id", data.HandleUpdateCompany)
			dataGroup.DELETE("/companies/:id", data.HandleDeleteCompany)
		}

		events := api.Group("/events")
		{
			events.GET("/plate", handlers.HandleGetPlateEvents)
			events.GET("/plate/:id", handlers.HandleGetPlateEventByID)
			events.POST("/plate",
				middleware.SystemEventLogger("plate"),
				handlers.HandlePlateEvent,
			)

			events.GET("/weight", handlers.HandleGetWeightEvents)
			events.GET("/weight/:id", handlers.HandleGetWeightEventByID)
			events.POST("/weight",
				middleware.SystemEventLogger("weight"),
				handlers.HandleWeightEvent,
			)

			events.PATCH("/plate/:id", handlers.HandlePatchPlateEvent)

			events.GET("/system", handlers.HandleGetSystemEvents)
			events.GET("/system/:id", handlers.HandleGetSystemEventByID)
		}

		permits := api.Group("/permits")
		{
			permits.GET("", handlers.HandleGetPermits)
			permits.GET("/:id", handlers.HandleGetPermitByID)
			permits.POST("", handlers.HandleCreatePermit)
			permits.PUT("/:id", handlers.HandleUpdatePermit)
		}

		users := api.Group("/users")
		{
			users.GET("", handlers.HandleListUsers)
			users.GET("/me", handlers.HandleGetMyProfile)
			users.PUT("/me", handlers.HandleUpdateMyProfile)
			users.GET("/:id", handlers.HandleGetUser)
			users.DELETE("/:id", handlers.HandleDeleteUser)
			users.GET("/by-auth-id/:authId", handlers.HandleGetUserByAuthID)
			users.POST("/", handlers.HandleCreateUser)
			users.PUT("/:id", handlers.HandleUpdateUser)
		}
	}

	repository.DB.FirstOrCreate(&models.SystemSetting{Key: "match_window_seconds", Value: "120"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
