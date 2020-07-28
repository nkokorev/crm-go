package base

import (
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/models"
	"github.com/nkokorev/crm-go/utils"
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

	pool.DropTableIfExists(models.DeliveryPickup{},models.DeliveryRussianPost{}, models.DeliveryCourier{})

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

func RefreshTablesPart_I() {

	pool := models.GetPool()

	// not there
	err := pool.Exec("drop table if exists orders_products, payment_methods_web_sites, payment_methods_payments").Error
	if err != nil {
		log.Fatal(err)
	}


	pool.DropTableIfExists(models.PaymentOption{}, models.DeliveryOrder{}, models.OrderChannel{}, models.Order{}, models.Payment{},
	models.PaymentAmount{})
	// pool.DropTableIfExists(models.PaymentMethod{}, models.DeliveryOrder{}, models.OrderChannel{}, models.PaymentSubject{},models.VatCode{},models.Order{}, models.Payment{},models.YandexPayment{},models.PaymentAmount{})


	pool.DropTableIfExists(models.Product{}, models.ProductCard{}, models.ProductGroup{})
	pool.DropTableIfExists(models.EmailNotification{}, models.EmailBox{}, models.WebSite{},models.EmailTemplate{})
	pool.DropTableIfExists(models.WebHook{}, models.EventListener{}, models.EventItem{},models.HandlerItem{}, models.DeliveryPickup{}, models.DeliveryRussianPost{}, models.DeliveryCourier{})

	pool.DropTableIfExists(models.Article{}, models.Storage{}, models.UnitMeasurement{}, models.ApiKey{}, models.AccountUser{}, models.User{})
	pool.DropTableIfExists(models.VatCode{}, models.PaymentSubject{})
	pool.DropTableIfExists(models.Role{}, models.UserVerificationMethod{}, models.Account{}, models.CrmSetting{})

	
	models.CrmSetting{}.PgSqlCreate()

	models.UserVerificationMethod{}.PgSqlCreate()
	models.Account{}.PgSqlCreate()
	models.Role{}.PgSqlCreate()
	models.User{}.PgSqlCreate()
	models.AccountUser{}.PgSqlCreate()
	models.ApiKey{}.PgSqlCreate()

	models.UnitMeasurement{}.PgSqlCreate()

	models.Storage{}.PgSqlCreate()
	models.Article{}.PgSqlCreate()

	/////////////////////

	models.WebSite{}.PgSqlCreate()
	models.EmailBox{}.PgSqlCreate()
	models.EmailTemplate{}.PgSqlCreate()

	models.ProductGroup{}.PgSqlCreate()
	models.ProductCard{}.PgSqlCreate()
	models.Product{}.PgSqlCreate()


	models.HandlerItem{}.PgSqlCreate()
	models.EventItem{}.PgSqlCreate()
	models.EventListener{}.PgSqlCreate()

	models.WebHook{}.PgSqlCreate()

	// Уведомления
	models.EmailNotification{}.PgSqlCreate()


	models.PaymentAmount{}.PgSqlCreate()
	models.Payment{}.PgSqlCreate()
	models.Order{}.PgSqlCreate()
	models.DeliveryOrder{}.PgSqlCreate()
	models.VatCode{}.PgSqlCreate()

	models.OrderChannel{}.PgSqlCreate()
	models.PaymentOption{}.PgSqlCreate()
	models.PaymentSubject{}.PgSqlCreate()

	models.DeliveryRussianPost{}.PgSqlCreate()
	models.DeliveryPickup{}.PgSqlCreate()
	models.DeliveryCourier{}.PgSqlCreate()
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
	roleClientMain, err := mAcc.GetRoleByTag(models.RoleClient)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	// 2. Создаем пользователя admin в main аккаунте
	timeNow := time.Now().UTC()
	owner, err := mAcc.CreateUser(
			models.User{
			Username:"admin",
			Email:"kokorevn@gmail.com",
			PhoneRegion: "RU",
			Phone: "89251952295",
			Password:"qwerty109#QW",
			Name:"Никита",
			Surname:"Кокорев",
			Patronymic:"Романович",
			EmailVerifiedAt:&timeNow,
			},
		*roleOwnerMain,
		)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать admin'a: ", err)
	}

	mex388, err := mAcc.CreateUser(
		models.User{
			Username:"mex388",
			Email:"nkokorev@rus-marketing.ru",
			PhoneRegion: "RU",
			Phone: "79251952222",
			Password:"qwerty109#QW",
			Name:"Никита",
			Surname:"Кокорев",
			Patronymic:"Романович",
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

	////////////////////////////////////

	// ######### 357 Грамм ############

	// 1. Создаем Василия (^_^)
	vpopov, err := mAcc.CreateUser(
		models.User{
			Username:"antiglot",
			// Email:"vp@357gr.ru",
			Email:"mail-test@ratus-dev.ru",
			PhoneRegion: "RU",
			Phone: "89055294696",
			Password:"qwerty109#QW",
			Name:"Василий",
			Surname:"Попов",
			Patronymic:"Николаевич",
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать admin'a: ", err)
	}
	
	dvc, err := models.GetUserVerificationTypeByCode(models.VerificationMethodEmailAndPhone)
	if err != nil || dvc == nil {
		log.Fatal("Не удалось получить верификацию...")
		return
	}
	
	// 2. создаем из-под Василия 357gr
	acc357, err := vpopov.CreateAccount( models.Account{
		Name:                                "357 грамм",
		Website:                             "https://357gr.ru/",
		Type:                                "store",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		UiApiAesEnabled:                     true,
		UiApiAuthMethods:                    pq.StringArray{"email, phone"}, // !!
		UiApiEnabledUserRegistration:        true,
		UiApiUserRegistrationInvitationOnly: false,
		UiApiUserRegistrationRequiredFields: pq.StringArray{"email, phone, name"}, // !! хз хз
		UiApiUserEmailDeepValidation:        true, // хз
		UserVerificationMethodId:            dvc.Id,
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

	// 3.2. добавляем кучу других клиентов
	if true {
		var clients []models.User

		for i:=0; i < 1 ;i++ {
			clients = append(clients, models.User{
				Name: fmt.Sprintf("Name #%d", i),
				Email: fmt.Sprintf("email%d@mail.ru", i),
				Phone: fmt.Sprintf("+7925195221%d", i),
				Password: "asdfg109#QW",
			})
		}
		for i,_ := range clients {
			_, err := acc357.CreateUser(clients[i], *roleClientMain)
			if err != nil {
				fmt.Println(err)
				log.Fatal("Не удалось добавить клиента id: ", i)
				return
			}
		}
	}


	// 4. Создаем домен для 357gr
	_webSite357, err := acc357.CreateEntity(&models.WebSite{
		Hostname: "357gr.ru",
		DKIMPublicRSAKey: ``,
		DKIMPrivateRSAKey: ``,
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
			Username:"ikomastas",
			Email:"sa-tolstov@yandex.ru",
			// Email:"info@rus-marketing.ru",
			PhoneRegion: "RU",
			Phone: "",
			Password:"qwerty123#Q",
			Name:"Станислав",
			Surname:"Толстов",
			Patronymic:"",
			EmailVerifiedAt:&timeNow,
		},
		*roleClientMain,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать admin'a: ", err)
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
		Name:                                "Brouser",
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
			Username:"korotaev",
			// Email:"sa-tolstov@yandex.ru",
			Email:"mailtest@ratus-dev.ru",
			PhoneRegion: "RU",
			Phone: "",
			Password:"qwerty109#QW",
			Name:"Максим",
			Surname:"Коротаев",
			Patronymic:"Валерьевич",
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
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2. Добавляем mex388 как админа
	_, err = airoClimat.AppendUser(*mex388, *roleAdminMain)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in brouser")
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
			Name: "Сайт по продаже бактерицидных рециркуляторов", Address: "г. Москва, р-н Текстильщики", Email: "info@airoclimate.ru", Phone: "+7 (4832) 77-03-73",
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
	_, err = webSiteAiro.CreateEmailBox(models.EmailBox{Default: false, Allowed: true, Name: "Отдел продаж AIRO Climate", Box: "sale"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}
	_, err = webSiteAiro.CreateEmailBox(models.EmailBox{Default: false, Allowed: true, Name: "Служба поддержки AIRO Climate", Box: "help"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для главного аккаунта: ", err)
	}
	
	groupAiroRoot, err := webSiteAiro.CreateProductGroup(
		models.ProductGroup{
			Code: "root", Name: "", URL: "/", IconName: "far fa-home", RouteName: "info.index",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}

	groupAiroCatalogRoot, err := groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog", Name: "Весь каталог", URL: "catalog", IconName: "far fa-th-large", RouteName: "catalog",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	groupAiro1, err := groupAiroCatalogRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog",Name: "Бактерицидные рециркуляторы", URL: "bactericidal-recirculators", IconName: "far fa-fan-table", RouteName: "catalog.recirculators",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}
	groupAiro2, err := groupAiroCatalogRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog",Name: "Бактерицидные камеры", URL: "bactericidal-chambers", IconName: "far fa-box-full", RouteName: "catalog.chambers",
			})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
		return
	}

	//////////////

	_, err = groupAiroRoot.CreateChild(

		models.ProductGroup{
			Code: "info", Name: "Статьи", URL: "articles", IconName: "far fa-books", RouteName: "info.articles", Order: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	deliveryGroupRoute, err := groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "delivery", Name: "Доставка товара", URL: "delivery", IconName: "far fa-shipping-fast", RouteName: "delivery", Order: 1,
		})
	_, err = deliveryGroupRoute.CreateChild(
		models.ProductGroup{
			Code: "delivery", Name: "Способы оплаты", URL: "payment", IconName: "far fa-hand-holding-usd", RouteName: "delivery.payment", Order: 2,
		})
	_, err = deliveryGroupRoute.CreateChild(
		models.ProductGroup{
			Code: "delivery", Name: "Возврат товара", URL: "moneyback", IconName: "far fa-exchange-alt", RouteName: "delivery.moneyback", Order: 3,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "О компании", URL: "about", IconName: "far fa-home-heart", RouteName: "info.about",      Order: 5,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "Политика конфиденциальности", URL: "privacy-policy", IconName: "far fa-home-heart", RouteName: "info.privacy-policy",      Order: 6,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "Контакты", URL: "contacts", IconName: "far fa-address-book", RouteName: "info.contacts",  Order: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}

	////////
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "cart", Name: "Корзина", URL: "cart", IconName: "far fa-cart-arrow-down", RouteName: "cart", Order: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat webSite: ", err)
	}

	// 6. Создаем карточки товара
	cards := []models.ProductCard{
		{Id: 0, URL: "airo-dez-adjustable-black", 	Label:"Рециркулятор AIRO-DEZ черный с регулировкой", Breadcrumb: "Рециркулятор AIRO-DEZ черный с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ черный с регулировкой",
			SwitchProducts: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			}},
		{Id: 0, URL: "airo-dez-black", 				Label:"Рециркулятор AIRO-DEZ черный", Breadcrumb: "Рециркулятор AIRO-DEZ черный", MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ черный"},
		{Id: 0, URL: "airo-dez-adjustable-white", 	Label:"Рециркулятор AIRO-DEZ белый с регулировкой", Breadcrumb: "Рециркулятор AIRO-DEZ белый с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ белый с регулировкой"},
		{Id: 0, URL: "airo-dez-white", 				Label:"Рециркулятор AIRO-DEZ белый", Breadcrumb: "Рециркулятор AIRO-DEZ белый",MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ белый"},
		{Id: 0, URL: "airo-dez-compact", 			Label: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", Breadcrumb: "Мобильный аиродезинфектор AIRO-DEZ COMPACT",MetaTitle: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", },

		{Id: 0, URL: "airo-dezpuf",			Label: "Бактерицидная камера пуф AIRO-DEZPUF", Breadcrumb: "Бактерицидная камера пуф AIRO-DEZPUF",MetaTitle: "Бактерицидная камера пуф AIRO-DEZPUF"},
		{Id: 0, URL: "airo-dezpuf-wenge", 	Label: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге", Breadcrumb: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге",MetaTitle: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"},
		
		{Id: 0, URL: "airo-dezbox", 		Label: "Бактерицидная камера AIRO-DEZBOX", Breadcrumb: "Бактерицидная камера AIRO-DEZBOX",MetaTitle: "Бактерицидная камера AIRO-DEZBOX", },
		{Id: 0, URL: "airo-dezbox-white", 	Label: "Бактерицидная камера AIRO-DEZBOX белая",Breadcrumb: "Бактерицидная камера AIRO-DEZBOX белая", MetaTitle: "Бактерицидная камера AIRO-DEZBOX белая", },
		{Id: 0, URL: "airo-deztumb", 		Label: "Тумба облучатель бактерицидный AIRO-DEZTUMB", Breadcrumb: "Тумба облучатель бактерицидный AIRO-DEZTUMB",MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB", },
		{Id: 0, URL: "airo-deztumb-big", 	Label: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", Breadcrumb: "Тумба облучатель бактерицидный AIRO-DEZTUMB big",MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", },

		{Id: 0, URL: "airo-deztumb-pine", 	Label: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина",Breadcrumb: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", MetaTitle: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", },

	}

	//metadata := json.RawMessage(`{"color": "white", "bodyMaterial": "металл", "filterType": "угольно-фотокаталитический"}`)

	// 7. Создаем список товаров
	products := []models.Product{
		{
			Model: ToStringPointer("AIRO-DEZ с регулировкой черный"), //,
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности черный", ShortName: "Рециркулятор AIRO-DEZ черный",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 19500.00, RetailDiscount: 1000,
			ShortDescription: "",Description: "",
			WeightKey: "grossWeight",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZ черный"),
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ черный", ShortName: "Рециркулятор AIRO-DEZ черный",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 17500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZ с регулировкой белый"),
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности белый",  ShortName: "Рециркулятор AIRO-DEZ белый",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 19500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZ белый"),
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ",  ShortName: "Рециркулятор AIRO-DEZ белый",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 17500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZ COMPACT"),
			Name:"Мобильный аиродезинфектор AIRO-DEZ COMPACT",  ShortName: "Аиродезинфектор AIRO-DEZ COMPACT",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 39000.00, RetailDiscount: 3000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		
		{
			Model: ToStringPointer("AIRO-DEZPUF"),
			Name:"Бактерицидная камера пуф AIRO-DEZPUF",  ShortName: "Камера пуф AIRO-DEZPUF",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 11000.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZPUF венге"),
			Name:"Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге",  ShortName: "Камера AIRO-DEZBOX",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 12000.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},

		{
			Model: ToStringPointer("AIRO-DEZBOX"),
			Name:"Бактерицидная камера AIRO-DEZBOX",  ShortName: "Камера AIRO-DEZBOX",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 7800.00, RetailDiscount: 800,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZBOX белая"),
			Name:"Бактерицидная камера AIRO-DEZBOX белая",  ShortName: "Камера AIRO-DEZBOX белая",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 7800.00, RetailDiscount: 800,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIRO-DEZTUMB"),
			Name:"Тумба облучатель бактерицидный AIRO-DEZTUMB",  ShortName: "Бактерицидная тумба AIRO-DEZTUMB",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		{
			Model: ToStringPointer("AIROTUMB big"),
			Name:"Тумба облучатель бактерицидный AIRO-DEZTUMB big",  ShortName: "Облучатель AIROTUMB big",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
		
		{
			Model: ToStringPointer("AIRO-DEZTUMB касцина"),
			Name:"Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина",  ShortName: "Бактерицидная тумба AIRO-DEZTUMB",
			PaymentSubjectId: 1, UnitMeasurementId: 1, VatCodeId: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: utils.MapToRawJson(map[string]interface{}{
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
				}),
			},
		},
	}
	
	
	// 7. Добавляем продукты в категории с созданием карточки товара
	for i,_ := range products {
		var group *models.ProductGroup
		if i < 4 {
			group = groupAiro1
		} else {
			group = groupAiro2
		}
		_, err = webSiteAiro.CreateProductWithCardAndGroup(products[i], cards[i], &group.Id)
		if err != nil {
			log.Fatal("Не удалось создать Product для airoClimat: ", err)
		}
	}


	return

}

func UploadTestDataPart_II() {

	// 1. Получаем AiroClimate аккаунт
	accountAiro, err := models.GetAccount(5)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	// 2. Получаем магазин
	var webSite models.WebSite
	err = accountAiro.LoadEntity(&webSite, 5)
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
	airoAccount, err := models.GetAccount(5)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	mainAccount, err := models.GetMainAccount()
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}


	// HandlerItem
	eventHandlers := []models.HandlerItem {
		{Name:"Вызов WebHook'а", Code: "WebHookCall", EntityType: "web_hooks", Enabled: true, Description: "Вызов указанного WebHook'а"},
		{Name:"Запуск email-уведомления", Code: "EmailNotificationRun", EntityType: "email_notification", Enabled: true, Description: "Отправка электронного письма. Адресат выбирается в зависимости от настроек уведомления и события. Если объект пользователь - то на его email. При отсутствии email'а, запуск уведомления не произойдет."},
		{Name:"Запуск email-серии", Code: "EmailQueueRun", EntityType: "email_queue", Enabled: true, Description: "Запуск автоматической серии писем. Адресат выбирается исходя из события. Если объект пользователь - то на его email. При отсутствии email'а, запуск серии не произойдет."},
	}
	for _,v := range eventHandlers {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Events
	eventItems := []models.EventItem{
		{Name: "Пользователь создан", 	Code: "UserCreated", Enabled: true, Description: "Создание пользователя в текущем аккаунте. Сам пользователь на момент вызова не имеет доступа к аккаунту (если вообще будет)."},
		{Name: "Пользователь обновлен", Code: "UserUpdated", Enabled: true, Description: "Какие-то данные в учетной записи пользователя обновились."},
		{Name: "Пользователь удален", 	Code: "UserDeleted", Enabled: true, Description: "Учетная запись пользователя удалена из системы RatusCRM."},

		{Name: "Пользователь добавлен в аккаунт", Code: "UserAppendedToAccount", Enabled: true, Description: "Пользователь получил доступ в текущий аккаунт с какой-то конкретно ролью."},
		{Name: "Пользователь удален из аккаунта", Code: "UserRemovedFromAccount", Enabled: true, Description: "У пользователя больше нет доступа к вашей системе из-под своей учетной записи."},

		{Name: "Товар создан", 		Code: "ProductCreated", Enabled: true, Description: "Создан новый товар или услуга."},
		{Name: "Товар обновлен", 	Code: "ProductUpdated", Enabled: true, Description: "Данные товара или услуга были обновлены. Сюда также входит обновление связанных данных: изображений, описаний, видео."},
		{Name: "Товар удален", 		Code: "ProductDeleted", Enabled: true, Description: "Товар или услуга удалены из системы со всеми связанными данными."},

		{Name: "Карточка товара создана", 	Code: "ProductCardCreated", Enabled: true, Description: "Карточка товара создана в системе"},
		{Name: "Карточка товара обновлена", Code: "ProductCardUpdated", Enabled: true, Description: "Данные карточки товара успешно обновлены."},
		{Name: "Карточка товара удалена", 	Code: "ProductCardDeleted", Enabled: true, Description: "Карточка товара удалена из системы"},

		{Name: "Раздел сайта создан", 	Code: "ProductGroupCreated", Enabled: true, Description: "Создан новый раздел, категория или страница на сайте."},
		{Name: "Раздел сайта обновлен", Code: "ProductGroupUpdated", Enabled: true, Description: "Данные раздела или категории сайта успешно обновлены."},
		{Name: "Раздел сайта удален", 	Code: "ProductGroupDeleted", Enabled: true, Description: "Раздел сайта или категория удалена из системы"},

		{Name: "Сайт создан", 	Code: "WebSiteCreated", Enabled: true, Description: "Создан новый сайт или магазин."},
		{Name: "Сайт обновлен", Code: "WebSiteUpdated", Enabled: true, Description: "Персональные данные сайта или магазина были успешно обновлены."},
		{Name: "Сайт удален", 	Code: "WebSiteDeleted", Enabled: true, Description: "Сайт или магазин удален из системы."},
		
		{Name: "Файл создан", 	Code: "StorageCreated", Enabled: true, Description: "В системе создан новый файл."},
		{Name: "Файл обновлен", Code: "StorageUpdated", Enabled: true, Description: "Какие-то данные файла успешно изменены."},
		{Name: "Файл удален", 	Code: "StorageDeleted", Enabled: true, Description: "Файл удален из системы."},

		{Name: "Статья создана", 	Code: "ArticleCreated", Enabled: true, Description: "В системе создана новая статья."},
		{Name: "Статья обновлена", 	Code: "ArticleUpdated", Enabled: true, Description: "Какие-то данные статьи были изменены. Учитываются также и смежные данные, вроде изображений и видео."},
		{Name: "Статья удалена", 	Code: "ArticleDeleted", Enabled: true, Description: "Статья со смежными данными удалена из системы."},
	}
	for _,v := range eventItems {
		_, err = mainAccount.CreateEntity(&v)
		if err != nil {
			log.Fatal(err)
		}
	}

	els := []models.EventListener {
		// товар
		{Name: "Добавление товара на сайт", EventId: 6, HandlerId: 2, EntityId: 3, Enabled: true},
		{Name: "Обновление товара на сайте", EventId: 7, HandlerId: 2, EntityId: 3, Enabled: true},
		{Name: "Обновление товара на сайте", EventId: 8, HandlerId: 2, EntityId: 6, Enabled: true},

		// Карточки товара
		{Name: "Обновление карточек товара", EventId: 9, HandlerId: 2, EntityId: 8, Enabled: true},
		{Name: "Обновление карточек товара", EventId: 10, HandlerId: 2, EntityId: 9, Enabled: true},
		{Name: "Обновление карточек товара", EventId: 11, HandlerId: 2, EntityId: 10, Enabled: true},

		// Магазин (WebSite)
		{Name: "Обновление данных магазина", EventId: 16, HandlerId: 2, EntityId: 2, Enabled: true},
		{Name: "Обновление данных магазина", EventId: 17, HandlerId: 2, EntityId: 3, Enabled: true},

		// Статьи
		{Name: "Обновление статей на сайте", EventId: 21, HandlerId: 2, EntityId: 16, Enabled: true},
		{Name: "Обновление статей на сайте", EventId: 22, HandlerId: 2, EntityId: 17, Enabled: true},
		{Name: "Обновление статей на сайте", EventId: 23, HandlerId: 2, EntityId: 18, Enabled: true},

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

		{Name: "Upload all webSite data", Code: models.EventUpdateAllShopData, URL: domainAiroSite + "/ratuscrm/webhooks/upload/all", HttpMethod: http.MethodGet},

		// WebSite
		{Name: "Update webSite", Code: models.EventShopUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/web-sites", HttpMethod: http.MethodPatch},
		{Name: "Delete webSite", Code: models.EventShopDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/web-sites", HttpMethod: http.MethodDelete},

		// Product
		{Name: "Create product", Code: models.EventProductCreated, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodPost},
		{Name: "Update product", Code: models.EventProductUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete product", Code: models.EventProductDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.productId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all products", Code: models.EventProductsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/products", HttpMethod: http.MethodGet},

		// ProductCard
		{Name: "Create product card", Code: models.EventProductCardCreated, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodPost},
		{Name: "Update product card", Code: models.EventProductCardUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete product card", Code: models.EventProductCardDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.productCardId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all product cards", Code: models.EventProductCardsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards", HttpMethod: http.MethodGet},

		// Groups
		{Name: "Create product group", Code: models.EventProductGroupCreated, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.productGroupId}}", HttpMethod: http.MethodPost},
		{Name: "Update product group", Code: models.EventProductGroupUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.productGroupId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete product group", Code: models.EventProductGroupDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.productGroupId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all product groups", Code: models.EventProductGroupsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups", HttpMethod: http.MethodGet},

		// Articles
		{Name: "Create article", Code: models.EventArticleCreated, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodPost},
		{Name: "Update article", Code: models.EventArticleUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodPatch},
		{Name: "Delete article", Code: models.EventArticleDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.articleId}}", HttpMethod: http.MethodDelete},
		{Name: "Upload all articles", Code: models.EventArticlesUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/articles", HttpMethod: http.MethodGet},

	}
	for i,_ := range webHooks {
		// _, err = airoAccount.CreateWebHook(webHooks[i])
		_, err = airoAccount.CreateEntity(&webHooks[i])
		if err != nil {
			log.Fatal("Не удалось создать webHook: ", err)
		}

	}

	// Добавляем шаблоны писем для синдиката и главного аккаунта
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
			Enabled: false, Delay: 0, Name:"Новый заказ", Description: "Оповещение менеджеров о новом заказе", EmailTemplateId: &numOne, SendingToFixedAddresses: true,
			RecipientList: postgres.Jsonb{RawMessage: utils.StringArrToRawJson([]string{"nkokorev@rus-marketing.ru"})},
			RecipientUsersList: postgres.Jsonb{RawMessage: utils.UINTArrToRawJson([]uint{2,6,7})},
			EmailBoxId: &num5,

		},
		{
			Enabled: false, Delay: 0, Name:"Ваш заказ получен!", Description: "Информирование клиента о принятом заказе", EmailTemplateId: &numOne, SendingToFixedAddresses: true,
			RecipientList: postgres.Jsonb{RawMessage: utils.StringArrToRawJson([]string{"mex388@gmail.com"})},
			RecipientUsersList: postgres.Jsonb{RawMessage: utils.UINTArrToRawJson([]uint{7})},
			EmailBoxId: &num6,
		},
		{
			Enabled: false, Delay: 0, Name:"*Ваш заказ отправлен по почте", Description: "Информирование клиента о принятом заказе", EmailTemplateId: &numOne, SendingToFixedAddresses: true,
			RecipientList: postgres.Jsonb{RawMessage: utils.StringArrToRawJson([]string{"nkokorev@rus-marketing.ru"})},
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

	account, err := models.GetAccount(5)
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

			mimeType, err := GetFileContentType(f)
			if err != nil {
				log.Fatalf("unable to mimeType file: %v", err)
			}

			fs := models.Storage{
				Name: strings.ToLower(file.Name()),
				Data: body,
				MIME: mimeType,
				Size: uint(file.Size()),
				Priority: 0,
			}
			// file, err := account.StorageCreateFile(&fs)
			file, err := account.CreateEntity(&fs)
			if err != nil {
				log.Fatalf("unable to create file: %v", err)
			}

			err = (models.Product{Id: uint(index)}).AppendAssociationImage(file)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
		}
	}

	fmt.Println("Данные загружены!")
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
	account, err := models.GetAccount(5)
	if err != nil {
		fmt.Println("Не удалось найти аккаунт для загрузки статей", err)
	}

	url := "/var/www/ratuscrm/files/airoclimate/articles/"
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

		articleNew := models.Article{
			Name: strings.ToLower(file.Name()),
			Public: true,
			Shared: true,
			Body: string(body),
		}
		article, err := account.CreateArticle(articleNew)
		if err != nil {
			log.Fatalf("unable to create file: %v", err)
		}

		fmt.Println("article:", article.Name)
		
	}
}

func LoadProductDescriptionAiroClimate()  {
	account, err := models.GetAccount(5)
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
		
		_, err = account.UpdateProduct(uint(fileId), map[string]interface{}{"Description":string(body)})
		if err != nil {
			log.Fatalf("unable to update product: %v", err)
		}

	}
}

func LoadProductCategoryDescriptionAiroClimate()  {
	/*account, err := models.GetAccount(5)
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

		group := models.ProductGroup{}

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

func RefreshTablesPart_IV() {
	pool := models.GetPool()

	err := pool.Exec("drop table if exists payment_options_delivery_pickups, payment_options_delivery_couriers, payment_options_delivery_russian_posts, " +
		"orders_products, payment_methods_web_sites, payment_methods_payments").Error
	if err != nil {
		log.Fatalf("Cant create tables -1: %v", err)
		return
	}
	pool.DropTableIfExists(

		models.CartItem{},
		models.OrderComment{},
		models.OrderChannel{},

		models.DeliveryOrder{},
		models.PaymentOption{},
		models.Order{},
		models.Payment{},
		models.PaymentAmount{},
		models.PaymentYandex{},
		models.PaymentCash{},

		)


	// А теперь создаем

	models.PaymentAmount{}.PgSqlCreate()
	models.PaymentOption{}.PgSqlCreate()
	models.CartItem{}.PgSqlCreate()
	// models.PaymentSubject{}.PgSqlCreate()
	// models.VatCode{}.PgSqlCreate()
	models.OrderComment{}.PgSqlCreate()
	models.OrderChannel{}.PgSqlCreate()
	models.Order{}.PgSqlCreate()
	models.DeliveryOrder{}.PgSqlCreate()
	models.PaymentCash{}.PgSqlCreate()
	models.PaymentYandex{}.PgSqlCreate()
	models.Payment{}.PgSqlCreate()

	pool.AutoMigrate(&models.DeliveryPickup{},&models.DeliveryCourier{},&models.DeliveryRussianPost{})

}

func UploadTestDataPart_IV()  {

	// 1. Получаем AiroClimate аккаунт
	airoAccount, err := models.GetAccount(5)
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	// Создаем методы оплаты
	paymentOptions := []models.PaymentOption {
		{Name:   "Оплата при получении",	Code: "cash"},
		{Name:   "Онлайн-оплата картой",	Code: "online"},
		// {Name:   "Покупка в кредит",		Code: "credit"},
	}
	for i := range(paymentOptions) {
		_, err := models.Account{Id: airoAccount.Id}.CreateEntity(&paymentOptions[i])
		if err != nil {
			log.Fatalf("Не удалось создать paymentMethods: ", err)
		}
	}

	var paymentCash models.PaymentOption
	if err := airoAccount.LoadEntity(&paymentCash, 1); err != nil {
		log.Fatal(err)
	}

	var paymentOnline models.PaymentOption
	if err := airoAccount.LoadEntity(&paymentOnline, 2); err != nil {
		log.Fatal(err)
	}

/////////
	var webSite models.WebSite
	if err := airoAccount.LoadEntity(&webSite, 5); err != nil { log.Fatal(err)}

	if err := webSite.AppendPaymentOptions([]models.PaymentOption{
		{AccountId: 5, Id: 1},{AccountId: 5, Id: 2},
	}); err != nil {
		log.Fatal(err)
	}
	////////////

	// Создаем способ оплаты YandexPayment
	entityPayment, err := airoAccount.CreateEntity(
		&models.PaymentYandex{
			Name:   "Прием платежей через интернет-магазин airoclimate.ru",
			ApiKey: "test_f56EEL_m2Ky7CJnnRjSpb4JLMhiGoGD3X6ScMHGPruM",
			ShopId: "730509",
			URL: "https://ui.api.ratuscrm.com/yandex-payment/dasdasdsa/notifications",
			ReturnUrl: "https://airoclimate.ru/payment-return",
			Enabled: true,
			Description: "-",
			SavePaymentMethod: false,
			Capture: false,
		})
	if err != nil {
		log.Fatalf("Не удалось создать entityPayment: ", err)
	}
	var _paymentYandex models.PaymentYandex
	if err = airoAccount.LoadEntity(&_paymentYandex,entityPayment.GetId()); err != nil {
		log.Fatalf("Не удалось найти entityPayment: ", err)
	}

	if err := _paymentYandex.SetPaymentOption(paymentOnline); err != nil {
		log.Fatal(err)
	}

	// Создаем способ оплаты PaymentCash
	entityPayment2, err := airoAccount.CreateEntity(
		&models.PaymentCash{
			Name:   "Прием платежей через в интернет-магазине airoclimate.ru",
			Enabled: true,
			Description: "Наличный способ оплаты при самовывозе",
		})
	if err != nil {
		log.Fatalf("Не удалось создать entityPayment: ", err)
	}
	var _paymentCash models.PaymentCash
	if err = airoAccount.LoadEntity(&_paymentCash,entityPayment2.GetId()); err != nil {
		log.Fatalf("Не удалось найти paymentCash: ", err)
	}

	if err := _paymentCash.SetPaymentOption(paymentOnline); err != nil {
		log.Fatal(err)
	}

	deliveries := webSite.GetDeliveryMethods()
	for i := range(deliveries) {
		if err := deliveries[i].AppendPaymentOptions([]models.PaymentOption{paymentCash, paymentOnline}); err != nil {
			fmt.Println(err)
			return
		}
	}



	///////////////



    fmt.Println("Объекты созданы создан: ")
}

func Migrate_I() {
	pool := models.GetPool()
	pool.AutoMigrate(&models.User{})
}