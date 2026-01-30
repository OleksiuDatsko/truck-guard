package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CustomsPost struct {
	gorm.Model
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default" gorm:"default:false"`

	Gates []Gate `gorm:"foreignKey:CustomsPostID" json:"gates,omitempty"`
}

type CustomsMode struct {
	gorm.Model
	Name        string `json:"name"`
	Code        string `gorm:"uniqueIndex;not null" json:"code"`
	Description string `json:"description"`
}

type Company struct {
	gorm.Model
	Name         string         `json:"name"`
	EDRPOU       string         `gorm:"uniqueIndex;not null" json:"edrpou"`
	Details      datatypes.JSON `gorm:"type:jsonb" json:"details"`
	LastSyncedAt *time.Time     `json:"last_synced_at"`
}

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
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsEntry       bool   `json:"is_entry" gorm:"default:false"`
	IsExit        bool   `json:"is_exit" gorm:"default:false"`
	CustomsPostID *uint  `json:"customs_post_id"`

	CustomsPost *CustomsPost   `gorm:"foreignKey:CustomsPostID" json:"customs_post,omitempty"`
	Cameras     []CameraConfig `gorm:"foreignKey:GateID" json:"cameras,omitempty"`
	Scales      []ScaleConfig  `gorm:"foreignKey:GateID" json:"scales,omitempty"`
	FlowStep    *FlowStep      `gorm:"foreignKey:GateID" json:"flow_step,omitempty"`
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

type VehicleType struct {
	gorm.Model
	Name        string  `json:"name"`
	Code        string  `gorm:"uniqueIndex;not null" json:"code"`
	Description string  `json:"description"`
	EntryPrice  float64 `json:"entry_price"`
	DailyPrice  float64 `json:"daily_price"`
	Color       string  `json:"color"`
}

type PaymentType struct {
	gorm.Model
	Name        string `json:"name"`
	Code        string `gorm:"uniqueIndex;not null" json:"code"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	Icon        string `json:"icon"`
}

type PermitPayer struct {
	gorm.Model
	PermitID  uint `gorm:"uniqueIndex:idx_permit_slot" json:"permit_id"`
	CompanyID uint `json:"company_id"`
	SlotIndex int  `gorm:"uniqueIndex:idx_permit_slot" json:"slot_index"` // 1, 2, 3, 4

	Company *Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

type PermitAudit struct {
	gorm.Model
	PermitID uint           `json:"permit_id"`
	UserID   *uint          `json:"user_id"`
	Action   string         `json:"action"`
	Changes  datatypes.JSON `gorm:"type:jsonb" json:"changes"`
	Comment  string         `json:"comment"`

	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Permit *Permit `gorm:"foreignKey:PermitID" json:"permit,omitempty"`
}

type Permit struct {
	gorm.Model
	// Flow & State
	FlowID              *uint `json:"flow_id"`
	Flow                *Flow `gorm:"foreignKey:FlowID" json:"flow,omitempty"`
	CurrentStepSequence int   `json:"current_step_sequence"`
	IsClosed            bool  `gorm:"default:false" json:"is_closed"`
	IsVoid              bool  `gorm:"default:false" json:"is_void"`

	// Customs & Document Info
	CustomsPostID     *uint        `json:"customs_post_id"`
	CustomsPost       *CustomsPost `gorm:"foreignKey:CustomsPostID" json:"customs_post,omitempty"`
	DeclarationNumber string       `json:"declaration_number"`
	CustomsModeCode   string       `json:"customs_mode_code"`
	CustomsMode       *CustomsMode `gorm:"foreignKey:CustomsModeCode;references:Code" json:"customs_mode,omitempty"`
	Notes             string       `json:"notes"`

	// Vehicle & Classification
	VehicleTypeID *uint        `json:"vehicle_type_id"`
	VehicleType   *VehicleType `gorm:"foreignKey:VehicleTypeID" json:"vehicle_type,omitempty"`
	PlateFront    string       `json:"plate_front"`
	PlateBack     string       `json:"plate_back"`
	TotalWeight   float64      `json:"total_weight"`

	// Financials
	PaymentTypeID *uint        `json:"payment_type_id"`
	PaymentType   *PaymentType `gorm:"foreignKey:PaymentTypeID" json:"payment_type,omitempty"`
	EntryFee      float64      `json:"entry_fee"`
	ExitFee       float64      `json:"exit_fee"`

	// Payers
	Payers []PermitPayer `gorm:"foreignKey:PermitID" json:"payers,omitempty"`

	// Time Tracking
	EntryTime      time.Time  `json:"entry_time"`
	ExitTime       *time.Time `json:"exit_time"`
	LastActivityAt time.Time  `json:"last_activity_at"`

	// Verification
	CreatedBy  *uint      `json:"created_by"`
	Creator    *User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	VerifiedBy *uint      `json:"verified_by"`
	Verifier   *User      `gorm:"foreignKey:VerifiedBy" json:"verifier,omitempty"`
	VerifiedAt *time.Time `json:"verified_at"`

	// Relations
	GateEvents  []GateEvent   `gorm:"foreignKey:PermitID" json:"gate_events"`
	AuditEvents []PermitAudit `gorm:"foreignKey:PermitID" json:"audit_events"`
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

type User struct {
	gorm.Model
	AuthID      uint   `gorm:"uniqueIndex" json:"auth_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	ThirdName   string `json:"third_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
	Notes       string `json:"notes"`
	Role        string `json:"role"`

	CustomsPostID *uint        `json:"customs_post_id"`
	CustomsPost   *CustomsPost `gorm:"foreignKey:CustomsPostID" json:"customs_post,omitempty"`

	UserSettings *UserSettings `gorm:"foreignKey:UserID" json:"user_settings,omitempty"`
}

type UserSettings struct {
	gorm.Model
	UserID   uint           `gorm:"uniqueIndex;not null" json:"user_id"`
	Settings datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"settings"`
}
