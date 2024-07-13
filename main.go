package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Message struct {
	Title   string
	Date    string
	Content string
}

type CmdOutput struct {
	Result []byte
	Err    error
}

type TemplateRender struct {
	templates *template.Template
}

func (t *TemplateRender) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

var (
	maxPage int = 0
)

var clients = make(map[chan string]struct{})

const (
	CutDirName    = "cut"
	CutDirPath    = "./cut"
	MergeDirName  = "merge"
	MergeDirPath  = "./merge"
	UpLoadDirName = "upload"
	UpLoadDirPath = "./upload"
	PythonPath    = "/opt/venv/bin/python3"
	// PythonPath = "python3.11"
)

func main() {
	userNameFlag := flag.String("user", "user", "BasicAuth user flag")
	passwordFlag := flag.String("password", "password", "BasicAuth password flag")
	portFlag := flag.Int("port", 8080, "Port flag")

	flag.Parse()

	_, err := os.Stat(CutDirName)
	if err != nil {
		err = os.Mkdir(CutDirName, 0755)
		if err != nil {
			log.Printf("[error] main os.Mkdir cut: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = os.Stat(MergeDirName)
	if err != nil {
		err = os.Mkdir(MergeDirName, 0755)
		if err != nil {
			log.Printf("[error] main os.Mkdir merge: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = os.Stat(UpLoadDirName)
	if err != nil {
		err = os.Mkdir(UpLoadDirName, 0755)
		if err != nil {
			log.Printf("[error] main os.Mkdir upload: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = os.Stat("message.txt")
	if err != nil {
		_, err = os.Create("message.txt")
		if err != nil {
			log.Printf("[error] main os.Create message: %v\n", err)
			os.Exit(1)
		}
	} else {
		err = os.Remove("message.txt")
		if err != nil {
			log.Printf("[error] main os.Remove message: %v\n", err)
			os.Exit(1)
		}
		_, err = os.Create("message.txt")
		if err != nil {
			log.Printf("[error] main os.Create message: %v\n", err)
			os.Exit(1)
		}
	}

	cuts, err := filepath.Glob(CutDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] main filepath.Glob cut : %v\n", err)
		os.Exit(1)
	} else if len(cuts) != 0 {
		for _, f := range cuts {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove cut: %v\n", err)
				os.Exit(1)
			}
		}
	}

	merge, err := filepath.Glob(MergeDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] main filepath.Glob merge: %v\n", err)
		os.Exit(1)
	} else if len(merge) != 0 {
		for _, f := range merge {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove merge: %v\n", err)
				os.Exit(1)
			}
		}
	}

	upload, err := filepath.Glob(UpLoadDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] main filepath.Glob upload: %v\n", err)
		os.Exit(1)
	} else if len(merge) != 0 {
		for _, f := range upload {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove upload: %v\n", err)
				os.Exit(1)
			}
		}
	}

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

	e.Static("/"+MergeDirName, MergeDirName)
	e.Static("/views", "views")

	e.GET("/", Index)
	e.GET("/pdf", ShowPDF)
	e.GET("/message", ShowMessage)

	e.GET("/sse", SSE)

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
	m.POST("/addmessage", AddMessage)

	port := strconv.Itoa(*portFlag)

	e.Logger.Fatal(e.Start(":" + port))
}

func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func ShowPDF(c echo.Context) error {
	pdfPath := filepath.Join(MergeDirPath, "merge.pdf")
	if _, err := os.ReadFile(pdfPath); err != nil {
		log.Printf("[error] ShowPDF os.ReadFile merge : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintln("ファイルがまだアップロードされていません。"),
		}
		return c.Render(http.StatusNotFound, "error.html", data)
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
		log.Printf("[error] ChangeMaxPage strconv.Atoi maxPage : %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("整数以外が入力されました。: %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	merge, err := filepath.Glob(MergeDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] ChangeMaxPage filepath.Glob merge: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if len(merge) != 0 {
		for _, f := range merge {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] ChangeMaxPage os.Remove merge: %v\n", err)
				data := map[string]string{
					"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
				}
				return c.Render(http.StatusServiceUnavailable, "error.html", data)
			}
		}
	}

	ch := make(chan CmdOutput)
	cmd := exec.Command(PythonPath, "pdf-merge.py", mp)
	go func(cmd *exec.Cmd) {
		result, err := cmd.CombinedOutput()
		ch <- CmdOutput{Result: result, Err: err}
	}(cmd)
	output := <-ch
	if string(output.Result) != "Done\n" {
		log.Printf("[error] ChangeMaxPage exec.Command.CombinedOutput: %v\n", string(output.Result))
		data := map[string]string{
			"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if output.Err != nil {
		log.Printf("[error] ChangeMaxPage exec.Command.CombinedOutput: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	fmt.Printf("The maximum page has been updated. %d\n", maxPage)

	SendEvent("update")

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("最大ページを更新しました。"),
		"CurrentPage": maxPage,
	})
}

func UpLoad(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[error] UpLoad c.FormFile: %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("ファイルのアップロードに失敗しました。 %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	src, err := file.Open()
	if err != nil {
		log.Printf("[error] UpLoad file.Open: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer src.Close()

	_, err = os.Stat(UpLoadDirName)
	if err != nil {
		log.Printf("[error] UpLoad os.Stat: %v\n", err)
		err = os.Mkdir(UpLoadDirName, 0755)
		if err != nil {
			log.Printf("[error] UpLoad os.Mkdir upload: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("アップロード先のディレクトリ作成に失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "error.html", data)
		}
	}

	dst, err := os.Create(filepath.Join(UpLoadDirPath, "upload.pdf"))
	if err != nil {
		log.Printf("[error] UpLoad os.Create upload: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		log.Printf("[error] UpLoad io.Copy upload: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのコピーに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	ch := make(chan CmdOutput)
	cmd := exec.Command(PythonPath, "pdf-cut.py", UpLoadDirPath+"/upload.pdf")
	go func(cmd *exec.Cmd) {
		result, err := cmd.CombinedOutput()
		ch <- CmdOutput{Result: result, Err: err}
	}(cmd)
	output := <-ch
	if string(output.Result) != "Done\n" {
		log.Printf("[error] UpLoad exec.Command.CombinedOutput: %v\n", string(output.Result))
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのカットに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if output.Err != nil {
		log.Printf("[error] UpLoad exec.Command.CombinedOutput: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのカットに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	maxPage = 1
	src2, err := os.Open(CutDirName + "/1.pdf")
	if err != nil {
		log.Printf("[error] UpLoad os.Open cut: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer src.Close()

	dst2, err := os.Create(filepath.Join(MergeDirPath, "merge.pdf"))
	if err != nil {
		log.Printf("[error] UpLoad os.Create merge: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	_, err = io.Copy(dst2, src2)
	if err != nil {
		log.Printf("[error] UpLoad io.Copy: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのコピーに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("ファイルのアップロードが完了しました。"),
		"CurrentPage": maxPage,
	})
}

func ShowMessage(c echo.Context) error {
	file, err := os.Open("message.txt")
	if err != nil {
		log.Printf("[error] ShowMessage os.Open: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer file.Close()

	var messages []Message
	scanner := bufio.NewScanner(file)
	var currentMessage *Message
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "TITLE:") {
			if currentMessage != nil {
				messages = append(messages, *currentMessage)
			}
			currentMessage = &Message{Title: line}
		} else if strings.HasPrefix(line, "DATE:") {
			currentMessage.Date = line
		} else if currentMessage != nil {
			if currentMessage.Content != "" {
				currentMessage.Content += "\n"
			}
			currentMessage.Content += line
		}
	}
	if currentMessage != nil {
		messages = append(messages, *currentMessage)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[error] ShowMessage scanner.Err: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージスキャンに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	return c.Render(http.StatusOK, "message.html", map[string]interface{}{
		"Message": messages,
	})
}

func AddMessage(c echo.Context) error {
	title := c.FormValue("title")
	content := c.FormValue("content")

	file, err := os.OpenFile("message.txt", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("[error] AddMessage os.Open: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("メッセージの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer file.Close()

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("[error] AddMessage time.LoadLocation: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("タイムゾーン変換に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	nowJST := time.Now().In(jst)

	fmt.Fprintln(file, "TITLE: "+title)
	fmt.Fprintln(file, "DATE: "+nowJST.Format(time.DateTime))
	fmt.Fprintln(file, content)
	fmt.Fprintln(file, "")

	SendEvent("update")

	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("メッセージの追記に成功しました。"),
		"CurrentPage": maxPage,
	})
}

func SSE(c echo.Context) error {
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		log.Println("[error] SSE c.Response")
		return c.String(http.StatusInternalServerError, "Streaming unsupported")
	}

	messageChan := make(chan string)
	clients[messageChan] = struct{}{}

	defer func() {
		delete(clients, messageChan)
		close(messageChan)
	}()

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	for {
		select {
		case msg := <-messageChan:
			fmt.Fprintf(c.Response(), "data: %s\n\n", msg)
			flusher.Flush()
		case <-c.Request().Context().Done():
			return nil
		}
	}
}

func SendEvent(message string) {
	for client := range clients {
		client <- message
	}
}
