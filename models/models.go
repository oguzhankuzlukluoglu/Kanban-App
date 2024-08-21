package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB
var dbMigrate *gorm.DB

func SetDB(connection string) {
	var err error
	db, err = gorm.Open("mysql", connection)
	if err != nil {
		panic(err)
	}
	db.SingularTable(true)
}
func SetDBMigrate(connection string) {
	var err error
	dbMigrate, err = gorm.Open("mysql", connection)
	if err != nil {
		panic(err)
	}
	dbMigrate.SingularTable(true)
}

func GetDB() *gorm.DB {
	return db
}

func GetDBMigrate() *gorm.DB {
	return dbMigrate
}
