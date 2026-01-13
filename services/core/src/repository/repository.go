package repository

import (
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
		&models.SystemSetting{},
		&models.Gate{},
		&models.ExcludedPlate{},
		&models.Permit{},
	)
	DB = db

}
