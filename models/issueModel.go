package models

import (
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type Issue struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Reporter    uint   `json:"reporter"`
	ProjectID   uint   `json:"project_id"`
	WorkSpaceID uint
	UserID      uint
	Comments    []Comment `gorm:"foreignKey:IssueID"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	DueDate     time.Time `json:"due_date"`
	User        User      `gorm:"foreignKey:UserID"`
	IssueInt    uint
	Project     Project
}

func (Issue) TableName() string {
	return "issues"
}

func (issue Issue) Migrate() {
	db := GetDB()

	db.AutoMigrate(&issue) 
}

func (issue Issue) Add(workspaceID uint) error {
	db := GetDB()

	issue.WorkSpaceID = workspaceID

	if err := db.Create(&issue).Error; err != nil {
		return err
	}

	return nil
}

func GetIssue(c *gin.Context) {
	db := GetDB()

	var issues []Issue
	db.Preload("Comments").Find(&issues)
	c.JSON(http.StatusOK, issues)
}
func (issue Issue) Get(where ...interface{}) Issue { 
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&issue)
	return issue
}
func (issue Issue) GetAll(where ...interface{}) []Issue {
	db := GetDB()

	var issues []Issue
	db.Preload("User").Find(&issues, where...)
	return issues
}
func (issue Issue) ModifiedGetAll(where ...interface{}) []Issue {
	db := GetDB()

	var issues []Issue
	db.Preload("Project").Find(&issues, where...)
	return issues
}
func (issue Issue) Update(colum string, value interface{}) {
	db := GetDB()

	db.Model(&issue).Update(colum, value)
}
func (issue Issue) Updates(data Issue) {
	db := GetDB()

	db.Model(&issue).Updates(data)
}
func (issue Issue) Delete() {
	db := GetDB()

	db.Delete(&issue, issue.ID)
}
