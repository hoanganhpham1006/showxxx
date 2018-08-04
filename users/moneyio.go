package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-martini/martini"

	l "github.com/daominah/livestream/language"
	m "github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
	"github.com/daominah/livestream/zdatabase"
	"github.com/daominah/livestream/zglobal"
)

func MoneyIOCharge(user_id int64, data map[string]interface{}) (
	map[string]interface{}, error) {
	charging_type := m.ReadString(data, "ChargingType")
	//	CardVendor := m.ReadString(data, "CardVendor")
	//	CardSerial := m.ReadString(data, "CardSerial")
	//	CardCode := m.ReadString(data, "CardCode")
	BankName := m.ReadString(data, "BankName")
	BankVndValue := m.ReadFloat64(data, "BankVndValue")
	switch charging_type {
	case "paytrust":
		bankCode, isIn := zglobal.MoneyIOPaytrustMapBankNameToBankCode[BankName]
		if !isIn {
			return nil, errors.New(l.Get(l.M038InvalidBankName))
		}
		in_app_value := BankVndValue * zglobal.MoneyIORateBankCharging
		_, _ = bankCode, in_app_value
		return map[string]interface{}{
			"UrlToPaymentSite": "TODO",
			"VndValue":         BankVndValue,
			"InAppValue":       in_app_value,
		}, nil
	default:
		return nil, nil
	}
}

func MoneyIOWithdraw(user_id int64, data map[string]interface{}) (
	map[string]interface{}, error) {
	// TODO
	return nil, nil
}

// return chargingId, error
func chargingInitDbRow(
	user_id int64, charging_type string, charging_input map[string]interface{}) (
	int64, error) {
	var chargingId int64
	charging_inputS, _ := json.Marshal(charging_input)
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO finance_charge (user_id, charging_type, charging_input)
		VALUES ($1, $2, $3) RETURNING id`,
		user_id, charging_type, string(charging_inputS))
	err := row.Scan(&chargingId)
	if err != nil {
		return 0, err
	}
	return chargingId, nil
}

// return error
func chargingSaveThirdPartyResponse(chargingId int64,
	http_request string, http_response string,
	vnd_value float64, transaction_id_3rd_party string,
	is_successful bool, error_message string) error {
	_, err := zdatabase.DbPool.Exec(
		`UPDATE finance_charge
        SET http_request = $1, http_response = $2, vnd_value = $3, 
            transaction_id_3rd_party = $4, is_successful = $5, error_message = $6,
            last_modified = $7
        WHERE id = $8`,
		http_request, http_response, vnd_value,
		transaction_id_3rd_party, is_successful, error_message,
		time.Now(), chargingId)
	return err
}

func chargingChangeInAppMoney(
	chargingId int64, user_id int64, in_app_value float64) error {
	_, err, money_log_id := ChangeUserMoney2(
		user_id, MT_CASH, in_app_value, REASON_CHARGE, false)
	if err != nil {
		return err
	}
	_, err = zdatabase.DbPool.Exec(
		`UPDATE finance_charge
        SET in_app_value = $1, money_log_id = $2, last_modified = $3
        WHERE id = $4`,
		in_app_value, money_log_id, time.Now(),
		chargingId)
	return err
}

// return withdrawingId, error
func withdrawingInitDbRow(
	user_id int64, withdrawing_type string, in_app_value float64, vnd_value float64) (
	int64, error) {
	_, err, money_log_id := ChangeUserMoney2(
		user_id, MT_CASH, -in_app_value, REASON_WITHDRAW, true)
	if err != nil {
		return 0, err
	}
	var withdrawingId int64
	row := zdatabase.DbPool.QueryRow(
		`INSERT INTO finance_withdraw
    		(user_id, withdrawing_type, in_app_value, vnd_value, money_log_id)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		user_id, withdrawing_type, in_app_value, vnd_value, money_log_id)
	err = row.Scan(&withdrawingId)
	if err != nil {
		return 0, err
	}
	return withdrawingId, nil
}

func withdrawingAdminDeny(withdrawingId int64, denied_reason string) error {
	tx, err := zdatabase.DbPool.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`SET TRANSACTION ISOLATION LEVEL Serializable`)
	if err != nil {
		tx.Rollback()
		return err
	}
	var user_id int64
	var in_app_value float64
	var is_denied_by_admin bool
	{
		stmt, err := tx.Prepare(
			`SELECT user_id, in_app_value, is_denied_by_admin
			FROM finance_withdraw WHERE id = $1`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		row := stmt.QueryRow(withdrawingId)
		err = row.Scan(&user_id, &in_app_value, &is_denied_by_admin)
		if err != nil {
			tx.Rollback()
			return err
		}
		if is_denied_by_admin {
			tx.Rollback()
			return errors.New(l.Get(l.M039CanOnlyDenyWithdrawOnce))
		}
	}
	{
		stmt, err := tx.Prepare(
			`UPDATE finance_withdraw
              SET is_denied_by_admin = true, denied_reason = $1, last_modified = $2
				WHERE id = $3`)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(denied_reason, time.Now(), withdrawingId)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	ChangeUserMoney(user_id, MT_CASH, in_app_value, REASON_ADMIN_DENY_WITHDRAW, false)
	return nil
}

// return error
func withdrawingSaveThirdPartyResponse(withdrawingId int64,
	http_request string, http_response string,
	transaction_id_3rd_party string, withdrawing_output string,
	is_successful bool, error_message string) error {
	_, err := zdatabase.DbPool.Exec(
		`UPDATE finance_withdraw
        SET http_request = $1, http_response = $2, withdrawing_output = $3, 
            transaction_id_3rd_party = $4, is_successful = $5, error_message = $6,
            last_modified = $7
        WHERE id = $8`,
		http_request, http_response, withdrawing_output,
		transaction_id_3rd_party, is_successful, error_message,
		time.Now(), withdrawingId)
	return err
}

// listen and handle instant payment msg from 3rd party
func ipnCreateRouter() *martini.ClassicMartini {
	r := martini.Classic()

	r.Post("/paytrust/charge", ipnPaytrustCharge)
	r.Post("/paytrust/withdraw", ipnPaytrustWithdraw)

	return r
}

func IpnListenAndServe() {
	r := ipnCreateRouter()
	fmt.Printf("Listening ipn on address %v\n", zconfig.IPNPort)
	go r.RunOnAddr(zconfig.IPNPort)
}

func ipnPaytrustCharge(
	r *http.Request, w http.ResponseWriter, p martini.Params) string {
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
	return ""
}

func ipnPaytrustWithdraw(
	r *http.Request, w http.ResponseWriter, p martini.Params) string {
	return ""
}
