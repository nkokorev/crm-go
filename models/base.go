package models

import (
	"fmt"
	"gorm.io/gorm/logger"
	"time"

	// "github.com/go-gorm/gorm"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"

	// _ "github.com/go-sql-driver/mysql"
	// _ "github.com/lib/pq"
	_ "github.com/jackc/pgx/v4"

	"os"
)

/*func init() {
	connectDb()
}*/

var db *gorm.DB

func GetDB() *gorm.DB {

	if db == nil {
		db = ConnectDb()
	}

	return db
}

func SetDB(p *gorm.DB) {
	db = p
}

func GetPool() *gorm.DB {
	return db
}

func ConnectDb() *gorm.DB {

	if db != nil {
		return nil
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

	// db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	// https://github.com/go-gorm/postgres
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Millisecond*200,   // Slow SQL threshold
			LogLevel:      logger.Error, // Уровни логирования GORM: Silent, Error, Warn, Info
			Colorful:      true,         // Disable color
		},
	)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: os.Getenv("DATABASE_URL"),
		// PreferSimpleProtocol: true, // disables implicit prepared statement usage
		PreferSimpleProtocol: false, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: dbLogger,
		SkipDefaultTransaction: false, // ожидать завершения записи
		// DisableForeignKeyConstraintWhenMigrating: true,
		/*NamingStrategy: schema.NamingStrategy{
			TablePrefix: "t_",   // префикс имен таблиц, таблица для `User` будет `t_users`
			SingularTable: true, // использовать именование в единственном числе, таблица для `User` будет `user` при включении этой опции, или `t_user` при TablePrefix = "t_"
		},*/
	})


	if err != nil {
		log.Fatal("Error connect to DB")
	}

	if db == nil {
		log.Fatal("Error connect to DB == nil")
	}

	// db.DB().SetConnMaxLifetime(0)
	// db.DB().SetMaxIdleConns(10)
	// db.DB().SetMaxOpenConns(100)
	

	//db = db.Set("gorm:auto_preload", true)

	// db = db.LogMode(true)

	SetDB(db)
	fmt.Println("DataBase init full!")
	return db
}
