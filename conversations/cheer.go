package conversations

import (
	"time"
	"errors"
	"fmt"

	// "github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/users"
	"github.com/daominah/livestream/zdatabase"
	"github.com/daominah/livestream/zglobal"

	"sync"
)

var GiftList = make([]Gift, 0)

type Gift struct {
	Id			int64
	Name		string
	Value		float64
	Image 	string

	Mutex sync.Mutex
}

func init() {
	go CJGetListGift()
}

func CJGetListGift() {
	for {
		GMutex.Lock()
		GiftList = nil
		rows, e := zdatabase.DbPool.Query(
			`SELECT id, name, val, image FROM gift`,
		)
		if e != nil {
			continue;
		}
		for rows.Next() {
			var id int64
			var name string
			var value float64
			var image string

			e1 := rows.Scan(&id, &name, &value, &image)

			if e1 != nil {
				fmt.Println(e1)
				break
			}

			gift := Gift{Id: id, Name:name, Value:value, Image:image}

			GiftList = append(GiftList, gift)
		}
		GMutex.Unlock()
		time.Sleep(30*time.Minute)
	}
} 

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

func CheerGift(conversation_id int64, cheerer_id int64, target_user_id int64, 
	cheer_type string, gift_id int64, cheer_message string) error {
		hadGiftId := false
		var gift Gift
		for _, tmp := range GiftList {
			if tmp.Id == gift_id {
				gift = tmp
				hadGiftId = true
				break
			}
		}
		if !hadGiftId {
			return errors.New("Khong ton tai gift_id")
		}
		val := gift.Value

		conversation, err := GetConversation(conversation_id)
		_ = conversation
		if err != nil {
			return err
		}
		_, _, err = users.TransferMoney(cheerer_id, target_user_id, users.MT_CASH,
			val, users.REASON_CHEER, zglobal.CheerTax)
		if err != nil {
			return err
		}
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
		_, _ = zdatabase.DbPool.Exec(
			`INSERT INTO conversation_cheer
					(conversation_id, cheerer_id, target_user_id, cheer_type,
					val, cheer_message, gift_id, team_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			conversation_id, cheerer_id, target_user_id, cheer_type,
			val, cheer_message, gift_id, team_id)
		cheerName, _ := users.GetProfilenameById(cheerer_id)
		targetName, _ := users.GetProfilenameById(target_user_id)
		cheer_info := fmt.Sprintf("%v cheer %v $%v:\n %v",
			cheerName, targetName, val, cheer_message)
		//
		CreateCheerMessage(conversation_id, cheerer_id, cheer_info, gift_id, DISPLAY_TYPE_CHEER)
		return nil

}

func LoadGiftsList() ([]map[string]interface{}, error) {
	rows, err := zdatabase.DbPool.Query(`SELECT id, name, val, image FROM gift`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	gifts := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var name, image string
		var val float64
		e := rows.Scan(&id, &name, &val, &image)
		if e != nil {
			return nil, err
		}
		gifts = append(gifts, map[string]interface{}{
			"Id": id, "Name": name, "Val": val, "Image": image,
		})
	}
	return gifts, nil
}
