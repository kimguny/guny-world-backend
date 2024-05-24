// database/database.go
package database

import (
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

var DB *sqlx.DB

func InitDB() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatal(err)
    }

    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dbName := os.Getenv("DB_NAME")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")

    connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)

    DB, err = sqlx.Open("mysql", connectionString)
    if err != nil {
        log.Fatal(err)
    }

    // 연결 풀 설정
    DB.SetMaxOpenConns(10)
    DB.SetMaxIdleConns(5)

    // 연결 체크
    err = DB.Ping()
    if err != nil {
        log.Fatal(err)
    }
}
