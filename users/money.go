package users

import (
	"encoding/json"
	"errors"
	//	"fmt"
	"time"

	l "github.com/daominah/livestream/language"
	"github.com/daominah/livestream/rank"
	"github.com/daominah/livestream/zdatabase"
)

//
func ViewMoneyLog(userId int64, fromTime time.Time, toTime time.Time) (
	[]map[string]interface{}, error) {
	rows, e := zdatabase.DbPool.Query(
		`SELECT money_type, changed_val, money_before, money_after,
    		reason, created_time
        FROM user_money_log
        WHERE user_id = $1 AND $2 <= created_time AND created_time <= $3
        ORDER BY created_time DESC `,
		userId, fromTime, toTime)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	rs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var changed_val, money_before, money_after float64
		var reason, money_type string
		var created_time time.Time
		e := rows.Scan(&money_type, &changed_val, &money_before, &money_after,
			&reason, &created_time)
		if e != nil {
			return nil, e
		}
		r := map[string]interface{}{
			"UserId":      userId,
			"CreatedTime": created_time,
			"MoneyType":   money_type,
			"ChangedVal":  changed_val,
			"MoneyBefore": money_before,
			"MoneyAfter":  money_after,
			"Reason":      reason,
		}
		rs = append(rs, r)
	}
	return rs, nil
}

// change money in database,
// return moneyAfterChanged, databaseError, logicError
func changeUserMoney(
	userId int64, moneyType string, change float64, reason string,
	constraintPositiveMoney bool) (
	float64, error, error) {
	tx, err := zdatabase.DbPool.Begin()
	if err != nil {
		return -1, err, nil
	}
	_, err = tx.Exec(`SET TRANSACTION ISOLATION LEVEL Serializable`)
	if err != nil {
		tx.Rollback()
		return -1, err, nil
	}
	var val float64
	{
		stmt, err := tx.Prepare(
			`SELECT val FROM user_money
			WHERE user_id = $1 AND money_type = $2 `)
		if err != nil {
			tx.Rollback()
			return -1, err, nil
		}
		defer stmt.Close()
		row := stmt.QueryRow(userId, moneyType)
		err = row.Scan(&val)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				tx.Rollback()
				return -1, nil, errors.New(l.Get(l.M019MoneyTypeNotExist))
			}
			tx.Rollback()
			return -1, err, nil
		}
	}
	if constraintPositiveMoney && (val+change < 0) {
		tx.Rollback()
		return -1, nil, errors.New(l.Get(l.M018NotEnoughMoney))
	}
	{
		stmt, err := tx.Prepare(
			`UPDATE user_money SET val = $1
			WHERE user_id = $2 AND money_type = $3 `)
		if err != nil {
			tx.Rollback()
			return -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(val+change, userId, moneyType)
		if err != nil {
			tx.Rollback()
			return -1, err, nil
		}
	}
	{
		stmt, err := tx.Prepare(
			`INSERT INTO user_money_log
			(user_id, money_type, changed_val, money_before, money_after, reason)
			VALUES ($1, $2, $3, $4, $5, $6) `)
		if err != nil {
			tx.Rollback()
			return -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(userId, moneyType, change, val, val+change, reason)
		if err != nil {
			tx.Rollback()
			return -1, err, nil
		}
	}
	return val + change, tx.Commit(), nil
}

// return moneyAfterChanged, error
func ChangeUserMoney(
	userId int64, moneyType string, change float64, reason string,
	constraintPositiveMoney bool) (
	float64, error) {
	user, e := GetUser(userId)
	if user == nil {
		return -1, e
	}
	timeout := time.Now().Add(5 * time.Second)
	var newVal float64
	var databaseError, logicError error
	for time.Now().Before(timeout) {
		newVal, databaseError, logicError = changeUserMoney(
			userId, moneyType, change, reason, constraintPositiveMoney)
		if databaseError == nil {
			break
		}
	}
	var resultError error
	if databaseError != nil {
		resultError = databaseError
	} else {
		resultError = logicError
	}
	// update cache
	if databaseError == nil && logicError == nil {
		GMutex.Lock()
		if MapIdToUser[userId] != nil {
			MapIdToUser[userId].Mutex.Lock()
			MapIdToUser[userId].MapMoney[moneyType] = newVal
			MapIdToUser[userId].Mutex.Unlock()
		}
		GMutex.Unlock()
	}
	//
	return newVal, resultError
}

// change money in database,
// input transferValue is a positive value, userId lose money, targetId gain money
// return moneyAfterSender, moneyAfterTarget, databaseError, logicError,
func transferMoney(
	userId int64, targetId int64,
	moneyType string, transferValue float64, reason string, tax float64) (
	float64, float64, error, error) {
	tx, err := zdatabase.DbPool.Begin()
	if err != nil {
		return -1, -1, err, nil
	}
	_, err = tx.Exec(`SET TRANSACTION ISOLATION LEVEL Serializable`)
	if err != nil {
		tx.Rollback()
		return -1, -1, err, nil
	}
	// change sender money
	var val float64
	{
		stmt, err := tx.Prepare(
			`SELECT val FROM user_money
			WHERE user_id = $1 AND money_type = $2 `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		row := stmt.QueryRow(userId, moneyType)
		err = row.Scan(&val)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
				tx.Rollback()
				return -1, -1, nil, errors.New(l.Get(l.M019MoneyTypeNotExist))
			}
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	if val-transferValue < 0 {
		tx.Rollback()
		return -1, -1, nil, errors.New(l.Get(l.M018NotEnoughMoney))
	}
	{
		stmt, err := tx.Prepare(
			`UPDATE user_money SET val = $1
			WHERE user_id = $2 AND money_type = $3 `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(val-transferValue, userId, moneyType)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	{
		stmt, err := tx.Prepare(
			`INSERT INTO user_money_log
			(user_id, money_type, changed_val, money_before, money_after, reason)
			VALUES ($1, $2, $3, $4, $5, $6) `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(
			userId, moneyType, -transferValue, val, val-transferValue, reason)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	// change target money
	var val2 float64
	receivedValue := transferValue * (1 - tax)
	{
		stmt, err := tx.Prepare(
			`SELECT val FROM user_money
			WHERE user_id = $1 AND money_type = $2 `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		row := stmt.QueryRow(targetId, moneyType)
		err = row.Scan(&val2)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	{
		stmt, err := tx.Prepare(
			`UPDATE user_money SET val = $1
			WHERE user_id = $2 AND money_type = $3 `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(val2+receivedValue, targetId, moneyType)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	{
		miscMap := map[string]interface{}{"tax": tax}
		miscBs, _ := json.Marshal(miscMap)
		misc := string(miscBs)
		stmt, err := tx.Prepare(
			`INSERT INTO user_money_log
			(user_id, money_type, changed_val, money_before, money_after, reason, misc)
			VALUES ($1, $2, $3, $4, $5, $6, $7) `)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
		defer stmt.Close()
		_, err = stmt.Exec(targetId, moneyType, receivedValue,
			val2, val2+receivedValue, reason, misc)
		if err != nil {
			tx.Rollback()
			return -1, -1, err, nil
		}
	}
	return val - transferValue, val2 + receivedValue, tx.Commit(), nil
}

// input transferValue is a positive value, userId lose money, targetId gain money,
// return moneyAfterSender, moneyAfterTarget, error
func TransferMoney(
	userId int64, targetId int64,
	moneyType string, transferValue float64, reason string, tax float64) (
	float64, float64, error) {
	timeout := time.Now().Add(5 * time.Second)
	var moneyAfterSender, moneyAfterTarget float64
	var databaseError, logicError error
	for time.Now().Before(timeout) {
		moneyAfterSender, moneyAfterTarget, databaseError, logicError = transferMoney(
			userId, targetId, moneyType, transferValue, reason, tax)
		if databaseError == nil {
			break
		}
	}
	var resultError error
	if databaseError != nil {
		resultError = databaseError
	} else {
		resultError = logicError
	}
	// update cache and rank
	if databaseError == nil && logicError == nil {
		GMutex.Lock()
		if MapIdToUser[userId] != nil {
			MapIdToUser[userId].Mutex.Lock()
			MapIdToUser[userId].MapMoney[moneyType] = moneyAfterSender
			MapIdToUser[userId].Mutex.Unlock()
		}
		if MapIdToUser[targetId] != nil {
			MapIdToUser[targetId].Mutex.Lock()
			MapIdToUser[targetId].MapMoney[moneyType] = moneyAfterTarget
			MapIdToUser[targetId].Mutex.Unlock()
		}
		GMutex.Unlock()
		//
		for _, rankId := range []int64{
			rank.RANK_SENT_CASH_DAY,
			rank.RANK_SENT_CASH_WEEK,
			rank.RANK_SENT_CASH_MONTH,
			rank.RANK_SENT_CASH_ALL} {
			rank.ChangeKey(rankId, userId, transferValue)
		}
		for _, rankId := range []int64{
			rank.RANK_RECEIVED_CASH_DAY,
			rank.RANK_RECEIVED_CASH_WEEK,
			rank.RANK_RECEIVED_CASH_MONTH,
			rank.RANK_RECEIVED_CASH_ALL} {
			rank.ChangeKey(rankId, targetId, transferValue)
		}
	}
	//
	return moneyAfterSender, moneyAfterTarget, resultError
}
