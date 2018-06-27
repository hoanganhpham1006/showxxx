package admintool

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/go-martini/martini"

	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/users"
)

func UserDetail(r *http.Request, w http.ResponseWriter, p martini.Params) string {
	playerId, _ := strconv.ParseInt(p["uid"], 10, 64)
	user, err := users.GetUser(playerId)
	if user == nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return user.ToString()
}

func UserChangeRole(
	r *http.Request, w http.ResponseWriter, p martini.Params) string {
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
	err = users.ChangeUserRole(userId, newRole)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return ""
	}
	return ""
}
