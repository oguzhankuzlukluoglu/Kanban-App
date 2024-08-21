package controllers

import (
	"html/template"
	"kanban/library"
	"kanban/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Dashboard struct {
}

// anasayfa Home
func (dashboard Dashboard) HomeIndex(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("home")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userID := GetCurrentUserID(c)

	ID := c.Params.ByName("ID")

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) InboxIndex(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("inbox")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userID := GetCurrentUserID(c)

	ID := c.Params.ByName("ID")

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) EverythingIndex(c *gin.Context) {
	view := GetTemplates()

	userID := GetCurrentUserID(c)

	ID := c.Params.ByName("ID")

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) NotificationAsRead(c *gin.Context) {
	var request map[string]interface{}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := request["id"].(string)

	notificationUpdate := models.Notifications{}.Get("ID = ?", id)

	notificationUpdate.Updates(models.Notifications{
		IsSeen: true,
	})
}

func (dashboard Dashboard) ShareIndex(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("share")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userID := GetCurrentUserID(c)

	if err := view.ExecuteTemplate(c.Writer, "index", library.SessionData(userID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// Settings Sayfası
func (dashboard Dashboard) SettingsIndex(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("settings")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	ID := c.Params.ByName("ID")
	userID := GetCurrentUserID(c)

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
func CheckUserProjectAccess(c *gin.Context, projectID uint64, userID uint) bool {
	db := models.GetDB()
	var projectUser models.ProjectUser
	if err := db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&projectUser).Error; err != nil {
		return false
	}
	return true
}

func (dashboard Dashboard) TeamSpaceDetailsById(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("team_space")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	ID := c.Params.ByName("ID")
	uintID, err := strconv.ParseUint(ID, 10, 32)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid ID format")
		return
	}

	userID := GetCurrentUserID(c)

	db := models.GetDB()
	var workSpaceControl models.WorkSpace

	db.Where("user_id = ? AND id = ?", userID, uintID).First(&workSpaceControl)
	if workSpaceControl.ID != uint(uintID) {
		library.SetAlert(c, "Bu projeye erişiminiz yok")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) TeamSpaceDetailsListById(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("team_space/list")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	ID := c.Params.ByName("ID")
	userID := GetCurrentUserID(c)
	projectID, err := strconv.ParseUint(ID, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	if !CheckUserProjectAccess(c, projectID, userID) {
		library.SetAlert(c, "Bu projeye erişiminiz yok.")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) TeamSpaceDetailsBoardById(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("team_space/board")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	ID := c.Params.ByName("ID")
	projectID, err := strconv.ParseUint(ID, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	if !CheckUserProjectAccess(c, projectID, userID) {
		library.SetAlert(c, "Bu projeye erişiminiz yok.")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func (dashboard Dashboard) UpdateIssueStatus(c *gin.Context) {
	var request map[string]interface{}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := request["id"].(string)
	status := request["status"].(string)

	issueUpdate := models.Issue{}.Get("ID = ?", id)
	issueUpdate.Updates(models.Issue{
		Status: status,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Update success", "id": id, "status": status})
}

func (dashboard Dashboard) TeamSpaceDetailsTableById(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("team_space/table")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userID := GetCurrentUserID(c)
	ID := c.Params.ByName("ID")
	projectID, err := strconv.ParseUint(ID, 10, 64)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid project ID")
		return
	}

	if !CheckUserProjectAccess(c, projectID, userID) {
		library.SetAlert(c, "Bu projeye erişiminiz yok.")
		c.Redirect(http.StatusSeeOther, "/")
		return
	}

	if err := view.ExecuteTemplate(c.Writer, "index", library.CombineSessionAndIdData(c, userID, ID)); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
