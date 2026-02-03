package data

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/truckguard/core/src/models"
	"github.com/truckguard/core/src/repository"
	"github.com/truckguard/core/src/utils"
)

func HandleListPosts(c *gin.Context) {
	var posts []models.CustomsPost
	var total int64
	limit, offset, page := utils.GetPagination(c)
	name := c.Query("name")

	query := repository.DB.WithContext(c.Request.Context()).Model(&models.CustomsPost{})
	if name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("id desc").Find(&posts)

	utils.SendPaginatedResponse(c, posts, total, page, limit)
}

func HandleCreatePost(c *gin.Context) {
	var input models.CustomsPost
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := repository.DB.WithContext(c.Request.Context()).Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func HandleUpdatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.CustomsPost
	if err := repository.DB.WithContext(c.Request.Context()).First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	var input models.CustomsPost
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	post.Name = input.Name
	post.IsDefault = input.IsDefault
	// Adding more fields if model changes

	repository.DB.WithContext(c.Request.Context()).Save(&post)
	c.JSON(http.StatusOK, post)
}

func HandleDeletePost(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DB.WithContext(c.Request.Context()).Delete(&models.CustomsPost{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}
