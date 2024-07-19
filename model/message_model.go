package model

import (
	"kiso-lab-support-tool/db"
	"kiso-lab-support-tool/entity"
	"time"
)

type MessageModel struct{}

func (mm MessageModel) GetAll() ([]entity.Message, error) {
	var messages []entity.Message

	db := db.GetDB()

	err := db.Find(&messages).Error
	if err != nil {
		return messages, err
	}

	return messages, nil
}

func (mm MessageModel) Create(title, content string, date time.Time) error {
	var message entity.Message

	db := db.GetDB()

	message.Title = title
	message.Content = content
	message.Date = date

	err := db.Create(&message).Error
	if err != nil {
		return err
	}

	return nil
}

func (mm MessageModel) Delete(id int) error {
	var message entity.Message

	db := db.GetDB()

	err := db.Where("id = ?", id).Delete(&message).Error
	if err != nil {
		return err
	}

	return nil
}
