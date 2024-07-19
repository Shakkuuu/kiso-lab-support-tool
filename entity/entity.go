package entity

import "time"

type Message struct {
	ID      int `gorm:"primary_key"`
	Title   string
	Content string
	Date    time.Time
}

type ViewMessage struct {
	ID      int `gorm:"primary_key"`
	Title   string
	Content string
	Date    string
}

type CmdOutput struct {
	Result []byte
	Err    error
}
