// Package zglobal contains global variables,
// these values update once per minute from database.
package zglobal

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/daominah/livestream/zdatabase"
)

var ExVar1 map[string]int64

var CheerTax float64
var CheerTeamMainProfit float64
var CheerTeamCaptainProfit float64

var MessageBigCost float64

var GameEggPayoutRate float64
var GameCarPayoutRate float64

var MoneyIOAvailableChargingTypes map[string]bool
var MoneyIORateBankCharging float64
var MoneyIOPaytrustMapBankNameToBankCode map[string]string
var MoneyIOPaytrustKey string

func init() {
	// default values
	ExVar1Default := map[string]int64{"a": 1, "b": 2}

	CheerTaxDefault := float64(0.05)
	CheerTeamMainProfitDefault := float64(0.85)
	CheerTeamCaptainProfitDefault := float64(0.05)

	MessageBigCostDefault := float64(1000)

	GameEggPayoutRateDefault := float64(0.90)
	GameCarPayoutRateDefault := float64(0.90)

	MoneyIOAvailableChargingTypesDefault := map[string]bool{
		"paytrust": true,
	}
	MoneyIORateBankChargingDefault := float64(1.0)
	MoneyIOPaytrustMapBankNameToBankCodeDefault := map[string]string{
		"VietinBank":  "5a8d9b3432bc7",
		"BIDV":        "5a8dc25912217",
		"TechComBank": "5a8ee643945a3",
		"SacomBank":   "5a8eec3fc74e6",
		"DongABank":   "5a904bc3775ba"}
	MoneyIOPaytrustKeyDefault := "Basic TjZsWTZxNjAxQll6WkdnSzhYMERtVU1DaUFjSEVDVFE6"

	// loop update values
	go func() {
		time.Sleep(5 * time.Second) // waiting for init record.dbPool

		for {
			var key, value string
			var err error
			//
			key = "ExVar1"
			value = zdatabase.LoadGlobalVar(key)
			err = json.Unmarshal([]byte(value), &ExVar1)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				ExVar1 = ExVar1Default
				temp, _ := json.Marshal(ExVar1Default)
				zdatabase.SaveGlobalVar(key, string(temp))
			}
			//
			key = "CheerTax"
			value = zdatabase.LoadGlobalVar(key)
			CheerTax, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				CheerTax = CheerTaxDefault
				temp := fmt.Sprintf("%v", CheerTaxDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}
			//
			key = "CheerTeamMainProfit"
			value = zdatabase.LoadGlobalVar(key)
			CheerTeamMainProfit, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				CheerTeamMainProfit = CheerTeamMainProfitDefault
				temp := fmt.Sprintf("%v", CheerTeamMainProfitDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}
			//
			key = "CheerTeamCaptainProfit"
			value = zdatabase.LoadGlobalVar(key)
			CheerTeamCaptainProfit, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				CheerTeamCaptainProfit = CheerTeamCaptainProfitDefault
				temp := fmt.Sprintf("%v", CheerTeamCaptainProfitDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}
			//
			key = "MessageBigCost"
			value = zdatabase.LoadGlobalVar(key)
			MessageBigCost, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MessageBigCost = MessageBigCostDefault
				temp := fmt.Sprintf("%v", MessageBigCostDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}
			//
			key = "GameEggPayoutRate"
			value = zdatabase.LoadGlobalVar(key)
			GameEggPayoutRate, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				GameEggPayoutRate = GameEggPayoutRateDefault
				temp := fmt.Sprintf("%v", GameEggPayoutRateDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}

			//
			key = "GameCarPayoutRate"
			value = zdatabase.LoadGlobalVar(key)
			GameCarPayoutRate, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				GameCarPayoutRate = GameCarPayoutRateDefault
				temp := fmt.Sprintf("%v", GameCarPayoutRateDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}

			//
			key = "MoneyIOAvailableChargingTypes"
			value = zdatabase.LoadGlobalVar(key)
			err = json.Unmarshal([]byte(value), &MoneyIOAvailableChargingTypes)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MoneyIOAvailableChargingTypes = MoneyIOAvailableChargingTypesDefault
				temp, _ := json.Marshal(MoneyIOAvailableChargingTypesDefault)
				zdatabase.SaveGlobalVar(key, string(temp))
			}
			//
			key = "MoneyIOPaytrustMapBankNameToBankCode"
			value = zdatabase.LoadGlobalVar(key)
			err = json.Unmarshal([]byte(value), &MoneyIOPaytrustMapBankNameToBankCode)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MoneyIOPaytrustMapBankNameToBankCode = MoneyIOPaytrustMapBankNameToBankCodeDefault
				temp, _ := json.Marshal(MoneyIOPaytrustMapBankNameToBankCodeDefault)
				zdatabase.SaveGlobalVar(key, string(temp))
			}
			//
			key = "MoneyIORateBankCharging"
			value = zdatabase.LoadGlobalVar(key)
			MoneyIORateBankCharging, err = strconv.ParseFloat(value, 64)
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MoneyIORateBankCharging = MoneyIORateBankChargingDefault
				temp := fmt.Sprintf("%v", MoneyIORateBankChargingDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}
			//
			key = "MoneyIOPaytrustKey"
			value = zdatabase.LoadGlobalVar(key)
			MoneyIOPaytrustKey = value
			if err != nil {
				fmt.Println("zglobal err", key, err)
				MoneyIOPaytrustKey = MoneyIOPaytrustKeyDefault
				temp := fmt.Sprintf("%v", MoneyIOPaytrustKeyDefault)
				zdatabase.SaveGlobalVar(key, temp)
			}

			//
			time.Sleep(5 * time.Second)
		}
	}()
}
