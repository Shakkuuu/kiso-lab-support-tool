package controller

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

// 各種ディレクトリ名とPython実行環境を指定
const (
	CutDirName          = "cut" // 分割した資料を入れるディレクトリ名
	CutDirPath          = "./cut"
	ViewDocumentDirName = "view-document" // クライアントに公開するディレクトリ名
	ViewDocumentDirPath = "./view-document"
	UpLoadDirName       = "upload" // アップロードされた資料を入れておくディレクトリ名
	UpLoadDirPath       = "./upload"
	PythonPath          = "/opt/venv/bin/python3" // Pythonの実行環境（実行コマンド）指定
	// PythonPath = "python3.11"
)

var (
	clients      = make(map[chan string]struct{}) // SSEのクライアントを管理するmap
	clientsMutex sync.Mutex                       // mapへの同時書き込みを制限する用
)

// indexページ表示
func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

// Managementページ表示
func Management(c echo.Context) error {
	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"CurrentPage": maxPage,
	})
}

// SSE接続とクライアント追加
func SSE(c echo.Context) error {
	// Fulusherを取得
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		log.Println("[error] SSE c.Response")
		return c.String(http.StatusInternalServerError, "Streaming unsupported")
	}

	messageChan := make(chan string)
	// 書き込みのためmapをロック
	clientsMutex.Lock()
	// クライアント追加
	clients[messageChan] = struct{}{}
	// ロック解除
	clientsMutex.Unlock()

	// クライアントとの接続終了後mapから削除
	defer func() {
		// 書き込みのためmapをロック
		clientsMutex.Lock()
		// クライアント削除
		delete(clients, messageChan)
		// ロック解除
		clientsMutex.Unlock()
		// チャネルを閉じる
		close(messageChan)
	}()

	// SSE用のヘッダー設定
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	// イベントの待機
	for {
		select {
		case msg := <-messageChan: // イベントが来た場合
			// レスポンスに書き込み
			fmt.Fprintf(c.Response(), "data: %s\n\n", msg)
			// クライアントに送信
			flusher.Flush()
		case <-c.Request().Context().Done(): // 接続が終了した場合
			// ループから抜ける
			return nil
		}
	}
}

// イベント送信
func SendEvent(message string) {
	// 書き込みのためmapをロック
	clientsMutex.Lock()
	// 関数終了後にロック解除
	defer clientsMutex.Unlock()
	// 各クライアントにイベント送信
	for client := range clients {
		client <- message
	}
}
