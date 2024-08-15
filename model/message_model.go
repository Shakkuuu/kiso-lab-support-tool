package model

import (
	"kiso-lab-support-tool/db"
	"kiso-lab-support-tool/entity"
	"time"
)

type MessageModel struct{}

// メッセージ全取得
func (mm MessageModel) GetAll() ([]entity.Message, error) {
	var messages []entity.Message

	// dbインスタンス取得
	db := db.GetDB()

	// メッセージ全取得
	err := db.Find(&messages).Error
	if err != nil {
		return messages, err
	}

	return messages, nil
}

// メッセージ作成
func (mm MessageModel) Create(title, content string, date time.Time) error {
	var message entity.Message

	// dbインスタンス取得
	db := db.GetDB()

	message.Title = title
	message.Content = content
	message.Date = date

	// メッセージ作成
	err := db.Create(&message).Error
	if err != nil {
		return err
	}

	return nil
}

// メッセージ削除
func (mm MessageModel) Delete(id int) error {
	var message entity.Message

	// dbインスタンス取得
	db := db.GetDB()

	// 指定したidに一致するメッセージを削除
	err := db.Where("id = ?", id).Delete(&message).Error
	if err != nil {
		return err
	}

	return nil
}
