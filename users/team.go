package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/zdatabase"
)

type Team struct {
	TeamId      int64
	TeamName    string
	TeamImage   string
	Summary     string
	CreatedTime time.Time
	// can be nil
	Captain *User
	Members map[int64]*User
	Mutex   sync.Mutex
}

// data sent to client,
// notice concurrently read and write user's map
func (team *Team) ToMap() map[string]interface{} {
	team.Mutex.Lock()
	defer team.Mutex.Unlock()
	members := make([]map[string]interface{}, 0)
	for _, member := range team.Members {
		members = append(members, member.ToShortMap())
	}
	captain := map[string]interface{}{}
	if team.Captain != nil {
		captain = team.Captain.ToShortMap()
	}
	result := map[string]interface{}{
		"TeamId":      team.TeamId,
		"TeamName":    team.TeamName,
		"TeamImage":   team.TeamImage,
		"Summary":     team.Summary,
		"CreatedTime": team.CreatedTime,
		"Captain":     captain,
		"Member":      members,
	}
	return result
}

func (team *Team) String() string {
	bs, e := json.Marshal(team.ToMap())
	if e != nil {
		return "{}"
	}
	return string(bs)
}

// return teamId, error
func CreateTeam(teamName string, teamImage string, teamSummary string) (
	int64, error) {
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO team
    		(team_name, team_image, summary)
		VALUES ($1, $2, $3) RETURNING team_id`,
		teamName, teamImage, teamSummary)
	var id int64
	e := row.Scan(&id)
	if e != nil {
		//		pqErr, isOk := e.(*pq.Error)
		//		if !isOk {
		//			return 0, e
		//		}
		//		fmt.Println(pqErr.Detail)
		return 0, errors.New(l.Get(l.M012DuplicateTeamName))
	}

	LoadTeam(id)

	return id, nil
}

// load team data from database to MapIdToTeam
func LoadTeam(teamId int64) (*Team, error) {
	var team_name, team_image, summary string
	var created_time time.Time
	row := zdatabase.DbPool.QueryRow(
		`SELECT team_name, team_image, summary, created_time
		FROM team
		WHERE team_id = $1`,
		teamId)
	err := row.Scan(&team_name, &team_image, &summary, &created_time)
	if err != nil {
		return nil, err
	}
	team := &Team{TeamId: teamId, TeamName: team_name, TeamImage: team_image,
		Summary: summary, CreatedTime: created_time,
		Members: make(map[int64]*User)}
	rows, e := zdatabase.DbPool.Query(
		`SELECT user_id, is_captain, joined_time
		FROM team_member
		WHERE team_id = $1`,
		teamId)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	for rows.Next() {
		var user_id int64
		var is_captain bool
		var joined_time time.Time
		e = rows.Scan(&user_id, &is_captain, &joined_time)
		if e != nil {
			return nil, e
		}
		user, e := GetUser(user_id)
		if e != nil {
			return nil, e
		}
		team.Mutex.Lock()
		team.Members[user_id] = user
		team.Mutex.Unlock()
		if is_captain {
			team.Captain = user
		}
	}
	GMutex.Lock()
	MapIdToTeam[teamId] = team
	GMutex.Unlock()
	return team, nil
}

// try to read data in ram,
// if cant: read data from database
func GetTeam(teamId int64) (*Team, error) {
	GMutex.Lock()
	t := MapIdToTeam[teamId]
	GMutex.Unlock()
	if t != nil {
		return t, nil
	} else {
		return LoadTeam(teamId)
	}
}

func LoadTeamIdByName(teamName string) (int64, error) {
	var teamId int64
	row := zdatabase.DbPool.QueryRow(
		`SELECT team_id FROM team WHERE team_name = $1`, teamName)
	e := row.Scan(&teamId)
	if e != nil {
		return 0, e
	}
	return teamId, nil
}

func AddTeamMember(teamId int64, userId int64) error {
	_, err := zdatabase.DbPool.Exec(
		`INSERT INTO team_member (team_id, user_id) VALUES ($1, $2)`,
		teamId, userId,
	)
	if err != nil {
		pqErr, isOk := err.(*pq.Error)
		if !isOk {
			return err
		}
		if pqErr.Code.Name() == "unique_violation" {
			_ = fmt.Print
			// fmt.Printf("oe %v\n%v\n%v\n", pqErr.Code.Name(), pqErr.Detail, pqErr.Constraint)
			return errors.New(l.Get(l.M015MemberMultipleTeam))
		} else {
			return err
		}
	}
	//
	team, err := GetTeam(teamId)
	if err != nil {
		return errors.New("AddTeamMember GetTeam " + err.Error())
	}
	user, err := GetUser(userId)
	if err != nil {
		return errors.New("AddTeamMember GetUser " + err.Error())
	}
	team.Mutex.Lock()
	team.Members[userId] = user
	team.Mutex.Unlock()
	user.TeamId = team.TeamId
	return nil
}

func RemoveTeamMember(teamId int64, userId int64) error {
	_, err := zdatabase.DbPool.Exec(
		`DELETE FROM team_member
		WHERE team_id = $1 AND user_id = $2`,
		teamId, userId,
	)
	if err != nil {
		return err
	}
	//
	team, err := GetTeam(teamId)
	if err != nil {
		return err
	}
	team.Mutex.Lock()
	delete(team.Members, userId)
	team.Mutex.Unlock()
	user, _ := GetUser(userId)
	if user == nil {
		return errors.New(l.Get(l.M022InvalidUserId))
	}
	user.TeamId = 0
	return nil
}

func SetTeamCaptain(teamId int64, userId int64) error {
	r, e := zdatabase.DbPool.Exec(
		`UPDATE team_member
		SET is_captain = TRUE
		WHERE  team_id = $1 AND user_id = $2`,
		teamId, userId,
	)
	if e != nil {
		return errors.New(l.Get(l.M016TeamMultipleCaptain))
	}
	nRowsAffected, _ := r.RowsAffected()
	if nRowsAffected == 0 {
		return errors.New(l.Get(l.M013SetTeamCaptainOutsider))
	}
	//
	team, err := GetTeam(teamId)
	if err != nil {
		return err
	}
	user, err := GetUser(userId)
	if err != nil {
		return err
	}
	team.Captain = user
	return nil
}

func RequestJoinTeam(teamId int64, userId int64) error {
	_, e := zdatabase.DbPool.Exec(
		`INSERT INTO team_joining_request
    		(team_id, user_id)
    	VALUES ($1, $2)`,
		teamId, userId,
	)
	if e != nil {
		return errors.New(l.Get(l.M014DuplicateTeamJoiningRequest))
	}
	return nil
}

func RemoveRequestJoinTeam(teamId int64, userId int64) error {
	_, e := zdatabase.DbPool.Exec(
		`DELETE FROM team_joining_request
    	WHERE team_id = $1 AND user_id = $2`,
		teamId, userId,
	)
	return e
}

func LoadTeamJoiningRequests(teamId int64) ([]map[string]interface{}, error) {
	rows, err := zdatabase.DbPool.Query(
		`SELECT user_id, created_time
        FROM team_joining_request
        WHERE team_id = $1`,
		teamId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userId int64
		var createdTime time.Time
		err = rows.Scan(&userId, &createdTime)
		if err != nil {
			return nil, err
		}
		result = append(result, map[string]interface{}{
			"TeamId":      teamId,
			"UserId":      userId,
			"CreatedTime": createdTime,
		})
	}
	return result, nil
}
