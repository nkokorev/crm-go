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
	models.ConnectDb()
	// defer db.Close()
	// defer pool.Close()


	// base.Test()

	if false {
		base.RefreshTablesPart_I()
		base.UploadTestDataPart_I()
		base.LoadImagesAiroClimate(13)
		base.LoadArticlesAiroClimate()
		base.LoadProductDescriptionAiroClimate()
		base.LoadProductCategoryDescriptionAiroClimate()
		base.UploadTestDataPart_II()
		base.UploadTestDataPart_III()
		base.UploadTestDataPart_IV()
		base.UploadBroUserData()

		base.UploadTestDataPart_V()
		base.Upload357grData()
	}

	if err := models.SettingsDb(); err != nil {
		log.Fatal(err)
	}


	// base.Migrate_I()
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
