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

// DBとの接続
func Init() {
	// sqliteのDBファイル展開
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.Printf("db Init error: %v\n", err)
		panic("failed to connect database")
	}

	// メッセージの構造体をマイグレーション（テーブル作成）
	db.AutoMigrate(&entity.Message{})
}

// DBのインスタンスを他パッケージに渡す用
func GetDB() *gorm.DB {
	return db
}

// DBとの接続を閉じる用
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
