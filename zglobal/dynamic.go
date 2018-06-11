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

func init() {
	// default values
	ExVar1Default := map[string]int64{"a": 1, "b": 2}
	CheerTaxDefault := float64(0.05)
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
			time.Sleep(5 * time.Second)
		}
	}()
}
