package entity

import "time"

// メッセージ（DB保存用）
type Message struct {
	ID      int `gorm:"primary_key"`
	Title   string
	Content string
	Date    time.Time
}

// メッセージ（クライアント用）
type ViewMessage struct {
	ID      int
	Title   string
	Content string
	Date    string
}

// Pythonスクリプトの実行結果を一時的に格納
type CmdOutput struct {
	Result []byte
	Err    error
}

// 資料のページパスをクライアントに渡す用
type PagePath struct {
	Path   string
	Number string
}

// Formからきた最大ページのバリデーション用
type MaxPageForm struct {
	MaxPage int `form:"maxpage" validate:"required,min=1,max=10000"`
}

// Formからきたメッセージのバリデーション用
type MessageForm struct {
	Title   string `form:"title" validate:"required,min=1,max=50"`
	Content string `form:"content" validate:"required,min=1,max=10000"`
}
