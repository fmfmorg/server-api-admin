package postgresdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"server-api-admin/config"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB
)

func init() {
	var err error
	DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s", config.DBHost, config.DBPort, config.DBUsername, config.DBName, config.DBSslMode))
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	sqlFile, err := os.ReadFile("../database/schema.sql")
	if err != nil {
		fmt.Println("error 1111111")
		panic(err)
	}

	_, err = DB.Exec(string(sqlFile))
	if err != nil {
		fmt.Println("error 2222222")
		panic(err)
	}

	fmt.Println("database init complete")
}
