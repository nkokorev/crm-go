package base

import (
	"fmt"
	"github.com/nkokorev/crm-go/models"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func RefreshTables() {


	var err error
	pool := models.GetPool()

	pool.Migrator().DropTable(models.DeliveryPickup{},models.DeliveryRussianPost{}, models.DeliveryCourier{})

	err = pool.Exec("drop table if exists web_hooks, articles").Error
	if err != nil {
		fmt.Println("Cant create tables -1: ", err)
		return
	}

	err = pool.Exec("drop table if exists  unit_measurements, product_card_products, product_cards").Error
	if err != nil {
		fmt.Println("Cant create tables 0: ", err)
		return
	}

	err = pool.Exec("drop table if exists  storage, products, product_groups").Error
	if err != nil {
		fmt.Println("Cant create tables 1: ", err)
		return
	}

	// err = pool.Exec("drop table if exists domains, email_boxes, email_senders, email_templates, api_keys").Error
	err = pool.Exec("drop table if exists api_keys").Error
	if err != nil {
		fmt.Println("Cant create tables 1: ", err)
		return
	}

	err = pool.Exec("drop table if exists crm_settings, roles, account_users, users").Error
	if err != nil {
		fmt.Println("Cant create tables 2: ", err)
		return
	}

	err = pool.Exec("drop table if exists user_verification_methods, accounts").Error
	if err != nil {
		fmt.Println("Cant create tables 3: ", err)
		return
	}



}

func Test() {

	pool := models.GetDB()
	if pool == nil {
		log.Fatal("Not db in crate models")
	}

	err := pool.Exec("DROP SCHEMA public CASCADE;\nCREATE SCHEMA public;").Error
	if err != nil {
		log.Fatal(err)
	}

	/*pool.Migrator().DropTable(
		&models.CrmSetting{},&models.UserVerificationMethod{},
		&models.UsersSegment{},&models.UserSegmentConditions{},
		&models.Storage{},&models.Account{},

		&models.MTABounced{},&models.MTAHistory{},&models.MTAWorkflow{},
		&models.EmailQueueEmailTemplate{},&models.EmailQueue{},&models.EmailCampaign{},&models.TaskScheduler{},&models.Payment2Delivery{},
		&models.DeliveryOrder{},&models.DeliveryStatus{},&models.OrderChannel{},&models.Order{},&models.OrderStatus{},
		&models.Payment{},&models.PaymentAmount{},&models.CartItem{},&models.Comment{},&models.PaymentMode{},&models.PaymentYandex{},
		&models.PaymentCash{},&models.Product{},&models.ProductCard{},&models.ProductGroup{},
		&models.EmailNotification{}, &models.EmailBox{}, &models.WebSite{},&models.EmailTemplate{},
		&models.WebHook{}, &models.EventListener{}, &models.EventItem{},&models.HandlerItem{}, &models.DeliveryPickup{},
		&models.DeliveryRussianPost{}, &models.DeliveryCourier{},
		&models.Article{},  &models.UnitMeasurement{}, &models.ApiKey{}, &models.AccountUser{}, &models.User{},
		&models.VatCode{}, &models.PaymentSubject{},
		&models.Role{},
	)*/

	models.CrmSetting{}.PgSqlCreate()
	models.UserVerificationMethod{}.PgSqlCreate()
	models.Account{}.PgSqlCreate()
	models.User{}.PgSqlCreate()
	models.AccountUser{}.PgSqlCreate()

	models.Role{}.PgSqlCreate()
	models.ApiKey{}.PgSqlCreate()
	models.MeasurementUnit{}.PgSqlCreate()

	models.WebSite{}.PgSqlCreate()
	models.WebPage{}.PgSqlCreate()

	models.PaymentMode{}.PgSqlCreate()
	models.PaymentAmount{}.PgSqlCreate()

	models.OrderChannel{}.PgSqlCreate()
	models.Payment2Delivery{}.PgSqlCreate()
	models.PaymentSubject{}.PgSqlCreate()
	models.PaymentYandex{}.PgSqlCreate()
	models.PaymentCash{}.PgSqlCreate()

	models.Product{}.PgSqlCreate()
	models.ProductCard{}.PgSqlCreate()


	models.EmailBox{}.PgSqlCreate()
	models.EmailTemplate{}.PgSqlCreate()
	models.Storage{}.PgSqlCreate()
	models.Article{}.PgSqlCreate()
	models.HandlerItem{}.PgSqlCreate()
	models.EventItem{}.PgSqlCreate()
	models.EventListener{}.PgSqlCreate()
	models.WebHook{}.PgSqlCreate()
	models.EmailNotification{}.PgSqlCreate()
	models.Comment{}.PgSqlCreate()
	models.EmailQueue{}.PgSqlCreate()
	models.EmailQueueEmailTemplate{}.PgSqlCreate()
	models.UsersSegment{}.PgSqlCreate()
	models.EmailCampaign{}.PgSqlCreate()
	models.TaskScheduler{}.PgSqlCreate()
	models.MTAWorkflow{}.PgSqlCreate()

	models.MTAHistory{}.PgSqlCreate()
	models.MTABounced{}.PgSqlCreate()
	models.UserSegmentCondition{}.PgSqlCreate()



	/*models.PaymentMode{}.PgSqlCreate()
	models.PaymentAmount{}.PgSqlCreate()

	models.OrderChannel{}.PgSqlCreate()
	models.Payment2Delivery{}.PgSqlCreate()
	models.PaymentSubject{}.PgSqlCreate()
	models.PaymentYandex{}.PgSqlCreate()
	models.PaymentCash{}.PgSqlCreate()*/

	models.DeliveryStatus{}.PgSqlCreate()
	models.OrderStatus{}.PgSqlCreate()

	models.DeliveryOrder{}.PgSqlCreate()
	models.VatCode{}.PgSqlCreate()

	models.DeliveryRussianPost{}.PgSqlCreate()
	models.DeliveryPickup{}.PgSqlCreate()
	models.DeliveryCourier{}.PgSqlCreate()


	models.Order{}.PgSqlCreate()
	models.CartItem{}.PgSqlCreate()
	models.Payment{}.PgSqlCreate()

}

func RefreshTablesPart_I() {

	pool := models.GetPool()

	err := pool.Exec("DROP SCHEMA public CASCADE;\nCREATE SCHEMA public;").Error
	if err != nil {
		log.Fatal(err)
	}

	models.CrmSetting{}.PgSqlCreate()
	models.UserVerificationMethod{}.PgSqlCreate()
	models.Account{}.PgSqlCreate()
	models.User{}.PgSqlCreate()
	models.Role{}.PgSqlCreate()
	models.AccountUser{}.PgSqlCreate()


	// не зависящие
	models.PaymentSubject{}.PgSqlCreate()
	models.PaymentMode{}.PgSqlCreate()
	models.VatCode{}.PgSqlCreate()
	models.OrderStatus{}.PgSqlCreate()
	models.MeasurementUnit{}.PgSqlCreate()
	// models.Role{}.PgSqlCreate()
	models.ApiKey{}.PgSqlCreate()
	models.Bank{}.PgSqlCreate()
	models.Storage{}.PgSqlCreate()

	models.WebSite{}.PgSqlCreate()
	models.WebPage{}.PgSqlCreate()

	models.PaymentAccount{}.PgSqlCreate()
	models.Company{}.PgSqlCreate()
	models.Manufacturer{}.PgSqlCreate()

	models.ProductType{}.PgSqlCreate()
	models.ProductCategory{}.PgSqlCreate()
	models.Warehouse{}.PgSqlCreate()
	models.Product{}.PgSqlCreate()
	models.ProductCard{}.PgSqlCreate()
	models.ProductCardProduct{}.PgSqlCreate()
	models.ProductCategoryProduct{}.PgSqlCreate()
	models.WebPageProductCategories{}.PgSqlCreate()

	models.WarehouseItem{}.PgSqlCreate()

	models.Inventory{}.PgSqlCreate()
	models.InventoryItem{}.PgSqlCreate()


	models.Shipment{}.PgSqlCreate()
	models.ShipmentItem{}.PgSqlCreate()

	models.EmailBox{}.PgSqlCreate()
	models.EmailTemplate{}.PgSqlCreate()
	// models.Storage{}.PgSqlCreate()
	models.Article{}.PgSqlCreate()
	models.HandlerItem{}.PgSqlCreate()
	models.EventItem{}.PgSqlCreate()
	models.EventListener{}.PgSqlCreate()
	models.WebHook{}.PgSqlCreate()
	models.EmailNotification{}.PgSqlCreate()
	models.Comment{}.PgSqlCreate()
	models.EmailQueue{}.PgSqlCreate()
	models.EmailQueueEmailTemplate{}.PgSqlCreate()
	models.UsersSegment{}.PgSqlCreate()
	models.EmailCampaign{}.PgSqlCreate()
	models.TaskScheduler{}.PgSqlCreate()
	models.MTAWorkflow{}.PgSqlCreate()

	models.MTAHistory{}.PgSqlCreate()
	models.MTABounced{}.PgSqlCreate()
	models.UserSegmentCondition{}.PgSqlCreate()

	models.PaymentAmount{}.PgSqlCreate()
	models.OrderChannel{}.PgSqlCreate()
	models.Payment2Delivery{}.PgSqlCreate()

	models.PaymentYandex{}.PgSqlCreate()
	models.PaymentCash{}.PgSqlCreate()

	models.DeliveryStatus{}.PgSqlCreate()

	models.DeliveryOrder{}.PgSqlCreate()


	models.DeliveryRussianPost{}.PgSqlCreate()
	models.DeliveryPickup{}.PgSqlCreate()
	models.DeliveryCourier{}.PgSqlCreate()


	models.Order{}.PgSqlCreate()
	models.CartItem{}.PgSqlCreate()
	models.Payment{}.PgSqlCreate()
}

// загрузка первоначальных данных в EAV-таблицы
func UploadEavData() {

	var err error
	pool := models.GetPool()

	// Добавляем в таблицу типов атрибутов EAV-моделей используемы типы (7). Частные типы - не предпологаются.
	err = pool.Exec("insert into eav_attr_type\n    (name, code, table_name, description)\nvalues\n    ('Текстовое поле', 'text_field', 'eav_attr_varchar', 'Текстовое поле для хранения кратких текстовых данных до 255 символов.'),\n    ('Текстовая область', 'text_area', 'eav_attr_varchar', 'Текстовая область для хранения кратких текстовых данных до 255 символов.'),\n    ('Текстовый редактор', 'text_editor', 'eav_attr_text', 'Редактируемый wysiwyg-редактором текст до 16383 символов.'),\n    ('Дата', 'date', 'eav_attr_date', 'Дата в формате UTC.'),\n    ('Да / Нет', 'bool', 'eav_attr_boolean', 'Логический формат данных, который может принимать значение ИСТИНА (1) и ЛОЖЬ (0).'),\n    ('Целое число', 'int', 'eav_attr_int', 'Целое число от -2147483648 до 2147483648.'),\n    ('Десятичное число', 'decimal', 'eav_attr_decimal', 'Знаковое десятичное число, 10 знаков до и после запятой.');").Error
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_type: ", err)
	}

	// Добавляем в таблицу атрибутов EAV-моделей системные атрибуты.
	/*err = pool.Exec("insert into eav_attributes\n    (account_id, eav_attr_type_code, label, code, required, multiple)\nvalues\n    (3,'text_field', 'Имя продукта', 'name', false, false),\n    (3,'text_field', 'Производитель', 'manufactures', false, false),\n    (3,'text_editor', 'Описание', 'description', false, false),\n    (3,'decimal', 'Цена', 'price', false, false),\n    (3,'date', 'Дата производства', 'manufacture_date', false, false),\n    (3,'text_field', 'Размер одежды', 'clothing_size', false, true),\n    (3,'text_field', 'Тип упаковки', 'pkg_type', false, true),\n    (3,'text_field', 'Состав', 'composition', false, false)\n    ").Error
	if err != nil {
		log.Fatal("Cant insert into table eav_attributes: ", err)
	}*/

	// загружаем значения varchar
	/*err = pool.Exec("insert into eav_varchar_values\n    (eav_attr_id, value)\nvalues\n    (6, 'S'), -- Размер одежды\n    (6, 'M'), -- Размер одежды\n    (6, 'L'), -- Размер одежды\n    (7, 'Без упаковки (без упаковки)'), -- Тип упаковки\n    (7, 'Подарочный пакет'), -- Тип упаковки\n    (8, 'хлопок 90%, эластан 10%'),-- Состав\n    (8, 'вискоза 89%, эластан 11%'), -- Состав\n    (8, 'вискоза 89%, эластан 11%'), -- Состав\n    (8, 'хлопок 100%') -- Состав\n    ").Error
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_values_varchar: ", err)
	}*/

}

func UploadTestDataPart_I() {

	// 1. Получаем главный аккаунт
	mAcc, err := models.GetMainAccount()
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	roleOwnerMain, err := mAcc.GetRoleByTag(models.RoleOwner)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}
	roleAdminMain, err := mAcc.GetRoleByTag(models.RoleAdmin)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}
	roleManagerMain, err := mAcc.GetRoleByTag(models.RoleManager)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}
	roleClientMain, err := mAcc.GetRoleByTag(models.RoleClient)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	// 2. Создаем пользователя admin в main аккаунте
	timeNow := time.Now().UTC()
	owner, err := mAcc.CreateUser(
			models.User{
			Username: 		utils.STRp("admin"),
			Email:			utils.STRp("kokorevn@gmail.com"),
			PhoneRegion: 	utils.STRp("RU"),
			Phone: 			utils.STRp("89251952295"),
			Password:		utils.STRp("qwerty109#QW"),
			Name:			utils.STRp("Никита"),
			Surname:		utils.STRp("Кокорев"),
			Patronymic:		utils.STRp("Романович"),
			EmailVerifiedAt:&timeNow,
			},
		*roleOwnerMain,
		)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать admin'a: ", err)
	}

	// Рабочий аккаунт (nkokorev)
	mex388, err := mAcc.CreateUser(
		models.User{
			Username:		utils.STRp("mex388"),
			Email:			utils.STRp("nkokorev@rus-marketing.ru"),
			PhoneRegion: 	utils.STRp("RU"),
			Phone: 			utils.STRp("79251952222"),
			Password:		utils.STRp("qwerty109#QW"),
			Name:			utils.STRp("Никита"),
			Surname:		utils.STRp("Кокорев"),
			Patronymic:		utils.STRp("Романович"),
			EmailVerifiedAt:&timeNow,
		},
		*roleAdminMain,
	)
	if err != nil || mex388 == nil {
		log.Fatal("Не удалось создать mex388'a: ", err)
	}

	// 3. Создаем домен для главного аккаунта
	_webSiteMain, err := mAcc.CreateEntity(&models.WebSite{
		Hostname: "ratuscrm.com",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC4dksLEYhARII4b77fe403uCJhD8x5Rddp9aUJCg1vby7d6QLOpP7uXpXKVLXxaxQcX7Kjw2kGzlvx7N+d2tToZ8+T3SUadZxLOLYDYkwalkP3vhmA3cMuhpRrwOgWzDqSWsDfXgr4w+p1BmNbScpBYCwCrRQ7B12/EXioNcioCQIdAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDB8BPdNbNwi3LA6VMp8BbOGKNrV1PxYZsxp6LvTSK9EgJcRIMw
C+Uc1GgnvcTNksF5GviVYcy2az/e8ACLvcKI6Lb1gUhk10SHIRcb5boK/Li9aOUu
F5ndGzzg0aBzsG2P0us+tkgFOTjc5MuBdlKOzraLegRbfL5MWUWe5SS3FQIdAQAB
AoGANIXli1Jg34kUsgQ+3qvEMVrg31BOTqAlnMQOz4pvbw8yjnSLpvaBvVYVQzYU
16v4M+lHC4XqIdlZmfIb47yns12ASHSoFUzPeUioRu9oWxaOlcHSqWkZBg5miEuM
pCgRrHG9eO3hoa3etgNTKzAUzqS5NhI2F4JXacHgJaQDT30CQQDuyOJfmTFzAz8I
d0IPNjdyuaoLB7Vtzf9b3ihALJx6pvogM7ZcEAgDRlYLfuONMfrsLm3VqNhuMnaX
O4iMyEbnAkEAz+t6qcosS/+J5MOvNQM0yFMLOdvAaJFVg019TSxc4Bp+DWIfUQXf
0rk5d5BmMI0/RRaqKaB5V/oDdh3EiJueowJBALkskdi/DUj64HvpOBJh4hgXAVYy
cTEpCfmtS5uQvPyk1t34HFhCmmQnvHyHt2F8u/FChCyoFsdGXQ8kvN0oR0sCQQCG
8DeinABVrlmq60j5acRGwoaFnVXpR3EtDwxkGoeINgla3DSg2+QgGW/vZfq8Rd8r
EoOLEofODgdTEAyt7/lrAkAJ9HC2mnLKThsXQi8HuU8PMolXv2OA2g45+mCcxkxg
JY0w37/g0vPnSkxvmjyeF8ARRR+FbfL/Tyzhn6r/kf7n
-----END RSA PRIVATE KEY-----`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}
	webSiteMain, ok := _webSiteMain.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать web-site для главного аккаунта: ", err)
	}

	// 4. Добавляем почтовые ящики в домен
	_, err = webSiteMain.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "RatusCRM", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}

	// 5. Создаем несколько API-ключей
	_, err = mAcc.ApiKeyCreate(models.ApiKey{Name:"Для сайта"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}
	_, err = mAcc.ApiKeyCreate(models.ApiKey{Name:"Postman test"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}
	_, err = mAcc.ApiKeyCreate(models.ApiKey{Name:"Bitrix24 export"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	dvc, err := models.GetUserVerificationTypeByCode(models.VerificationMethodEmailAndPhone)
	if err != nil || dvc == nil {
		log.Fatal("Не удалось получить верификацию...")
		return
	}

	// ######### Ratus Media Account ############

	// 2. создаем из-под mex388 RatusMedia
	ratusMediaAcc, err := owner.CreateAccount( models.Account{
		Name:                                "Ratus Media",
		Website:                             "ratus.media",
		Type:                                "service",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            &dvc.Id,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	// Создаем API ключ
	_, err = ratusMediaAcc.ApiKeyCreate(models.ApiKey{Name:"Интеграция сайта на Rust с CRM"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 3. добавляем mex388 как админа
	_, err = ratusMediaAcc.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}
	// 3. создаем MarkPlatov
	markPlatov, err := ratusMediaAcc.CreateUser(
		models.User{
			Username:	utils.STRp("MarkPlatov"),
			Email:		utils.STRp("markPlatov@rus-marketing.ru"),
			PhoneRegion: utils.STRp("RU"),
			Phone: 		utils.STRp("79777201164"),
			Password:	utils.STRp("daUw#92QapZ"),
			Name:		utils.STRp("Михаил"),
			Surname:	utils.STRp("Коротченко"),
			Patronymic:	utils.STRp("-"),
			EmailVerifiedAt:&timeNow,
		},
		*roleManagerMain,
	)
	if err != nil {
		log.Fatal("Не удалось создать markPlatov'a: ", err)
	}


	// 4. Создаем домен для ratus.media
	_webSiteRatusMedia, err := ratusMediaAcc.CreateEntity(&models.WebSite{
		Hostname: "ratus.media",
		DKIMPublicRSAKey: `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDYq5m0HLzmuGrIvghDA3uHR8rF
JTmhGutraXmqrHT3dLx4en15H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0
pjg2ee1PxllVduwzzwzIUfo3k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIp
FeTiqGsEKHqAbsiq0QIDAQAB
-----END PUBLIC KEY-----
`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDYq5m0HLzmuGrIvghDA3uHR8rFJTmhGutraXmqrHT3dLx4en15
H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0pjg2ee1PxllVduwzzwzIUfo3
k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIpFeTiqGsEKHqAbsiq0QIDAQAB
AoGBAIwqFnipzpoC2zGZnYacjiDplIcMmcdavsjxpr9+aPxcDEB0HQ9qPutt5TZy
QSis8Vzfp2oLXhDPvicLgyvJSmy1lsevb8KtXR7sUR6OOsSk9UDxdkziSpCxn3b1
1cTHeS+iH5vT7oNAMl4bczA/Mf+5HbYYZOVBkAK0XZgfPlgBAkEA+HgODNGFcakr
npk9nIYO3fQ47j3yE0WH2a7LXcMaqcqSc7+rOPI1xNqmNY1wInYKrE3nVIbt94Hp
zEYvbo1JYQJBAN88z23O09SOK0vDNztqHLs5CONXaulHjhsumxRz/sFnJEA+39IB
H7cLKY0bNj32naSPvlMvpOXrUs3RvoGtp3ECQA3+XPQ13KkhGPbOtJybJD014vrv
/DE/qj4FTSjTsy263fKs6czEP2j903ySoa0fLsyzeoLVMpPM+kjp7wPv6yECQQDC
1SH8ply+0G39K53S3yPAS2D0eeE3RFpEYwK6cRpQ1aIP6FATyyNQ8A8Ntkg7ADhJ
3aKwm46BlEaql3p7V4IhAkBTq1wkY0mGjIfgfN5sTg8B4SbL/dhfLrl9xeRbc5qA
AJnnVkwI9ntl6+d3uML4VA7hUloxsufH7fZ3lmaR+453
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для ratus media: ", err)
	}
	webSiteRatusMedia, ok := _webSiteRatusMedia.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для ratus media: ", err)
	}

	// 5. Добавляем почтовые ящики в домен ratus.media
	_, err = webSiteRatusMedia.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "Ratus Media", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Ratus Media: ", err)
	}
	

	// ######### Test Account ############

	// 2. создаем из-под mex388 TestAccount
	testAcc, err := mex388.CreateAccount( models.Account{
		Name:                                "Test account",
		Website:                             "example.com",
		Type:                                "store",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            &dvc.Id,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	_, err = testAcc.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с сайтом"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 3. добавляем меня как админа
	_, err = testAcc.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr: ", err)
		return
	}
	// 3. добавляем MarkPlatov
	_, err = testAcc.AppendUser(*markPlatov, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in test account")
		return
	}


	// 4. Создаем домен для example.com
	_webSiteTest, err := testAcc.CreateEntity(&models.WebSite{
		Hostname: "example.com",
		DKIMPublicRSAKey: `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDYq5m0HLzmuGrIvghDA3uHR8rF
JTmhGutraXmqrHT3dLx4en15H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0
pjg2ee1PxllVduwzzwzIUfo3k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIp
FeTiqGsEKHqAbsiq0QIDAQAB
-----END PUBLIC KEY-----
`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDYq5m0HLzmuGrIvghDA3uHR8rFJTmhGutraXmqrHT3dLx4en15
H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0pjg2ee1PxllVduwzzwzIUfo3
k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIpFeTiqGsEKHqAbsiq0QIDAQAB
AoGBAIwqFnipzpoC2zGZnYacjiDplIcMmcdavsjxpr9+aPxcDEB0HQ9qPutt5TZy
QSis8Vzfp2oLXhDPvicLgyvJSmy1lsevb8KtXR7sUR6OOsSk9UDxdkziSpCxn3b1
1cTHeS+iH5vT7oNAMl4bczA/Mf+5HbYYZOVBkAK0XZgfPlgBAkEA+HgODNGFcakr
npk9nIYO3fQ47j3yE0WH2a7LXcMaqcqSc7+rOPI1xNqmNY1wInYKrE3nVIbt94Hp
zEYvbo1JYQJBAN88z23O09SOK0vDNztqHLs5CONXaulHjhsumxRz/sFnJEA+39IB
H7cLKY0bNj32naSPvlMvpOXrUs3RvoGtp3ECQA3+XPQ13KkhGPbOtJybJD014vrv
/DE/qj4FTSjTsy263fKs6czEP2j903ySoa0fLsyzeoLVMpPM+kjp7wPv6yECQQDC
1SH8ply+0G39K53S3yPAS2D0eeE3RFpEYwK6cRpQ1aIP6FATyyNQ8A8Ntkg7ADhJ
3aKwm46BlEaql3p7V4IhAkBTq1wkY0mGjIfgfN5sTg8B4SbL/dhfLrl9xeRbc5qA
AJnnVkwI9ntl6+d3uML4VA7hUloxsufH7fZ3lmaR+453
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}
	webSiteTest, ok := _webSiteTest.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 5. Добавляем почтовые ящики в домен 357gr
	_, err = webSiteTest.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "Example", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}

	// ######### Demonstration Account ############

	// 2. создаем из-под mex388 TestAccount
	demoAcc, err := mex388.CreateAccount( models.Account{
		Name:                                "Demo account",
		Website:                             "example.com",
		Type:                                "store",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            &dvc.Id,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	_, err = demoAcc.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с сайтом"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 3. добавляем меня как админа
	_, err = demoAcc.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}
	// 3. добавляем MarkPlatov
	_, err = demoAcc.AppendUser(*markPlatov, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}


	// 4. Создаем домен для example.com
	_webSiteDemo, err := demoAcc.CreateEntity(&models.WebSite{
		Hostname: "demo.com",
		DKIMPublicRSAKey: ``,
		DKIMPrivateRSAKey: ``,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}
	webSiteDemo, ok := _webSiteDemo.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 5. Добавляем почтовые ящики в домен 357gr
	_, err = webSiteDemo.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "Demo account", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}


	// ######### 357 Грамм ############

	// 1. Создаем Василия (^_^)
	vpopov, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("antiglot"),
			// Email:"vp@357gr.ru",
			Email:		utils.STRp("mail-test@ratus-dev.ru"),
			PhoneRegion: utils.STRp("RU"),
			Phone: 		utils.STRp("89055294696"),
			Password:	utils.STRp("qwerty109#QW"),
			Name:		utils.STRp("Василий"),
			Surname:	utils.STRp("Попов"),
			Patronymic:	utils.STRp("Николаевич"),
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать vpopov'a: ", err)
	}
	

	
	// 2. создаем из-под Василия 357gr
	acc357, err := vpopov.CreateAccount( models.Account{
		Name:                                "357 грамм",
		Website:                             "https://357gr.ru/",
		Type:                                "store",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"email","phone"})), 
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"email","phone","name"})),
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            &dvc.Id,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil || acc357 == nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	_, err = acc357.ApiKeyCreate(models.ApiKey{Name:"Для сайта"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 3. добавляем меня как админа
	_, err = acc357.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	_, err = acc357.AppendUser(*markPlatov, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя markPlatov in 357gr")
		return
	}

	// 3.2. добавляем кучу других клиентов
	if false {
		var clients []models.User

		for i:=1; i < 200 ;i++ {
			clients = append(clients, models.User{
				Name: utils.STRp(fmt.Sprintf("Name #%d", i)),
				Email: utils.STRp(fmt.Sprintf("email%d@mail.ru", i)),
				Phone: utils.STRp(fmt.Sprintf("+7925195221%d", i)),
				Password: utils.STRp("asdfg109#QW"),
			})
		}
		for i,_ := range clients {
			_, err := acc357.CreateUser(clients[i], *roleClientMain)
			if err != nil {
				log.Printf("Не удалось добавить клиента id: %v", i)
				return
			}
		}
	}


	// 4. Создаем домен для 357gr
	_webSite357, err := acc357.CreateEntity(&models.WebSite{
		Hostname: "357gr.ru",
		DKIMPublicRSAKey: `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDYq5m0HLzmuGrIvghDA3uHR8rF
JTmhGutraXmqrHT3dLx4en15H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0
pjg2ee1PxllVduwzzwzIUfo3k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIp
FeTiqGsEKHqAbsiq0QIDAQAB
-----END PUBLIC KEY-----
`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDYq5m0HLzmuGrIvghDA3uHR8rFJTmhGutraXmqrHT3dLx4en15
H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0pjg2ee1PxllVduwzzwzIUfo3
k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIpFeTiqGsEKHqAbsiq0QIDAQAB
AoGBAIwqFnipzpoC2zGZnYacjiDplIcMmcdavsjxpr9+aPxcDEB0HQ9qPutt5TZy
QSis8Vzfp2oLXhDPvicLgyvJSmy1lsevb8KtXR7sUR6OOsSk9UDxdkziSpCxn3b1
1cTHeS+iH5vT7oNAMl4bczA/Mf+5HbYYZOVBkAK0XZgfPlgBAkEA+HgODNGFcakr
npk9nIYO3fQ47j3yE0WH2a7LXcMaqcqSc7+rOPI1xNqmNY1wInYKrE3nVIbt94Hp
zEYvbo1JYQJBAN88z23O09SOK0vDNztqHLs5CONXaulHjhsumxRz/sFnJEA+39IB
H7cLKY0bNj32naSPvlMvpOXrUs3RvoGtp3ECQA3+XPQ13KkhGPbOtJybJD014vrv
/DE/qj4FTSjTsy263fKs6czEP2j903ySoa0fLsyzeoLVMpPM+kjp7wPv6yECQQDC
1SH8ply+0G39K53S3yPAS2D0eeE3RFpEYwK6cRpQ1aIP6FATyyNQ8A8Ntkg7ADhJ
3aKwm46BlEaql3p7V4IhAkBTq1wkY0mGjIfgfN5sTg8B4SbL/dhfLrl9xeRbc5qA
AJnnVkwI9ntl6+d3uML4VA7hUloxsufH7fZ3lmaR+453
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}
	webSite357, ok := _webSite357.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 5. Добавляем почтовые ящики в домен 357gr
	_, err = webSite357.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "357 Грамм", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}

	// 6. Api key
	_, err = acc357.ApiKeyCreate(models.ApiKey{Name:"Для Postman"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", acc357.Name, err)
	}
	

	//////// SyndicAd

	// 1. Создаем Станислава
	stas, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("ikomastas"),
			Email:		utils.STRp("sa-tolstov@yandex.ru"),
			// Email:"info@rus-marketing.ru",
			PhoneRegion: utils.STRp("RU"),
			// Phone: nil,
			Password:	utils.STRp("qwerty123#Q"),
			Name:		utils.STRp("Станислав"),
			Surname:	utils.STRp("Толстов"),
			// Patronymic: nil,
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать stas'a: ", err)
	}
	
	// 1. Создаем синдикат из-под Станислава
	accSyndicAd, err := stas.CreateAccount(models.Account{
		Name:                                "SyndicAd",
		Website:                             "syndicad.com",
		Type:                                "internet-service",
		ApiEnabled:                          true,
		UiApiEnabled:                        false,
		VisibleToClients:                    false,
	})
	if err != nil || accSyndicAd == nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	// 2. добавляем меня как админа
	_, err = accSyndicAd.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2 Добавляем Mex388
	_, err = accSyndicAd.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in 357gr")
		return
	}

	_, err = accSyndicAd.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с системой"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 2. Создаем домен для синдиката
	_webSiteSynd, err := accSyndicAd.CreateEntity(&models.WebSite{
		Hostname: "syndicad.com",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyxCGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+ABtf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIdAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyx
CGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+A
Btf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIdAQAB
AoGAIR9YdelFBhrtM2WEVb/bnX+7vJ2mm+OLxTMyFuuvuvsiw6TBnHgXncYZBk/D
Zm9uhfCKU1loRIGd6gxY+dx+hVCFHh4tyQ+xvb+siTsDO3VXhHCq+XZpstDanrS0
kEjDPx95QYgJ3taG55Agu2Ql/cgevyFevOhXUPrZ6lStdcUCQQDxpSPUywPgOas5
CFMWB5k5+DRAz9CygH5L7i53RnitwPL3jHvwOHs5JD25lD9IfKVyGuJtYeUTPenp
FlIxzv+TAkEA0HAuDHrCItg1x/UDO9N+IafTFN5+31Me9POiOGkghXfbWJCfxaBW
wJWLTPI7p+PT07/sRusQpGRiGi0RagZbowJAVqXsr0UM4r5LE2xUvrWC0DKcKhFa
uGcy4m9J4iM26rchaHrLhlv6c4b3SzBJcOihOsVBJA/SYI/27EnAt3OOWQJAXhjm
kPeyQKy+ysBPb2iw3ly3LAqt1//cT9TU/QZoihhry3WuyzbxMwvP0TLhv49Yh5Vz
AykHYE95AjwqSmUIZQJAaRJMuw5gVSjQaLz/qoiMVEQO7vmazsiB9/YKTPp18I+4
pBRlD1bMcxJEBYvc/tLA1LqyGGhd1mabVQ7iYPq45w==
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для Синдиката: ", err)
	}
	webSiteSynd, ok := _webSiteSynd.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 3. Добавляем почтовые ящики
	_, err = webSiteSynd.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "SyndicAd", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Синдиката: ", err)
	}


	// Brouser.com
	// 1. Создаем аккаунт из-под Станислава
	brouser, err := stas.CreateAccount(models.Account{
		Name:                                "BroUser",
		Website:                             "www.brouser.com",
		Type:                                "internet-service",
		ApiEnabled:                          true,
		UiApiEnabled:                        false,
		VisibleToClients:                    false,
	})
	if err != nil || accSyndicAd == nil {
		log.Fatal("Не удалось создать аккаунт Brouser")
		return
	}

	// 2. добавляем меня как админа
	_, err = brouser.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2. Добавляем mex388
	_, err = brouser.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in brouser")
		return
	}

	_, err = brouser.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с главной системой"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 2. Создаем домен для BroUser
	_webSiteBro, err := brouser.CreateEntity(&models.WebSite{
		Name: "Сайт-визитка",
		Hostname: "brouser.com",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDXVD+X2Jja2cckCCYTg9UURSPb9Qx9c8idTcFqmpJVxKjKPvryklToXJATsKVzvOwbmrt9FVn2VnB9VQgmUyifF1RYqt0OgLRn+LB0o8x2WbzBKXHcumqZvEA+ZEFq5CzBGpW+4WWyPGIrKXst5A77EHhNgVskzrvcoaCrOT9MJQIDAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDXVD+X2Jja2cckCCYTg9UURSPb9Qx9c8idTcFqmpJVxKjKPvry
klToXJATsKVzvOwbmrt9FVn2VnB9VQgmUyifF1RYqt0OgLRn+LB0o8x2WbzBKXHc
umqZvEA+ZEFq5CzBGpW+4WWyPGIrKXst5A77EHhNgVskzrvcoaCrOT9MJQIDAQAB
AoGAIIBS6PSEfeQJLuMb/C4V521YMEcYj4b+bN/jpdeW5uM8JurCrgJwVnJCPPaY
wpNtf+0nB4ZFge0iJYjEJiS/KJ1YT50fEKqMPx/GVm9UULDvUsWsLFONGr1+hP2+
XaU4ik/+ym3SQ9Ir+VAq6qyBeOwZlpRBySezCGJ+UpluIrECQQDrItv+oYR8QzzA
4G3ZaP3PclwPOVWIJyvxM6E0zgPRR4JQO80MVEj0IcaZUl/7EsgqOkRorye0Tba1
eJmrZbu7AkEA6m94LzePJslSqGcAiU7eyJuqBQbkKaJmK0nVFAkAf4hm1om1DSgk
iPShiBQ79vTP5T7l2j20miqm+E00CDBpnwJAT7jF9hM1JBx34L03AVuDkm4noFHE
GiGN2H20zn569N3V5PYhk2iQQ5WgDCPNvwajLw4KW6PnRk6DAAwfrekUOQJAcG0W
oOYvE3W22yXSXwbg1im4poKAhurnvljBA8OxZne+gaI2nmGi678NfBngC/WpgZHh
XwD6jHhp7GfxzP+SlwJBALL6Mmgkk9i5m5k2hocMR8U8+CMM3yHtHZRec7AdRv0c
3/m5b5CLpflEX58hz9NeWHfoNJ2QXj3bkYDzZ1vnzJw=
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для Brouser: ", err)
	}
	webSiteBro, ok := _webSiteBro.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 3. Добавляем почтовые ящики
	_, err = webSiteBro.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "Brouser", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Brouser: ", err)
	}
	

	// AiroClimate

	// 1. Создаем аккаунт из-под Станислава
	// 1. Создаем Коротаева
	korotaev, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("korotaev"),
			// Email:"sa-tolstov@yandex.ru",
			Email:	utils.STRp("korotaev@vtvent.ru"),
			PhoneRegion: utils.STRp("RU"),
			Phone: nil,
			Password:	utils.STRp("jv92sA#qpx2S"),
			Name:		utils.STRp("Максим"),
			Surname:	utils.STRp("Коротаев"),
			Patronymic:	utils.STRp("Валерьевич"),
			Subscribed: false,
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать korotaev'a: ", err)
	}

	ivlev, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("ivlev"),
			// Email:"sa-tolstov@yandex.ru",
			Email:	utils.STRp("ivlev@vtvent.ru"),
			PhoneRegion: 	utils.STRp("RU"),
			Phone: nil,
			Password:utils.STRp("uisNKs82#Mr2A"),
			Name:	utils.STRp("Владислав"),
			Surname:	utils.STRp("Ивлев"),
			Patronymic:	nil,
			Subscribed: false,
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать korotaev'a: ", err)
	}

	airoClimat, err := korotaev.CreateAccount(models.Account{
		Name:                                "AIRO Climate",
		Website:                             "https://airoclimate.ru",
		Type:                                "shop",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		VisibleToClients:                    false,
	})
	if err != nil || airoClimat == nil {
		log.Fatal("Не удалось создать аккаунт AIRO")
		return
	}

	// 2. добавляем меня как админа
	_, err = airoClimat.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in airo")
		return
	}

	// 2.2. Добавляем mex388 как админа
	_, err = airoClimat.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in airo")
		return
	}

	_, err = airoClimat.AppendUser(*ivlev, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in airo")
		return
	}

	_, err = airoClimat.ApiKeyCreate(models.ApiKey{Name:"Для сайта"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
		return
	}

	///////////////////////////////////////

	///////////////////////////////////////
	// 4. !!! Создаем магазин
	airoShopE, err := airoClimat.CreateEntity(
		&models.WebSite{
			Name: "Сайт по продаже бактерицидных рециркуляторов", Address: utils.STRp("г. Москва, р-н Текстильщики"),
			Email: utils.STRp("info@airoclimate.ru"), Phone: utils.STRp("+7 (4832) 77-03-73"),
			Hostname: "airoclimate.ru",
			URL: "https://airoclimate.ru",
			DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDFS3EibqbaeWQvH8+2CRw5ijKV
1UOoR1Uzi/wNjOIlAxQJfBnocmLtmLVcpTW/ZmjES6iV2e3WkOICzgxLT44UlXFj
Fox0sQ+qWzKAFjz5SWWZ2vTFrMicGweps48TQ+L9ZX6yRIxuJQGN0uGd0MH49Wzc
+kOepVTv5oxkqAUjFQIDAQAB`,
			DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXgIBAAKBgQDFS3EibqbaeWQvH8+2CRw5ijKV1UOoR1Uzi/wNjOIlAxQJfBno
cmLtmLVcpTW/ZmjES6iV2e3WkOICzgxLT44UlXFjFox0sQ+qWzKAFjz5SWWZ2vTF
rMicGweps48TQ+L9ZX6yRIxuJQGN0uGd0MH49Wzc+kOepVTv5oxkqAUjFQIDAQAB
AoGBALeXWWLaJugcmA6GAqp5Vctxf1sQRlI8dtttwxH07KfWcnnVAuLcNpS0Sug4
UIiYSpuHcAxp7DmDPt2vUZ9vG10FWDptoc731TrRDbp83nEJpvS1Me95KHNyKKk5
k531CX7lhUqmLjqLvSCXDqbhP/QdRp8AUvB3b0BhcqW8c1ChAkEA+vD6UsOBGMGm
12CId0uB7Od3x7j1kk3HiH8olZtBAxIDy5NelZdu/ViigHCN4wzYcnRQm1Y5p/Bb
vLv7cWhlWQJBAMlFnLzdtx6RrcU6L9OHXlEuBp8GR8S2LPhnYv0dM1T2sBpLTbmO
jk16kPXeGWDLOPIdrSFb71AH8p0ymH4B6B0CQQCMd4nX/EHuZrAKzal2BZlD0Em3
TayA6fLwUCWaoR5iJppjQSnn2K2zOQM1nEuANfePEdbxLPH3NM9VNXDJiaN5AkEA
s+Wdh44gi5koGV29u7KF0cdysZa6dQ9juI8oAhakd++aTZY7HXxWotfHU4s1Ybei
6X0u7t8uUnkYF/tOI2pu3QJAbpjBRWfkYBM7Nxwd2UVCVQRR0KA+bSbrxUwyQV+V
TsAWKRB/H4nLPV8gbADJAwlz75F035Z/E7SN4RdruEX6TA==
-----END RSA PRIVATE KEY-----
`,
			DKIMSelector: "dk1",
			/*PaymentMethods: []models.PaymentMethod{
				{AccountId: 1, Id: 1},{AccountId: 1, Id: 3},{AccountId: 1, Id: 2},
			},*/
		})
	if err != nil {
		log.Fatal("Не удалось создать WebSite для airoClimat: ", err)
		return
	}
	webSiteAiro, ok := airoShopE.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось преобразовать WebSite для airoClimat: ", err)
		return
	}


	// 3. Добавляем почтовые ящики
	_, err = webSiteAiro.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "AIRO Climate", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для AiroClimate: ", err)
	}

	mPage, err := airoClimat.CreateEntity(&models.WebPage{
		AccountId: airoClimat.Id, WebSiteId: &webSiteAiro.Id, Label: utils.STRp("Главная"), Code: utils.STRp("root"),Path: utils.STRp("/"),
		MetaTitle: utils.STRp("Главная :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-home"), RouteName:  utils.STRp("info.index"),
	})
	if err != nil {
		log.Fatal("Не удалось создать mPage для airoClimat webSite: ", err)
		return
	}
	webPageRoot := mPage.(*models.WebPage)

	catE, err := webPageRoot.CreateChild(models.WebPage {
		AccountId: airoClimat.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Весь каталог"), Path:  utils.STRp("catalog"),
		MetaTitle: utils.STRp("Каталог :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-th-large"), RouteName:  utils.STRp("catalog"),
	})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}

	webPageCatalogRoot := catE.(*models.WebPage)

	_webPageCatalog1, err := webPageCatalogRoot.CreateChild(models.WebPage{
		AccountId: airoClimat.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Бактерицидные рециркуляторы"), Path:  utils.STRp("bactericidal-recirculators"),
		MetaTitle: utils.STRp("Бактерицидные рециркуляторы :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.recirculators"),
	})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	webPageCatalog1 := _webPageCatalog1.(*models.WebPage)
	_webPageCatalog2, err := webPageCatalogRoot.CreateChild(models.WebPage{
		AccountId: airoClimat.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Бактерицидные камеры"), Path:  utils.STRp("bactericidal-chambers"),
		MetaTitle: utils.STRp("Бактерицидные камеры :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.chambers"),
	})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	webPageCatalog2 := _webPageCatalog2.(*models.WebPage)

	//////////////

	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Статьи"), Path:  utils.STRp("articles"),
			MetaTitle: utils.STRp("Статьи :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-books"), RouteName:  utils.STRp("articles"), Priority: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	deliveryGrE, err := webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Доставка товара"), Path:  utils.STRp("delivery"),
			MetaTitle: utils.STRp("Доставка товара :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-shipping-fast"), RouteName:  utils.STRp("delivery"), Priority: 1,
		})
	if err != nil {
		log.Fatal(err)
	}
	deliveryGroupRoute := deliveryGrE.(*models.WebPage)
	_, err = deliveryGroupRoute.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Способы оплаты"), Path:  utils.STRp("payment"),
			MetaTitle: utils.STRp("Способы оплаты :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-hand-holding-usd"), RouteName:  utils.STRp("delivery.payment"), Priority: 2,
		})
	_, err = deliveryGroupRoute.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Возврат товара"), Path:  utils.STRp("moneyback"),
			MetaTitle: utils.STRp("Возврат товара :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-exchange-alt"), RouteName:  utils.STRp("delivery.moneyback"), Priority: 3,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("info"), Label:  utils.STRp("О компании"), Path:  utils.STRp("about"),
			MetaTitle: utils.STRp("О компании :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-home-heart"), RouteName:  utils.STRp("info.about"),      Priority: 5,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Политика конфиденциальности"), Path:  utils.STRp("privacy-policy"),
			MetaTitle: utils.STRp("Политика конфиденциальности :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-home-heart"), RouteName:  utils.STRp("info.privacy-policy"),      Priority: 6,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Контакты"), Path:  utils.STRp("contacts"),
			MetaTitle: utils.STRp("Контакты :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
			IconName:  utils.STRp("far fa-address-book"), RouteName:  utils.STRp("info.contacts"),  Priority: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}

	////////
	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: airoClimat.Id, Code:  utils.STRp("cart"), Label:  utils.STRp("Корзина"), Path:  utils.STRp("cart"),
			MetaTitle: utils.STRp("Корзина :: AiroClimate"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
				IconName:  utils.STRp("far fa-cart-arrow-down"), RouteName:  utils.STRp("cart"), Priority: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}

	// 5* Создаем категории товаров

	_CategoryRoot, err := airoClimat.CreateEntity(&models.ProductCategory{
		Code:  utils.STRp("catalog"), Label: utils.STRp("Каталог"),LabelPlural: utils.STRp("Каталог"),
	})
	CategoryRoot := _CategoryRoot.(*models.ProductCategory)

	_catGr1, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("recirculators"), Label: utils.STRp("Бактерицидный рециркулятор"),LabelPlural: utils.STRp("Бактерицидные рециркуляторы"),
	})
	catGr1 := _catGr1.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	_catGr2, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("chambers"), Label: utils.STRp("Бактерицидная камера"),LabelPlural: utils.STRp("Бактерицидные камеры"),
	})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	catGr2 := _catGr2.(*models.ProductCategory)

	// А можно добавить категорию 1 и категорию 2
	if err := webPageCatalogRoot.AppendProductCategory(CategoryRoot); err != nil {
		log.Fatal(err)
	}
	if err := webPageCatalog1.AppendProductCategory(catGr1); err != nil {
		log.Fatal(err)
	}
	if err := webPageCatalog2.AppendProductCategory(catGr2); err != nil {
		log.Fatal(err)
	}

	// 5.5 Создаем продуктовые типы
	productTypes := []models.ProductType{
		{Name: utils.STRp("Бактерицидный облучатель"), Code: utils.STRp("bactericidal_irradiator")},
		{Name: utils.STRp("Бактерицидный рециркулятор"), Code: utils.STRp("bactericidal_recirculator")},
	}

	for i := range productTypes {
		_,_ = airoClimat.CreateEntity(&productTypes[i])
	}

	// 6. Создаем карточки товара
	cards := []models.ProductCard{
		{Id: 0, RouteName: utils.STRp("catalog.recirculators.card"), Path: utils.STRp("airo-dez-adjustable-black"),	Label: utils.STRp("Рециркулятор AIRO-DEZ черный с регулировкой"),Breadcrumb: utils.STRp("Рециркулятор AIRO-DEZ черный с регулировкой"), 	MetaTitle:  utils.STRp("Рециркулятор AIRO-DEZ черный с регулировкой")},
		{Id: 0, RouteName: utils.STRp("catalog.recirculators.card"),	Path: utils.STRp("airo-dez-black"), 				Label: utils.STRp("Рециркулятор AIRO-DEZ черный"), 				Breadcrumb: utils.STRp("Рециркулятор AIRO-DEZ черный"), 					MetaTitle: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ черный")},
		{Id: 0, RouteName: utils.STRp("catalog.recirculators.card"),	Path: utils.STRp("airo-dez-adjustable-white"), 	Label: utils.STRp("Рециркулятор AIRO-DEZ белый с регулировкой"), 	Breadcrumb: utils.STRp("Рециркулятор AIRO-DEZ белый с регулировкой"),	MetaTitle: utils.STRp("Рециркулятор AIRO-DEZ белый с регулировкой")},
		{Id: 0, RouteName: utils.STRp("catalog.recirculators.card"),	Path: utils.STRp("airo-dez-white"), 				Label: utils.STRp("Рециркулятор AIRO-DEZ белый"), 				Breadcrumb: utils.STRp("Рециркулятор AIRO-DEZ белый"),					MetaTitle: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ белый")},
		{Id: 0, RouteName: utils.STRp("catalog.recirculators.card"),	Path: utils.STRp("airo-dez-compact"), 			Label: utils.STRp("Мобильный аиродезинфектор AIRO-DEZ COMPACT"), 	Breadcrumb: utils.STRp("Мобильный аиродезинфектор AIRO-DEZ COMPACT"),	MetaTitle: utils.STRp("Мобильный аиродезинфектор AIRO-DEZ COMPACT")},

		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-dezpuf"),					Label: utils.STRp("Бактерицидная камера пуф AIRO-DEZPUF"), 				Breadcrumb: utils.STRp("Бактерицидная камера пуф AIRO-DEZPUF"),			MetaTitle: utils.STRp("Бактерицидная камера пуф AIRO-DEZPUF")},
		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-dezpuf-wenge"), 			Label: utils.STRp("Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"), Breadcrumb: utils.STRp("Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"),MetaTitle: utils.STRp("Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге")},
		
		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-dezbox"), 				Label: utils.STRp("Бактерицидная камера AIRO-DEZBOX"), 					Breadcrumb: utils.STRp("Бактерицидная камера AIRO-DEZBOX"),				MetaTitle: utils.STRp("Бактерицидная камера AIRO-DEZBOX")},
		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-dezbox-white"), 			Label: utils.STRp("Бактерицидная камера AIRO-DEZBOX белая"),				Breadcrumb: utils.STRp("Бактерицидная камера AIRO-DEZBOX белая"), 		MetaTitle: utils.STRp("Бактерицидная камера AIRO-DEZBOX белая")},
		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-deztumb"), 				Label: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB"), 		Breadcrumb: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB"),	MetaTitle: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB")},
		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-deztumb-big"), 			Label: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB big"), 	Breadcrumb: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB big"),MetaTitle: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB big")},

		{Id: 0, RouteName: utils.STRp("catalog.chambers.card"),		Path: utils.STRp("airo-deztumb-pine"), 			Label: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина"),	Breadcrumb: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина"), MetaTitle: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина")},
	}

	// 7. Создаем список товаров
	products := []models.Product{
		{
			Model: utils.STRp("AIRO-DEZ с регулировкой черный"), Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности черный"), ShortLabel: utils.STRp("Рециркулятор AIRO-DEZ черный"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1.0, ProductTypeId: utils.UINTp(1),
			RetailPrice: utils.FL64p(19500.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				"bodyMaterial":"металл",
				"filterType":"угольно-фотокаталитический",
				"performance":150, // m3/час
				"rangeUVRadiation":"250-260Hm",
				"powerLampRecirculator":10.8, // Вт/m2
				"powerConsumption":60, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"overallDimensions":"690х250х250мм", //Габаритные размеры(ВхШхГ)
				"noiseLevel":35, //дБ
				"grossWeight": 5.5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZ черный"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ черный"), ShortLabel: utils.STRp("Рециркулятор AIRO-DEZ черный"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1.0,  ProductTypeId: utils.UINTp(1),
			RetailPrice: utils.FL64p(17500.00), RetailDiscount: utils.FL64p(1000), RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				"bodyMaterial":"металл",
				"filterType":"угольно-фотокаталитический",
				"performance":150, // m3/час
				"rangeUVRadiation":"250-260Hm",
				"powerLampRecirculator":10.8, // Вт/m2
				"powerConsumption":60, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":35, //дБ
				"overallDimensions":"690х250х250мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5.5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZ с регулировкой белый"), Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности белый"),  ShortLabel: utils.STRp("Рециркулятор AIRO-DEZ белый"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(1),
			RetailPrice: utils.FL64p(19500.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"белый",
				"bodyMaterial":"металл",
				"filterType":"угольно-фотокаталитический",
				"performance":150, // m3/час
				"rangeUVRadiation":"250-260Hm",
				"powerLampRecirculator":10.8, // Вт/m2
				"powerConsumption":60, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":35, //дБ
				"overallDimensions":"690х250х250мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5.5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZ белый"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Рециркулятор воздуха бактерицидный AIRO-DEZ"),  ShortLabel: utils.STRp("Рециркулятор AIRO-DEZ белый"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(1),
			RetailPrice: utils.FL64p(17500.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"белый",
				"bodyMaterial":"металл",
				"filterType":"угольно-фотокаталитический",
				"performance":150, // m3/час
				"rangeUVRadiation":"250-260Hm",
				"powerLampRecirculator":10.8, // Вт/m2
				"powerConsumption":60, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":35, //дБ
				"overallDimensions":"690х250х250мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5.5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZ COMPACT"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Мобильный аиродезинфектор AIRO-DEZ COMPACT"),  ShortLabel: utils.STRp("Аиродезинфектор AIRO-DEZ COMPACT"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(1),
			RetailPrice: utils.FL64p(39000.00), RetailDiscount: utils.FL64p(3000), RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				"bodyMaterial":"металл",
				"filterType":"угольно-фотокаталитический",
				"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				"powerLampRecirculator":19, // Вт/m2
				"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":135, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":45, //дБ
				"overallDimensions":"300х610х150мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 6.8, // Брутто, кг
			})),
		},
		
		{
			Model: utils.STRp("AIRO-DEZPUF"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Бактерицидная камера пуф AIRO-DEZPUF"),  ShortLabel: utils.STRp("Камера пуф AIRO-DEZPUF"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(11000.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":25, //дБ
				"overallDimensions":"480х500х320мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZPUF венге"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"), ShortLabel: utils.STRp("Камера AIRO-DEZBOX"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(12000.00), RetailDiscount: utils.FL64p(1000), RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"венге",
				"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":25, //дБ
				"overallDimensions":"500х500х320мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},

		{
			Model: utils.STRp("AIRO-DEZBOX"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Бактерицидная камера AIRO-DEZBOX"),  ShortLabel: utils.STRp("Камера AIRO-DEZBOX"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(7800.00), RetailDiscount: utils.FL64p(800), RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":25, //дБ
				"overallDimensions":"400х500х320мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZBOX белая"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Бактерицидная камера AIRO-DEZBOX белая"),  ShortLabel: utils.STRp("Камера AIRO-DEZBOX белая"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(7800.00), RetailDiscount: utils.FL64p(800), RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"белый",
				"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":25, //дБ
				"overallDimensions":"400х500х320мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIRO-DEZTUMB"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB"),  ShortLabel: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(11500.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"черный",
				//"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт мощность устр-ва
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":5, //дБ
				"overallDimensions":"560х450х400мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
		{
			Model: utils.STRp("AIROTUMB big"),Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Тумба облучатель бактерицидный AIRO-DEZTUMB big"),  ShortLabel: utils.STRp("Облучатель AIROTUMB big"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(11500.00), RetailDiscount: utils.FL64p(1000),
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"белый",
				//"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт мощность устр-ва
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":5, //дБ
				"overallDimensions":"670х450х400мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
		
		{
			Model: utils.STRp("AIRO-DEZTUMB касцина"), Brand: utils.STRp("AIRO-Climate"),
			Label: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина"),  ShortLabel: utils.STRp("Бактерицидная тумба AIRO-DEZTUMB"),
			PaymentSubjectId: 1, MeasurementUnitId: utils.UINTp(1), VatCodeId: 1, ProductTypeId: utils.UINTp(2),
			RetailPrice: utils.FL64p(11500.00), RetailDiscount: utils.FL64p(1000),RetailSale: true,
			ShortDescription: "",Description: "",
			Attributes: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
				"color":"светлая сосна",
				"bodyMaterial":"металл",
				//"filterType":"угольно-фотокаталитический",
				//"performance":220, // m3/час
				"rangeUVRadiation":"250-260Hm",
				//"powerLampRecirculator":19, // Вт/m2
				//"powerLampIrradiator":10.8, // Вт/m2
				"powerConsumption":10, // Вт мощность устр-ва
				"lifeTimeDevice":100000, // часов
				"lifeTimeLamp":9000, // часов
				"baseTypeLamp":"G13", //Тип цоколя лампы
				"degreeProtection":"IP20",
				"supplyVoltage":"175-265В",
				"temperatureMode":"+2...+50C",
				"noiseLevel":25, //дБ
				"overallDimensions":"460х500х320мм", //Габаритные размеры(ВхШхГ)
				"grossWeight": 5, // Брутто, кг
			})),
		},
	}


	var productCategory models.ProductCategory
	// 7. Добавляем продукты в категории с созданием карточки товара
	for i := range products {
		// var groupId uint

		if i <= 4 {
			// groupId = catGr1.GetId()
			productCategory = *catGr1
		} else {
			// groupId = catGr2.GetId()
			productCategory = *catGr2
		}
		
		// создаем товар, карточку товара и добавляем их в группу
		_, err = webSiteAiro.CreateProductWithProductCard(products[i], cards[i], productCategory)
		if err != nil {
			log.Fatal("Не удалось создать Product для airoClimat: ", err)
		}
	}


	return
}

func UploadTestDataPart_II() {

	// 1. Получаем AiroClimate аккаунт
	accountAiro, err := models.GetAccount(8)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	// 2. Получаем магазин
	var webSite models.WebSite
	err = accountAiro.LoadEntity(&webSite, accountAiro.Id,nil)
	if err != nil {
		log.Fatalf("Не удалось найти webSite: %v", err)
	}

	// Создаем вариант доставки "Почтой россии"
	entityRussianPost, err := accountAiro.CreateEntity(
		&models.DeliveryRussianPost{
			Name: "Доставка Почтой России",
			Enabled: true,
			AccessToken: "b07bk92rzBXosriAgmR5key1IpHq1Tpn",
			XUserAuthorization: "a29yb3RhZXZAdnR2ZW50LnJ1OmpIeXc2MnIzODNKc3F6aQ==",
			PostalCodeFrom: "109390",
			MailCategory: "ORDINARY",
			MailType: "POSTAL_PARCEL",
			PaymentSubjectId: 10, // Платеж
			MaxWeight: 20.0,
			Fragile: false,
			WithElectronicNotice: true,
			WithOrderOfNotice: true,
			WithSimpleNotice: false,
		})
	if err != nil {
		log.Fatalf("Не удалось получить DeliveryRussianPost: %v", err)
	}
	if err := webSite.AppendDeliveryMethod(entityRussianPost); err != nil {
		log.Fatalf("Не удалось добавить метод доставки в магазин: %v\n", err)
	}


	entityPickup, err := accountAiro.CreateEntity(&models.DeliveryPickup{Name: "Самовывоз из г. Москва, м. Текстильщики", Enabled: true,PaymentSubjectId: 10})
	if err != nil {
		log.Fatalf("Не удалось получить entityPickup: %v", err)
	}
	if err := webSite.AppendDeliveryMethod(entityPickup); err != nil {
		log.Fatalf("Не удалось добавить метод доставки в магазин: %v\n", err)
	}

	entityCourier, err := accountAiro.CreateEntity(
		&models.DeliveryCourier{
			Name: "Доставка курьером по г. Москва (в пределах МКАД)",
			Enabled: true,
			Price: 500,
			MaxWeight: 40.0,
			PaymentSubjectId: 10,
		})
	if err != nil {
		log.Fatalf("Не удалось получить entityCourier: %v", err)
	}
	if err := webSite.AppendDeliveryMethod(entityCourier); err != nil {
		log.Fatalf("Не удалось добавить метод доставки в магазин: %v\n", err)
	}

}

func UploadTestDataPart_III() {
	// 1. Получаем главный аккаунт
	airoAccount, err := models.GetAccount(8)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	els := []models.EventListener {
		// товар
		{Name: "Добавление товара на сайт",  EventId: 9,  HandlerId: 1, EntityId: 3, Enabled: true},
		{Name: "Обновление товара на сайте", EventId: 10, HandlerId: 1, EntityId: 4, Enabled: true},
		{Name: "Удаление товара с сайта", 	 EventId: 11, HandlerId: 1, EntityId: 5, Enabled: true},

		// Карточки товара
		{Name: "Добавление карточки товара", EventId: 12, HandlerId: 1, EntityId: 7, Enabled: true},
		{Name: "Обновление карточки товара", EventId: 13, HandlerId: 1, EntityId: 8, Enabled: true},
		{Name: "Удаление карточки товара",   EventId: 14, HandlerId: 1, EntityId: 9, Enabled: true},

		// Магазин (WebSite)
		{Name: "Обновление данных магазина", EventId: 19, HandlerId: 1, EntityId: 2, Enabled: true},

		// Статьи
		{Name: "Создание статьи на сайте", 	 EventId: 24, HandlerId: 1, EntityId: 15, Enabled: true},
		{Name: "Обновление статьи на сайте", EventId: 25, HandlerId: 1, EntityId: 16, Enabled: true},
		{Name: "Удаление статьи на сайте",   EventId: 26, HandlerId: 1, EntityId: 17, Enabled: true},

	}
	for i := range els {
		_, err = airoAccount.CreateEntity(&els[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	// 9. Создаем вебхкуи
	domainAiroSite := ""
	AppEnv := os.Getenv("APP_ENV")

	switch AppEnv {
	case "local":
		domainAiroSite = "http://airoclimate.me"
	case "public":
		domainAiroSite = "https://airoclimate.ru"
	default:
		domainAiroSite = "https://airoclimate.ru"
	}

	webHooks := []models.WebHook {

		// WebSite
		{Name: "Update Web Site", Code: models.EventShopUpdated, URL: domainAiroSite + "/api/ratuscrm/webhooks/web-sites/{{.webSiteId}}", HttpMethod: http.MethodPatch},

		// Product
		{Name: "Create product", Code: models.EventProductCreated, URL: domainAiroSite + "/api/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodPost},
		{Name: "Update product", Code: models.EventProductUpdated, URL: domainAiroSite + "/api/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete product", Code: models.EventProductDeleted, URL: domainAiroSite + "/api/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all products", Code: models.EventProductsUpdate, URL: domainAiroSite + "/api/ratuscrm/webhooks/products", HttpMethod: http.MethodGet},

		// ProductCard
		{Name: "Create product card", Code: models.EventProductCardCreated, URL: domainAiroSite + "/api/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodPost},
		{Name: "Update product card", Code: models.EventProductCardUpdated, URL: domainAiroSite + "/api/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete product card", Code: models.EventProductCardDeleted, URL: domainAiroSite + "/api/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all product cards", Code: models.EventProductCardsUpdate, URL: domainAiroSite + "/api/ratuscrm/webhooks/product-cards", HttpMethod: http.MethodGet},

		// Groups
		{Name: "Create web page", Code: models.EventWebPageCreated, URL: domainAiroSite + "/api/ratuscrm/webhooks/web-pages/{{.webPageId}}", HttpMethod: http.MethodPost},
		{Name: "Update web page", Code: models.EventWebPageUpdated, URL: domainAiroSite + "/api/ratuscrm/webhooks/web-pages/{{.webPageId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete web page", Code: models.EventWebPageDeleted, URL: domainAiroSite + "/api/ratuscrm/webhooks/web-pages/{{.webPageId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all web pages", Code: models.EventWebPagesUpdate, URL: domainAiroSite + "/api/ratuscrm/webhooks/web-pages", HttpMethod: http.MethodGet},

		// Articles
		{Name: "Create article", Code: models.EventArticleCreated, URL: domainAiroSite + "/api/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodPost},
		{Name: "Update article", Code: models.EventArticleUpdated, URL: domainAiroSite + "/api/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete article", Code: models.EventArticleDeleted, URL: domainAiroSite + "/api/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all articles", Code: models.EventArticlesUpdate, URL: domainAiroSite + "/api/ratuscrm/webhooks/articles", HttpMethod: http.MethodGet},

		{Name: "Upload all webSite data", Code: models.EventUpdateAllShopData, URL: domainAiroSite + "/api/ratuscrm/webhooks/upload/all", HttpMethod: http.MethodGet},
	}
	for i,_ := range webHooks {
		// _, err = airoAccount.CreateWebHook(webHooks[i])
		_, err = airoAccount.CreateEntity(&webHooks[i])
		if err != nil {
			log.Fatal("Не удалось создать webHook: ", err)
		}

	}

	// Добавляем шаблоны писем для airoclimate
	data, err := ioutil.ReadFile("/var/www/ratuscrm/files/airoclimate/emails/example.html")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	emailTemplates := []models.EmailTemplate{
		{Name: "Спасибо за ваш заказ", Description: "Уведомление клиента о заказе, который не оплачен.", HTMLData: string(data)},
		{Name: "Новый заказ", Description: "Уведомление о новом заказе для менеджеров", HTMLData: string(data)},
		{Name: "Ваш заказ отправлен", Description: "Уведомление для клиента об отправке заказа по почте.", HTMLData: string(data)},
		{Name: "Благодарим за покупку", Description: "Письмо-благодарность для клиента, после оплаты.", HTMLData: string(data)},
	}
	for i := range emailTemplates {
		_, err = airoAccount.CreateEntity(&emailTemplates[i])
		if err != nil {log.Fatal(err)}
	}

	// =================================


	numOne := uint(1)
	num5 := uint(5)
	num6 := uint(6)
	num7 := uint(7)
	emailNotifications := []models.EmailNotification {
		{
			Status: models.WorkStatusPending, DelayTime: 0, Name:"Оповещение менеджера", Subject: utils.STRp("Поступил новый заказ"), EmailTemplateId: &numOne,
			RecipientUsersList: datatypes.JSON(utils.UINTArrToRawJson([]uint{2})),
			EmailBoxId: &num5,

		},
		{
			Status: models.WorkStatusPending, DelayTime: 0, Name:"Оповещение клиента", Subject: utils.STRp("Ваш заказ получен"), EmailTemplateId: &numOne,
			RecipientUsersList: datatypes.JSON( utils.UINTArrToRawJson([]uint{7})),
			EmailBoxId: &num6,
		},
		{
			Status: models.WorkStatusPending, DelayTime: 0, Name:"Оповещение об отправке заказа",Subject: utils.STRp("Ваш заказ отправлен по почте"),  EmailTemplateId: &numOne,
			EmailBoxId: &num7,
		},

	}
	for _,v := range emailNotifications {
		_, err = airoAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}
}


func ToStringPointer(s string) *string {
	return &s
}

func LoadImagesAiroClimate(count int)  {

	account, err := models.GetAccount(8)
	if err != nil {
		fmt.Println("Не удалось загрузить изображения для аккаунта", err)
	}

	for  index := 1; index < count; index++ {
		url := "/var/www/ratuscrm/files/airoclimate/images/" + strconv.Itoa(index) + "/"
		files, err := ioutil.ReadDir(url)
		if err != nil {
			log.Fatal(err)
		}

		// идем по файлам
		for fileId, file := range files {

			//fmt.Println("Open: ", url + file.Name())
			f, err := os.Open(url + file.Name())
			if err != nil {
				panic(err)
			}
			defer f.Close()


			body, err := ioutil.ReadFile(url + file.Name())
			if err != nil {
				log.Fatalf("unable to read file: %v", err)
			}

			mimeType, err := GetFileContentType(f)
			if err != nil {
				log.Fatalf("unable to mimeType file: %v", err)
			}

			fs := models.Storage{
				AccountId: account.Id,
				Name: strings.ToLower(file.Name()),
				Data: body,
				MIME: mimeType,
				Size: file.Size(),
			}

			/*if err = fs.SetAutoPriority(); err != nil {
				log.Fatal(err)
			}*/

			// file, err := account.StorageCreateFile(&fs)
			file, err := account.CreateEntity(&fs)
			if err != nil {
				log.Fatalf("unable to create file: %v", err)
			}

			err = (models.Product{Id: uint(index)}).AppendAssociationImage(file)
			if err != nil {
				log.Fatalf("Error product: %v", err)
			}

			// fmt.Printf("index: %v, fileId: %v\n", index, fileId)
			// Добавляем изображение для карточки товара
			if fileId == 0 {
				err = (models.ProductCard{Id: uint(index)}).AppendAssociationImage(file)
				if err != nil {
					log.Fatalf("Error ProductCard AppendAssociationImage:  %v", err)
				}
			}
		}
	}
}

func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func LoadArticlesAiroClimate()  {
	account, err := models.GetAccount(8)
	if err != nil {
		fmt.Println("Не удалось найти аккаунт для загрузки статей", err)
	}

	url := "/var/www/ratuscrm/files/airoclimate/articles/"
	files, err := ioutil.ReadDir(url)
	if err != nil {
		log.Fatal(err)
	}

	// идем по файлам
	for i, file := range files {

		//fmt.Println("Open: ", url + file.Name())
		f, err := os.Open(url + file.Name())
		if err != nil {
			panic(err)
		}
		defer f.Close()

		body, err := ioutil.ReadFile(url + file.Name())
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		Title 	:= utils.STRp("Бактерицидные облучатели")
		Path 	:= utils.STRp("bactericidal-irradiators")
		Breadcrumb := utils.STRp("Бактерицидные облучатели")
		Name 	:= utils.STRp("Бактерицидные облучатели")
		
		if i ==1 {
			Title 	= utils.STRp("Бактерицидные рециркуляторы")
			Path 	= utils.STRp("bactericidal-recirculators")
			Breadcrumb = utils.STRp("Бактерицидные рециркуляторы")
			Name 	= utils.STRp("Бактерицидные рециркуляторы")
		}


		articleNew := models.Article{
			WebSiteId: utils.UINTp(account.Id),
			// Name: utils.STRp(strings.ToLower(file.Name())),
			MetaTitle: Title,
			Path: Path,
			Breadcrumb: Breadcrumb,
			Name: Name,
			Public: true,
			Shared: true,
			Body: utils.STRp(string(body)),
		}
		_, err = account.CreateEntity(&articleNew)
		if err != nil {
			log.Fatalf("unable to create file: %v", err)
		}
		/*_, ok := articleEntity.(*models.Article)
		if !ok {
			log.Fatal("article, ok := articleEntity.(*models.Article)")
		}*/

	}

}

func LoadProductDescriptionAiroClimate()  {
	account, err := models.GetAccount(8)
	if err != nil {
		fmt.Println("Не удалось найти аккаунт для загрузки статей", err)
	}

	url := "/var/www/ratuscrm/files/airoclimate/products/"
	files, err := ioutil.ReadDir(url)
	if err != nil {
		log.Fatal(err)
	}

	// идем по файлам
	for _, file := range files {

		//fmt.Println("Open: ", url + file.Name())
		f, err := os.Open(url + file.Name())
		if err != nil {
			panic(err)
		}
		defer f.Close()

		body, err := ioutil.ReadFile(url + file.Name())
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		// fmt.Println("article:", file.Name())
		split := strings.Split(file.Name(), ".")
		fileId, err := strconv.ParseUint(split[0], 10, 64)
		if err != nil {
			log.Fatalf("unable to read id file name: %v", err)
		}
		
		/*_, err = account.UpdateProduct(uint(fileId), map[string]interface{}{"Description":string(body)})
		if err != nil {
			log.Fatalf("unable to update product: %v", err)
		}*/
		err = account.UpdateEntity(&models.Product{Id: uint(fileId), AccountId: account.Id},map[string]interface{}{"Description":string(body)},nil)
		if err != nil {
			log.Fatalf("unable to update product: %v", err)
		}

	}



}

func LoadProductCategoryDescriptionAiroClimate()  {
	/*account, err := models.GetAccount(8)
	if err != nil {
		fmt.Println("Не удалось найти аккаунт для загрузки статей", err)
	}*/

	url := "/var/www/ratuscrm/files/airoclimate/categories/"
	files, err := ioutil.ReadDir(url)
	if err != nil {
		log.Fatal(err)
	}

	// идем по файлам
	for _, file := range files {

		//fmt.Println("Open: ", url + file.Name())
		f, err := os.Open(url + file.Name())
		if err != nil {
			panic(err)
		}
		defer f.Close()

		body, err := ioutil.ReadFile(url + file.Name())
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}

		// fmt.Println("article:", file.Name())
		split := strings.Split(file.Name(), ".html")
		routeName := split[0]

		// fmt.Println(routeName)
		// return

		group := models.WebPage{}

		if err := models.GetDB().First(&group, "route_name = ?", routeName).Error; err != nil {
			log.Fatalf("cant find group by route name: %v", err)
		}

		mapUpd := map[string]interface{}{"Description":string(body)}
		err = models.GetDB().Model(group).Omit("id").Updates( mapUpd ).Error
		if err != nil {
			log.Fatalf("unable to update product group descr: %v", err)
		}
	}


}

func UploadTestDataPart_IV()  {

	// 1. Получаем AiroClimate аккаунт
	airoAccount, err := models.GetAccount(8)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	var webSite models.WebSite
	if err := airoAccount.LoadEntity(&webSite, airoAccount.Id,nil); err != nil { log.Fatal(err)}

	payment2Deliveries := []models.Payment2Delivery {
		{AccountId: airoAccount.Id, WebSiteId: webSite.Id, PaymentId: 1, PaymentType: "payment_cashes", DeliveryId: 1, DeliveryType: "delivery_pickups"},
		{AccountId: airoAccount.Id, WebSiteId: webSite.Id, PaymentId: 1, PaymentType: "payment_yandexes", DeliveryId: 1, DeliveryType: "delivery_pickups"},
		{AccountId: airoAccount.Id, WebSiteId: webSite.Id, PaymentId: 1, PaymentType: "payment_yandexes", DeliveryId: 1, DeliveryType: "delivery_couriers"},
		{AccountId: airoAccount.Id, WebSiteId: webSite.Id, PaymentId: 1, PaymentType: "payment_yandexes", DeliveryId: 1, DeliveryType: "delivery_russian_posts"},
	}

	for _,v := range(payment2Deliveries) {
		if err := webSite.AppendPayment2Delivery(v.PaymentId, v.PaymentType, v.DeliveryId, v.DeliveryType); err != nil {
			log.Fatal(err)
		}
	}

	// Создаем способ оплаты YandexPayment
	entityPayment, err := airoAccount.CreateEntity(
		&models.PaymentYandex{
			Name:   "Онлайн-оплата на сайте",
			Label:   "Онлайн-оплата банковской картой",
			ApiKey: "test_f56EEL_m2Ky7CJnnRjSpb4JLMhiGoGD3X6ScMHGPruM",
			ShopId: 730509,
			ReturnUrl: "https://airoclimate.ru",
			Enabled: true,
			WebSiteId: webSite.Id,
			SavePaymentMethod: false,
			Capture: false,
		})
	if err != nil {
		log.Fatalf("Не удалось создать entityPayment: ", err)
	}
	var _paymentYandex models.PaymentYandex
	if err = airoAccount.LoadEntity(&_paymentYandex,entityPayment.GetId(),nil); err != nil {
		log.Fatalf("Не удалось найти entityPayment: ", err)
	}

	// Создаем способ оплаты PaymentCash
	entityPayment2, err := airoAccount.CreateEntity(
		&models.PaymentCash{
			Name:   "Оплата наличными при самовывозе",
			Label:   "Оплата наличными при получении",
			WebSiteId: webSite.Id,
			Enabled: true,
		})
	if err != nil {
		log.Fatalf("Не удалось создать entityPayment: ", err)
	}
	var _paymentCash models.PaymentCash
	if err = airoAccount.LoadEntity(&_paymentCash,entityPayment2.GetId(),nil); err != nil {
		log.Fatalf("Не удалось найти paymentCash: ", err)
	}

	deliveries := webSite.GetDeliveryMethods()
	for i,v := range(deliveries) {
		if v.GetCode() == "russianPost" {
			if err := deliveries[i].AppendPaymentMethods([]models.PaymentMethod{&_paymentYandex}); err != nil {
				return
			}
		} else {
			if err := deliveries[i].AppendPaymentMethods([]models.PaymentMethod{&_paymentCash, &_paymentYandex}); err != nil {
				return
			}
		}

	}

}

func UploadTestDataPart_V() {

	// 1. Получаем главный аккаунт
	mAcc, err := models.GetMainAccount()
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	owner, err := mAcc.GetUser(1)
	if err != nil {
		log.Fatalf("Не удалось найти owner: %v", err)
	}
	mex388, err := mAcc.GetUser(2)
	if err != nil {
		log.Fatalf("Не удалось найти mex388: %v", err)
	}
	
	roleAdminMain, err := mAcc.GetRoleByTag(models.RoleAdmin)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	roleClientMain, err := mAcc.GetRoleByTag(models.RoleClient)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	timeNow := time.Now().UTC()

	// ######### Stan-Prof ############

	// 1. Создаем Романа
	romanUfa, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("roman_s"),
			Email:		utils.STRp("roman_s@stan-prof.ru"),
			PhoneRegion: utils.STRp("RU"),
			Phone: 		utils.STRp("+79872519935"),
			Password:	utils.STRp("qwerty109#QW"),
			Name:		utils.STRp("Роман"),
			Surname:	utils.STRp(""),
			Patronymic:	utils.STRp(""),
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil {
		log.Fatal("Не удалось создать romanUfa'a: ", err)
	}

	dvc, err := models.GetUserVerificationTypeByCode(models.VerificationMethodEmailAndPhone)
	if err != nil || dvc == nil {
		log.Fatal("Не удалось получить верификацию...")
		return
	}

	// 2. создаем из-под Романа Stan-Prof
	stanProf, err := romanUfa.CreateAccount( models.Account{
		Name:                                "StanProf",
		Website:                             "www.stan-prof.ru",
		Type:                                "store",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"email"})),
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            &dvc.Id,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil || stanProf == nil {
		log.Fatal("Не удалось создать аккаунт StanProf")
		return
	}

	_, err = stanProf.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с сайтом"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 3. добавляем меня как админа
	_, err = stanProf.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}


	// 4. Создаем домен для StanProf
	_webSiteStanProf, err := stanProf.CreateEntity(&models.WebSite{
		Hostname: "stan-prof.ru",
		DKIMPublicRSAKey: `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDYq5m0HLzmuGrIvghDA3uHR8rF
JTmhGutraXmqrHT3dLx4en15H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0
pjg2ee1PxllVduwzzwzIUfo3k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIp
FeTiqGsEKHqAbsiq0QIDAQAB
-----END PUBLIC KEY-----
`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDYq5m0HLzmuGrIvghDA3uHR8rFJTmhGutraXmqrHT3dLx4en15
H8y7ml37dLrqUraDQTcm7Xmi/zJaJl5i9WLOUui0pjg2ee1PxllVduwzzwzIUfo3
k6Z9I+RiTLWtjtUCGvR1eJ7K7uzUdQOVv94M4nIpFeTiqGsEKHqAbsiq0QIDAQAB
AoGBAIwqFnipzpoC2zGZnYacjiDplIcMmcdavsjxpr9+aPxcDEB0HQ9qPutt5TZy
QSis8Vzfp2oLXhDPvicLgyvJSmy1lsevb8KtXR7sUR6OOsSk9UDxdkziSpCxn3b1
1cTHeS+iH5vT7oNAMl4bczA/Mf+5HbYYZOVBkAK0XZgfPlgBAkEA+HgODNGFcakr
npk9nIYO3fQ47j3yE0WH2a7LXcMaqcqSc7+rOPI1xNqmNY1wInYKrE3nVIbt94Hp
zEYvbo1JYQJBAN88z23O09SOK0vDNztqHLs5CONXaulHjhsumxRz/sFnJEA+39IB
H7cLKY0bNj32naSPvlMvpOXrUs3RvoGtp3ECQA3+XPQ13KkhGPbOtJybJD014vrv
/DE/qj4FTSjTsy263fKs6czEP2j903ySoa0fLsyzeoLVMpPM+kjp7wPv6yECQQDC
1SH8ply+0G39K53S3yPAS2D0eeE3RFpEYwK6cRpQ1aIP6FATyyNQ8A8Ntkg7ADhJ
3aKwm46BlEaql3p7V4IhAkBTq1wkY0mGjIfgfN5sTg8B4SbL/dhfLrl9xeRbc5qA
AJnnVkwI9ntl6+d3uML4VA7hUloxsufH7fZ3lmaR+453
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}
	webSiteStanProf, ok := _webSiteStanProf.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 5. Добавляем почтовые ящики в домен 357gr
	_, err = webSiteStanProf.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "СтанПроф", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}

	//////// Cs-Garant

	// 1. Создаем Ярослава
	yaroslav, err := mAcc.CreateUser(
		models.User{
			Username:	utils.STRp("yaroslavAnanev"),
			Email:		utils.STRp("mnggarant@bk.ru"),
			PhoneRegion: utils.STRp("RU"),
			Phone: 		utils.STRp("89221927108"),
			Password:	utils.STRp("qwerty123#Q"),
			Name:		utils.STRp("Ярослав"),
			Surname:	utils.STRp("Ананьев"),
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil {
		log.Fatal("Не удалось создать stas'a: ", err)
	}

	// 1. Создаем синдикат из-под Станислава
	accCsGarant, err := yaroslav.CreateAccount(models.Account{
		Name:                                "CS-Garant",
		Website:                             "https://cs-garant.ru/",
		Type:                                "service",
		ApiEnabled:                          true,
		UiApiEnabled:                        false,
		VisibleToClients:                    false,
	})
	if err != nil {
		log.Fatal("Не удалось создать аккаунт CS-Garant")
		return
	}

	// 2. добавляем меня как админа
	_, err = accCsGarant.AppendUser(*owner, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2 Добавляем Mex388
	_, err = accCsGarant.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in accCsGarant")
		return
	}

	_, err = accCsGarant.ApiKeyCreate(models.ApiKey{Name:"Для интеграции с системой"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 2. Создаем домен для Гаранта
	_webSiteGarant, err := accCsGarant.CreateEntity(&models.WebSite{
		Hostname: "cs-garant.ru",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyxCGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+ABtf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIdAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyx
CGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+A
Btf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIdAQAB
AoGAIR9YdelFBhrtM2WEVb/bnX+7vJ2mm+OLxTMyFuuvuvsiw6TBnHgXncYZBk/D
Zm9uhfCKU1loRIGd6gxY+dx+hVCFHh4tyQ+xvb+siTsDO3VXhHCq+XZpstDanrS0
kEjDPx95QYgJ3taG55Agu2Ql/cgevyFevOhXUPrZ6lStdcUCQQDxpSPUywPgOas5
CFMWB5k5+DRAz9CygH5L7i53RnitwPL3jHvwOHs5JD25lD9IfKVyGuJtYeUTPenp
FlIxzv+TAkEA0HAuDHrCItg1x/UDO9N+IafTFN5+31Me9POiOGkghXfbWJCfxaBW
wJWLTPI7p+PT07/sRusQpGRiGi0RagZbowJAVqXsr0UM4r5LE2xUvrWC0DKcKhFa
uGcy4m9J4iM26rchaHrLhlv6c4b3SzBJcOihOsVBJA/SYI/27EnAt3OOWQJAXhjm
kPeyQKy+ysBPb2iw3ly3LAqt1//cT9TU/QZoihhry3WuyzbxMwvP0TLhv49Yh5Vz
AykHYE95AjwqSmUIZQJAaRJMuw5gVSjQaLz/qoiMVEQO7vmazsiB9/YKTPp18I+4
pBRlD1bMcxJEBYvc/tLA1LqyGGhd1mabVQ7iYPq45w==
-----END RSA PRIVATE KEY-----
`,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для Синдиката: ", err)
	}
	webSiteSynd, ok := _webSiteGarant.(*models.WebSite)
	if !ok {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 3. Добавляем почтовые ящики
	_, err = webSiteSynd.CreateEmailBox(models.EmailBox{Default: true, Allowed: true, Name: "Центр сертификации Гарант", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Синдиката: ", err)
	}

	return

}

func UploadBroUserData() {
	// Получаем BroUser
	account, err := models.GetAccount(7)
	if err != nil {
		log.Fatalf("Не удалось найти BroUser аккаунт: %v", err)
	}

	data, err := ioutil.ReadFile("/var/www/ratuscrm/files/brouser/emails/example.html")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	emailTemplates := []models.EmailTemplate{
		{Name: "Шаблон для онбординг серии - 1", Description: "1-е письмо в серии.", HTMLData: string(data)},
		{Name: "Шаблон для онбординг серии - 2", Description: "2-е письмо в серии.", HTMLData: string(data)},
		{Name: "Шаблон для онбординг серии - 3", Description: "3-е письмо в серии", HTMLData: string(data)},
	}
	for i := range emailTemplates {
		_, err = account.CreateEntity(&emailTemplates[i])
		if err != nil {log.Fatal(err)}
	}

	emailQueueE, _ := account.CreateEntity(&models.EmailQueue{
		Name: "Onboarding",
		Status: models.WorkStatusPending,
	})

	_, _ = account.CreateEntity(&models.EmailQueueEmailTemplate{
		AccountId: account.Id,
		EmailQueueId: emailQueueE.GetId(),
		Enabled: true,
		Step: 1,
		EmailTemplateId: utils.UINTp(5),
		EmailBoxId: &account.Id,
		DelayTime: 0,
		Subject: utils.STRp("Тема письма 1"),
		CreatedAt: time.Now().UTC(),
	})
	_, _ = account.CreateEntity(&models.EmailQueueEmailTemplate{
		AccountId: account.Id,
		EmailQueueId: emailQueueE.GetId(),
		Enabled: false,
		Step: 2,
		EmailTemplateId: utils.UINTp(6),
		EmailBoxId: &account.Id,
		DelayTime: 0,
		Subject: utils.STRp("Тема письма 2"),
		CreatedAt: time.Now().UTC(),
	})
	_, _ = account.CreateEntity(&models.EmailQueueEmailTemplate{
		AccountId: account.Id,
		EmailQueueId: emailQueueE.GetId(),
		Enabled: false,
		Step: 3,
		EmailTemplateId: utils.UINTp(7),
		EmailBoxId: &account.Id,
		DelayTime: 0,
		Subject: utils.STRp("Тема письма 3"),
		CreatedAt: time.Now().UTC(),
	})
}

func Upload357grData() {

	account, err := models.GetAccount(5)
	if err != nil || account.Name != "357 грамм" {
		log.Fatal("Ошибка поиска аккаунта 357 грамм: ", err)
	}

	Page, err := account.CreateEntity(&models.WebPage {
		AccountId: account.Id, WebSiteId: &account.Id, Label: utils.STRp("Главная"), Code: utils.STRp("root"),Path: utils.STRp("/"),
		MetaTitle: utils.STRp("Главная :: 357 грамм"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-home"), RouteName:  utils.STRp("home"),
	})
	if err != nil {
		log.Fatal("Не удалось создать mPage для airoClimat webSite: ", err)
		return
	}
	webPageRoot := Page.(*models.WebPage)

	_rootCatalogPage, err := webPageRoot.CreateChild(models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Весь каталог"), Path:  utils.STRp("catalog"),
		MetaTitle: utils.STRp("Каталог :: 357 грамм"),MetaKeywords: utils.STRp(""),MetaDescription: utils.STRp(""),
		IconName:  utils.STRp("far fa-th-large"), RouteName:  utils.STRp("catalog"),
	})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
		return
	}

	rootPageCatalog := _rootCatalogPage.(*models.WebPage)

	// ################# Каталог #################
	
	// Страница с чаем
	_webPageTea, err := rootPageCatalog.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чай"), Path:  utils.STRp("tea"),
		MetaTitle: utils.STRp("Каталог чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageTea := _webPageTea.(*models.WebPage)

	// Подстраницы с чаем

	// Китайский чай
	_, err = webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Китайский чай"), Path:  utils.STRp("china-tea"),
		MetaTitle: utils.STRp("Китайский чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.china-tea"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// красный чай
	_, err = webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Красный чай"), Path:  utils.STRp("red-tea"),
		MetaTitle: utils.STRp("Красный чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.red"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_webPageGreenTea, err := webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Зеленый чай"), Path:  utils.STRp("green-tea"),
		MetaTitle: utils.STRp("Зеленый чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.green"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageGreenTea := _webPageGreenTea.(*models.WebPage)

	_webPageOolongTea, err := webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Улун"), Path:  utils.STRp("oolong-tea"),
		MetaTitle: utils.STRp("Улуны :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.oolong"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageOolongTea := _webPageOolongTea.(*models.WebPage)

	_webPagePuerTea, err := webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Пуэр"), Path:  utils.STRp("puers"),
		MetaTitle: utils.STRp("Пуэры :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPagePuerTea := _webPagePuerTea.(*models.WebPage)
	_, err = webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Травяной чай"), Path:  utils.STRp("herbal-tea"),
		MetaTitle: utils.STRp("Травяной чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.herbal"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чайные добавки"), Path:  utils.STRp("additions-tea"),
		MetaTitle: utils.STRp("Чайные добавки :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.additions"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// подстраницы Зеленого чая ---
	_, err = webPageGreenTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Китайский зеленый чай"), Path:  utils.STRp("china"),
		MetaTitle: utils.STRp("Китайский зеленый чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.green.china"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageGreenTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Жасминовый чай"), Path:  utils.STRp("jasmine"),
		MetaTitle: utils.STRp("Жасминовый чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.green.jasmine"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageGreenTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Связанный чай"), Path:  utils.STRp("related"),
		MetaTitle: utils.STRp("Связанный чай :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.green.related"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	// подстраницы Улунского чая ---
	_, err = webPageOolongTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Южнофуцзянские улуны"), Path:  utils.STRp("china"),
		MetaTitle: utils.STRp("Южнофуцзянские улуны :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.oolong.china"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageOolongTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Уишаньские улуны"), Path:  utils.STRp("wuyishan"),
		MetaTitle: utils.STRp("Уишаньские улуны :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.oolong.wuyishan"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageOolongTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Тайваньские улуны"), Path:  utils.STRp("taiwanese"),
		MetaTitle: utils.STRp("Тайваньские улуны :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.oolong.taiwanese"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageOolongTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Гуандунские улуны"), Path:  utils.STRp("related"),
		MetaTitle: utils.STRp("Гуандунские улуны :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.green.related"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// подстраницы Пуэра чая ---
	_, err = webPagePuerTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Шу Пуэр пресованный"), Path:  utils.STRp("shu"),
		MetaTitle: utils.STRp("Шу Пуэр пресованный :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer.shu-pressed"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPagePuerTea.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Шу Пуэр рассыпной"), Path:  utils.STRp("shu-loose"),
		MetaTitle: utils.STRp("Шу Пуэр рассыпной :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer.shu-loose"),
	})
	
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPagePuerTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Шэн Пуэр пресованный"), Path:  utils.STRp("shen"),
		MetaTitle: utils.STRp("Шэн Пуэр пресованный :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer.shen-pressed"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPagePuerTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Шэн Пуэр рассыпной"), Path:  utils.STRp("shen-loose"),
		MetaTitle: utils.STRp("Шэн Пуэр рассыпной :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer.shen-loose"),
	})
	
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPagePuerTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Белый Пуэр"), Path:  utils.STRp("white"),
		MetaTitle: utils.STRp("Белый Пуэр :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea.puer.white"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// ---

	// Страница с кофе
	_webPageCoffee, err := rootPageCatalog.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Кофе"), Path:  utils.STRp("coffee"),
		MetaTitle: utils.STRp("Каталог кофе :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.coffee"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageCoffee для 357gr webSite: ", err)
		return
	}
	webPageCoffee := _webPageCoffee.(*models.WebPage)


	// подстраницы Кофе ---
	_, err = webPageCoffee.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Зеленый кофе"), Path:  utils.STRp("beans"),
		MetaTitle: utils.STRp("Зеленый кофе :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.green"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_webPageCoffeeBeans, err := webPageCoffee.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Кофе в зернах"), Path:  utils.STRp("beans"),
		MetaTitle: utils.STRp("Кофе в зернах :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.beans"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageCoffeeBeans := _webPageCoffeeBeans.(*models.WebPage)
	// Подстраницы для кофе
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Классический"), Path:  utils.STRp("beans"),
		MetaTitle: utils.STRp("Классические смеси кофе :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.beans.classic"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Фирменные смеси"), Path:  utils.STRp("classic"),
		MetaTitle: utils.STRp("Фирменные смеси кофе :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.beans.classic"),
	})
	if err != nil {
		log.Fatal("Не удалось создать webPageCoffeeBeans для 357gr webSite: ", err)
		return
	}


	// Страница с подарками
	_webPageGifts, err := rootPageCatalog.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Подарки"), Path:  utils.STRp("gifts"),
		MetaTitle: utils.STRp("Каталог подарков :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.gifts"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageGifts для 357gr webSite: ", err)
		return
	}
	webPageGifts := _webPageGifts.(*models.WebPage)

	// подстраницы под подарки
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чайные сувениры"), Path:  utils.STRp("souvenirs"),
		MetaTitle: utils.STRp("Чайные сувениры :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.gifts.souvenirs"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Подарочные корзины"), Path:  utils.STRp("baskets"),
		MetaTitle: utils.STRp("Подарочные корзины :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.gifts.baskets"),
	})
	if err != nil {
		log.Fatal("Не удалось создать webPageCoffeeBeans для 357gr webSite: ", err)
		return
	}
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Подарочные букеты"), Path:  utils.STRp("bouquets"),
		MetaTitle: utils.STRp("Подарочные букеты :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.gifts.bouquets"),
	})
	if err != nil {
		log.Fatal("Не удалось создать webPageCoffeeBeans для 357gr webSite: ", err)
		return
	}
	_, err = webPageCoffeeBeans.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Подарки руководителю"), Path:  utils.STRp("manager"),
		MetaTitle: utils.STRp("Подарки руководителю :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.gifts.manager"),
	})
	if err != nil {
		log.Fatal("Не удалось создать webPageCoffeeBeans для 357gr webSite: ", err)
		return
	}


	// Страница с посудой
	_webPageTeaThings, err := rootPageCatalog.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Посуда и аксессуары"), Path:  utils.STRp("tea-things"),
		MetaTitle: utils.STRp("Посуда и аксессуары :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.tea-things"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTeaThings для 357gr webSite: ", err)
		return
	}
	webPageTeaThings := _webPageTeaThings.(*models.WebPage)

	// подстраницы под посуду
	_webPageTeaBrewing, err := webPageTeaThings.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Для заваривания"), Path:  utils.STRp("brewing"),
		MetaTitle: utils.STRp("Посуда для заваривания чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.tea-things.brewing"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// под подстраницы под посуду для заваривания чая
	webPageTeaBrewing := _webPageTeaBrewing.(*models.WebPage)
	_, err = webPageTeaBrewing.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Традиционная"), Path:  utils.STRp("traditional"),
		MetaTitle: utils.STRp("Традиционная посуда для заваривания чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.tea-things.brewing.traditional"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaBrewing.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Европейская"), Path:  utils.STRp("european"),
		MetaTitle: utils.STRp("Европейская посуда для заваривания чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.tea-things.brewing.european"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaBrewing.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Типоды"), Path:  utils.STRp("gunfu"),
		MetaTitle: utils.STRp("Типоды (чайники с кнопкой) для заваривания чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.tea-things.brewing.gunfu"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaBrewing.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Сифон"), Path:  utils.STRp("siphons"),
		MetaTitle: utils.STRp("Сифоны для варки чая и кофе :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.coffee.tea-things.brewing.siphons"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// под подстраницы под "Для чаепития"
	_webPageForTea, err := webPageTeaThings.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Для чаепития"), Path:  utils.STRp("for-tea"),
		MetaTitle: utils.STRp("Посуда для чаепития :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.tea-things.for-tea"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTeaThings для 357gr webSite: ", err)
		return
	}
	webPageForTea := _webPageForTea.(*models.WebPage)

	// под подстраницы "Для чаепития"
	_, err = webPageForTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Для чайной церемонии"), Path:  utils.STRp("ceremony"),
		MetaTitle: utils.STRp("Посуда для чайной церемонии :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.ceremony"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	// под подстраницы под чаепитие
	_, err = webPageForTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Сервизы"), Path:  utils.STRp("sets"),
		MetaTitle: utils.STRp("Чайные сервизы :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.for-tea.sets"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageForTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чашки"), Path:  utils.STRp("cups"),
		MetaTitle: utils.STRp("Чашки для чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.for-tea.cups"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageForTea.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Пиалы"), Path:  utils.STRp("pialas"),
		MetaTitle: utils.STRp("Пиалы для чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.for-tea.pialas"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	
	// Подстраницы "Для хранения чая"
	_webPageTeaStoring, err := webPageTeaThings.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Для хранения чая"), Path:  utils.STRp("storing"),
		MetaTitle: utils.STRp("Посуда для хранения чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageTeaStoring := _webPageTeaStoring.(*models.WebPage)

	// под подстраницы "Для хранения чая"
	_webPageTeaStoringBanks, err := webPageTeaStoring.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Банки для хранения чая"), Path:  utils.STRp("banks"),
		MetaTitle: utils.STRp("Банки для хранения чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing.banks"),
	})
	webPageTeaStoringBanks := _webPageTeaStoringBanks.(*models.WebPage)

	_, err = webPageTeaStoringBanks.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Банки металлические"), Path:  utils.STRp("metal"),
		MetaTitle: utils.STRp("Банки металлические :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing.banks.metal"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaStoringBanks.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Банки картонные"), Path:  utils.STRp("cardboard"),
		MetaTitle: utils.STRp("Банки картонные :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing.banks.cardboard"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaStoring.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Упаковка для пуэров"), Path:  utils.STRp("packaging-puer"),
		MetaTitle: utils.STRp("Упаковка для пуэров :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing.packaging-puer"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaStoring.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чайницы"), Path:  utils.STRp("caddys"),
		MetaTitle: utils.STRp("Чайницы для чая :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.storing.caddys"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTeaStoring для 357gr webSite: ", err)
		return
	}

	// Подстраницы Посуда: "Разное"
	_webPageTeaOthers, err := webPageTeaThings.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Разное"), Path:  utils.STRp("others"),
		MetaTitle: utils.STRp("Посуда и аксессуары: разное :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	webPageTeaOthers := _webPageTeaOthers.(*models.WebPage)

	_, err = webPageTeaOthers.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Благовония"), Path:  utils.STRp("incenses"),
		MetaTitle: utils.STRp("Благовония :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others.incenses"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaOthers.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чайные фигурки"), Path:  utils.STRp("figurines"),
		MetaTitle: utils.STRp("Чайные фигурки :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others.figurines"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaOthers.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Электрические плитки"), Path:  utils.STRp("electric-tiles"),
		MetaTitle: utils.STRp("Электрические плитки :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others.electric-tiles"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaOthers.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чистящие средства"), Path:  utils.STRp("cleaners"),
		MetaTitle: utils.STRp("Чистящие средства :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others.cleaners"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}
	_, err = webPageTeaOthers.CreateChild( models.WebPage {
		AccountId: account.Id, Code:  utils.STRp("catalog"),Label:  utils.STRp("Чайный инструмент"), Path:  utils.STRp("tools"),
		MetaTitle: utils.STRp("Чайный инструмент :: 357 грамм"),
		IconName:  utils.STRp("far fa-fan-table"), RouteName:  utils.STRp("catalog.tea-things.others.tools"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTea для 357gr webSite: ", err)
		return
	}

	
	// Новинки
	_webPageNewProducts, err := rootPageCatalog.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Новинки"), Path:  utils.STRp("new-products"),
		MetaTitle: utils.STRp("Новые поступления :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.new-products"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageNewProducts для 357gr webSite: ", err)
		return
	}
	webPageNewProducts := _webPageNewProducts.(*models.WebPage)

	// Акции и скидки
	_webPagePromotions, err := rootPageCatalog.CreateChild(models.WebPage{
		AccountId: account.Id, Code:  utils.STRp("catalog"), Label:  utils.STRp("Акции и скидки"), Path:  utils.STRp("promotions"),
		MetaTitle: utils.STRp("Акции и скидки :: 357 грамм"),
		IconName:  utils.STRp("far fa-box-full"), RouteName:  utils.STRp("catalog.promotions"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _webPageTeaThings для 357gr webSite: ", err)
		return
	}
	webPagePromotions := _webPagePromotions.(*models.WebPage)

	fmt.Println(webPagePromotions, webPageNewProducts,webPageTeaThings,webPageGifts,webPageCoffee,webPageTea)

	// ################# Базовые страницы #################

	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Контакты"), Path:  utils.STRp("contacts"),
			MetaTitle: utils.STRp("Контакты :: 357 грамм"),
			IconName:  utils.STRp("far fa-address-book"), RouteName:  utils.STRp("info.contacts"),  Priority: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}

	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("info"), Label:  utils.STRp("О магазине"), Path:  utils.STRp("about"),
			MetaTitle: utils.STRp("О магазине :: 357 грамм"),
			IconName:  utils.STRp("far fa-home-heart"), RouteName:  utils.STRp("info.about"),  Priority: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}

	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Полезные материалы"), Path:  utils.STRp("articles"),
			MetaTitle: utils.STRp("Полезные материалы :: 357 грамм"),
			IconName:  utils.STRp("far fa-books"), RouteName:  utils.STRp("articles"),  Priority: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}
	
	deliveryGrE, err := webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Доставка товара"), Path:  utils.STRp("delivery"),
			MetaTitle: utils.STRp("Доставка товара :: 357 грамм"),
			IconName:  utils.STRp("far fa-shipping-fast"), RouteName:  utils.STRp("delivery"), Priority: 1,
		})
	if err != nil {
		log.Fatal(err)
	}
	deliveryGroupRoute := deliveryGrE.(*models.WebPage)
	_, err = deliveryGroupRoute.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Способы оплаты"), Path:  utils.STRp("payment"),
			MetaTitle: utils.STRp("Способы оплаты :: 357 грамм"),
			IconName:  utils.STRp("far fa-hand-holding-usd"), RouteName:  utils.STRp("delivery.payment"), Priority: 2,
		})
	_, err = deliveryGroupRoute.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("delivery"), Label:  utils.STRp("Возврат товара"), Path:  utils.STRp("moneyback"),
			MetaTitle: utils.STRp("Возврат товара :: 357 грамм"),
			IconName:  utils.STRp("far fa-exchange-alt"), RouteName:  utils.STRp("delivery.moneyback"), Priority: 3,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}

	_, err = webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("info"), Label:  utils.STRp("Политика конфиденциальности"), Path:  utils.STRp("privacy-policy"),
			MetaTitle: utils.STRp("Политика конфиденциальности :: 357 грамм"),
			IconName:  utils.STRp("far fa-home-heart"), RouteName:  utils.STRp("info.privacy-policy"),      Priority: 6,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}
	
	// ################# Корзина страницы #################

	_cartWebPage, err := webPageRoot.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("cart"), Label:  utils.STRp("Корзина"), Path:  utils.STRp("cart"),
			MetaTitle: utils.STRp("Корзина :: 357 грамм"),
			IconName:  utils.STRp("far fa-cart-arrow-down"), RouteName:  utils.STRp("cart"), Priority: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для 357 грамм webSite: ", err)
	}
	cartWebPage := _cartWebPage.(*models.WebPage)

	_, err = cartWebPage.CreateChild(
		models.WebPage{
			AccountId: account.Id, Code:  utils.STRp("cart"), Label:  utils.STRp("Благодарим за заказ"), Path:  utils.STRp("checkout"),
			MetaTitle: utils.STRp("Благодарим за заказ :: 357 грамм"),
			IconName:  utils.STRp("far fa-cart-arrow-down"), RouteName:  utils.STRp("cart"), Priority: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать checkout для 357 грамм webSite: ", err)
	}



	// ################# Категории товаров #################

	
	// 5* Создаем категории товаров

	// псевдо категория
	_CategoryRoot, err := account.CreateEntity(&models.ProductCategory{
		Code:  utils.STRp("catalog"), Label: utils.STRp("Весь каталог"),
	})
	CategoryRoot := _CategoryRoot.(*models.ProductCategory)

	// категория: Новинки
	_, err = CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("news"), Label: utils.STRp("Новинки"),
	})
	if err != nil {
		log.Fatal("Не удалось создать news для 357gr category: ", err)
		return
	}

	// категория: участие в акции
	_, err = CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("promotions"), Label: utils.STRp("Акции"),
	})
	if err != nil {
		log.Fatal("Не удалось создать promotions для 357gr category: ", err)
		return
	}

	// категория: весь чай
	_catTea, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("tea"), Label: utils.STRp("Чай"),
	})
	catTea := _catTea.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// категория: красный чай
	_, err = catTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("red-tea"), Label: utils.STRp("Красный чай"),
	})
	// catRedTea := _catRedTea.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// категория: зеленый чай
	_catGreenTea, err := catTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Зеленый чай"),
	})
	catGreenTea := _catGreenTea.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// подкатегория: китайский зеленый чай
	_, err = catGreenTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Китайский зеленый чай"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	// подкатегория: жасминовый зеленый чай
	_, err = catGreenTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Жасминовый чай"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	// подкатегория: связанный зеленый чай
	_, err = catGreenTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Связанный чай"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// категория: улунский чай
	_catUlunTea, err := catTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("oolong-tea"), Label: utils.STRp("Улунский чай"),
	})
	catUlunTea := _catUlunTea.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// подкатегория: Южнофуцзянские улуны
	_, err = catUlunTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Южнофуцзянский улун"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	// подкатегория: Уишаньские улуны
	_, err = catUlunTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Уишаньский улун"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	// подкатегория: Тайваньские улуны
	_, err = catUlunTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Тайваньский улун"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	// подкатегория: Гуандунские улуны
	_, err = catUlunTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-tea"), Label: utils.STRp("Гуандунский улун"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// категория: Пуэр
	_catPuer, err := catTea.CreateChild(models.ProductCategory{
		Code:  utils.STRp("puer-tea"), Label: utils.STRp("Пуэр"),
	})
	catPuer := _catPuer.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// подкатегория: Шу Пуэр
	_catShuPuer, err := catPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shu-puer"), Label: utils.STRp("Шу Пуэр"),
	})
	catShuPuer := _catShuPuer.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// под подкатегория: Шу Пуэр пресованный
	_, err = catShuPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shu-puer-pressed"), Label: utils.STRp("Шу Пуэр прессованный"),
	})
	// под подкатегория: Шу Пуэр рассыпной
	_, err = catShuPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shu-puer-loose"), Label: utils.STRp("Шу Пуэр рассыпной"),
	})

	// подкатегория: Шэн Пуэр
	_catShenPuer, err := catPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shen-puer"), Label: utils.STRp("Шэн Пуэр"),
	})
	catShenPuer := _catShenPuer.(*models.ProductCategory)
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}

	// подкатегория: Шэн Пуэр пресованный
	_, err = catShenPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shen-puer-pressed"), Label: utils.STRp("Шэн Пуэр прессованный"),
	})
	// подкатегория: Шэн Пуэр рассыпной
	_, err = catShenPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("shen-puer-loose"), Label: utils.STRp("Шэн Пуэр рассыпной"),
	})

	// подкатегория: Белый Пуэр
	_, err = catPuer.CreateChild(models.ProductCategory{
		Code:  utils.STRp("white-puer"), Label: utils.STRp("Белый Пуэр"),
	})
	if err != nil {
		log.Fatal("Не удалось создать catTea для 357 catalog: ", err)
		return
	}
	
	// категория: весь кофе
	_catCoffee, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("coffee"), Label: utils.STRp("Кофе"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catCoffee для 357gr category: ", err)
		return
	}
	catCoffee := _catCoffee.(*models.ProductCategory)

	// подкатегория: Зеленый кофе
	_, err = catCoffee.CreateChild(models.ProductCategory{
		Code:  utils.STRp("green-coffee"), Label: utils.STRp("Зеленый кофе"),
	})

	// подкатегория: Кофе в зернах
	_, err = catCoffee.CreateChild(models.ProductCategory{
		Code:  utils.STRp("beans-coffee"), Label: utils.STRp("Кофе в зернах"),
	})


	// категория: Подарки
	_catGifts, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("gifts"), Label: utils.STRp("Подарки"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catGifts для 357gr category: ", err)
		return
	}
	catGifts := _catGifts.(*models.ProductCategory)
	// подкатегория: Чайные сувениры
	_, err = catGifts.CreateChild(models.ProductCategory{
		Code:  utils.STRp("gifts"), Label: utils.STRp("Чайные сувениры"),
	})
	_, err = catGifts.CreateChild(models.ProductCategory{
		Code:  utils.STRp("gifts"), Label: utils.STRp("Подарочные корзины"),
	})
	_, err = catGifts.CreateChild(models.ProductCategory{
		Code:  utils.STRp("gifts"), Label: utils.STRp("Подарочные букеты"),
	})
	_, err = catGifts.CreateChild(models.ProductCategory{
		Code:  utils.STRp("gifts"), Label: utils.STRp("Подарки руководителю"),
	})

	// категория: Посуда и аксессуары
	_catAccessories, err := CategoryRoot.CreateChild(models.ProductCategory{
		Code:  utils.STRp("accessories"), Label: utils.STRp("Посуда и аксессуары"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catAccessories для 357gr category: ", err)
		return
	}
	catAccessories := _catAccessories.(*models.ProductCategory)

	
	// под подкатегория: Для заваривания
	_catBrewing, err := catAccessories.CreateChild(models.ProductCategory{
		Code:  utils.STRp("accessories.brewing"), Label: utils.STRp("Для заваривания"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catBrewing для 357gr category: ", err)
		return
	}
	catBrewing := _catBrewing.(*models.ProductCategory)
	
	// под подкатегория: Для чаепития
	_catForTea, err := catAccessories.CreateChild( models.ProductCategory{
		Code:  utils.STRp("accessories.for-tea"), Label: utils.STRp("Для чаепития"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catForTea для 357gr category: ", err)
		return
	}
	catForTea := _catForTea.(*models.ProductCategory)

	// под подкатегория: Для хранения
	_catTeaStorage, err := catAccessories.CreateChild( models.ProductCategory{
		Code:  utils.STRp("accessories.storage"), Label: utils.STRp("Для хранения чая"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catBrewing для 357gr category: ", err)
		return
	}
	catTeaStorage := _catTeaStorage.(*models.ProductCategory)

	// под подкатегория: Разное
	_catOthers, err := catAccessories.CreateChild( models.ProductCategory{
		Code:  utils.STRp("accessories.others"), Label: utils.STRp("Разное"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catOthers для 357gr category: ", err)
		return
	}
	catOthers := _catOthers.(*models.ProductCategory)

	// под под под категория посуда и акксесуары

	// для заваривания - Традиционная посуда
	_, err = catBrewing.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.traditional"), Label: utils.STRp("Традиционная посуда"),LabelPlural: utils.STRp("Традиционная посуда"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для заваривания - Европейская посуда
	_, err = catBrewing.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.european"), Label: utils.STRp("Европейская посуда"),LabelPlural: utils.STRp("Европейская посуда"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для заваривания - Типоды
	_, err = catBrewing.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.gunfu"), Label: utils.STRp("Типод"),LabelPlural: utils.STRp("Типоды"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для заваривания - Сифоны
	_, err = catBrewing.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.siphons"), Label: utils.STRp("Сифон"),LabelPlural: utils.STRp("Сифоны"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}

	
	// Для чаепития - Для чайной церемонии
	_, err = catForTea.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.traditional"), Label: utils.STRp("Для чайной церемонии"), LabelPlural: utils.STRp("Для чайной церемонии"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Для чаепития - Сервизы
	_, err = catForTea.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.european"), Label: utils.STRp("чайный сервиз"), LabelPlural: utils.STRp("Сервизы"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Для чаепития - Чашки
	_, err = catForTea.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.gunfu"), Label: utils.STRp("Чашка"), LabelPlural: utils.STRp("Чашки"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Для чаепития - Пиалы
	_, err = catForTea.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.tableware.siphons"), Label: utils.STRp("Пиала"), LabelPlural: utils.STRp("Пиалы"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}


	
	// для хранения - Банки металлические
	_, err = catTeaStorage.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.storage.metal"), Label: utils.STRp("Банка металлическая"), LabelPlural: utils.STRp("Банки металлические"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для хранения - Банки картонные
	_, err = catTeaStorage.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.storage.cartons"), Label: utils.STRp("Банка картонная"), LabelPlural: utils.STRp("Банки картонные"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для хранения - Упаковка для пуэров
	_, err = catTeaStorage.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.storage.caps"), Label: utils.STRp("Упаковка для пуэра"), LabelPlural: utils.STRp("Упаковки для пуэров"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// для хранения - Чайницы
	_, err = catTeaStorage.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.storage.pialas"), Label: utils.STRp("Чайница"), LabelPlural: utils.STRp("Чайницы"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}


	// Разное - Благовония
	_, err = catOthers.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.for-tea.ceremonies"), Label: utils.STRp("Благовонье"),LabelPlural: utils.STRp("Благовония"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Разное - Чайные фигурки
	_, err = catOthers.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.for-tea.services"), Label: utils.STRp("Чайная фигурка"),LabelPlural: utils.STRp("Чайные фигурки"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Разное - Электрические плитки
	_, err = catOthers.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.for-tea.caps"), Label: utils.STRp("Электрическая плитка"),LabelPlural: utils.STRp("Электрические плитки"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Разное - Чистящие средства
	_, err = catOthers.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.for-tea.pialas"), Label: utils.STRp("Чистящее средство"),LabelPlural: utils.STRp("Чистящие средства"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}
	// Разное - Чайный инструмент
	_, err = catOthers.CreateChild( models.ProductCategory {
		Code:  utils.STRp("accessories.for-tea.pialas"), Label: utils.STRp("Чайный инструмент"),LabelPlural: utils.STRp("Чайный инструмент"),
	})
	if err != nil {
		log.Fatal("Не удалось создать _catTableware для 357gr category: ", err)
		return
	}

	// ...

	fmt.Println("Загрузка данных закончена")
	
	return
}

func Migrate_I() {
	// pool := models.GetPool()
	
	// pool.Raw("")
	// pool.Exec("alter table email_campaigns alter column email_template_id drop not null;\nalter table email_campaigns alter column email_box_id drop not null;\nalter table email_campaigns alter column users_segment_id drop not null;\n\nALTER TABLE email_campaigns\n    ADD CONSTRAINT email_campaigns_email_template_id_fkey FOREIGN KEY (email_template_id) REFERENCES email_templates(id) ON DELETE SET NULL ON UPDATE CASCADE,\n    ADD CONSTRAINT email_campaigns_email_box_id_fkey FOREIGN KEY (email_box_id) REFERENCES email_boxes(id) ON DELETE SET NULL ON UPDATE CASCADE,\n    ADD CONSTRAINT email_campaigns_users_segment_id_fkey FOREIGN KEY (users_segment_id) REFERENCES users_segments(id) ON DELETE SET NULL ON UPDATE CASCADE;")
	/*err := pool.Exec("drop table if exists user_segments_user_segment_conditions").Error
	if err != nil {
		log.Fatal(err)
	}

	pool.DropTableIfExists(
		models.UsersSegment{},
		models.UserSegmentConditions{},
		models.EmailCampaign{},
		models.TaskScheduler{},
	)

	models.UsersSegment{}.PgSqlCreate()
	models.UserSegmentConditions{}.PgSqlCreate()
	models.EmailCampaign{}.PgSqlCreate()
	models.TaskScheduler{}.PgSqlCreate()

	account,err := models.GetAccount(4)
	if err != nil { log.Fatal(err)}

	userSegment := []models.UsersSegment {
		{AccountId: account.Id, Name: "Все подписчики", StrictMatching: true},
		{AccountId: account.Id, Name: "Все клиенты", StrictMatching: true},
	}

	for i := range userSegment {
		_, _ = account.CreateEntity(&userSegment[i])
	}

	emailCampaigns := []models.EmailCampaign{
		{Name: "Оповещение пользователей", Enabled: true, ScheduleRun: time.Now().UTC(), Subject: "Установите наш модуль!", PreviewText: "", EmailTemplateId: 7, EmailBoxId: 4,
			UsersSegmentId: 1},
	}

	for i := range emailCampaigns {
		_, _ = account.CreateEntity(&emailCampaigns[i])
	}*/


	return

}