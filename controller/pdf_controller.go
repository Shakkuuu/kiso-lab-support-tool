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

	"github.com/labstack/echo/v4"

	"github.com/go-playground/validator/v10"
)

type DocumentController struct{}

var (
	// 現在の最大ページ
	maxPage int = 0
	// バリデータのインスタンス作成
	validate = validator.New()
)

// 資料表示
func (dc DocumentController) ShowDocument(c echo.Context) error {
	// パスパラメータで指定ページを取得
	currentPage := c.Param("currentPage")

	// 指定ページをintに変換
	intCurrentPage, err := strconv.Atoi(currentPage)
	if err != nil {
		log.Printf("[error] ShowDocument strconv.Atoi : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintln("ページを正しく指定してください。"),
		}
		return c.Render(http.StatusBadRequest, "error.html", data)
	}

	// 資料がアップロードされているか確認
	_, err = os.Stat(ViewDocumentDirPath + "/1.jpg")
	if err != nil {
		log.Printf("[error] ShowDocument os.Stat : %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintln("ファイルがまだアップロードされていません。"),
		}
		return c.Render(http.StatusNotFound, "error.html", data)
	}

	// 指定されたページ番号が最大ページを超えていないか確認（指定したページのjpgファイルが存在するかで範囲外を指定していないか確認）
	_, err = os.Stat(ViewDocumentDirPath + "/" + currentPage + ".jpg")
	if err != nil {
		log.Printf("[error] ShowDocument os.Stat : %v\n", err)
		log.Println("範囲外のページが選択されました。")
		// 範囲外の場合は強制的に1ページ目を表示
		currentPage = "1"
	}

	// 指定したページの静的ファイル資料を取得するためのパス作成
	currentPagePath := "/static" + "/" + ViewDocumentDirName + "/" + currentPage + ".jpg"

	// 最大ページまでの資料取得のパスも作成
	var pagePathList []entity.PagePath
	for i := 1; i <= maxPage; i++ {
		filePath := "/static" + "/" + ViewDocumentDirName + "/" + strconv.Itoa(i) + ".jpg"
		pp := entity.PagePath{
			Path:   filePath,
			Number: strconv.Itoa(i),
		}
		pagePathList = append(pagePathList, pp)
	}

	/*
		以下で資料表示のHTMLテンプレートを返している
		Renderに渡しているデータは以下の通り
		PagePathList：最大ページまでの資料を取得するためのパスのリスト
		CurrentPagePath：指定されたページ番号の資料を取得するためのリスト
		BackPageNumber：指定されたページの1つ前のページ番号
		NextPageNumber：指定されたページの1つ後のページ番号
		BackShow：戻りページのボタンを表示させるか
		NextShow：次ページのボタンを表示させるか
	*/

	if maxPage == 1 { // 最大ページが1だった場合は次ページや戻りページのボタンを表示させない
		return c.Render(http.StatusOK, "document-view.html", map[string]interface{}{
			"PagePathList":    pagePathList,
			"CurrentPagePath": currentPagePath,
			"BackPageNumber":  intCurrentPage - 1,
			"NextPageNumber":  intCurrentPage + 1,
			"BackShow":        false,
			"NextShow":        false,
		})
	} else if currentPage == "1" { // 指定されたページが1だった場合は次ページのボタンは表示するが戻りページのボタンは表示させない
		return c.Render(http.StatusOK, "document-view.html", map[string]interface{}{
			"PagePathList":    pagePathList,
			"CurrentPagePath": currentPagePath,
			"BackPageNumber":  intCurrentPage - 1,
			"NextPageNumber":  intCurrentPage + 1,
			"BackShow":        false,
			"NextShow":        true,
		})
	} else if intCurrentPage == maxPage { // 指定されたページが最大ページと同じだった場合は戻りページのボタンは表示するが次ページのボタンは表示させない
		return c.Render(http.StatusOK, "document-view.html", map[string]interface{}{
			"PagePathList":    pagePathList,
			"CurrentPagePath": currentPagePath,
			"BackPageNumber":  intCurrentPage - 1,
			"NextPageNumber":  intCurrentPage + 1,
			"BackShow":        true,
			"NextShow":        false,
		})
	}

	// それ以外の場合は、次ページも戻りページもボタン表示
	return c.Render(http.StatusOK, "document-view.html", map[string]interface{}{
		"PagePathList":    pagePathList,
		"CurrentPagePath": currentPagePath,
		"BackPageNumber":  intCurrentPage - 1,
		"NextPageNumber":  intCurrentPage + 1,
		"BackShow":        true,
		"NextShow":        true,
	})
}

// 最大ページの更新
func (dc DocumentController) ChangeMaxPage(c echo.Context) error {
	var err error
	// Formからきた最大ページのデータを構造体にバインド
	maxPageForm := new(entity.MaxPageForm)
	err = c.Bind(maxPageForm)
	if err != nil {
		log.Printf("[error] ChangeMaxPage c.Bind : %v\n", err)
		data := map[string]interface{}{
			"Message": fmt.Sprintf("Formの取得に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// 最大ページのバリデーション
	err = validate.Struct(maxPageForm)
	if err != nil {
		log.Printf("[error] ChangeMaxPage validate.Struct : %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("整数以外あるいは値が1以上10000以下になっていません。: %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	// 最大ページ更新
	maxPage = maxPageForm.MaxPage

	// クライアント向け公開ディレクトリに資料が入っているか
	viewDocuments, err := filepath.Glob(ViewDocumentDirPath + "/*.jpg")
	if err != nil {
		log.Printf("[error] ChangeMaxPage filepath.Glob view-document: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if len(viewDocuments) != 0 { // クライアントに表示させるjpgファイルが入っていたら削除
		for _, f := range viewDocuments {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] ChangeMaxPage os.Remove viewDocuments: %v\n", err)
				data := map[string]string{
					"Message": fmt.Sprintf("最大ページ更新処理に失敗しました。: %v\n", err),
				}
				return c.Render(http.StatusServiceUnavailable, "error.html", data)
			}
		}
	}

	// 最大ページまでの資料のjpgファイルをクライアント向け公開ディレクトリにコピー
	for i := 1; i <= maxPage; i++ {
		// 分割されたり資料から指定されたページのjpgファイル展開
		src, err := os.Open(CutDirName + "/" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			log.Printf("[error] ChangeMaxPage os.Open cut: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "error.html", data)
		}
		defer src.Close()

		// クライアント向け公開ディレクトリにjpgファイル作成
		dst, err := os.Create(filepath.Join(ViewDocumentDirPath, strconv.Itoa(i)+".jpg"))
		if err != nil {
			log.Printf("[error] ChangeMaxPage os.Create view-document: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "error.html", data)
		}

		// jpgファイルの中身をコピー
		_, err = io.Copy(dst, src)
		if err != nil {
			log.Printf("[error] ChangeMaxPage io.Copy: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("ファイルのコピーに失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "error.html", data)
		}
	}

	fmt.Printf("The maximum page has been updated. %d\n", maxPage)

	// 最大ページを更新したことを、Documentページを開いているクライアントに告知（SSE送信）
	SendEvent("DocumentUpdate")

	// HTMLテンプレートのFormの最大ページのinputのvalue用にmaxPageを渡している
	return c.Render(http.StatusOK, "management.html", map[string]interface{}{
		"Message":     fmt.Sprintln("最大ページを更新しました。"),
		"CurrentPage": maxPage,
	})
}

// 資料(PDF)のアップロード
func (dc DocumentController) UpLoad(c echo.Context) error {
	// Formからアップロードされたファイルを受け取る
	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[error] UpLoad c.FormFile: %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("ファイルのアップロードに失敗しました。 %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	// アップロードされたファイルを展開
	src, err := file.Open()
	if err != nil {
		log.Printf("[error] UpLoad file.Open: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer src.Close()

	// 先頭の512byteぶんを持ってくる
	buf := make([]byte, 512)
	_, err = src.Read(buf)
	if err != nil && err != io.EOF {
		log.Printf("[error] UpLoad src.Read: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// ファイルポインタを元に戻す
	src.Seek(0, io.SeekStart)

	// 先頭の512byteを見てContentTypeがpdfかを確認する
	contentType := http.DetectContentType(buf)
	if contentType != "application/pdf" {
		log.Printf("[error] UpLoad http.DetectContentType: %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("PDFではないファイルがアップロードされました。 %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	// ファイルサイズが適切かを確認する。
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if file.Size > maxFileSize {
		log.Printf("[error] UpLoad FileSizeOver: %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("PDFのファイルサイズが大きすぎます。 %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	} else if file.Size <= 0 {
		log.Printf("[error] UpLoad FileSizeLess: %v\n", err)
		data := map[string]interface{}{
			"Message":     fmt.Sprintf("PDFのファイルサイズが小さすぎます。 %v\n", err),
			"CurrentPage": maxPage,
		}
		return c.Render(http.StatusBadRequest, "management.html", data)
	}

	// クライアント向け公開ディレクトリに資料が入っているか
	viewDocuments, err := filepath.Glob(ViewDocumentDirPath + "/*.jpg")
	if err != nil {
		log.Printf("[error] UpLoad filepath.Glob view-document: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if len(viewDocuments) != 0 { // クライアントに表示させるjpgファイルが入っていたら削除
		for _, f := range viewDocuments {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] UpLoad os.Remove viewDocuments: %v\n", err)
				data := map[string]string{
					"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
				}
				return c.Render(http.StatusServiceUnavailable, "error.html", data)
			}
		}
	}

	// cutディレクトリに資料が入っているか
	cuts, err := filepath.Glob(CutDirPath + "/*.jpg")
	if err != nil {
		log.Printf("[error] Upload filepath.Glob cut: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if len(cuts) != 0 { // cutディレクトリにjpgファイルが入っていたら削除
		for _, f := range cuts {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] UpLoad os.Remove cut: %v\n", err)
				data := map[string]string{
					"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
				}
				return c.Render(http.StatusServiceUnavailable, "error.html", data)
			}
		}
	}

	// アップロード先のディレクトリが存在するか確認
	_, err = os.Stat(UpLoadDirName)
	if err != nil {
		// なかったら作成
		log.Printf("[error] UpLoad os.Stat: %v\n", err)
		err = os.Mkdir(UpLoadDirName, 0444)
		if err != nil {
			log.Printf("[error] UpLoad os.Mkdir upload: %v\n", err)
			data := map[string]string{
				"Message": fmt.Sprintf("アップロード先のディレクトリ作成に失敗しました。 %v\n", err),
			}
			return c.Render(http.StatusServiceUnavailable, "error.html", data)
		}
	}

	// uploadディレクトリに資料が入っているか
	uploads, err := filepath.Glob(UpLoadDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] Upload filepath.Glob upload: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if len(uploads) != 0 { // uploadディレクトリにpdfファイルが入っていたら削除
		for _, f := range uploads {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] UpLoad os.Remove upload: %v\n", err)
				data := map[string]string{
					"Message": fmt.Sprintf("アップロードに失敗しました。: %v\n", err),
				}
				return c.Render(http.StatusServiceUnavailable, "error.html", data)
			}
		}
	}

	// アップロード先のディレクトリにコピー用のpdfを作成
	dst, err := os.Create(filepath.Join(UpLoadDirPath, "upload.pdf"))
	if err != nil {
		log.Printf("[error] UpLoad os.Create upload: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// コピー用のpdfにアップロードされたpdfをコピー
	_, err = io.Copy(dst, src)
	if err != nil {
		log.Printf("[error] UpLoad io.Copy upload: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのコピーに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// Pythonスクリプトの結果を受け取るチャネルの作成
	ch := make(chan entity.CmdOutput)
	// Pythonスクリプトを実行するコマンドを生成
	cmd := exec.Command(PythonPath, "pdf-cut.py", UpLoadDirPath+"/upload.pdf", CutDirName)
	// 処理が重くなる可能性があるため並行処理で実行
	go func(cmd *exec.Cmd) {
		// コマンドを実行して結果を受け取る
		result, err := cmd.CombinedOutput()
		// 結果をチャネルに送信
		ch <- entity.CmdOutput{Result: result, Err: err}
	}(cmd)
	// 結果をチャネルから受けとり
	output := <-ch
	// 結果を確認
	if string(output.Result) != "Done\n" { // 実行されたが、想定通りの処理がされなかった場合（想定通りの処理がされた場合は、Pythonスクリプト内でprintされる Done のみ受け取る）
		log.Printf("[error] UpLoad exec.Command.CombinedOutput: %v\n", string(output.Result))
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのカットに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	} else if output.Err != nil { // 何かしらのエラーをう受け取った場合
		log.Printf("[error] UpLoad exec.Command.CombinedOutput: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルのカットに失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// 最初のページだけ表示させるため、最大ページを1ページに更新
	maxPage = 1
	// 分割された資料から1ページ目だけ開く
	src2, err := os.Open(CutDirName + "/1.jpg")
	if err != nil {
		log.Printf("[error] UpLoad os.Open cut: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの展開に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}
	defer src.Close()

	// クライアント向けの公開ディレクトリに1ページ目のコピー用jpgを作成
	dst2, err := os.Create(filepath.Join(ViewDocumentDirPath, strconv.Itoa(maxPage)+".jpg"))
	if err != nil {
		log.Printf("[error] UpLoad os.Create view-document: %v\n", err)
		data := map[string]string{
			"Message": fmt.Sprintf("ファイルの作成に失敗しました。 %v\n", err),
		}
		return c.Render(http.StatusServiceUnavailable, "error.html", data)
	}

	// 1ページ目をコピー
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
