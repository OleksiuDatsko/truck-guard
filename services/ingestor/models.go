package main

import (
	"encoding/json"
	"time"
)

type IngestEvent struct {
	DeviceID string    `json:"device_id"`
	ImageKey string    `json:"image_key"`
	Payload  string    `json:"payload"` 
	At       time.Time `json:"at"`
}

func (e *IngestEvent) ToJSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}