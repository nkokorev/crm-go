package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
	"github.com/nkokorev/crm-go/routes"
	"github.com/ttacon/libphonenumber"
	"log"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error load .env file", err)
	}
}

func main() {
	pool := models.Connect()
	defer pool.Close()

	pool.DB().SetConnMaxLifetime(0)
	pool.DB().SetMaxIdleConns(10)
	pool.DB().SetMaxOpenConns(100)

	// base.RefreshTablesPart_I()
	// base.UploadTestDataPart_I()
	// base.LoadImagesAiroClimate(13)
	// base.LoadArticlesAiroClimate()
	// base.LoadProductDescriptionAiroClimate()
	// base.LoadProductCategoryDescriptionAiroClimate()
	// base.UploadTestDataPart_II()
	// base.UploadTestDataPart_III()

	// yandex payment
	// base.RefreshTablesPart_IV()
	// base.UploadTestDataPart_IV()
	
	base.Migrate_I()

	if err := (models.EventListener{}).ReloadEventHandlers(); err != nil {
		log.Fatal(fmt.Sprintf("Не удалось зарегистрировать EventHandler: %v", err))
	}

	models.RunHttpServer(routes.Handlers())
	// controllers.Keymaker("/home/mex388/go/src/github.com/nkokorev/crm-go/")
}

func runMigration(line string) {
	switch line {
	case "full":
		base.RefreshTables()
	}
}
func examplePhone(numToParse string) {

	//num, err := libphonenumber.Get
	num, err := libphonenumber.Parse(numToParse, "RU")
	if err != nil {
		// Handle error appropriately.
		log.Fatal("Err: ", err)
	}
	formattedNum := libphonenumber.Format(num, libphonenumber.NATIONAL)

	//fmt.Println("Num: ", num)
	fmt.Println("CountryCode: ", *num.CountryCode)
	fmt.Println("National Number: ", *num.NationalNumber)
	fmt.Println("National Formatted: ", formattedNum)
	fmt.Println("RFC3966: ", libphonenumber.Format(num, libphonenumber.RFC3966))
	fmt.Println("INTERNATIONAL: ", libphonenumber.Format(num, libphonenumber.INTERNATIONAL)) // наиболее популярный
	fmt.Println("E164: ", libphonenumber.Format(num, libphonenumber.E164))

	// num is a *libphonenumber.PhoneNumber

}
func SendMail() error {

	// 1. Получаем аккаунт
	acc, err := models.GetAccount(4)
	if err != nil {
		return err
	}

	// 2. Загружаем шаблон из БД
	var et models.EmailTemplate
	err = acc.LoadEntity(&et, 4)
	if err != nil {
		return err
	}

	// 3. Выбираем MailBox
	var ebox4 models.EmailBox
	err = acc.LoadEntity(&ebox4, 4)
	if err != nil {
		return err
	}

	

	// 4. Отправляем шаблон из MailBox
	// err = et.Send(*mb, models.User{Email: "aix27249@yandex.ru"}, "Тест return path")
	err = et.Send(ebox4, models.User{Email: "nkokorev@rus-marketing.ru"}, "Тест return path")
	if err != nil {
		return err
	}

	return nil
}