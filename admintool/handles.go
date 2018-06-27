package admintool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
)

/*
400 Bad Request
401 Unauthorized
*/

func UserLogin(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	username := misc.ReadString(data, "Username")
	password := misc.ReadString(data, "Password")
	user, cookie, err := users.LoginByPassword(username, password)
	if user == nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	w.Header().Set("Set-Cookie", fmt.Sprintf("login_session=%v", cookie))
	return ""
}

func checkIsAdmin(r *http.Request) error {
	cookie, e := r.Cookie("login_session")
	if e != nil {
		return e
	}
	login_session := cookie.Value
	u, e := users.LoginByCookie(login_session)
	if e != nil {
		return e
	}
	if u.Role != users.ROLE_ADMIN {
		return errors.New(l.Get(l.M031OperationNotPermitted))
	}
	return nil
}

func UserDetail(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	playerId, _ := strconv.ParseInt(p["uid"], 10, 64)
	user, err := users.GetUser(playerId)
	if user == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return user.ToString()
}

func UserChangeRole(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	userId, _ := strconv.ParseInt(p["uid"], 10, 64)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	newRole := misc.ReadString(data, "NewRole")
	_ = fmt.Println
	err = users.ChangeUserRole(userId, newRole)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}

func UserChangeCash(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	userId, _ := strconv.ParseInt(p["uid"], 10, 64)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	change := misc.ReadFloat64(data, "Change")
	_ = fmt.Println
	_, err = users.ChangeUserMoney(userId, users.MT_CASH, change,
		users.REASON_ADMIN_CHANGE, false)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}

func UserSuspend(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	userId, _ := strconv.ParseInt(p["uid"], 10, 64)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	isSuspended := misc.ReadBool(data, "IsSuspended")
	err = users.SuspendUser(userId, isSuspended)
	return ""
}

func UserOnlineStat(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	users.GMutex.Lock()
	defer users.GMutex.Unlock()
	nOnline := 0
	nOnlineBroadcaster := 0
	nBroadcasting := 0
	for _, user := range users.MapIdToUser {
		if user.StatusL1 != users.STATUS_OFFLINE {
			nOnline += 1
			if user.Role == users.ROLE_BROADCASTER {
				nOnlineBroadcaster += 1
				if user.StatusL1 == users.STATUS_BROADCASTING {
					nBroadcasting += 1
				}
			}
		}
	}
	res := map[string]interface{}{
		"NOnline":            nOnline,
		"NOnlineBroadcaster": nOnlineBroadcaster,
		"NBroadcasting":      nBroadcasting,
	}
	resB, _ := json.MarshalIndent(res, "", "    ")
	return string(resB)
}
