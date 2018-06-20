// This package provides functionality for saving static content
//
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/daominah/livestream/zconfig"
)

func main() {
	go func() {
		fmt.Printf("Listening http message on address host%v%v\n",
			zconfig.StaticUploadPort, zconfig.StaticUploadPath)
		err := http.ListenAndServe(zconfig.StaticUploadPort, nil)
		if err != nil {
			fmt.Printf("Fail to listen http message on address host%v%v\n %v\n",
				zconfig.StaticUploadPort, zconfig.StaticUploadPath, err.Error())
		}
	}()
	//
	http.HandleFunc(zconfig.StaticUploadPath, func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r)
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", 405)
			return
		}
		limitNBytes := int64(10) * 1024 * 1024
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, limitNBytes))
		if err != nil {
			http.Error(w, "Invalid body", 406)
			return
		}
		if int64(len(body)) == limitNBytes {
			http.Error(w,
				fmt.Sprintf("File must be less than %v bytes", limitNBytes),
				406)
			return
		}
		now := time.Now()
		folderName := fmt.Sprintf("%04d%02d%02d", now.Year(), int(now.Month()), now.Day())
		fileName := fmt.Sprintf("%v", now.UnixNano())
		folderPath := fmt.Sprintf("%v/%v", zconfig.StaticFolder, folderName)
		filePath := fmt.Sprintf("%v/%v/%v", zconfig.StaticFolder, folderName, fileName)

		err = ioutil.WriteFile(filePath, body, 0777)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				os.Mkdir(folderPath, 0777)
				err = ioutil.WriteFile(filePath, body, 0777)
			} else {
				http.Error(w, err.Error(), 500)
				return
			}
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		respBody := fmt.Sprintf(`/%v/%v`, folderName, fileName)
		w.Write([]byte(respBody))
		return
	})
	//
	select {}
}
