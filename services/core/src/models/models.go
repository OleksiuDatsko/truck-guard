package models

import (
	"gorm.io/gorm"
	"time"
)

type CameraPreset struct {
	gorm.Model
	Name         string `json:"name"`
	Format       string `json:"format"`
	RunANPR      bool   `json:"run_anpr"`
	FieldMapping string `json:"field_mapping"`
}

type CameraConfig struct {
	gorm.Model
	CameraID    string `gorm:"uniqueIndex;not null" json:"camera_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	PresetID *uint         `json:"preset_id"`
	Preset   *CameraPreset `json:"preset,omitempty"`

	Format       string `json:"format"`
	RunANPR      *bool  `json:"run_anpr"`
	FieldMapping string `json:"field_mapping"`
}

type SystemEvent struct {
	gorm.Model
	Type      string    `json:"type"`
	SourceID  string    `json:"source_id"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type RawPlateEvent struct {
	gorm.Model
	CameraID      string      `json:"camera_id"`
	CameraName    string      `json:"camera_name"`
	Plate         string      `json:"plate"`
	ImageKey      string      `json:"image_key"`
	Timestamp     time.Time   `json:"timestamp"`
	SystemEventID uint        `json:"system_event_id"`
	SystemEvent   SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type RawWeightEvent struct {
	gorm.Model
	ScaleID       string      `json:"scale_id"`
	Weight        float64     `json:"weight"`
	Timestamp     time.Time   `json:"timestamp"`
	SystemEventID uint        `json:"system_event_id"`
	SystemEvent   SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
