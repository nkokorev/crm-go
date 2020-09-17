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
		logger.Config {
			SlowThreshold: time.Millisecond*200,   // Slow SQL threshold
			LogLevel:      logger.Silent, // Уровни логирования GORM: Silent, Error, Warn, Info
			Colorful:      true,         // Disable color
		},
	)
	db, err := gorm.Open(postgres.New( postgres.Config {
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

	
	/*if err := SettingsDb(); err != nil {
		log.Println("Ошибка настроек БД", err)
	}*/
	SetDB(db)
	fmt.Println("DataBase init full!")
	return db
}

func SettingsDb() error {
	err := db.SetupJoinTable(&ProductCard{}, "Products", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "ProductCards", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&ProductCategory{}, "ProductCards", &ProductCategoryProductCard{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&ProductCard{}, "ProductCategories", &ProductCategoryProductCard{})
	if err != nil {
		log.Fatal(err)
	}


	if err := db.SetupJoinTable(&Account{}, "Users", &AccountUser{}); err != nil {
		log.Fatal(err)
	}
	if err := db.SetupJoinTable(&User{}, "Accounts", &AccountUser{}); err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Warehouse{}, "Products", &WarehouseItem{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "Warehouses", &WarehouseItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Shipment{}, "Products", &ShipmentItem{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "Shipments", &ShipmentItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&User{}, "Companies", &CompanyUser{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Company{}, "Users", &CompanyUser{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&WebPage{}, "ProductCategories", &WebPageProductCategories{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&ProductCategory{}, "WebPages", &WebPageProductCategories{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&Inventory{}, "Products", &InventoryItem{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "Inventories", &InventoryItem{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&ProductCategory{}, "Products", &ProductCategoryProduct{})
	if err != nil {
		log.Fatal(err)
	}
	err = db.SetupJoinTable(&Product{}, "ProductCategories", &ProductCategoryProduct{})
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
