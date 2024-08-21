package controllers

import (
	"fmt"
	"kanban/library"
	"kanban/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Issue struct{}

func (issue Issue) IssueAdd(c *gin.Context) {

	name := c.PostForm("name")
	description := c.PostForm("description")
	due_date := c.PostForm("due_date")
	priority := c.PostForm("priority")
	projectIDStr := c.PostForm("project_id")
	checkboxes := c.PostFormArray("checkboxes")

	projectID, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil || projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz proje ID"})
		return
	}

	layout := "2006-01-02"
	parsedDueDate, err := time.Parse(layout, due_date)
	if err != nil {
		library.SetAlert(c, "Geçersiz tarih formatı")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	userID := GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz erişim"})
		return
	}

	db := models.GetDB()
	var project models.Project
	if err := db.Model(&models.Project{}).Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Proje bilgileri alınırken hata oluştu"})
		return
	}
	var lastIssue models.Issue //bu kısım issue listelemesi için custom edildi frontla haberleşmek için gizli bir int
	if err := db.Last(&lastIssue).Error; err != nil {
		fmt.Println("Fetching last notification is failed:", err)
	}
	convert := lastIssue.ID + 1

	for _, idStr := range checkboxes {
		selectedUserID, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		newIssue := models.Issue{
			UserID:      uint(selectedUserID),
			Title:       name,
			Description: description,
			DueDate:     parsedDueDate,
			Priority:    priority,
			ProjectID:   uint(projectID),
			WorkSpaceID: project.WorkSpaceID,
			Status:      "card1",
			IssueInt:    convert,
			Reporter:    GetCurrentUserID(c),
		}
		var currentProjectid = newIssue.ProjectID
		var projectuser models.ProjectUser

		if err := db.Where("user_id = ? AND project_id = ?", userID, currentProjectid).First(&projectuser).Error; err != nil {
			continue
		}

		if projectuser.Role != "manager" && projectuser.Role != "owner" {

			library.SetAlert(c, "Erişiminiz yok")

			c.Abort()
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", newIssue.ProjectID))
			return
		}

		if err := newIssue.Add(project.WorkSpaceID); err != nil {
			library.SetAlert(c, "Görev eklenirken hata oluştu")
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", projectID))
			return
		}
		whoidM := GetCurrentUserID(c)
		var whouserM models.User
		if err := db.Where("id = ?", whoidM).First(&whouserM).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Current user details alınırken hata oluştu"})
			return
		}
		message := fmt.Sprintf(" \"%s\" isimli projede \"%s\" tarafından \"%s\" isimli issue'ya atandın", project.Name, whouserM.Email, name)
		err = models.SendNotification(uint(selectedUserID), message, project.ID)
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}
		var selectedUser models.User
		db.Where("id=?", selectedUserID).Find(&selectedUser)

		go models.SendMailSimpleHTMLForAssignment(
			"Kanban Uygulaması",
			"./views/mail/issueAssignment.html",
			[]string{selectedUser.Email}, // Email gönderimi için kullanılıyor
			selectedUser.Username,        // Kullanıcı adı
			whouserM.Email,               // sessiondaki userin emaili
			project.Name,                 //projenin ismi
			name,                         // issue adı
			project.ID,                   // Proje ID
			newIssue.DueDate,             //issue'nun expire'ı
		)
	}

	library.SetAlert(c, "Görev Başarıyla Eklendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", projectID))
}

// List sayfası issue yanında düzenleme butonu form issue düzenleme
func (issue Issue) IssueEdit(c *gin.Context) {
	issueIDStr := c.PostForm("issue_id")

	issueUpdate := models.Issue{}.GetAll("issue_int = ?", issueIDStr)
	session := sessions.Default(c)
	userID := session.Get("UserID")

	db := models.GetDB()

	var currentProjectID uint
	authorized := false

	for _, issue := range issueUpdate {
		currentProjectID = issue.ProjectID

		var projectUser models.ProjectUser
		if err := db.Where("user_id = ? AND project_id = ?", userID, currentProjectID).First(&projectUser).Error; err != nil {
			continue
		}

		if projectUser.Role == "manager" || projectUser.Role == "owner" {
			authorized = true
			break
		}
	}

	if !authorized {
		library.SetAlert(c, "Erişiminiz yok")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", currentProjectID))
		return
	}

	name := c.PostForm("name")
	dueDateStr := c.PostForm("due_date")
	priority := c.PostForm("priority")
	status := c.PostForm("status")
	checkboxes := c.PostFormArray("checkboxes")

	fmt.Println("******************")
	fmt.Println(status)

	layout := "2006-01-02"
	parsedDueDate, err := time.Parse(layout, dueDateStr)
	if err != nil {
		library.SetAlert(c, "Geçersiz tarih formatı")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if status == "ToDo" {
		status = "card1"
	}
	if status == "Working" {
		status = "card2"
	}
	if status == "Done" {
		status = "card3"
	}

	for _, issue := range issueUpdate {
		for _, idStr := range checkboxes {
			userID, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			db.Model(&issue).Updates(models.Issue{
				Title:    name,
				DueDate:  parsedDueDate,
				Priority: priority,
				UserID:   uint(userID),
				Status:   status,
			})
		}
	}

	library.SetAlert(c, "Görev başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", currentProjectID))
}

func (issue Issue) IssueDelete(c *gin.Context) {
	issueIDStr := c.PostForm("issue_id")
	issueUpdate := models.Issue{}.Get("ID = ?", issueIDStr)
	session := sessions.Default(c)
	userID := session.Get("UserID")
	currentProjectID := issueUpdate.ProjectID
	db := models.GetDB()
	var projectUser models.ProjectUser
	if err := db.Where("user_id = ? AND project_id = ?", userID, currentProjectID).First(&projectUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in project"})

		c.Abort()
		return
	}
	if projectUser.Role != "manager" && projectUser.Role != "owner" {

		library.SetAlert(c, "Erişiminiz yok")

		c.Abort()
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", issueUpdate.ProjectID))
		return
	} else {
		library.SetAlert(c, "Görev başarıyla silindi")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", issueUpdate.ProjectID))
	}

	issueDelete := models.Issue{}.GetAll("issue_int = ?", issueIDStr)

	for i := 0; i < len(issueDelete); i++ {
		issueDelete[i].Delete()
	}

	library.SetAlert(c, "Görev başarıyla silindi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", currentProjectID))
}
