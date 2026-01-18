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

	Cameras []CameraConfig `gorm:"foreignKey:PresetID" json:"cameras,omitempty"`
}

type CameraConfig struct {
	gorm.Model
	SourceID    string `gorm:"column:camera_id;uniqueIndex;not null" json:"camera_id"`
	Name        string `json:"name"`
	Description string `json:"description"`

	PresetID *uint         `json:"preset_id"`
	Preset   *CameraPreset `gorm:"foreignKey:PresetID" json:"preset,omitempty"`

	GateID *uint `json:"gate_id"`
	Gate   *Gate `gorm:"foreignKey:GateID" json:"gate,omitempty"`

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
	Gate   *Gate `gorm:"foreignKey:GateID" json:"gate,omitempty"`
}

type Gate struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	IsEntry     bool   `json:"is_entry" gorm:"default:false"`
	IsExit      bool   `json:"is_exit" gorm:"default:false"`

	Cameras  []CameraConfig `gorm:"foreignKey:GateID" json:"cameras,omitempty"`
	Scales   []ScaleConfig  `gorm:"foreignKey:GateID" json:"scales,omitempty"`
	FlowStep *FlowStep      `gorm:"foreignKey:GateID" json:"flow_step,omitempty"`
}

type Flow struct {
	gorm.Model
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []FlowStep `gorm:"foreignKey:FlowID" json:"steps"`
}

type FlowStep struct {
	gorm.Model
	FlowID   uint `json:"flow_id"`
	GateID   uint `json:"gate_id"`
	Sequence int  `json:"sequence"`

	Gate *Gate `gorm:"foreignKey:GateID" json:"gate,omitempty"`
	Flow *Flow `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
}

type Permit struct {
	gorm.Model
	FlowID              *uint      `json:"flow_id"`
	Flow                *Flow      `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	PlateFront          string     `json:"plate_front"`
	PlateBack           string     `json:"plate_back"`
	TotalWeight         float64    `json:"total_weight"`
	EntryTime           time.Time  `json:"entry_time"`
	ExitTime            *time.Time `json:"exit_time"`
	LastActivityAt      time.Time  `json:"last_activity_at"`
	IsClosed            bool       `gorm:"default:false" json:"is_closed"`
	CurrentStepSequence int        `json:"current_step_sequence"`

	GateEvents []GateEvent `gorm:"foreignKey:PermitID" json:"gate_events"`
}

type GateEvent struct {
	gorm.Model
	GateID uint `json:"gate_id"`
	Gate   Gate `gorm:"foreignKey:GateID" json:"gate,omitempty"`

	PermitID  *uint     `json:"permit_id"`
	Timestamp time.Time `json:"timestamp"`

	PlateEvents  []RawPlateEvent  `gorm:"foreignKey:GateEventID" json:"plate_events,omitempty"`
	WeightEvents []RawWeightEvent `gorm:"foreignKey:GateEventID" json:"weight_events,omitempty"`
}

type SystemEvent struct {
	gorm.Model
	Type      string    `json:"type"`
	SourceID  string    `json:"source_id"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`

	PlateEvent  *RawPlateEvent  `gorm:"foreignKey:SystemEventID" json:"plate_event,omitempty"`
	WeightEvent *RawWeightEvent `gorm:"foreignKey:SystemEventID" json:"weight_event,omitempty"`
}

type RawPlateEvent struct {
	gorm.Model
	CameraSourceID   string    `gorm:"column:camera_source_id" json:"camera_source_id"`
	CameraSourceName string    `gorm:"column:camera_source_name" json:"camera_source_name"`
	CameraID         string    `gorm:"column:camera_id" json:"camera_id"`
	Plate            string    `json:"plate"`
	PlateCorrected   string    `json:"plate_corrected"`
	CorrectedBy      string    `json:"corrected_by"`
	IsManual         bool      `gorm:"default:false" json:"is_manual"`
	ImageKey         string    `json:"image_key"`
	Timestamp        time.Time `json:"timestamp"`
	Suggestions      string    `gorm:"type:jsonb" json:"suggestions"`

	Camera CameraConfig `gorm:"references:CameraID" json:"-"`

	SystemEventID uint         `json:"system_event_id"`
	SystemEvent   *SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	GateEventID *uint      `json:"gate_event_id"`
	GateEvent   *GateEvent `gorm:"foreignKey:GateEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type RawWeightEvent struct {
	gorm.Model
	ScaleSourceID string    `json:"scale_source_id"`
	ScaleID       string    `gorm:"column:scale_id" json:"scale_id"`
	Weight        float64   `json:"weight"`
	Timestamp     time.Time `json:"timestamp"`

	Scale ScaleConfig `gorm:"references:ScaleID" json:"-"`

	SystemEventID uint         `json:"system_event_id"`
	SystemEvent   *SystemEvent `gorm:"foreignKey:SystemEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`

	GateEventID *uint      `json:"gate_event_id"`
	GateEvent   *GateEvent `gorm:"foreignKey:GateEventID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

type SystemSetting struct {
	gorm.Model
	Key   string `gorm:"uniqueIndex;not null" json:"key"`
	Value string `json:"value"`
}

type ExcludedPlate struct {
	gorm.Model
	Plate   string `gorm:"uniqueIndex;not null" json:"plate"`
	Comment string `json:"comment"`
}
