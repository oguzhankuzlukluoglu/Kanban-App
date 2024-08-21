package routers

import (
	"kanban/config"
	"kanban/controllers"
	"kanban/models"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func Load(router *gin.Engine) {
	router.SetHTMLTemplate(controllers.GetTemplates())
	store := cookie.NewStore([]byte(config.GetConfig().SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	})
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	secretKey := os.Getenv("SECRET_KEY")

	store = cookie.NewStore([]byte(secretKey))
	router.Use(sessions.Sessions("kanban", store))
	router.Use(gin.Recovery())

	router.GET("/login", controllers.Userops{}.Index)
	router.POST("/login/do-login", controllers.Userops{}.Login)
	router.POST("/login/do-signup", controllers.Userops{}.Signup)
	router.GET("/logout", controllers.Userops{}.Logout)

	// Authenticated routes
	authRoutes := router.Group("/")
	authRoutes.Use(controllers.AuthMiddleware)
	{
		authRoutes.GET("/", controllers.Dashboard{}.HomeIndex)
		authRoutes.GET("/inbox", controllers.Dashboard{}.InboxIndex)
		authRoutes.GET("/settings", controllers.Dashboard{}.SettingsIndex)
		authRoutes.GET("/share", controllers.Dashboard{}.ShareIndex)
		authRoutes.GET("/everything", controllers.Dashboard{}.EverythingIndex)

		authRoutes.GET("/team_space/:ID", controllers.Dashboard{}.TeamSpaceDetailsById)
		authRoutes.GET("/team_space/list/:ID", controllers.Dashboard{}.TeamSpaceDetailsListById)
		authRoutes.GET("/team_space/board/:ID", controllers.Dashboard{}.TeamSpaceDetailsBoardById)
		authRoutes.GET("/team_space/table/:ID", controllers.Dashboard{}.TeamSpaceDetailsTableById)
		authRoutes.GET("/GetCommentsByIssue", models.GetCommentByIssue)

		authRoutes.POST("/user-edit", controllers.Userops{}.UserUpdate)
		authRoutes.POST("/user-delete", controllers.Userops{}.UserDelete)
		authRoutes.POST("/password-change", controllers.Userops{}.PasswordChange)

		authRoutes.POST("/workSpaces-yeni-ekle", controllers.Project{}.WorkSpaceAdd)
		authRoutes.POST("/workSpaces-edit", controllers.Project{}.WorkSpaceNameUpdate)
		authRoutes.POST("/workSpaces-delete", controllers.Project{}.WorkSpaceDelete)

		authRoutes.POST("/project-yeni-ekle", controllers.Project{}.ProjectAdd)
		authRoutes.POST("/project/set-user-role", controllers.SetUserRole)
		authRoutes.POST("/project-user-ekle", controllers.Project{}.ProjectUserAdd)
		authRoutes.POST("/project-edit", controllers.Project{}.ProjectUpdate)
		authRoutes.POST("/project-delete", controllers.Project{}.ProjectDelete)

		authRoutes.POST("/issue-yeni-ekle", controllers.Issue{}.IssueAdd)
		authRoutes.POST("/issue-edit", controllers.Issue{}.IssueEdit)
		authRoutes.POST("/issue-delete", controllers.Issue{}.IssueDelete)

		authRoutes.POST("/comment-ekle", controllers.AddCommentHandler)

		authRoutes.POST("/update-issue-status", controllers.Dashboard{}.UpdateIssueStatus)
		authRoutes.POST("/notification-read", controllers.Dashboard{}.NotificationAsRead)
		authRoutes.POST("/notifications-read-all", controllers.MarkAllNotificationsReadHandler)
		authRoutes.POST("/notification/:id", controllers.DeleteNotification)
		authRoutes.POST("/notification/delete-all-noti", controllers.DeleteAllNotification)
		authRoutes.POST("/deleteUser", controllers.ModifiedDeleteProjectUser)
		authRoutes.POST("delete-comment", controllers.DeleteCommentHandler)

	}

	// Static files
	router.Static("/assets", "./assets")
	router.Static("/uploads", "./uploads")
}
