package users

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/sha3"

	//	"github.com/daominah/livestream/connections"
	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zdatabase"
)

const (
	ROLE_ADMIN       = "ROLE_ADMIN"
	ROLE_BROADCASTER = "ROLE_BROADCASTER"
	ROLE_USER        = "ROLE_USER"

	// money types
	MT_CASH               = "MT_CASH"
	MT_EXPERIENCE         = "MT_EXPERIENCE"
	MT_ONLINE_DURATION    = "MT_ONLINE_DURATION"
	MT_BROADCAST_DURATION = "MT_BROADCAST_DURATION"

	// money log reasons
	REASON_ADMIN_CHANGE             = "REASON_ADMIN_CHANGE"
	REASON_TRANSFER                 = "REASON_TRANSFER"
	REASON_CHEER                    = "REASON_CHEER"
	REASON_CHEER_TEAM_SPLIT_MEMBER  = "REASON_CHEER_TEAM_SPLIT_MEMBER"
	REASON_CHEER_TEAM_SPLIT_CAPTAIN = "REASON_CHEER_TEAM_SPLIT_CAPTAIN"
	REASON_CHAT_BIG                 = "REASON_CHAT_BIG"

	STATUS_OFFLINE      = "STATUS_OFFLINE"
	STATUS_ONLINE       = "STATUS_ONLINE"
	STATUS_BROADCASTING = "STATUS_BROADCASTING"
	STATUS_WATCHING     = "STATUS_WATCHING"
	STATUS_PLAYING_GAME = "STATUS_PLAYING_GAME"

	SEX_MALE      = "SEX_MALE"
	SEX_FEMALE    = "SEX_FEMALE"
	SEX_UNDEFINED = "SEX_UNDEFINED"
)

// mofify this list to add money types
var MONEY_TYPES []string

// remember to lock when read/write this map
var MapIdToUser map[int64]*User
var MapIdToTeam map[int64]*Team

// locker for MapUsers
var GMutex sync.Mutex

func init() {
	MONEY_TYPES = []string{
		MT_CASH, MT_EXPERIENCE, MT_ONLINE_DURATION, MT_BROADCAST_DURATION,
	}
	MapIdToUser = make(map[int64]*User)
	MapIdToTeam = make(map[int64]*Team)
}

type User struct {
	Id       int64
	Username string
	// ROLE_ADMIN, ROLE_BROADCASTER, ROLE_USER
	Role        string
	IsSuspended bool
	RealName    string
	NationalId  string
	// SEX_MALE, SEX_FEMALE, SEX_UNDEFINED
	Sex   string
	Phone string
	Email string
	// ISO 3166-1 alpha-2: VN, US, GB,..
	Country      string
	Address      string
	ProfileName  string
	ProfileImage string
	Summary      string
	// json: {"skype": "daominah"}
	Misc        string
	CreatedTime time.Time

	TeamId int64
	// map moneyType to value, this map is only for caching;
	// if u want change user money, use func changeUserMoney
	MapMoney map[string]float64

	NFollowers int
	NFollowing int
	// base on MT_ONLINE_DURATION
	Level int
	// base on purchased cash in month
	LevelVip int

	// StatusL1 will be assign in other packages
	StatusL1 string
	// json: {"Game": "GAME_TAIXIU"}, {"Video": 92}
	StatusL2 string
	//
	Mutex sync.Mutex
}

func (u *User) String() string {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()
	bs, e := json.MarshalIndent(u, "", "    ")
	if e != nil {
		return "{}"
	}
	return string(bs)
}

func (u *User) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	s := u.String()
	json.Unmarshal([]byte(s), &result)
	return result
}

func (u *User) ToShortMap() map[string]interface{} {
	result := map[string]interface{}{
		"Id":           u.Id,
		"ProfileName":  u.ProfileName,
		"ProfileImage": u.ProfileImage,
		"TeamId":       u.TeamId,
		"Sex":          u.Sex,
		"StatusL1":     u.StatusL1,
		"StatusL2":     u.StatusL2,
	}
	return result
}

// name is list of alphanumeric characters or _ or @
func NormalizeName(name string) string {
	alphanumericChars :=
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_@"
	// set of alphanumeric chars
	mapChars := map[string]bool{}
	for _, r := range alphanumericChars {
		c := fmt.Sprintf("%c", r)
		mapChars[c] = true
	}
	cs := []string{}
	for _, r := range name {
		c := fmt.Sprintf("%c", r)
		if mapChars[c] {
			cs = append(cs, c)
		}
	}
	good := strings.Join(cs, "")
	return good
}

// use sha3_512 algo, return hex encoding
func HassPassword(password string) string {
	h := sha3.New512()
	h.Write([]byte(password))
	resultBs := h.Sum(nil)
	resultHex := hex.EncodeToString(resultBs)
	return resultHex
}

// return userId, error
func CreateUser(username string, password string) (int64, error) {
	username = NormalizeName(username)
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO "user"
    		(username, hashed_password, login_session, profile_name)
		VALUES ($1, $2, 'hohohaha', $3) RETURNING id`,
		username, HassPassword(password), username)
	var id int64
	e := row.Scan(&id)
	if e != nil {
		return 0, errors.New(l.Get(l.M001DuplicateUsername))
	}

	ts := []string{}
	args := []interface{}{}
	for i, moneyType := range MONEY_TYPES {
		ts = append(ts, fmt.Sprintf("($%v, $%v, $%v)", 3*i+1, 3*i+2, 3*i+3))
		args = append(args, []interface{}{id, moneyType, float64(0)}...)
	}
	queryPart := strings.Join(ts, ", ")
	_, e = zdatabase.DbPool.Exec(fmt.Sprintf(
		`INSERT INTO user_money (user_id, money_type, val)
	    VALUES %v`, queryPart),
		args...)
	if e != nil {
		return 0, e
	}

	LoadUser(id)

	return id, nil
}

// load user data from database to MapIdToUser, this map is only for caching,
// this func creates new moneyType from MONEY_TYPES if necessary
func LoadUser(id int64) (*User, error) {
	var username, role, real_name, national_id, sex, phone, email, country string
	var address, profile_name, profile_image, summary, misc string
	var is_suspended bool
	var created_time time.Time
	row := zdatabase.DbPool.QueryRow(
		`SELECT username, role, real_name, national_id, sex, phone, email, country, 
		    address, profile_name, profile_image, summary, misc, 
		    is_suspended, created_time
    	FROM "user"
    	WHERE id = $1 `, id)
	e := row.Scan(
		&username, &role, &real_name, &national_id, &sex, &phone, &email, &country,
		&address, &profile_name, &profile_image, &summary, &misc,
		&is_suspended, &created_time)
	if e != nil {
		return nil, errors.New(l.Get(l.M022InvalidUserId))
	}
	user := &User{Id: id, Username: username, Role: role, RealName: real_name,
		NationalId: national_id, Sex: sex, Phone: phone, Email: email,
		Country: country, Address: address, ProfileName: profile_name,
		ProfileImage: profile_image, Summary: summary, Misc: misc,
		IsSuspended: is_suspended, CreatedTime: created_time,
		StatusL1: STATUS_OFFLINE, StatusL2: "{}",
	}
	//
	user.MapMoney = make(map[string]float64)
	for _, moneyType := range MONEY_TYPES {
		var val float64
		row = zdatabase.DbPool.QueryRow(
			`SELECT val FROM user_money
    			WHERE user_id = $1 AND money_type = $2 `,
			id, moneyType,
		)
		e := row.Scan(&val)
		if e == sql.ErrNoRows {
			zdatabase.DbPool.Exec(
				`INSERT INTO user_money (user_id, money_type, val)
            	    VALUES ($1, $2, $3)`, id, moneyType, 0)
		} else if e != nil {
			return nil, e
		}
		user.Mutex.Lock()
		user.MapMoney[moneyType] = val
		user.Mutex.Unlock()
	}
	//
	row = zdatabase.DbPool.QueryRow(
		`SELECT team_id FROM team_member WHERE user_id = $1`, id)
	row.Scan(&user.TeamId)
	// calculated field
	user.Level = int(user.MapMoney[MT_ONLINE_DURATION])
	user.LevelVip = 15

	user.NFollowers = len(LoadFollowers(id))
	user.NFollowing = len(LoadFollowing(id))

	GMutex.Lock()
	MapIdToUser[id] = user
	GMutex.Unlock()

	return user, nil
}

// try to read data in ram,
// if cant: read data from database
func GetUser(userId int64) (*User, error) {
	GMutex.Lock()
	u := MapIdToUser[userId]
	GMutex.Unlock()
	if u != nil {
		return u, nil
	} else {
		return LoadUser(userId)
	}
}

func LoadFollowers(userId int64) []int64 {
	result := make([]int64, 0)
	rows, err := zdatabase.DbPool.Query(
		`SELECT user_id_1 FROM user_following
		WHERE user_id_2 = $1`,
		userId)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var follower int64
		e := rows.Scan(&follower)
		if e != nil {
			fmt.Println("ERROR GetFollowers", e)
		}
		result = append(result, follower)
	}
	return result
}

func LoadFollowing(userId int64) []int64 {
	result := make([]int64, 0)
	rows, err := zdatabase.DbPool.Query(
		`SELECT user_id_2 FROM user_following
		WHERE user_id_1 = $1`,
		userId)
	if err != nil {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var following int64
		e := rows.Scan(&following)
		if e != nil {
			fmt.Println("ERROR GetFollowing", e)
		}
		result = append(result, following)
	}
	return result
}

// return userObj, cookie, error
func LoginByPassword(username string, password string) (
	*User, string, error) {
	row := zdatabase.DbPool.QueryRow(
		`SELECT id
		FROM "user"
		WHERE username = $1 AND hashed_password = $2`,
		NormalizeName(username), HassPassword(password))
	var id int64
	e := row.Scan(&id)
	if e != nil {
		return nil, "", errors.New(l.Get(l.M002InvalidLogin))
	}

	cookieData := map[string]string{
		"userId":    fmt.Sprintf("%v", id),
		"loginTime": time.Now().Format(time.RFC3339Nano)}
	cookiePlainBs, _ := json.Marshal(cookieData)
	// cookiePlain := string(cookiePlainBs)
	cookie := hex.EncodeToString(cookiePlainBs)
	zdatabase.DbPool.Exec(
		`UPDATE "user" SET login_session = $1 WHERE id = $2`,
		cookie, id)
	//
	u, e := LoadUser(id)
	if u == nil {
		return nil, "", e
	}
	if u.IsSuspended {
		return nil, "", errors.New(l.Get(l.M025UserSuspended))
	}
	return u, cookie, nil
}

// return userObj, error
func LoginByCookie(login_session string) (*User, error) {
	row := zdatabase.DbPool.QueryRow(
		`SELECT id
		FROM "user"
		WHERE login_session = $1 `,
		login_session)
	var id int64
	e := row.Scan(&id)
	if e != nil {
		return nil, errors.New(l.Get(l.M002InvalidLogin))
	}
	//
	u, e := LoadUser(id)
	if u == nil {
		return nil, e
	}
	if u.IsSuspended {
		return nil, errors.New(l.Get(l.M025UserSuspended))
	}
	return u, nil

}

//
func RecordLogin(
	userId int64, networkAddress, deviceName string, appName string) (
	int64, error) {
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO user_login
    		(user_id, network_address, device_name, app_name)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		userId, networkAddress, deviceName, appName)
	var id int64
	e := row.Scan(&id)
	if e != nil {
		return 0, e
	}
	return id, nil
}

// change online duration
func RecordLogout(loginId int64) error {
	row := zdatabase.DbPool.QueryRow(
		`SELECT user_id, login_time FROM user_login WHERE id = $1`, loginId)
	var user_id int64
	var login_time time.Time
	e := row.Scan(&user_id, &login_time)
	if e != nil {
		return e
	}
	now := time.Now()
	_, e = ChangeUserMoney(
		user_id, MT_ONLINE_DURATION, now.Sub(login_time).Seconds(), "", false)
	if e != nil {
		return e
	}
	_, e = zdatabase.DbPool.Exec(
		`UPDATE user_login SET logout_time = $1 WHERE id = $2`,
		now, loginId)
	return e
}

//
func SuspendUser(userId int64, isSuspended bool) error {
	_, e := zdatabase.DbPool.Exec(
		`UPDATE "user" SET is_suspended = $1 WHERE id = $2`,
		isSuspended, userId)
	// update cache
	if e == nil {
		GMutex.Lock()
		if MapIdToUser[userId] != nil {
			MapIdToUser[userId].IsSuspended = isSuspended
		}
		GMutex.Unlock()
	}
	return e
}

func ChangeUserRole(userId int64, newRole string) error {
	if misc.FindStringInSlice(
		newRole, []string{ROLE_ADMIN, ROLE_BROADCASTER, ROLE_USER}) == -1 {
		return errors.New(l.Get(l.M030InvalidRole))
	}
	r, e := zdatabase.DbPool.Exec(
		`UPDATE "user" SET role = $1 WHERE id = $2`,
		newRole, userId)
	// update cache
	if e != nil {
		return e
	}
	if nRowAff, _ := r.RowsAffected(); nRowAff == 0 {
		return errors.New(l.Get(l.M022InvalidUserId))
	}
	GMutex.Lock()
	if MapIdToUser[userId] != nil {
		MapIdToUser[userId].Role = newRole
	}
	GMutex.Unlock()
	return nil
}

func ChangeUserInfo(userId int64, RealName string, NationalId string, Sex string,
	Country string, Address string, ProfileName string, ProfileImage string,
	Summary string) error {
	sexTypes := []string{SEX_MALE, SEX_FEMALE, SEX_UNDEFINED}
	if Sex != "" && misc.FindStringInSlice(Sex, sexTypes) == -1 {
		return errors.New(l.Get(l.M020InvalidSex))
	}
	_, isIn := countries[Country]
	if Country != "" && !isIn {
		return errors.New(l.Get(l.M021InvalidCountry))
	}
	//
	columns := make([]string, 0)
	args := make([]interface{}, 0)
	holders := make([]int, 0)
	counter := 1
	if RealName != "" {
		columns = append(columns, "real_name")
		args = append(args, RealName)
		holders = append(holders, counter)
		counter += 1
	}
	if NationalId != "" {
		columns = append(columns, "national_id")
		args = append(args, NationalId)
		holders = append(holders, counter)
		counter += 1
	}
	if Sex != "" {
		columns = append(columns, "sex")
		args = append(args, Sex)
		holders = append(holders, counter)
		counter += 1
	}
	if Country != "" {
		columns = append(columns, "country")
		args = append(args, Country)
		holders = append(holders, counter)
		counter += 1
	}
	if Address != "" {
		columns = append(columns, "address")
		args = append(args, Address)
		holders = append(holders, counter)
		counter += 1
	}
	if ProfileName != "" {
		columns = append(columns, "profile_name")
		args = append(args, ProfileName)
		holders = append(holders, counter)
		counter += 1
	}
	if ProfileImage != "" {
		columns = append(columns, "profile_image")
		args = append(args, ProfileImage)
		holders = append(holders, counter)
		counter += 1
	}
	if Summary != "" {
		columns = append(columns, "summary")
		args = append(args, Summary)
		holders = append(holders, counter)
		counter += 1
	}
	queryParts := make([]string, 0)
	for i := 0; i < len(columns); i++ {
		queryParts = append(queryParts,
			fmt.Sprintf("%v = $%v", columns[i], holders[i]))
	}
	queryPart := strings.Join(queryParts, ", ")
	query := fmt.Sprintf(
		`UPDATE "user" 
		SET %v
		WHERE id = %v`, queryPart, userId)
	//	fmt.Println("ChangeUserInfo query", query)
	_, err := zdatabase.DbPool.Exec(query, args...)
	if err != nil {
		return err
	}
	_, err = LoadUser(userId)
	return err
}

// return file path on static server
func UploadFile(content []byte) (string, error) {
	reqBodyB := content
	client := &http.Client{}
	requestUrl := fmt.Sprintf("http://%v%v%v",
		zconfig.StaticHost, zconfig.StaticUploadPort, zconfig.StaticUploadPath)
	//	fmt.Println("requestUrl", requestUrl)
	reqBody := bytes.NewBufferString(string(reqBodyB))
	req, e := http.NewRequest("POST", requestUrl, reqBody)
	if e != nil {
		return "", e
	}
	resp, e := client.Do(req)
	if e != nil {
		return "", e
	}
	respBody, e := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if e != nil {
		return "", e
	}
	// fmt.Println("respBody ", string(respBody))
	if resp.StatusCode != 200 {
		return "", errors.New(string(respBody))
	}
	return string(respBody), nil
}

func ChangeUserProfileImage(userId int64, newProfileImage []byte) (
	string, error) {
	imgPath, e := UploadFile(newProfileImage)
	if e != nil {
		return "", errors.New(l.Get(l.M023StaticServerDown))
	}
	e = ChangeUserInfo(userId, "", "", "", "", "", "", imgPath, "")
	if e != nil {
		return "", e
	}
	return imgPath, nil
}

func Follow(userId int64, targetId int64) error {
	_, e := zdatabase.DbPool.Exec(
		`INSERT INTO user_following (user_id_1, user_id_2)
		VALUES ($1, $2)`,
		userId, targetId)
	// update cache and rank
	if e == nil {
		GMutex.Lock()
		if MapIdToUser[userId] != nil {
			MapIdToUser[userId].NFollowing += 1
		}
		if MapIdToUser[targetId] != nil {
			MapIdToUser[targetId].NFollowers += 1
		}
		GMutex.Unlock()
		//
		for _, rankId := range []int64{
			rank.RANK_N_FOLLOWERS_WEEK,
			rank.RANK_N_FOLLOWERS_ALL,
		} {
			rank.ChangeKey(rankId, targetId, 1)
		}
	}
	return e
}

func Unfollow(userId int64, targetId int64) error {
	r, e := zdatabase.DbPool.Exec(
		`DELETE FROM user_following 
		WHERE user_id_1 = $1 AND user_id_2 = $2`,
		userId, targetId)
	// update cache
	if e == nil {
		nRowsAffected, _ := r.RowsAffected()
		if nRowsAffected != 0 { // update cache and rank
			GMutex.Lock()
			if MapIdToUser[userId] != nil {
				MapIdToUser[userId].NFollowing -= 1
			}
			if MapIdToUser[targetId] != nil {
				MapIdToUser[targetId].NFollowers -= 1
			}
			GMutex.Unlock()
			//
			for _, rankId := range []int64{
				rank.RANK_N_FOLLOWERS_WEEK,
				rank.RANK_N_FOLLOWERS_ALL,
			} {
				_ = rankId
				rank.ChangeKey(rankId, targetId, -1)
			}
		}
	}
	return e
}

func CheckIsFollowing(userId int64, targetId int64) bool {
	row := zdatabase.DbPool.QueryRow(
		`SELECT created_time FROM user_following
		WHERE user_id_1 = $1 AND user_id_2 = $2`,
		userId, targetId)
	var followingTime time.Time
	e := row.Scan(&followingTime)
	if e == nil {
		return true
	}
	// sql: no rows in result set
	return false
}

func GetUsernameById(userId int64) (string, error) {
	u, e := GetUser(userId)
	if e != nil {
		return "", e
	}
	return u.Username, nil
}

func GetProfilenameById(userId int64) (string, error) {
	u, e := GetUser(userId)
	if e != nil {
		return "", e
	}
	return u.ProfileName, nil
}

// simple search using sql LIKE 'key%',
// TODO: implement full text search
func Search(key string) ([]map[string]interface{}, error) {
	columns := []string{"profile_name", "username"}
	nRowLimit := 10
	duplicateIdChecker := make(map[int64]bool)
	result := make([]map[string]interface{}, 0)
	//
	keyInt64, e := strconv.ParseInt(key, 10, 64)
	if e == nil {
		row := zdatabase.DbPool.QueryRow(
			`SELECT id FROM "user" WHERE id = $1`, keyInt64)
		var uid int64
		e := row.Scan(&uid)
		if e != nil {
			return nil, e
		}
		user, _ := GetUser(uid)
		if user != nil {
			result = append(result, user.ToShortMap())
			duplicateIdChecker[uid] = true
		}
	}
	//
	uids := make([]int64, 0)
	for _, column := range columns {
		rows, e := zdatabase.DbPool.Query(fmt.Sprintf(
			`SELECT id FROM "user"
	        WHERE %v LIKE $1
	        LIMIT %v`, column, nRowLimit),
			fmt.Sprintf("%v%%", key))
		if e != nil {
			return nil, e
		}
		defer rows.Close()
		for rows.Next() {
			var uid int64
			rows.Scan(&uid)
			if _, isIn := duplicateIdChecker[uid]; !isIn {
				uids = append(uids, uid)
				duplicateIdChecker[uid] = true
			}
		}
	}
	for _, uid := range uids {
		user, _ := GetUser(uid)
		if user != nil {
			result = append(result, user.ToShortMap())
		}
	}
	return result, nil
}
