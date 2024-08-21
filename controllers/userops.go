package controllers

import (
	"crypto/sha256"
	"fmt"
	"html/template"
	"kanban/library"
	"kanban/models"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Userops struct{}

func (userops Userops) Index(c *gin.Context) {
	view, err := template.ParseFiles(library.Include("login")...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	data := make(map[string]interface{})
	data["Alert"] = library.GetAlert(c)
	if err := view.ExecuteTemplate(c.Writer, "index", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

}
func (userops Userops) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := fmt.Sprintf("%x", sha256.Sum256([]byte(c.PostForm("password"))))

	user := models.User{}.Get("email = ? AND password = ?", email, password)

	if user.Email == email && user.Password == password {

		userIDStr := strconv.FormatUint(uint64(user.ID), 10)

		token, err := GenerateToken(userIDStr)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token was not created"})
			return
		}
		session := sessions.Default(c)
		session.Set("token", token)
		session.Set("UserID", userIDStr)
		if err := session.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Sessions was not saved"})
			return
		}
		fmt.Println(userIDStr + "true")
		library.SetAlert(c, "Hoşgeldiniz")
		c.Redirect(http.StatusSeeOther, "/")

	} else {
		library.SetAlert(c, "Yanlış Kullanıcı Adı veya Şifre...")
		c.Redirect(http.StatusSeeOther, "/login")

	}
}

func (userops Userops) Signup(c *gin.Context) {
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := fmt.Sprintf("%x", sha256.Sum256([]byte(c.PostForm("password"))))

	user := models.User{
		Username: username,
		Password: password,
		Email:    email,
	}
	reUserControl := models.User{}.Get("email=?",email)
	if email == reUserControl.Email {
		library.SetAlert(c, "Bu email daha önce alınmış")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	if email != "" && username != "" && password != "" {
		user.Add()
		library.SetAlert(c, "Kayıt başarılı. Lütfen giriş yapınız")
		c.Redirect(http.StatusSeeOther, "/login")
		return

	} else {

		library.SetAlert(c, "Kayıt işleminde bir hata oluştu. Boş alanları kontrol edin")
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
}

func (userops Userops) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("token")
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session could not be cleared"})
		return
	}
	library.SetAlert(c, "Hoşçakalın...")
	c.Redirect(http.StatusSeeOther, "/login")

}
func GetCurrentUserID(c *gin.Context) uint {
	session := sessions.Default(c)
	userID := session.Get("UserID")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		c.Abort()
		return 0
	}

	userIDUint, err := strconv.ParseUint(userID.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		c.Abort()
		return 0
	}
	return uint(userIDUint)
}

func (userops Userops) UserUpdate(c *gin.Context) {
	userID := GetCurrentUserID(c)
	name := c.PostForm("name")
	email := c.PostForm("email")
	bio := c.PostForm("bio")

	UserUpdater := models.User{}.Get("ID = ?", userID)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ":Unauthorized"})
		return
	}

	UserUpdater.Updates(models.User{
		Username: name,
		Email:    email,
		Bio:      bio,
	})

	library.SetAlert(c, "Kullanıcı bilgileri başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, "/settings")
}

func (userops Userops) PasswordChange(c *gin.Context) {
	userID := GetCurrentUserID(c)
	password := fmt.Sprintf("%x", sha256.Sum256([]byte(c.PostForm("password"))))

	PasswordChanger := models.User{}.Get("ID = ?", userID)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ":Unauthorized"})
		return
	}

	PasswordChanger.Updates(models.User{
		Password: password,
	})

	library.SetAlert(c, "Şifre başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, "/settings")
}

func (userops Userops) UserDelete(c *gin.Context) {
	userID := GetCurrentUserID(c)

	UserDeleter := models.User{}.Get("ID = ?", userID)
	workspaceDelete := models.WorkSpace{}.GetAll("user_id = ?", userID)
	projectDelete := models.Project{}.GetAll("user_id = ?", userID)
	projectUsersDelete := models.ProjectUser{}.GetAll("reporter_id = ?", userID)
	IssueDelete := models.Issue{}.GetAll("reporter = ?", userID)

	UserDeleter.Delete()
	for _, workspace := range workspaceDelete {
		workspace.Delete()
	}
	for _, proje := range projectDelete {
		proje.Delete()
	}
	for _, projeuser := range projectUsersDelete {
		projeuser.Delete()
	}
	for _, issue := range IssueDelete {
		issue.Delete()
	}

	library.SetAlert(c, "Kullanıcı başarıyla silindi")
	c.Redirect(http.StatusSeeOther, "/login")
}
