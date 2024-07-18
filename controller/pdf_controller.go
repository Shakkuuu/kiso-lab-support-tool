package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"kiso-lab-support-tool/entity"

	"github.com/labstack/echo"
)

type PDFController struct{}

var (
	maxPage int = 0
)

func (pc PDFController) ShowPDF(c echo.Context) error {
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

func (pc PDFController) ChangeMaxPage(c echo.Context) error {
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

	ch := make(chan entity.CmdOutput)
	cmd := exec.Command(PythonPath, "pdf-merge.py", mp, MergeDirName+"/merge.pdf", CutDirName)
	go func(cmd *exec.Cmd) {
		result, err := cmd.CombinedOutput()
		ch <- entity.CmdOutput{Result: result, Err: err}
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

func (pc PDFController) UpLoad(c echo.Context) error {
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

	ch := make(chan entity.CmdOutput)
	cmd := exec.Command(PythonPath, "pdf-cut.py", UpLoadDirPath+"/upload.pdf", CutDirName)
	go func(cmd *exec.Cmd) {
		result, err := cmd.CombinedOutput()
		ch <- entity.CmdOutput{Result: result, Err: err}
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
