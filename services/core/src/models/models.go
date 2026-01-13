package models

import (
	"time"

	"gorm.io/gorm"
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
	SourceID    string `gorm:"column:camera_id;uniqueIndex;not null" json:"camera_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	PresetID *uint         `json:"preset_id"`
	Preset   *CameraPreset `json:"preset,omitempty"`

	GateID *uint `json:"gate_id"`
	Gate   *Gate `json:"gate,omitempty"`

	Format       string `json:"format"`
	RunANPR      *bool  `json:"run_anpr"`
	FieldMapping string `json:"field_mapping"`
}

type ScaleConfig struct {
	gorm.Model
	SourceID    string `gorm:"column:scale_id;uniqueIndex;not null" json:"scale_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Format       string `json:"format"`
	FieldMapping string `json:"field_mapping"`

	GateID *uint `json:"gate_id"`
	Gate   *Gate `json:"gate,omitempty"`
}

type SystemSetting struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null" json:"key"`
	Value string `json:"value"`
}

type Gate struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExcludedPlate struct {
	gorm.Model
	Plate   string `gorm:"uniqueIndex;not null" json:"plate"`
	Comment string `json:"comment"`
}

type Permit struct {
	gorm.Model
	GateID       uint             `json:"gate_id"`
	Gate         Gate             `gorm:"foreignKey:GateID" json:"gate"`
	PlateFront   string           `json:"plate_front"`
	PlateBack    string           `json:"plate_back"`
	TotalWeight  float64          `json:"total_weight"`
	EntryTime    time.Time        `json:"entry_time"`
	IsClosed     bool             `gorm:"default:false" json:"is_closed"`
	PlateEvents  []RawPlateEvent  `gorm:"many2many:permit_plate_events;" json:"plate_events,omitempty"`
	WeightEvents []RawWeightEvent `gorm:"many2many:permit_weight_events;" json:"weight_events,omitempty"`
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
	SourceID       string      `gorm:"column:camera_id" json:"camera_id"`
	SourceName     string      `gorm:"column:camera_name" json:"camera_name"`
	Plate          string      `json:"plate"`
	PlateCorrected string      `json:"plate_corrected"`
	CorrectedBy    string      `json:"corrected_by"`
	IsManual       bool        `gorm:"default:false" json:"is_manual"`
	ImageKey       string      `json:"image_key"`
	Timestamp      time.Time   `json:"timestamp"`
	Suggestions    string      `gorm:"type:jsonb" json:"suggestions"`
	SystemEventID  uint        `json:"system_event_id"`
	SystemEvent    SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type RawWeightEvent struct {
	gorm.Model
	ScaleID       string      `json:"scale_id"`
	Weight        float64     `json:"weight"`
	Timestamp     time.Time   `json:"timestamp"`
	SystemEventID uint        `json:"system_event_id"`
	SystemEvent   SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}
