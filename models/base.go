package models

import (
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"log"

	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"os"
)

var db *gorm.DB

func GetDB() *gorm.DB {

	if db == nil { return Connect()	}

	return db
}

func SetDB(p *gorm.DB)  {
	db = p
}

func GetPool() *gorm.DB {
	return db
}


func Connect() *gorm.DB {

	if db != nil {
		return GetDB()
	}

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

	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error connect to DB")
	}
	//db = db.Set("gorm:auto_preload", true)

	//db = db.LogMode(true)

	SetDB(db)

	return db
}





