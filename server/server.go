package server

import (
	"io"
	"kiso-lab-support-tool/controller"
	"net/http"
	"strconv"
	"text/template"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type TemplateRender struct {
	templates *template.Template
}

var (
	pc controller.PDFController
	mc controller.MessageController
)

func (t *TemplateRender) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Init(un, pw string, p int) {
	echo.NotFoundHandler = func(c echo.Context) error {
		return c.Render(http.StatusNotFound, "404.html", nil)
	}

	echo.MethodNotAllowedHandler = func(c echo.Context) error {
		return c.Render(http.StatusMethodNotAllowed, "405.html", nil)
	}

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
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

	renderer := &TemplateRender{
		templates: template.Must(template.ParseGlob("./views/*.html")),
	}
	e.Renderer = renderer

	e.Static("/"+controller.MergeDirName, controller.MergeDirName)
	e.Static("/views", "views")

	e.GET("/", controller.Index)
	e.GET("/pdf", pc.ShowPDF)
	e.GET("/message", mc.ShowMessage)

	e.GET("/sse", controller.SSE)

	m := e.Group("/management")

	m.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == un && password == pw {
			return true, nil
		}
		return false, c.Render(http.StatusUnauthorized, "401.html", nil)
	}))

	m.GET("", controller.Management)
	m.POST("/maxpage", pc.ChangeMaxPage)
	m.POST("/upload", pc.UpLoad)
	m.POST("/addmessage", mc.AddMessage)
	m.GET("/message", mc.ManagementMessage)
	m.GET("/deletemessage/:id", mc.DeleteMessage)

	port := strconv.Itoa(p)

	e.Logger.Fatal(e.Start(":" + port))
}
