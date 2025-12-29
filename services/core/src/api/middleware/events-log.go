package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
)

func SystemEventLogger(eventType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}

		if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
			c.Next()
			return
		}

		sourceID, _ := body["camera_id"].(string)
		if sourceID == "" {
			sourceID, _ = body["scale_id"].(string)
		}

		ts := time.Now()
		if tsStr, ok := body["timestamp"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				ts = parsed
			}
		}

		payloadMap := make(map[string]interface{})
		for k, v := range body {
			if k != "camera_id" && k != "scale_id" && k != "timestamp" && k != "camera_name" {
				payloadMap[k] = v
			}
		}
		payloadJSON, _ := json.Marshal(payloadMap)

		repository.DB.Create(&models.SystemEvent{
			Type:      eventType,
			SourceID:  sourceID,
			Payload:   string(payloadJSON),
			Timestamp: ts,
		})

		c.Next()
	}
}
