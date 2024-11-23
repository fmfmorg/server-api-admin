package postgresdb

import (
	"database/sql"
	"fmt"
	"log"
	"server-api-admin/config"

	_ "github.com/lib/pq"
)

var (
	DB *sql.DB
)

func init() {
	fmt.Println("host: ", config.DBHost)
	fmt.Println("port: ", config.DBPort)
	fmt.Println("password: ", config.DBPassword)
	fmt.Println("username: ", config.DBUsername)
	fmt.Println("db name: ", config.DBName)
	fmt.Println("sslmode: ", config.DBSslMode)
	var err error
	DB, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%s password=%s user=%s dbname=%s sslmode=%s", config.DBHost, config.DBPort, config.DBPassword, config.DBUsername, config.DBName, config.DBSslMode))
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("database init complete")
}
