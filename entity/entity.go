package entity

type Message struct {
	ID      int `gorm:"primary_key"`
	Title   string
	Date    string
	Content string
}

type CmdOutput struct {
	Result []byte
	Err    error
}
