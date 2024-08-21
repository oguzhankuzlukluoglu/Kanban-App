package models

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	IssueID   uint
	Text      string
	UserID    uint
	ProjectID uint
	IssueInt  uint
	Issue     Issue `gorm:"foreignKey:IssueID;references:ID"`
	User      User  `gorm:"foreignKey:UserID"`
}

func (Comment) TableName() string {
	return "comments"
}

func (comment Comment) Migrate() {
	db := GetDB()

	db.AutoMigrate(&comment)
}

func (comment Comment) Add() {
	db := GetDB()

	db.Create(&comment)
}
func GetComment(c *gin.Context) {
	db := GetDB()

	var comments []Comment
	db.Find(&comments)
	c.JSON(http.StatusOK, comments)
}

func GetCommentByIssue(c *gin.Context) {
	db := GetDB()

	issueInt, err := strconv.Atoi(c.Query("issue_int"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz issue_int parametresi"})
		return
	}

	var comments []Comment
	db.Where("issue_int = ?", issueInt).Preload("User").Find(&comments)

	c.JSON(http.StatusOK, comments)
}
