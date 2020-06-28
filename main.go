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
	pool.DB().SetMaxOpenConns(10)

	//runMigration("full")
	//base.LoadImagesAiroClimate()
	base.LoadArticlesAiroClimate()

	models.RunHttpServer(routes.Handlers())
	//controllers.Keymaker("/home/mex388/go/src/github.com/nkokorev/crm-go/")
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
	et, err := acc.GetEmailTemplate(4)
	if err != nil {
		return err
	}

	// 3. Выбираем MailBox
	mb, err := acc.GetEmailBox(4)
	if err != nil {
		return err
	}

	/*user, err := acc.GetUserById(2)
	if err != nil {
		return err
	}*/

	// 4. Отправляем шаблон из MailBox
	// err = et.Send(*mb, models.User{Email: "aix27249@yandex.ru"}, "Тест return path")
	err = et.Send(*mb, models.User{Email: "nkokorev@rus-marketing.ru"}, "Тест return path")
	if err != nil {
		return err
	}
	
	return nil
}

