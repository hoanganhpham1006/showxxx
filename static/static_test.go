package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/daominah/livestream/zconfig"
)

func Test01(t *testing.T) {
	// fmt.Println("oe")
	client := &http.Client{}
	requestUrl := fmt.Sprintf("http://localhost%v%v",
		zconfig.StaticUploadPort, UPLOADING_PATH)
	//	fmt.Println("requestUrl", requestUrl)
	reqBodyB, e := ioutil.ReadFile("./test.jpg")
	if e != nil {
		t.Error(e)
		return
	}
	reqBody := bytes.NewBufferString(string(reqBodyB))
	req, e := http.NewRequest("POST", requestUrl, reqBody)
	if e != nil {
		t.Error(e)
		return
	}
	resp, e := client.Do(req)
	if e != nil {
		t.Error(e)
		return
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	_ = string(respBody)
	if resp.StatusCode != 200 {
		t.Error()
	}
}
