package models

import (
	"gorm.io/gorm"
)

type WorkSpace struct {
	gorm.Model
	Name         string `json:"name"`
	UserID       uint
	Issues       []Issue       `gorm:"foreignKey:WorkSpaceID"`
	ProjectUsers []ProjectUser `gorm:"foreignKey:WorkSpaceID"`
	Project      []Project     `gorm:"foreignKey:WorkSpaceID"`
}

func (WorkSpace) TableName() string {
	return "work_spaces"
}

func (workSpace WorkSpace) Migrate() {
	db := GetDB()

	db.AutoMigrate(&workSpace) 

}

func (workSpace *WorkSpace) Add(userID uint) error {
	db := GetDB()

	workSpace.UserID = userID
	if err := db.Create(&workSpace).Error; err != nil {
		return err
	}
	return nil
}

func (workSpace WorkSpace) Get(where ...interface{}) WorkSpace {
	db := GetDB()

	db.Where(where[0], where[1:]...).First(&workSpace)
	return workSpace
}

func (workSpace WorkSpace) GetAll(where ...interface{}) []WorkSpace {
	db := GetDB()

	var workSpaces []WorkSpace
	db.Preload("Project").Preload("Project.Issues").Find(&workSpaces, where...)
	return workSpaces
}

func (workSpace WorkSpace) Update(column string, value interface{}) {
	db := GetDB()

	db.Model(workSpace).Update(column, value)
}

func (workSpace WorkSpace) Updates(data WorkSpace) {
	db := GetDB()

	db.Model(workSpace).Updates(data)
}

func (workSpace WorkSpace) Delete() {
	db := GetDB()

	db.Delete(workSpace, workSpace.ID)
}
