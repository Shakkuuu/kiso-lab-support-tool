package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"

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
	mutex   sync.Mutex
)

func main() {
	_, err := os.Stat("cut")
	if err != nil {
		err = os.Mkdir("cut", 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir cut: %v\n", err)
			os.Exit(0)
		}
	}

	_, err = os.Stat("merge")
	if err != nil {
		err = os.Mkdir("merge", 0755)
		if err != nil {
			log.Printf("[error] os.Mkdir merge: %v\n", err)
			os.Exit(0)
		}
	}

	cuts, err := filepath.Glob("./cut/*.pdf")
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

	merge, err := filepath.Glob("./merge/*.pdf")
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

	time.Sleep(2 * time.Second)

	cmd := exec.Command("/opt/venv/bin/python3", "pdf-cut.py", os.Args[1])
	err = cmd.Start()
	if err != nil {
		os.Exit(0)
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

	e.Static("/merge", "merge")

	e.GET("/", index)
	// e.GET("/pdf/:page", pdf)
	e.GET("/pdf", pdf)

	// goroutineで入力受付
	go func() {
		time.Sleep(2 * time.Second)
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter max page number: ")
			input, _ := reader.ReadString('\n')
			input = input[:len(input)-1] // 改行を取り除く
			if num, err := strconv.Atoi(input); err == nil {
				mutex.Lock()
				maxPage = num
				mutex.Unlock()

				merge, err := filepath.Glob("./merge/*.pdf")
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
				page := strconv.Itoa(maxPage)
				cmd := exec.Command("python3.11", "pdf-merge.py", page)
				err = cmd.Start()
				if err != nil {
					log.Printf("[error] exec.Command: %v\n", err)
				}
				fmt.Printf("The maximum page has been updated. %d\n", maxPage)
			} else {
				fmt.Println("Please enter an integer.")
			}
		}
	}()

	// サーバーの起動
	go func() {
		e.Logger.Fatal(e.Start(":8080"))
	}()

	// メインgoroutineを停止させない
	select {}
}

func index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func pdf(c echo.Context) error {
	pdfPath := filepath.Join("./merge", "merge.pdf")
	if _, err := os.ReadFile(pdfPath); err != nil {
		data := map[string]string{
			"Message": fmt.Sprintln("ページが見つかりませんでした"),
		}
		return c.Render(http.StatusNotFound, "message.html", data)
	}

	return c.Render(http.StatusOK, "pdf-view.html", map[string]interface{}{
		"PDFPath": "/merge/merge.pdf",
	})
}

// func management(c echo.Context) error {

// }
