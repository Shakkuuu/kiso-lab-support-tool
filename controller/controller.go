package controller

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
)

const (
	CutDirName     = "cut"
	CutDirPath     = "./cut"
	ViewPDFDirName = "view-pdf"
	ViewPDFDirPath = "./view-pdf"
	UpLoadDirName  = "upload"
	UpLoadDirPath  = "./upload"
	PythonPath     = "/opt/venv/bin/python3"
	// PythonPath = "python3.11"
)

var (
	clients      = make(map[chan string]struct{})
	clientsMutex sync.Mutex
)

func Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", nil)
}

func Management(c echo.Context) error {
	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
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
	clientsMutex.Lock()
	clients[messageChan] = struct{}{}
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients, messageChan)
		clientsMutex.Unlock()
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
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		client <- message
	}
}
