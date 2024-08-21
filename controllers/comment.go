package controllers

import (
	"fmt"
	"kanban/library"
	"kanban/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddCommentHandler(c *gin.Context) {
	commentText := c.PostForm("comment")
	issueID := c.PostForm("issue_int")
	db := models.GetDB()

	if commentText == "" || issueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comment text and issue ID are required"})
		return
	}

	issueIDUint, err := strconv.ParseUint(issueID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid issue ID"})
		return
	}
	var issue models.Issue
	err = db.Select("project_id").Where("id = ?", issueID).First(&issue).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project ID", "details": err.Error()})
		return
	}
	comment := models.Comment{
		Text: commentText,
		UserID:    GetCurrentUserID(c),
		IssueInt:  uint(issueIDUint),
		ProjectID: issue.ProjectID,
	}

	if err := db.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add comment"})
		return
	}

	library.SetAlert(c, "Yorum başarıyla eklendi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", issue.ProjectID))
}
func DeleteCommentHandler(c *gin.Context) {
	comment_id := c.PostForm("comment_id")
	db := models.GetDB()
	var comment models.Comment

	if err := db.Where("id = ?", comment_id).First(&comment).Error; err != nil {
		library.SetAlert(c, "Yorum bulunamadı")
		c.Redirect(http.StatusSeeOther, "/team_space/list/")
		return
	}

	sessionUser := GetCurrentUserID(c)

	if sessionUser != comment.UserID {
		library.SetAlert(c, "Erişiminiz yok")
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", comment.ProjectID))
		return
	}

	db.Delete(&comment)
	fmt.Println(comment.ProjectID)

	library.SetAlert(c, "Yorum başarıyla silindi")
	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/team_space/list/%d", comment.ProjectID))
}
