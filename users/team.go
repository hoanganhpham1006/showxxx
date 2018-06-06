package users

import (
//	"fmt"
//	"encoding/json"
)

type Team struct {
	TeamId    int64
	CreaterId int64
	TeamName  string
	TeamImage string
	Summary   string
	Members   map[int64]*User
}

//
func (team *Team) ToMap() map[string]interface{} {
	result := make(map[string]interface{})
	return result
}

func CreateTeam(createrId int64, teamName string) {

}

func RequestJoinTeam(userId int64, teamId int64) {

}

func AddUserToTeam(requestId int64) {

}
