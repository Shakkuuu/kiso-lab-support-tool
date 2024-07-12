package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRender struct {
	templates *template.Template
}

func (t *TemplateRender) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var (
	maxPage int
)

const (
	CutDirName    = "cut"
	CutDirPath    = "./cut"
	MergeDirName  = "merge"
	MergeDirPath  = "./merge"
	UpLoadDirName = "upload"
	UpLoadDirPath = "./upload"
	// PythonPath    = "/opt/venv/bin/python3"
	PythonPath = "python3.11"
)

func main() {
	userNameFlag := flag.String("user", "user", "BasicAuth user flag")
	passwordFlag := flag.String("password", "password", "BasicAuth password flag")

	flag.Parse()

	_, err := os.Stat(CutDirName)
	if err != nil {
		err = os.Mkdir(CutDirName, 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir cut: %v\n", err)
			os.Exit(0)
		}
	}

	_, err = os.Stat(MergeDirName)
	if err != nil {
		err = os.Mkdir(MergeDirName, 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir merge: %v\n", err)
			os.Exit(0)
		}
	}

	_, err = os.Stat(UpLoadDirName)
	if err != nil {
		err = os.Mkdir(UpLoadDirName, 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir upload: %v\n", err)
			os.Exit(0)
		}
	}

	cuts, err := filepath.Glob(CutDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] filepath.Glob cut : %v\n", err)
	} else if len(cuts) != 0 {
		for _, f := range cuts {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] os.Remove cut: %v\n", err)
			}
		}
	}

	merge, err := filepath.Glob(MergeDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] filepath.Glob merge: %v\n", err)
	} else if len(merge) != 0 {
		for _, f := range merge {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] os.Remove merge: %v\n", err)
			}
		}
	}

	upload, err := filepath.Glob(UpLoadDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] filepath.Glob upload: %v\n", err)
	} else if len(merge) != 0 {
		for _, f := range upload {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] os.Remove upload: %v\n", err)
			}
		}
	}

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `time: ${time_rfc3339_nano}` + ", " +
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

	e.Static("/"+MergeDirName, MergeDirName)

	e.GET("/", Index)
	e.GET("/pdf", ShowPDF)

	m := e.Group("/management")

	m.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == *userNameFlag && password == *passwordFlag {
			return true, nil
		}
		return false, nil
	}))

	m.GET("", Management)
	m.POST("/maxpage", ChangeMaxPage)
	m.POST("/upload", UpLoad)

	e.Logger.Fatal(e.Start(":8080"))
}

func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func ShowPDF(c echo.Context) error {
	pdfPath := filepath.Join(MergeDirPath, "merge.pdf")
	if _, err := os.ReadFile(pdfPath); err != nil {
		data := map[string]string{
			"Message": fmt.Sprintln("ページが見つかりませんでした"),
		}
		return c.Render(http.StatusNotFound, "message.html", data)
	}

	return c.Render(http.StatusOK, "pdf-view.html", map[string]interface{}{
		"PDFPath": "/" + MergeDirName + "/merge.pdf",
	})
}

func Management(c echo.Context) error {
	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"CurrentPage": maxPage,
	})
}

func ChangeMaxPage(c echo.Context) error {
	mp := c.FormValue("maxpage")

	var err error
	maxPage, err = strconv.Atoi(mp)
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("整数以外が入力されました。: %v\n", err),
		}
		return c.Render(http.StatusNotFound, "management.html", data)
	}

	merge, err := filepath.Glob(MergeDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] filepath.Glob : %v\n", err)
	} else if len(merge) != 0 {
		for _, f := range merge {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] os.Remove merge: %v\n", err)
			}
		}
	}
	cmd := exec.Command(PythonPath, "pdf-merge.py", mp)
	err = cmd.Start()
	if err != nil {
		log.Printf("[error] exec.Command: %v\n", err)
	}
	fmt.Printf("The maximum page has been updated. %d\n", maxPage)

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("最大ページを更新しました。"),
		"CurrentPage": maxPage,
	})
}

func UpLoad(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのアップロードに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusBadRequest, "message.html", data)
	}

	src, err := file.Open()
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "message.html", data)
	}
	defer src.Close()

	_, err = os.Stat(UpLoadDirName)
	if err != nil {
		err = os.Mkdir(UpLoadDirName, 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir upload: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("アップロード先のディレクトリ作成に失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "message.html", data)
		}
	}

	dst, err := os.Create(filepath.Join(UpLoadDirPath, "upload.pdf"))
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "message.html", data)
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのコピーに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "message.html", data)
	}

	cmd := exec.Command(PythonPath, "pdf-cut.py", UpLoadDirPath+"/upload.pdf")
	err = cmd.Start()
	if err != nil {
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのカットに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "message.html", data)
	}

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("ファイルのアップロードが完了しました。"),
		"CurrentPage": maxPage,
	})
}
