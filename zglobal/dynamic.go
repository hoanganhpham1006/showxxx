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

func init() {
	// default values
	ExVar1Default := map[string]int64{"a": 1, "b": 2}

	CheerTaxDefault := float64(0.05)
	CheerTeamMainProfitDefault := float64(0.85)
	CheerTeamCaptainProfitDefault := float64(0.05)

	MessageBigCostDefault := float64(1000)

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
			time.Sleep(5 * time.Second)
		}
	}()
}
