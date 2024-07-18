package controller

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"kiso-lab-support-tool/model"

	"github.com/labstack/echo"
)

type MessageController struct{}

var mm model.MessageModel

func (mc MessageController) ShowMessage(c echo.Context) error {
	messages, err := mm.GetAll()
	if err != nil {
		log.Printf("[error] ShowMessage mm.GetAll: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの取得に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	return c.Render(http.StatusOK, "message.html", map[string]interface{}{
		"Message":    messages,
		"Management": false,
	})
}

func (mc MessageController) ManagementMessage(c echo.Context) error {
	messages, err := mm.GetAll()
	if err != nil {
		log.Printf("[error] ShowMessage mm.GetAll: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの取得に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	return c.Render(http.StatusOK, "message.html", map[string]interface{}{
		"Message":    messages,
		"Management": true,
	})
}

func (mc MessageController) AddMessage(c echo.Context) error {
	title := c.FormValue("title")
	content := c.FormValue("content")

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("[error] AddMessage time.LoadLocation: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("タイムゾーン変換に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	nowJST := time.Now().In(jst)

	err = mm.Create(title, nowJST.Format(time.DateTime), content)
	if err != nil {
		log.Printf("[error] AddMessage mm.Create: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの追加に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	SendEvent("update")

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("メッセージの送信に成功しました。"),
		"CurrentPage": maxPage,
	})
}

func (mc MessageController) DeleteMessage(c echo.Context) error {
	id := c.Param("id")

	intID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("[error] DeleteMessage strconv.Atoi : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの削除に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	err = mm.Delete(intID)
	if err != nil {
		log.Printf("[error] DeleteMessage mm.Delete : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの削除に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	SendEvent("update")

	return c.Redirect(http.StatusOK, "/message")
}