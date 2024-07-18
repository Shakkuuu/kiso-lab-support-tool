package db

import (
	"kiso-lab-support-tool/entity"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	err error
)

func Init() {
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Printf("db Init error: %v\n", err)
		panic("failed to connect database")
	}

	db.AutoMigrate(&entity.Message{})
}

func GetDB() *gorm.DB {
	return db
}

func Close() {
	if sqlDB, err := db.DB(); err != nil {
		log.Printf("db Close error: %v\n", err)
		panic(err)
	} else {
		if err := sqlDB.Close(); err != nil {
			log.Printf("db Close error: %v\n", err)
			panic(err)
		}
	}
}
