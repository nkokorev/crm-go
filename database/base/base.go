package base

import (
	//_ "database/sql"
	//_ "database/sql/driver"
	"fmt"
	//_ "github.com/jackc/pgx"
	"github.com/jinzhu/gorm"
	//_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}


func init() {

	//fmt.Println("Инициализация БД CRM-GO")

	if os.Getenv("ENV_VAR") == "test" {
		e := godotenv.Load("/home/mex388/go/src/github.com/nkokorev/crm-go/.env.test")
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
	dbType := os.Getenv("db_type")

	//dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username, dbName, password)
	dbUri := fmt.Sprintf("%s:%s@/%s?parseTime=true&charset=utf8", username,password,dbName )
	//dbUri := fmt.Sprintf("ratuscrm:pestik@/crm_test?charset=utf8&parseTime=True&loc=Local")
	//fmt.Println(dbUri)
	//conn, err := gorm.Open("postgres", dbUri)
	conn, err := gorm.Open(dbType, dbUri)
	if err != nil {
		fmt.Print(err)
	}
	conn.LogMode(false)
	db = conn
	if err != nil {
		fmt.Print(err)
	}
	//conn.LogMode(false)
	db = conn

}




