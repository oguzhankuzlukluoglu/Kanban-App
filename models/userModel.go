package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model

	Username, Password, Email, Bio string
	Project                        []Project   `gorm:"foreignKey:UserID"`
	WorkSpace                      []WorkSpace `gorm:"foreignKey:UserID"`
	Issue                          []Issue     `gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

func (user User) Migrate() {
	db := GetDB()

	db.AutoMigrate(&user)
}

func (user User) Add() {
	db := GetDB()

	db.Create(&user)
}

func (user User) Get(where ...interface{}) User {
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&user)
	return user
}

func (user User) GetAll(where ...interface{}) []User {
	db := GetDB()

	var users []User
	db.Find(&users, where...)
	return users
}

func (user User) Update(colum string, value interface{}) {
	db := GetDB()

	db.Model(&user).Update(colum, value)
}

func (user User) Updates(data User) {
	db := GetDB()

	db.Model(&user).Updates(data)
}

func (user User) Delete() {
	db := GetDB()

	db.Delete(&user, user.ID)
}
