package controllers

import (
	"fmt"
	"kanban/config"
	"kanban/library"
	"kanban/models"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Project struct{}

// Sidebar Create a Space form workSpace ekleme
func (project Project) WorkSpaceAdd(c *gin.Context) {
	name := c.PostForm("name")

	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	newWorkSpace := models.WorkSpace{
		Name: name,
	}

	if err := newWorkSpace.Add(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add workspace"})
		return
	}

	library.SetAlert(c, "Alan başarıyla oluşturuldu")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/%d", newWorkSpace.ID))
}

// Sidebar Create a Space form workSpace güncelleme
func (project Project) WorkSpaceNameUpdate(c *gin.Context) {
	workspaceIDStr := c.PostForm("workspace_id")
	UpdateWorkSpace := models.WorkSpace{}.Get("ID = ?", workspaceIDStr)

	name := c.PostForm("name")

	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	UpdateWorkSpace.Updates(models.WorkSpace{
		Name: name,
	})

	library.SetAlert(c, "Alan başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/%d", UpdateWorkSpace.ID))
}

// Sidebar Create a Space form workSpace güncelleme
func (project Project) WorkSpaceDelete(c *gin.Context) {
	workspaceIDStr := c.PostForm("workspace_id")

	workspaceDelete := models.WorkSpace{}.Get("ID = ?", workspaceIDStr)
	projectDelete := models.Project{}.GetAll("work_space_id = ?", workspaceIDStr)
	projectUsersDelete := models.ProjectUser{}.GetAll("work_space_id = ?", workspaceIDStr)
	issuesDelete := models.Issue{}.GetAll("work_space_id = ?", workspaceIDStr)
	workspaceDelete.Delete()
	for _, proje := range projectDelete {
		proje.Delete()
	}

	for _, projeuser := range projectUsersDelete {
		projeuser.Delete()
	}

	for _, issue := range issuesDelete {
		issue.Delete()
	}

	library.SetAlert(c, "Alan başarıyla silindi")
	c.Redirect(http.StatusSeeOther, "/")
}

func (project Project) ProjectAdd(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	workspaceIDStr := c.PostForm("workspace_id")

	selectedCheckboxesEmails := c.PostFormArray("checkboxes")

	workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ":Unauthorized"})
		return
	}

	newProject := models.Project{
		Name:        name,
		Description: description,
		WorkSpaceID: uint(workspaceID),
	}

	if err := newProject.Add(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Added to project failed"})
		return
	}

	projectID := newProject.ID

	addUserRequest := struct {
		ProjectID uint     `json:"project_id"`
		Emails    []string `json:"email"`
		Role      string
	}{
		ProjectID: projectID, 
		Emails:    selectedCheckboxesEmails,
	}

	err = models.AddUserToProjectByEmail(c, addUserRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There is a problem occured while added user to project"})
		return
	}

	library.SetAlert(c, "Proje başarıyla oluşturuldu")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/%d", workspaceID))

}

func (project Project) ProjectUserAdd(c *gin.Context) {
	projectIDStr := c.PostForm("project_id")

	selectedCheckboxesEmails := c.PostFormArray("checkboxes")
	fmt.Println(projectIDStr)
	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ":Unauthorized"})
		return
	}

	gecici := models.ProjectUser{}.Get("ID = ?", projectID)

	gecici.Updates(models.ProjectUser{
		Role: "member",
	})
	db := models.GetDB()
	var control models.ProjectUser
	db.Where("user_id = ? AND project_id = ?", userID, projectID).First(&control)

	if control.Role != "manager" && control.Role != "owner" {
		library.SetAlert(c, "Erişiminiz yok")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", projectID))
		return
	}

	addUserRequest := struct {
		ProjectID uint     `json:"project_id"`
		Emails    []string `json:"email"`
		Role      string
	}{
		ProjectID: uint(projectID),
		Emails:    selectedCheckboxesEmails,
		Role:      "member",
	}

	err = models.AddUserToProjectByEmail(c, addUserRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "There was a problem adding users to the project"})
		return
	}

	library.SetAlert(c, "Projeye yeni kullanıcı başarıyla eklendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", projectID))
}
func ModifiedDeleteProjectUser(c *gin.Context) {
	var requestData struct {
		UserID    string `json:"user_id"`
		ProjectID string `json:"project_id"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz JSON verisi"})
		return

	}
	fmt.Println("userid" + requestData.UserID)
	fmt.Println("projectid" + requestData.ProjectID)
	userID, err := strconv.ParseUint(requestData.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz kullanıcı ID'si"})
		return
	}
	projectID, err := strconv.ParseUint(requestData.ProjectID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID'si"})
		return
	}

	db := models.GetDB()
	sessionuserID := GetCurrentUserID(c)
	var roleController models.ProjectUser
	if err := db.Where("user_id=? AND project_id=?", sessionuserID, projectID).First(&roleController).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Erişim reddedildi"})
		return
	}

	if roleController.Role != "owner" {
		library.SetAlert(c, "Sadece proje sahibi kullanıcı silebilir")
		c.Redirect(http.StatusSeeOther, "/team_space/list/"+strconv.FormatUint(projectID, 10))
		return
	}
	var willdelete models.ProjectUser
	if err := db.Where("user_id=? AND project_id=?", userID, projectID).Delete(&willdelete).Error; err != nil {
		return
	}
	var who models.User
	db.Where("id=?", sessionuserID).First(&who)
	var which models.Project
	db.Where("id=?", projectID).First(&which)
	message := fmt.Sprintf("%s tarafından \"%s\" isimli projeden çıkarıldın", who.Email, which.Name)
	var to models.User
	db.Where("id=?", userID).First(&to)

	go models.SendNotification(uint(userID), message, uint(projectID))

	go models.SendMailSimpleHTMLforDiscard(
		"Kanban Uygulaması",
		"./views/mail/discardToProject.html",
		[]string{to.Email},
		who.Username,
		who.Email,
		which.Name,
		uint64(projectID),
	)
	library.SetAlert(c, "Kullanıcı bu projeden başarıyla silindi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", projectID))

}

func (project Project) ProjectUpdate(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	projectsIDStr := c.PostForm("project_id")

	session := sessions.Default(c)
	userID := session.Get("UserID")
	issueUpdate := models.ProjectUser{}.Get("project_id = ?", projectsIDStr)

	UpdateProject := models.Project{}.Get("id = ?", projectsIDStr)
	UpdateProjectUser := models.ProjectUser{}.GetAll("project_id = ?", projectsIDStr)

	currentProjectID := issueUpdate.ProjectID
	db, err := gorm.Open(mysql.Open(config.GetConnectionString()), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection failed"})
		c.Abort()
		return
	}
	var projectUser models.ProjectUser
	if err := db.Where("user_id = ? AND project_id = ?", userID, currentProjectID).First(&projectUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in project"})

		c.Abort()
		return
	}
	if projectUser.Role != "manager" && projectUser.Role != "owner" {

		library.SetAlert(c, "Erişiminiz yok")

		c.Abort()
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", UpdateProject.ID))
		return
	} else {
		library.SetAlert(c, "Proje başarıyla güncellendi")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", UpdateProject.ID))
	}

	UpdateProject.Updates(models.Project{
		Name:        name,
		Description: description,
	})

	for _, projectUser := range UpdateProjectUser {
		projectUser.Updates(models.ProjectUser{
			Name:        name,
			Description: description,
		})
	}

	library.SetAlert(c, "Proje başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", UpdateProject.ID))
}

// BodyNavbar Proje silme
func (project Project) ProjectDelete(c *gin.Context) {
	projectIDStr := c.PostForm("project_id")

	userID := GetCurrentUserID(c)
	db := models.GetDB()
	var roleControl models.ProjectUser
	db.Where("project_id=? AND user_id=?", projectIDStr, userID).First(&roleControl)
	if roleControl.Role != "manager" && roleControl.Role != "owner" {
		library.SetAlert(c, "erişiminiz yok")
		c.Abort()
		c.Redirect(http.StatusSeeOther, "/team_space/list/"+projectIDStr)
		return

	}
	projectDeleteUsers := models.ProjectUser{}.GetAll("project_id = ?", projectIDStr)
	projectDelete := models.Project{}.Get("ID = ?", projectIDStr)
	issueDelete := models.Issue{}.GetAll("project_id = ?", projectIDStr)

	projectDelete.Delete()
	for _, pu := range projectDeleteUsers {
		pu.Delete()
	}
	for _, iu := range issueDelete {
		iu.Delete()
	}

	library.SetAlert(c, "Proje başarıyla silindi")
	c.Redirect(http.StatusSeeOther, "/")
}
