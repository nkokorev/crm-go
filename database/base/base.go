package base


import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var Db *sql.DB

func GetDB() *sql.DB {
	return Db
}


func Connect() *sql.DB{


	env := os.Getenv("ENV_VAR");

	if len(env) > 0 {
		e := godotenv.Load("/home/mex388/go/src/github.com/nkokorev/crm-go/.env." + env)
		if e != nil {
			log.Fatal("Error loading test .env file")
		}
	} else {
		e := godotenv.Load()
		if e != nil {
			log.Fatal("Error loading .env file")
		}
	}

	username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	//dbHost := os.Getenv("db_host")
	//dbType := os.Getenv("db_type")

	dbUri := fmt.Sprintf("%s:%s@/%s?parseTime=true&charset=utf8", username,password,dbName )

	db, err := sql.Open("mysql", dbUri)

	if err != nil {
		log.Fatal("Cant open sql connection", err)
	}

	// Пингуем
	if err = db.Ping(); err != nil {
		log.Fatal("Нет пинга к БД",err)
	}

	Db = db

	return db
}




