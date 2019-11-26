package models


import (
	"fmt"
	//"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/jinzhu/gorm"
	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"log"
	"os"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	return db
}

func SetDB(p *gorm.DB)  {
	db = p
}

func GetPool() *gorm.DB {
	return db
}


func Connect() *gorm.DB {

	env := os.Getenv("ENV_VAR");

	if len(env) > 0 {
		e := godotenv.Load("/home/mex388/go/src/githu2b.com/nkokorev/crm-go/.env." + env)
		if e != nil {
			log.Fatal("Error loading test .env file")
		}
	} else {
		e := godotenv.Load()
		if e != nil {
			log.Fatal("Error loading .env file")
		}
	}

	/*username := os.Getenv("db_user")
	password := os.Getenv("db_pass")
	dbName := os.Getenv("db_name")
	dbHost := os.Getenv("db_host")*/
	//dbType := os.Getenv("db_type")

	//dbUri := fmt.Sprintf("%s:%s@/%s?parseTime=true&charset=utf8", username,password,dbName )
	//dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username,dbName, password)

	//db, err := sql.Open("mysql", dbUri)

	/*conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}*/
	//defer conn.Close(context.Background())

	fmt.Println(os.Getenv("DATABASE_URL"))
	//fmt.Println(dbUri)

	//db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}
	SetDB(db)

	return db
}





