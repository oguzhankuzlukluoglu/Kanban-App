package models

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Notifications struct {
	gorm.Model
	UserID    uint   `json:"user_id"`
	Message   string `json:"message"`
	IsSeen    bool   `json:"isSeen" gorm:"default:false"`
	ProjectID uint
}

func (Notifications) TableName() string {
	return "notifications"
}
func (notification Notifications) Migrate() {
	db := GetDB()

	db.AutoMigrate(&notification)
}

func (notification Notifications) Get(where ...interface{}) Notifications {
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&notification)
	return notification
}

func (notification Notifications) GetAll(where ...interface{}) []Notifications {
	db := GetDB()

	var notifications []Notifications
	db.Find(&notifications, where...)
	return notifications
}

func (notification Notifications) Updates(data Notifications) {
	db := GetDB()

	db.Model(&notification).Updates(data)
}

func SendNotification(userID uint, message string, projectid uint) error {
	db := GetDB()
	notification := Notifications{
		UserID:    userID,
		Message:   message,
		ProjectID: projectid,
	}
	if err := db.Create(&notification).Error; err != nil {
		return err
	}
	fmt.Printf("Notification sent to user %d: %s\n", userID, message)
	return nil

}
func GetNotificationByUserId(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	db := GetDB()
	var notification []Notifications
	err = db.Where("user_id=?", userID).Find(&notification).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, notification)

}
func (notification Notifications) GetUnseenCount(userID uint) int64 {
	db := GetDB()

	var count int64
	db.Model(&Notifications{}).Where("user_id = ? AND is_seen = ?", userID, false).Count(&count)

	return count
}
func MarkAllNotificationsAsRead(userID uint) error {
	db := GetDB()
	err := db.Model(&Notifications{}).Where("user_id = ?", userID).Update("is_seen", true).Error
	if err != nil {
		return err
	}
	return nil
}
