package base


import (
	//"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	_ "github.com/jackc/pgx/v4"
	"log"
	"os"
)

var pool *sqlx.DB

func GetDB() *sqlx.DB {
	return pool
}

func SetDB(p *sqlx.DB)  {
	pool = p
}

func GetPool() *sqlx.DB {
	return pool
}


func Connect() *sqlx.DB {


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
	dbHost := os.Getenv("db_host")
	//dbType := os.Getenv("db_type")

	//dbUri := fmt.Sprintf("%s:%s@/%s?parseTime=true&charset=utf8", username,password,dbName )
	dbUri := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s", dbHost, username,dbName, password)

	//db, err := sql.Open("mysql", dbUri)

	db, err := sqlx.Connect("postgres", dbUri)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(dbUri)

	// Пингуем
	if err = db.Ping(); err != nil {
		log.Fatal("Нет пинга к БД",err)
	}

	SetDB(db)

	return db
}





