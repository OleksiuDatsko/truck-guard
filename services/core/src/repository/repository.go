package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/truckguard/core/src/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dsn string) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to Core Database")
	}

	db.AutoMigrate(
		&models.SystemEvent{},
		&models.RawPlateEvent{},
		&models.RawWeightEvent{},
		&models.CameraConfig{},
		&models.ScaleConfig{},
		&models.CameraPreset{},
		&models.Gate{},
		&models.Flow{},
		&models.FlowStep{},
		&models.SystemSetting{},
		&models.ExcludedPlate{},
		&models.Permit{},
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
