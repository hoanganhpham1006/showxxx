package zdatabase

import (
	//	"database/sql"
	"time"
)

func LoadGlobalVar(key string) string {
	var result string
	row := DbPool.QueryRow(
		"SELECT zvalue FROM zglobal_var WHERE zkey = $1",
		key)
	_ = row.Scan(&result)
	return result
}

func SaveGlobalVar(key string, value string) {
	DbPool.Exec(
		"INSERT INTO zglobal_var (zkey, zvalue, last_modified) "+
			"VALUES ($1, $2, $3) "+
			"ON CONFLICT (zkey) DO UPDATE "+
			"  SET zvalue = EXCLUDED.zvalue, last_modified=EXCLUDED.last_modified ",
		key, value, time.Now())
}
