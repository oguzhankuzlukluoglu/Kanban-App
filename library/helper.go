package library

import (
	"fmt"
	"kanban/config"
	"kanban/models"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func SessionData(userID interface{}) map[string]interface{} {
	data := make(map[string]interface{})

	data["WorkSpaces"] = models.WorkSpace{}.GetAll("user_id = ?", userID)
	data["User"] = models.User{}.GetAll("ID = ?", userID)
	data["OtherUsers"] = models.User{}.GetAll("ID != ?", userID)
	data["Notification"] = models.Notifications{}.GetAll(" user_id = ?", userID)
	data["Share"] = models.ProjectUser{}.GetAll(" user_id = ? AND is_share = ?", userID, true)

	var count int64
	db := models.GetDB()
	db.Model(&models.Notifications{}).Where("user_id = ? AND is_seen = ?", userID, false).Count(&count)
	data["UnSeenNotification"] = count

	data["MyProjects"] = models.Project{}.GetAll("user_id = ?", userID)
	data["SharedWithMe"] = models.ProjectUser{}.GetAll("user_id = ? AND is_share = ?", userID, true)

	data["MyIssues"] = models.Issue{}.ModifiedGetAll("user_id = ?", userID) //ModifiedGetAll ile proje ve issueları preload yaptık.

	data["AllProjects"] = models.ProjectUser{}.GetAll("user_id = ?", userID)

	return data
}

func IdData(ID interface{}) map[string]interface{} {
	db, err := gorm.Open(mysql.Open(config.GetConnectionString()), &gorm.Config{})
	if err != nil {
		fmt.Println(err)

	}
	data := make(map[string]interface{})
	var userIDs []uint
	var emails []string
	var projectEmailsExceptMe []string
	var allEmailsExceptProject []string

	data["ProjectId"] = ID
	data["WorkSpace"] = models.WorkSpace{}.GetAll("ID = ?", ID)
	data["Project"] = models.Project{}.GetAll("ID = ?", ID)
	data["Issues"] = models.Issue{}.GetAll("project_id = ?", ID)
	data["Projects"] = models.Project{}.GetAll("work_space_id = ?", ID)
	data["ToDo"] = models.Issue{}.GetAll("project_id = ? AND status = ?", ID, "card1")
	data["Working"] = models.Issue{}.GetAll("project_id = ? AND status = ?", ID, "card2")
	data["Done"] = models.Issue{}.GetAll("project_id = ? AND status = ?", ID, "card3")

	var userRoles []struct {
		Email  string
		Role   string
		UserID string
	}

	db.Table("project_users").
		Select("users.email, project_users.role,user_id").
		Joins("join users on users.id = project_users.user_id").
		Where("project_users.project_id = ? AND project_users.deleted_at IS NULL", ID).
		Scan(&userRoles)

	data["UserRoles"] = userRoles

	db.Model(&models.ProjectUser{}).Where("project_id = ? AND user_id != ?", ID, 0).Pluck("DISTINCT user_id", &userIDs)
	db.Model(&models.User{}).Where("id IN ?", userIDs).Pluck("email", &emails)
	data["Emails"] = emails

	db.Model(&models.User{}).Where("id IN ? AND ID != ?", userIDs, ID).Pluck("email", &projectEmailsExceptMe)
	data["ProjectEmailsExceptMe"] = projectEmailsExceptMe

	db.Model(&models.User{}).Where("email NOT IN ?", emails).Pluck("email", &allEmailsExceptProject)
	data["AllEmailsExceptProject"] = allEmailsExceptProject
	return data
}

func CombineSessionAndIdData(c *gin.Context, userID, ID interface{}) map[string]interface{} {
	data := SessionData(userID)

	idData := IdData(ID)

	data["Alert"] = GetAlert(c)

	for key, value := range idData {
		data[key] = value
	}

	return data
}
