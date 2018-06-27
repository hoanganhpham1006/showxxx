package admintool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	l "github.com/daominah/livestream/language"
)

func Test01(t *testing.T) {
	_ = fmt.Println
	fmt.Println("hihi")
}

func Test02(t *testing.T) {
	type Case struct {
		Path     string
		RespCode int
		RespErrS string
	}
	server := httptest.NewServer(CreateRouter())
	defer server.Close()
	client := &http.Client{}
	for _, c := range []*Case{
		&Case{"/users/1", 200, ""},
		&Case{"/users/-1", 400, l.Get(l.M022InvalidUserId)},
	} {
		requestUrl := server.URL + c.Path
		reqBodyB, e := json.Marshal(map[string]interface{}{})
		if e != nil {
			t.Error(e)
		}
		reqBody := bytes.NewBufferString(string(reqBodyB))
		req, e := http.NewRequest("GET", requestUrl, reqBody)
		if e != nil {
			t.Error(e)
		}
		resp, e := client.Do(req)
		if e != nil {
			t.Error(e)
		}
		if resp.StatusCode != c.RespCode {
			t.Error("resp.StatusCode != c.RespCode", resp.StatusCode, c.RespCode)
		}
		bodyB, e := ioutil.ReadAll(resp.Body)
		body := string(bodyB)
		defer resp.Body.Close()
		if (resp.StatusCode != 200) && (body != c.RespErrS+"\n") {
			t.Errorf("body != c.RespErrS %v %v", (body), (c.RespErrS))
		}
	}
}

func Test03(t *testing.T) {

}
