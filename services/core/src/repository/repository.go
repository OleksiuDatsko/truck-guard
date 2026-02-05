package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/truckguard/core/src/models"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to Core Database")
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		panic(err)
	}

	db.AutoMigrate(
		&models.SystemEvent{},
		&models.PlateEvent{},
		&models.WeightEvent{},
		&models.CameraConfig{},
		&models.ScaleConfig{},
		&models.SystemSetting{},
		&models.ExcludedPlate{},
		&models.Permit{},
		&models.PermitCustomsData{},
		&models.PermitAudit{},
		&models.User{},
	)
	DB = db
}

var RDB *redis.Client

func InitRedis(addr string) {
	RDB = redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if _, err := RDB.Ping(context.Background()).Result(); err != nil {
		panic("Failed to connect to Redis")
	}
}
