package controller

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"kiso-lab-support-tool/entity"
	"kiso-lab-support-tool/model"

	"github.com/labstack/echo/v4"
	"github.com/microcosm-cc/bluemonday"
)

type MessageController struct{}

var mm model.MessageModel

// メッセージ表示
func (mc MessageController) ShowMessage(c echo.Context) error {
	// Modelからメッセージを全取得
	messages, err := mm.GetAll()
	if err != nil {
		log.Printf("[error] ShowMessage mm.GetAll: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの取得に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// 最新のメッセージが上に来るように並び替え
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})

	// Dateをフォーマット変換して再度スライスに格納
	var viewMessages []entity.ViewMessage
	for _, message := range messages {
		viewMessage := entity.ViewMessage{
			ID:      message.ID,
			Title:   message.Title,
			Content: message.Content,
			Date:    message.Date.Format(time.DateTime),
		}
		viewMessages = append(viewMessages, viewMessage)
	}

	return c.Render(http.StatusOK, "message.html", map[string]interface{}{
		"Message":    viewMessages,
		"Management": false,
	})
}

// Management用のメッセージ表示
func (mc MessageController) ManagementMessage(c echo.Context) error {
	// Modelからメッセージを全取得
	messages, err := mm.GetAll()
	if err != nil {
		log.Printf("[error] ShowMessage mm.GetAll: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの取得に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// 最新のメッセージが上に来るように並び替え
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Date.After(messages[j].Date)
	})

	// Dateをフォーマット変換して再度スライスに格納
	var viewMessages []entity.ViewMessage
	for _, message := range messages {
		viewMessage := entity.ViewMessage{
			ID:      message.ID,
			Title:   message.Title,
			Content: message.Content,
			Date:    message.Date.Format(time.DateTime),
		}
		viewMessages = append(viewMessages, viewMessage)
	}

	// Managementをtrueにして、HTMLテンプレートでメッセージにDeleteボタンを出現するようにする
	return c.Render(http.StatusOK, "message.html", map[string]interface{}{
		"Message":    viewMessages,
		"Management": true,
	})
}

// メッセージ追加
func (mc MessageController) AddMessage(c echo.Context) error {
	// メッセージインスタンス作成
	messageForm := new(entity.MessageForm)
	// Formから来たメッセージをバインド
	err := c.Bind(messageForm)
	if err != nil {
		log.Printf("[error] AddMessage c.Bind : %v\n", err)
		data := map[string]interface{}{
			"Message": fmt.Sprintf("Formの取得に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// バリデーションチェック
	err = validate.Struct(messageForm)
	if err != nil {
		log.Printf("[error] AddMessage validate.Struct : %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("タイトルは1文字以上50文字以下、コンテンツは1文字以上10000文字以下にしてください。: %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	// タイトルに不正な文字列が入れられないように変換
	policy := bluemonday.UGCPolicy()
	safeTitle := policy.Sanitize(messageForm.Title)

	// 本文でHTMLのコードを入れるとブラウザが解釈してクライアントに表示されてしまうため、そのまま文字列として表示できるように変換
	escapedContent := template.HTMLEscapeString(messageForm.Content)

	// タイムゾーン設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("[error] AddMessage time.LoadLocation: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("タイムゾーン変換に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	// 現在の日時を取得
	nowJST := time.Now().In(jst)

	// Modelでメッセージ作成
	err = mm.Create(safeTitle, escapedContent, nowJST)
	if err != nil {
		log.Printf("[error] AddMessage mm.Create: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの追加に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// メッセージが追加されたことを、Messageページを開いているクライアントに告知（SSE送信）
	SendEvent("MessageUpdate")

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("メッセージの送信に成功しました。"),
		"CurrentPage": maxPage,
	})
}

// メッセージ削除
func (mc MessageController) DeleteMessage(c echo.Context) error {
	// パスパラメータからメッセージのIDを取得
	id := c.Param("id")

	// IDをintに変換
	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[error] DeleteMessage strconv.Atoi : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの削除に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// Modelでメッセージを削除
	err = mm.Delete(intID)
	if err != nil {
		log.Printf("[error] DeleteMessage mm.Delete : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの削除に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// メッセージが削除されたことを、Messageページを開いているクライアントに告知（SSE送信）
	SendEvent("MessageUpdate")

	// Messageページにリダイレクト
	return c.Redirect(http.StatusOK, "/message")
}
