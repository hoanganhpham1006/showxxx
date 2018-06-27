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
	// fmt.Println("hihi")
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
	for i, c := range []*Case{
		&Case{"/users/1", 200, ""},
		&Case{"/users/-1", 400, l.Get(l.M022InvalidUserId)},
		&Case{"/stat/online", 200, ""},
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
			t.Error("case %v: resp.StatusCode != c.RespCode %v %v", i, resp.StatusCode, c.RespCode)
		}
		bodyB, e := ioutil.ReadAll(resp.Body)
		body := string(bodyB)
		defer resp.Body.Close()
		if (resp.StatusCode != 200) && (body != c.RespErrS+"\n") {
			t.Errorf("case %v body != c.RespErrS %v %v", i, (body), (c.RespErrS))
		}
		//		fmt.Println("resp.Body", body)
	}
}

func Test03(t *testing.T) {
	type Case struct {
		Body     map[string]interface{}
		Method   string
		Path     string
		RespCode int
		RespErrS string
	}
	server := httptest.NewServer(CreateRouter())
	defer server.Close()
	client := &http.Client{}
	accessToken := ""
	for i, c := range []*Case{
		&Case{map[string]interface{}{"Username": "daominah", "Password": "123qwe"},
			"POST", "/users/login", 200, ""},
		&Case{map[string]interface{}{"Username": "daominah", "Password": "daominah"},
			"POST", "/users/login", 401, l.Get(l.M002InvalidLogin)},
		&Case{map[string]interface{}{"NewRole": "ROLE_BROADCASTER"},
			"PUT", "/users/-1/role", 400, l.Get(l.M022InvalidUserId)},
		&Case{map[string]interface{}{"NewRole": "ROLE_HIHI"},
			"PUT", "/users/1/role", 400, l.Get(l.M030InvalidRole)},
		&Case{map[string]interface{}{"NewRole": "ROLE_BROADCASTER"},
			"PUT", "/users/7/role", 200, ""},
		&Case{map[string]interface{}{"NewRole": "ROLE_USER"},
			"PUT", "/users/7/role", 200, ""},
		&Case{map[string]interface{}{"Change": 10},
			"PATCH", "/users/-1/cash", 400, l.Get(l.M022InvalidUserId)},
		&Case{map[string]interface{}{"Change": 10},
			"PATCH", "/users/7/cash", 200, ""},
		&Case{map[string]interface{}{"Change": -1000000.5},
			"PATCH", "/users/7/cash", 200, ""},
		&Case{map[string]interface{}{"IsSuspended": true},
			"PUT", "/users/7/suspend", 200, ""},
	} {
		requestUrl := server.URL + c.Path
		reqBodyB, e := json.Marshal(c.Body)
		if e != nil {
			t.Error(e)
		}
		reqBody := bytes.NewBufferString(string(reqBodyB))
		req, e := http.NewRequest(c.Method, requestUrl, reqBody)
		req.Header.Set("Cookie", fmt.Sprintf("login_session=%v", accessToken))
		if e != nil {
			t.Error(e)
		}
		resp, e := client.Do(req)
		if e != nil {
			t.Error(e)
		}
		if resp.StatusCode != c.RespCode {
			t.Errorf("case %v: resp.StatusCode != c.RespCode %v %v", i, resp.StatusCode, c.RespCode)
		}
		bodyB, e := ioutil.ReadAll(resp.Body)
		body := string(bodyB)
		defer resp.Body.Close()
		if (resp.StatusCode != 200) && (body != c.RespErrS+"\n") {
			t.Errorf("case %v: body != c.RespErrS %v %v", i, (body), (c.RespErrS))
		}
		//
		if c.Path == "/users/login" {
			for _, cookie := range resp.Cookies() {
				if cookie.Name == "login_session" {
					accessToken = cookie.Value
				}
			}
		}
	}
}
