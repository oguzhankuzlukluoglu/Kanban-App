package controllers

import (
	"kanban/library"
	"kanban/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func MarkAllNotificationsReadHandler(c *gin.Context) {

	userID := GetCurrentUserID(c)
	models.MarkAllNotificationsAsRead(userID)

}
func DeleteNotification(c *gin.Context) {
	notiID := c.Param("id")

	db := models.GetDB()
	var notification models.Notifications
	if err := db.First(&notification, notiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking notification existence"})
		}
		return
	}
	if err := db.Delete(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}
	library.SetAlert(c, "Bildirim başarıyla silindi")

}
func DeleteAllNotification(c *gin.Context) {
	UserID := GetCurrentUserID(c)

	db := models.GetDB()
	db.Where("user_id=?", UserID).Delete(&models.Notifications{})

	library.SetAlert(c, "Tüm bildirimler başarıyla silindi")

}
