package library

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("123456789"))

func SetAlert(c *gin.Context, message string) error {
	session, err := store.Get(c.Request, "go-alert")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	session.AddFlash(message)
	err = session.Save(c.Request, c.Writer)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return err
	}
	return nil
}

func GetAlert(c *gin.Context) map[string]interface{} {
	session, err := store.Get(c.Request, "go-alert")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return nil
	}

	data := make(map[string]interface{})
	flashes := session.Flashes()

	if len(flashes) > 0 {
		data["is_alert"] = true
		data["message"] = flashes[0]
	} else {
		data["is_alert"] = false
		data["message"] = nil
	}

	err = session.Save(c.Request, c.Writer)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return nil
	}

	return data
}
