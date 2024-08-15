package server

import (
	"io"
	"kiso-lab-support-tool/controller"
	"net/http"
	"strconv"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	dc controller.DocumentController
	mc controller.MessageController
)

// HTMLテンプレートのレンダリング
type TemplateRender struct {
	templates *template.Template
}

func (t *TemplateRender) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

// サーバ実行
func Init(un, pw string, p int) {
	// URLパスが存在しなかった時に返すページ指定
	echo.NotFoundHandler = func(c echo.Context) error {
		return c.Render(http.StatusNotFound, "404.html", nil)
	}

	// 対応外のメソッドでアクセスされた時に返すページ指定
	echo.MethodNotAllowedHandler = func(c echo.Context) error {
		return c.Render(http.StatusMethodNotAllowed, "405.html", nil)
	}

	// echoのインスタンス作成
	e := echo.New()

	// 何かしらでpanicが起きた際に復帰できるようにミドルウェア設定
	e.Use(middleware.Recover())

	// HTMLテンプレートを入れてあるディレクトリを指定
	renderer := &TemplateRender{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
	e.Renderer = renderer

	// 認証無しURLパスグループ作成
	access := e.Group("")

	// アクセスログ用ログ出力のカスタマイズミドルウェア作成
	access.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[access] ` +
			`time: ${time_rfc3339_nano}` + ", " +
			`method: ${method}` + ", " +
			`remote_ip: ${remote_ip}` + ", " +
			`host: ${host}` + ", " +
			`uri: ${uri}` + ", " +
			`status: ${status}` + ", " +
			`error: ${error}` + ", " +
			`latency: ${latency}(${latency_human})` + "\n",
	}))

	// indexページ
	access.GET("/", controller.Index)
	// 資料表示（パスパラメータでページ番号指定）
	access.GET("/document/:currentPage", dc.ShowDocument)
	// メッセージ表示
	access.GET("/message", mc.ShowMessage)

	// SSEクライアントに参加
	access.GET("/sse", controller.SSE)

	// 静的ファイル配信グループ作成
	static := e.Group("/static")

	// 静的ファイル用ログ出力のカスタマイズミドルウェア作成
	static.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `[static] ` +
			`time: ${time_rfc3339_nano}` + ", " +
			`method: ${method}` + ", " +
			`remote_ip: ${remote_ip}` + ", " +
			`host: ${host}` + ", " +
			`uri: ${uri}` + ", " +
			`status: ${status}` + ", " +
			`error: ${error}` + ", " +
			`latency: ${latency}(${latency_human})` + "\n",
	}))

	// クライアントに表示させるページのjpgファイルの配信
	static.Static("/"+controller.ViewDocumentDirName, controller.ViewDocumentDirName)
	// HTMLテンプレートなどの静的ファイルの配信
	static.Static("/views", "views")

	// 認証を必要とするURLパスグループ作成
	m := access.Group("/management")

	// managementグループに対するBasic認証のミドルウェア
	m.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		// ユーザ名とパスワードが一致するか確認（ユーザ名とパスワードはフラグで指定しているものをmainからもらっている）
		if username == un && password == pw {
			return true, nil
		}
		// 正しくなかった場合は、認証エラーページを返す
		return false, c.Render(http.StatusUnauthorized, "401.html", nil)
	}))

	// 管理画面表示
	m.GET("", controller.Management)
	// 最大ページ更新
	m.POST("/maxpage", dc.ChangeMaxPage)
	// 資料のアップロード
	m.POST("/upload", dc.UpLoad)
	// メッセージ追加
	m.POST("/addmessage", mc.AddMessage)
	// メッセージ表示（管理版）
	m.GET("/message", mc.ManagementMessage)
	// メッセージ削除（パスパラメータでメッセージのID指定）
	m.GET("/deletemessage/:id", mc.DeleteMessage)

	// mainからもらったフラグで指定されたポート番号をstringに変換
	port := strconv.Itoa(p)

	// 指定したポートでサーバ起動
	e.Logger.Fatal(e.Start(":" + port))
}
