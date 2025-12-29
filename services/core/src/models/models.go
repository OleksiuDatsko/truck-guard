package models

import (
	"gorm.io/gorm"
	"time"
)

type SystemEvent struct {
	gorm.Model
	Type      string    `json:"type"`
	SourceID  string    `json:"source_id"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type RawPlateEvent struct {
	gorm.Model
	CameraID   string    `json:"camera_id"`
	CameraName string    `json:"camera_name"`
	Plate      string    `json:"plate"`
	ImageKey   string    `json:"image_key"`
	Timestamp  time.Time `json:"timestamp"`
}

type RawWeightEvent struct {
	gorm.Model
	ScaleID   string    `json:"scale_id"`
	Weight    float64   `json:"weight"`
	Timestamp time.Time `json:"timestamp"`
}
