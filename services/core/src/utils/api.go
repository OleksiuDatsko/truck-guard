package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPagination(c *gin.Context) (int, int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit
	return limit, offset, page
}

func SendPaginatedResponse(c *gin.Context, data interface{}, total int64, page, limit int) {
	totalPages := 0
	if limit > 0 {
		totalPages = int(total) / limit
		if int(total)%limit > 0 {
			totalPages++
		}
	}

	c.JSON(200, gin.H{
		"data": data,
		"metadata": gin.H{
			"total_items":  total,
			"total_pages":  totalPages,
			"current_page": page,
			"limit":        limit,
		},
	})
}
