package conversations

import (
	//	"errors"
	"fmt"

	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zdatabase"
	"github.com/daominah/livestream/zglobal"
)

func Cheer(conversation_id int64, cheerer_id int64, target_user_id int64,
	cheer_type string, val float64, cheer_message string, misc string) error {
	conversation, err := GetConversation(conversation_id)
	_ = conversation
	if err != nil {
		return err
	}
	//	fmt.Println("hihi", zglobal.CheerTax, zglobal.CheerTeamMainProfit, zglobal.CheerTeamCaptainProfit)
	_, _, err = users.TransferMoney(cheerer_id, target_user_id, users.MT_CASH,
		val, users.REASON_CHEER, zglobal.CheerTax)
	if err != nil {
		return err
	}
	//
	var team_id int64
	valAfterTax := val * (1 - zglobal.CheerTax)
	if cheer_type == CHEER_FOR_TEAM {
		targetUser, _ := users.GetUser(target_user_id)
		if targetUser != nil {
			if targetUser.TeamId != 0 {
				team, _ := users.GetTeam(targetUser.TeamId)
				if team != nil {
					team_id = targetUser.TeamId
					team.Mutex.Lock()
					moneyPerMember := valAfterTax * (1 - zglobal.CheerTeamMainProfit - zglobal.CheerTeamCaptainProfit) / float64(len(team.Members))
					for uid, _ := range team.Members {
						users.TransferMoney(target_user_id, uid, users.MT_CASH,
							moneyPerMember, users.REASON_CHEER_TEAM_SPLIT_MEMBER, 0)
					}
					team.Mutex.Unlock()
					if team.Captain != nil {
						users.TransferMoney(target_user_id, team.Captain.Id,
							users.MT_CASH, valAfterTax*zglobal.CheerTeamCaptainProfit,
							users.REASON_CHEER_TEAM_SPLIT_CAPTAIN, 0)
					}
				}
			}
		}
	}
	//
	_, _ = zdatabase.DbPool.Exec(
		`INSERT INTO conversation_cheer
    		(conversation_id, cheerer_id, target_user_id, cheer_type,
    		val, cheer_message, misc, team_id)
    	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		conversation_id, cheerer_id, target_user_id, cheer_type,
		val, cheer_message, misc, team_id)
	cheerName, _ := users.GetProfilenameById(cheerer_id)
	targetName, _ := users.GetProfilenameById(target_user_id)
	cheer_info := fmt.Sprintf("%v cheer %v $%v:\n %v",
		cheerName, targetName, val, cheer_message)
	//
	CreateMessage(conversation_id, cheerer_id, cheer_info, DISPLAY_TYPE_CHEER)
	return nil
}
