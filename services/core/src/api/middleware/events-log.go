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

		sourceID := c.GetHeader("X-Source-ID")
		if sourceID == "" {
			sourceID = c.GetHeader("X-Scale-ID")
		}

		ts := time.Now()
		if tsStr, ok := body["timestamp"].(string); ok {
			if parsed, err := time.Parse(time.RFC3339, tsStr); err == nil {
				ts = parsed
			}
		}

		payloadJSON, _ := json.Marshal(body)

		sysEvent := models.SystemEvent{
			Type:      eventType,
			SourceID:  sourceID,
			Payload:   string(payloadJSON),
			Timestamp: ts,
		}

		repository.DB.Create(&sysEvent)

		c.Set("system_event_id", sysEvent.ID)

		c.Next()
	}
}
