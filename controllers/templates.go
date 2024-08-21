package controllers

import (
	"fmt"
	"html/template"
	"kanban/library"
	"kanban/models"
	"time"
)

var tmpl *template.Template

func LoadTemplates() {

	funcMap := template.FuncMap{
		"add":             Count,
		"formatDate":      formatDate,
		"getProjectEmail": getProjectEmail,
	}

	var files []string
	files = append(files, library.Include("everything")...)

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		panic(err)
	}
}

func Count(count int) int {
	count = count + 1
	return count
}

func formatDate(t time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d", t.Day(), t.Month(), t.Year())
}

func getProjectEmail(projectID uint) []string {
	var data []string

	projectUsers := models.ProjectUser{}.GetAll("project_id = ?", projectID)

	for _, projectUser := range projectUsers {
		user := models.User{}.Get("id = ?", projectUser.UserID)
		if user.ID != 0 && user.Email != "" {
			data = append(data, user.Email)
		}
	}

	return data
}

func GetTemplates() *template.Template {
	return tmpl
}
