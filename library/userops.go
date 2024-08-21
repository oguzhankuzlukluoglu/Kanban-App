package library

import (
	"kanban/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func SetUser(c *gin.Context, email, password string) error {
	session, err := store.Get(c.Request, "kanban")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	session.Values["email"] = email
	session.Values["password"] = password

	if err := session.Save(c.Request, c.Writer); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func CheckUser(w http.ResponseWriter, r *http.Request) bool {
	session, err := store.Get(r, "kanban")
	if err != nil {
		println(err)
		return false
	}
	username := session.Values["username"]
	password := session.Values["password"]

	user := models.User{}.Get("username = ? AND password = ?", username, password)

	if (user.Username == username) && (user.Password == password) {
		return true
	}
	return false
}

func RemoveUser(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "kanban")
	if err != nil {
		println(err)
		return err
	}

	session.Options.MaxAge = -1

	return sessions.Save(r, w)
}
