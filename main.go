package main

import (
	"flag"
	"kiso-lab-support-tool/controller"
	"kiso-lab-support-tool/db"
	"kiso-lab-support-tool/server"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// 実行時のフラグでManagementページへアクセスする際のBasic認証のユーザ名&パスワードと、サーバ起動ポートを指定
	userNameFlag := flag.String("user", "user", "BasicAuth user flag")
	passwordFlag := flag.String("password", "password", "BasicAuth password flag")
	portFlag := flag.Int("port", 8080, "Port flag")

	flag.Parse()

	// 分割されてjpgに変換されたページを入れておくディレクトリの存在確認
	_, err := os.Stat(controller.CutDirName)
	if err != nil {
		// なければ作成
		err = os.Mkdir(controller.CutDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir cut: %v\n", err)
			os.Exit(1)
		}
	}

	// クライアントに表示させるページのjpgファイルを入れておくディレクトリの存在確認
	_, err = os.Stat(controller.ViewDocumentDirName)
	if err != nil {
		// なければ作成
		err = os.Mkdir(controller.ViewDocumentDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir view-document: %v\n", err)
			os.Exit(1)
		}
	}

	// アップロードされる元のPDFファイルを入れておくディレクトリの存在確認
	_, err = os.Stat(controller.UpLoadDirName)
	if err != nil {
		// なければ作成
		err = os.Mkdir(controller.UpLoadDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir upload: %v\n", err)
			os.Exit(1)
		}
	}

	// cutディレクトリにjpgファイルが残っていれば削除
	cuts, err := filepath.Glob(controller.CutDirPath + "/*.jpg")
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

	// view-documentディレクトリにjpgファイルが残っていれば削除
	viewDocuments, err := filepath.Glob(controller.ViewDocumentDirPath + "/*.jpg")
	if err != nil {
		log.Printf("[error] main filepath.Glob view-document: %v\n", err)
		os.Exit(1)
	} else if len(viewDocuments) != 0 {
		for _, f := range viewDocuments {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove view-document: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// uploadディレクトリにpdfファイルが残っていれば削除
	upload, err := filepath.Glob(controller.UpLoadDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] main filepath.Glob upload: %v\n", err)
		os.Exit(1)
	} else if len(upload) != 0 {
		for _, f := range upload {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove upload: %v\n", err)
				os.Exit(1)
			}
		}
	}

	// DB接続
	db.Init()

	// サーバ起動
	server.Init(*userNameFlag, *passwordFlag, *portFlag)
}
