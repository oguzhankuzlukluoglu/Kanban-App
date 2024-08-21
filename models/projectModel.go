package models

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name         string `json:"name"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	UserID       uint
	WorkSpaceID  uint
	Issues       []Issue
	ProjectUsers []ProjectUser
}

func (Project) TableName() string {
	return "projects"
}

type ProjectUser struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      uint
	ProjectID   uint
	WorkSpaceID uint
	IsShare     bool `json:"isShare" gorm:"default:false"`
	ReporterID  uint
	Role        string `gorm:"default:'member'"`
}

func (ProjectUser) TableName() string {
	return "project_users"
}

func (project Project) Migrate() {
	db := GetDB()

	db.AutoMigrate(&project) 

}

func (project *Project) Add(userID uint) error {
	db := GetDB()

	project.UserID = userID

	if err := db.Create(&project).Error; err != nil {
		return err
	}
	projectUser := ProjectUser{
		ProjectID:   project.ID,
		UserID:      userID,
		WorkSpaceID: project.WorkSpaceID,
		Name:        project.Name,
		Description: project.Description,
		IsShare:     false,
		ReporterID:  project.UserID,
		Role:        "owner",
	}
	if err := db.Create(&projectUser).Error; err != nil {
		return err
	}

	return nil
}

func GetProjects(c *gin.Context) {
	db := GetDB()
	var projects []Project
	db.Preload("Issues").Find(&projects)
	c.JSON(http.StatusOK, projects)
}

func (project Project) Get(where ...interface{}) Project {
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&project)
	return project
}

func (project Project) GetAll(where ...interface{}) []Project {
	db := GetDB()

	var projects []Project
	db.Find(&projects, where...)
	return projects
}

func (project Project) Update(column string, value interface{}) {
	db := GetDB()

	db.Model(project).Update(column, value)
}

func (project Project) Updates(data Project) {
	db := GetDB()

	db.Model(project).Updates(data)
}

func (project Project) Delete() {
	db := GetDB()

	db.Delete(project, project.ID)
}


func (pu ProjectUser) Migrate() {
	db := GetDB()

	db.AutoMigrate(&pu)
}
func (pu ProjectUser) Add() {
	db := GetDB()

	db.Create(&pu)
}

func (pu ProjectUser) Get(where ...interface{}) ProjectUser {
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&pu)
	return pu
}

func (pu ProjectUser) GetAll(where ...interface{}) []ProjectUser {
	db := GetDB()

	var ProjectUsers []ProjectUser
	db.Find(&ProjectUsers, where...)
	return ProjectUsers
}
func (pu ProjectUser) Updates(data ProjectUser) {
	db := GetDB()

	db.Model(&pu).Updates(data)
}
func (pu ProjectUser) Delete() {
	db := GetDB()

	db.Delete(pu, pu.ID)
}
func (project *Project) AddUser(userID uint) error {
	db := GetDB()

	projectUser := ProjectUser{
		ProjectID:   project.ID,
		UserID:      userID,
		Name:        project.Name,
		WorkSpaceID: project.WorkSpaceID,
		Description: project.Description,
		IsShare:     true,
		ReporterID:  project.UserID,
	}

	return db.Create(&projectUser).Error
}

func (project Project) GetAllByUser(userID uint) ([]Project, error) {
	db := GetDB()

	var projects []Project
	err := db.Joins("JOIN project_users ON project_users.project_id = projects.id").
		Where("project_users.user_id = ?", userID).
		Find(&projects).Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}
func GetUserByEmail(email []string) (User, error) {
	var user User
	db := GetDB()

	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return user, result.Error
	}

	return user, nil
}

func GetUserIDsByEmails(emails []string) ([]uint, error) {
	db := GetDB()

	var users []User
	var ids []uint
	if len(emails) == 0 {
		return ids, nil
	}

	result := db.Where("email IN (?)", emails).Find(&users)

	if result.Error != nil {
		return nil, result.Error
	}
	for _, user := range users {
		ids = append(ids, user.ID)
	}

	return ids, nil

}

func AddUserToProjectByEmail(c *gin.Context, request struct {
	ProjectID uint     `json:"project_id"`
	Emails    []string `json:"email"`
	Role      string
}) error {

	userIDs, err := GetUserIDsByEmails(request.Emails)
	if err != nil {
		return err
	}

	var project Project
	db := GetDB()

	if err := db.First(&project, request.ProjectID).Error; err != nil {
		return err
	}

	session := sessions.Default(c)
	userID := session.Get("UserID")
	user1 := &User{}
	db.Where("id =?", userID).First(&user1)

	for _, id := range userIDs {
		if err := project.AddUser(id); err != nil {
			return err
		}
		
		message := fmt.Sprintf("%s tarafından \"%s\" isimli projeye eklendin", user1.Email, project.Name)
		SendNotification(id, message, project.ID)
		var lastNotification Notifications
		if err := db.Last(&lastNotification).Error; err != nil {
			fmt.Println("Fetching last notification is failed:", err)
			continue
		}

		var userN User
		if err := db.Where("id = ?", lastNotification.UserID).First(&userN).Error; err != nil {
			fmt.Println("There was a problem occured while fetcing user:", err)
			continue
		}

		var projectN Project
		if err := db.First(&projectN, request.ProjectID).Error; err != nil {
			return err
		}
		go SendMailSimpleHTML(
			"Kanban Uygulaması",
			"./views/mail/addedToProject.html",
			[]string{userN.Email},
			userN.Username,
			user1.Email,
			projectN.Name,
			request.ProjectID,
		)
	}

	return nil
}

func GetProjectsByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var project Project
	projects, err := project.GetAllByUser(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}
func GetProjectsByUserID(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	db := GetDB()

	var projects []Project
	err = db.Joins("JOIN project_users ON project_users.project_id = projects.id").
		Where("project_users.user_id = ?", userID).
		Find(&projects).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, projects)
}
