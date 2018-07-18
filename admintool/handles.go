package admintool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"

	"github.com/daominah/livestream/conversations"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/streams"
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
	return user.String()
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
	return user.String()
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

func TeamDetail(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	teamId, _ := strconv.ParseInt(p["tid"], 10, 64)
	team, err := users.GetTeam(teamId)
	if team == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return team.String()
}

func TeamAddMember(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	teamId, _ := strconv.ParseInt(p["tid"], 10, 64)
	team, err := users.GetTeam(teamId)
	if team == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
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
	userId := misc.ReadInt64(data, "UserId")
	user, err := users.GetUser(userId)
	if user == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	err = users.AddTeamMember(teamId, userId)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}

func TeamRemoveMember(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	teamId, _ := strconv.ParseInt(p["tid"], 10, 64)
	team, err := users.GetTeam(teamId)
	if team == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	userId, _ := strconv.ParseInt(p["uid"], 10, 64)
	user, err := users.GetUser(userId)
	if user == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	err = users.RemoveTeamMember(teamId, userId)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}

func StreamAllSummaries(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	filterReported := r.URL.Query().Get("filter_reported")
	var d []map[string]interface{}
	if filterReported == "true" {
		d = streams.StreamAllSummaries(true)
	} else {
		d = streams.StreamAllSummaries(false)
	}
	bs, e := json.Marshal(d)
	if e != nil {
		return "[]"
	}
	return string(bs)
}

func StreamDetail(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	broadcasterId, _ := strconv.ParseInt(p["uid"], 10, 64)
	s, err := streams.GetStream(broadcasterId)
	if s == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return s.String()
}
func StreamChat(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	err := checkIsAdmin(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return ""
	}
	broadcasterId, _ := strconv.ParseInt(p["uid"], 10, 64)
	s, err := streams.GetStream(broadcasterId)
	if s == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
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
	message := misc.ReadString(data, "Message")
	err = conversations.CreateMessage(
		s.ConversationId, 1, message, conversations.DISPLAY_TYPE_ADMIN)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}
