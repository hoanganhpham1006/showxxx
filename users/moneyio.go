package users

import (
	//	"errors"
	//	"time"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/go-martini/martini"

	m "github.com/daominah/livestream/misc"
	"github.com/daominah/livestream/zconfig"
	//	"github.com/daominah/livestream/zdatabase"
	//	"github.com/daominah/livestream/zglobal"
)

func MoneyIOCharge(user_id int64, data map[string]interface{}) (
	map[string]interface{}, error) {
	charging_type := m.ReadString(data, "ChargingType")
	//	card_vendor := m.ReadString(data, "CardVendor")
	//	card_serial := m.ReadString(data, "CardSerial")
	//	card_code := m.ReadString(data, "CardCode")
	//	bank_name := m.ReadString(data, "BankName")
	//	bank_vnd_value := m.ReadFloat64(data, "BankVndValue")
	switch charging_type {
	case "paytrust":

	}
	return nil, nil
}

func MoneyIOWithdraw(user_id int64, data map[string]interface{}) (
	map[string]interface{}, error) {
	// TODO
	return nil, nil
}

// listen and handle instant payment msg from 3rd party
func IpnCreateRouter() *martini.ClassicMartini {
	r := martini.Classic()

	r.Post("/paytrust/charge")
	r.Post("/paytrust/withdraw")

	return r
}

func IpnListenAndServe() {
	r := IpnCreateRouter()
	fmt.Printf("Listening ipn on address %v\n", zconfig.IPNPort)
	go r.RunOnAddr(zconfig.IPNPort)
}

func IpnPaytrustCharge(
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

func IpnPaytrustWithdraw(
	r *http.Request, w http.ResponseWriter, p martini.Params) string {
	return ""
}
