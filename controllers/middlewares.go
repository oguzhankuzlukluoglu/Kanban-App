package controllers

import (
	"fmt"
	"kanban/config"
	"kanban/library"
	"kanban/models"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(c *gin.Context) {
	session := sessions.Default(c)
	tokenString := session.Get("token")

	if tokenString == nil {
		library.SetAlert(c, "Henüz Giriş Yapmadınız.")
		c.Redirect(http.StatusSeeOther, "/login")
		c.Abort()
		return
	}

	token, err := jwt.ParseWithClaims(tokenString.(string), &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTKey, nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		c.Abort()
		return
	}

	log.Println("Authenticated userID:", claims.UserID)

	c.Set("userID", claims.UserID)
	c.Next()
}

func SetUserRole(c *gin.Context) {
	projectID := c.PostForm("project_id")
	currentUserID := GetCurrentUserID(c)

	db, err := gorm.Open(mysql.Open(config.GetConnectionString()), &gorm.Config{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
		return
	}

	var project models.Project
	if err := db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.UserID != uint(currentUserID) {
		library.SetAlert(c, "Sadece proje sahibi yetki değiştirebilir")
		c.Redirect(http.StatusSeeOther, "/team_space/list/"+projectID)
		return
	}

	for i := 0; ; i++ {
		userIDStr := c.PostForm(fmt.Sprintf("user_id_%d", i))
		if userIDStr == "" {
			break
		}

		role := c.PostForm(fmt.Sprintf("authority_%d", i))

		var projectUser models.ProjectUser
		if err := db.Where("user_id = ? AND project_id = ?", userIDStr, projectID).First(&projectUser).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User %s not part of the project", userIDStr)})
			continue
		}

		if projectUser.Role == "owner" {
			continue
		}
		if projectUser.Role == role {
			continue
		}

		if projectUser.Role != role {
			projectUser.Role = role
			if err := db.Save(&projectUser).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to update role for user %s", userIDStr)})
				continue
			}
		}
		var user models.User //buradaki userIDStr, formdan alınan useridler,bunlar da for ile alındı
		if err := db.Select("email, username").Where("id = ?", userIDStr).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user details"})
			return
		}
		whoidM := GetCurrentUserID(c)
		var whouserM models.User
		if err := db.Select("username, email").Where("id = ?", whoidM).First(&whouserM).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current user details"})
			return
		}

		message := fmt.Sprintf(" \"%s\" isimli projede rolün %s tarafından %s olarak güncellendi", project.Name, whouserM.Email, role)
		err = models.SendNotification(projectUser.UserID, message, project.ID)
		if err != nil {
			fmt.Println("Error sending notification:", err)
		}

		whoid := GetCurrentUserID(c)
		var whouser models.User //whouser dediğimiz kişi buradaki sessionda olan ve işlemi gerçekleştiren kişi
		if err := db.Select("username, email").Where("id = ?", whoid).First(&whouser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve current user details"})
			return
		}
		go models.SendMailSimpleHTMLforRole(
			"Kanban Uygulaması",
			"./views/mail/roleSet.html",
			[]string{user.Email}, // Email gönderimi için kullanılıyor
			user.Username,        // Kullanıcı adı
			whouser.Email,        // sessiondaki userin emaili
			project.Name,         // Proje adı
			project.ID,           // Proje ID'si
			role,                 // Postformla aldığımız role kısmı
		)
	}

	library.SetAlert(c, "Ayarlar başarıyla güncellendi")
	c.Redirect(http.StatusSeeOther, "/team_space/list/"+projectID)
}
