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

type PagePath struct {
	Path   string
	Number string
}

type MaxPageForm struct {
	MaxPage int `form:"maxpage" validate:"required,min=1,max=10000"`
}

type MessageForm struct {
	Title   string `form:"title" validate:"required,min=1,max=50"`
	Content string `form:"content" validate:"required,min=1,max=10000"`
}
