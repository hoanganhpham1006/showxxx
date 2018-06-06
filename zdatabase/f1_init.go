package zdatabase

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/daominah/livestream/zconfig"
	_ "github.com/lib/pq"
)

var DbPool *sql.DB

func init() {
	createDbPool()
}

func createDbPool() {
	var err error
	dataSource := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		zconfig.PostgresUsername, zconfig.PostgresPassword,
		zconfig.PostgresAddress, zconfig.PostgresDatabaseName)
	DbPool, err = sql.Open("postgres", dataSource)
	if err != nil {
		panic(err)
	}
	DbPool.SetMaxIdleConns(20)
	DbPool.SetMaxOpenConns(40)
	err = DbPool.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Initialized dbPool %+v\n", dataSource)
}

func InitTables() {
	queriesB, err := ioutil.ReadFile(zconfig.PostgresInitTablesFile)
	if err != nil {
		fmt.Println("ERROR InitTables1", err)
	}
	queries := string(queriesB)
	for _, query := range strings.Split(queries, ";\n") {
		fmt.Println("___________________________________")
		fmt.Println("query: ", strings.Replace(query, "\n", " ", -1))
		_, err = DbPool.Exec(query)
		if err != nil {
			fmt.Println("ERROR InitTables2 ", err)
		}
	}
	fmt.Println("Initialized tables")
}
