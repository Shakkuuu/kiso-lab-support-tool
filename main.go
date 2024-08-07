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
	userNameFlag := flag.String("user", "user", "BasicAuth user flag")
	passwordFlag := flag.String("password", "password", "BasicAuth password flag")
	portFlag := flag.Int("port", 8080, "Port flag")

	flag.Parse()

	_, err := os.Stat(controller.CutDirName)
	if err != nil {
		err = os.Mkdir(controller.CutDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir cut: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = os.Stat(controller.ViewPDFDirName)
	if err != nil {
		err = os.Mkdir(controller.ViewPDFDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir view-pdf: %v\n", err)
			os.Exit(1)
		}
	}

	_, err = os.Stat(controller.UpLoadDirName)
	if err != nil {
		err = os.Mkdir(controller.UpLoadDirName, 0444)
		if err != nil {
			log.Printf("[error] main os.Mkdir upload: %v\n", err)
			os.Exit(1)
		}
	}

	cuts, err := filepath.Glob(controller.CutDirPath + "/*.pdf")
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

	viewPDF, err := filepath.Glob(controller.ViewPDFDirPath + "/*.pdf")
	if err != nil {
		log.Printf("[error] main filepath.Glob view-pdf: %v\n", err)
		os.Exit(1)
	} else if len(viewPDF) != 0 {
		for _, f := range viewPDF {
			err = os.Remove(f)
			if err != nil {
				log.Printf("[error] main os.Remove view-pdf: %v\n", err)
				os.Exit(1)
			}
		}
	}

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

	db.Init()

	server.Init(*userNameFlag, *passwordFlag, *portFlag)
}
