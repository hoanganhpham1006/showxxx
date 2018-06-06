package users

import (
	"time"

	"github.com/daominah/livestream/zdatabase"
)

// change money in database,
// return error if it has a concurrent update
func changeUserMoney(
	userId int64, moneyType string, change float64, reason string) (
	float64, error) {
	tx, err := zdatabase.DbPool.Begin()
	if err != nil {
		return -1, err
	}
	_, err = tx.Exec(`SET TRANSACTION ISOLATION LEVEL Serializable`)
	if err != nil {
		tx.Rollback()
		return -1, err
	}
	var val float64
	{
		stmt, err := tx.Prepare(
			`SELECT val FROM user_money
			WHERE user_id = $1 AND money_type = $2 `)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
		defer stmt.Close()
		row := stmt.QueryRow(userId, moneyType)
		err = row.Scan(&val)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}
	{
		stmt, err := tx.Prepare(
			`UPDATE user_money SET val = $1
			WHERE user_id = $2 AND money_type = $3 `)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(val+change, userId, moneyType)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}
	{
		stmt, err := tx.Prepare(
			`INSERT INTO user_money_log
			(user_id, money_type, changed_val, money_before, money_after, reason)
			VALUES ($1, $2, $3, $4, $5, $6) `)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
		defer stmt.Close()
		_, err = stmt.Exec(userId, moneyType, change, val, val+change, reason)
		if err != nil {
			tx.Rollback()
			return -1, err
		}
	}
	return val + change, tx.Commit()
}

//
func ChangeUserMoney(
	userId int64, moneyType string, change float64, reason string) error {
	timeout := time.Now().Add(5 * time.Second)
	var newVal float64
	var e error
	for time.Now().Before(timeout) {
		newVal, e = changeUserMoney(userId, moneyType, change, reason)
		if e == nil {
			break
		}
	}
	// update cache
	if e == nil {
		GMutex.Lock()
		if MapIdToUser[userId] != nil {
			MapIdToUser[userId].Mutex.Lock()
			MapIdToUser[userId].MapMoney[moneyType] = newVal
			MapIdToUser[userId].Mutex.Unlock()
		}
		GMutex.Unlock()
	}
	//
	return e
}

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
