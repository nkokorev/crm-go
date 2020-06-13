package base

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/models"
	"log"
	"time"
)

func RefreshTables() {

	var err error
	pool := models.GetPool()

	// дропаем системные таблицы
	// err = pool.Exec("drop table if exists product_card_products, unit_measurements, product_cards, products, product_groups, shops").Error
	/*err = pool.Exec("drop table if exists eav_attribute_types, eav_attributes_varchar, eav_attributes_int, eav_attributes").Error
	if err != nil {
		fmt.Println("Cant create tables 0: ", err)
		return
	}*/

	err = pool.Exec("drop table if exists  unit_measurements, product_card_products, product_cards").Error
	if err != nil {
		fmt.Println("Cant create tables 0: ", err)
		return
	}

	err = pool.Exec("drop table if exists  storage, products, shops, product_groups").Error
	if err != nil {
		fmt.Println("Cant create tables 1: ", err)
		return
	}

	err = pool.Exec("drop table if exists domains, email_boxes, email_senders, email_templates, api_keys").Error
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
	
	err = models.CrmSetting{}.PgSqlCreate()
	if err != nil {
		log.Fatal(err)
	}

	models.UserVerificationMethod{}.PgSqlCreate()
	models.Account{}.PgSqlCreate()
	models.Role{}.PgSqlCreate()
	models.User{}.PgSqlCreate()
	models.AccountUser{}.PgSqlCreate()
	models.ApiKey{}.PgSqlCreate()

	models.UnitMeasurement{}.PgSqlCreate()
	
	models.Shop{}.PgSqlCreate()
	models.ProductGroup{}.PgSqlCreate()
	models.ProductCard{}.PgSqlCreate()
	models.Product{}.PgSqlCreate()

	//models.EavAttrType{}.PgSqlCreate()
	//models.EavAttribute{}.PgSqlCreate()
	//models.EavAttrVarchar{}.PgSqlCreate()
	//models.EavAttrInt{}.PgSqlCreate()
	//models.EavAttrDecimal{}.PgSqlCreate()

	models.Domain{}.PgSqlCreate()
	models.EmailBox{}.PgSqlCreate()
	models.EmailTemplate{}.PgSqlCreate()

	models.Storage{}.PgSqlCreate()

	UploadTestData()
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

func UploadTestData() {
	
	// 1. Получаем главный аккаунт
	mAcc, err := models.GetMainAccount()
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
			//DefaultAccountID:null,
			EmailVerifiedAt:&timeNow,
			},
		models.RoleOwner,
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
			//DefaultAccountID:null,
			EmailVerifiedAt:&timeNow,
		},
		models.RoleAdmin,
	)
	if err != nil || mex388 == nil {
		log.Fatal("Не удалось создать mex388'a: ", err)
	}

	// 3. Создаем домен для главного аккаунта
	domainMain, err := mAcc.CreateDomain(models.Domain{
		Hostname: "ratuscrm.com",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC4dksLEYhARII4b77fe403uCJhD8x5Rddp9aUJCg1vby7d6QLOpP7uXpXKVLXxaxQcX7Kjw2kGzlvx7N+d2tToZ8+T3SUadZxLOLYDYkwalkP3vhmA3cMuhpRrwOgWzDqSWsDfXgr4w+p1BmNbScpBYCwCrRQ7B12/EXioNcioCQIDAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDB8BPdNbNwi3LA6VMp8BbOGKNrV1PxYZsxp6LvTSK9EgJcRIMw
C+Uc1GgnvcTNksF5GviVYcy2az/e8ACLvcKI6Lb1gUhk10SHIRcb5boK/Li9aOUu
F5ndGzzg0aBzsG2P0us+tkgFOTjc5MuBdlKOzraLegRbfL5MWUWe5SS3FQIDAQAB
AoGANIXli1Jg34kUsgQ+3qvEMVrg31BOTqAlnMQOz4pvbw8yjnSLpvaBvVYVQzYU
16v4M+lHC4XqIDlZmfIb47yns12ASHSoFUzPeUioRu9oWxaOlcHSqWkZBg5miEuM
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

	// 4. Добавляем почтовые ящики в домен
	_, err = domainMain.AddMailBox(models.EmailBox{Default: true, Allowed: true, Name: "RatusCRM", Box: "info"})
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
		models.RoleClient,
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
		UserVerificationMethodID:            dvc.ID,
		UiApiEnabledLoginNotVerifiedUser:    true, // really?
		VisibleToClients:                    false,
	})
	if err != nil || acc357 == nil {
		log.Fatal("Не удалось создать аккаунт 357 грамм")
		return
	}

	// 3. добавляем меня как админа
	_, err = acc357.AppendUser(*owner, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 3.2. добавляем кучу других клиентов
	if true {
		var clients []models.User

		for i:=0; i < 10 ;i++ {
			clients = append(clients, models.User{
				Name: fmt.Sprintf("Name #%d", i),
				Email: fmt.Sprintf("email%d@mail.ru", i),
				Phone: fmt.Sprintf("+7925195221%d", i),
			})
		}
		for i,_ := range clients {
			_, err := acc357.CreateUser(clients[i])
			if err != nil {
				fmt.Println(err)
				log.Fatal("Не удалось добавить клиента id: ", i)
				return
			}
		}
	}


	// 4. Создаем домен для 357gr
	domain357gr, err := acc357.CreateDomain(models.Domain{
		Hostname: "357gr.ru",
		DKIMPublicRSAKey: ``,
		DKIMPrivateRSAKey: ``,
		DKIMSelector: "dk1",
	})
	if err != nil {
		log.Fatal("Не удалось создать домены для главного аккаунта: ", err)
	}

	// 5. Добавляем почтовые ящики в домен 357gr
	_, err = domain357gr.AddMailBox(models.EmailBox{Default: true, Allowed: true, Name: "357 Грамм", Box: "info"})
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
			// Email:"sa-tolstov@yandex.ru",
			Email:"info@rus-marketing.ru",
			PhoneRegion: "RU",
			Phone: "",
			Password:"qwerty109#QW",
			Name:"Станислав",
			Surname:"Толстов",
			Patronymic:"",
			EmailVerifiedAt:&timeNow,
		},
		models.RoleClient,
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
	_, err = accSyndicAd.AppendUser(*owner, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2 Добавляем Mex388
	_, err = accSyndicAd.AppendUser(*mex388, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in 357gr")
		return
	}

	// 2. Создаем домен для синдиката
	domainSynd, err := accSyndicAd.CreateDomain(models.Domain{
		Hostname: "syndicad.com",
		DKIMPublicRSAKey: `MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyxCGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+ABtf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIDAQAB`,
		DKIMPrivateRSAKey: `-----BEGIN RSA PRIVATE KEY-----
MIICWwIBAAKBgQDEwBDUBhnVcb+wPoyj6UrobwhKp0bIMzl9znfS127PdLqeGEyx
CGy6CTT7coAturzb2dw33e3OhzzOvvBjnzSamRfpAj3vuBiSWtykS4JH17EN/4+A
Btf7VOqfRWwB7F80VJ+3/Xv7TzkmNcAg+ksgDzk//BCXfcVFfx56Jxf7mQIDAQAB
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

	// 3. Добавляем почтовые ящики
	_, err = domainSynd.AddMailBox(models.EmailBox{Default: true, Allowed: true, Name: "SyndicAd", Box: "info"})
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
	_, err = brouser.AppendUser(*owner, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2. Добавляем mex388
	_, err = brouser.AppendUser(*mex388, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in brouser")
		return
	}

	// 2. Создаем домен для синдиката
	domainBrouser, err := brouser.CreateDomain(models.Domain{
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

	// 3. Добавляем почтовые ящики
	_, err = domainBrouser.AddMailBox(models.EmailBox{Default: true, Allowed: true, Name: "Brouser", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Brouser: ", err)
	}
	
	// Добавляем шаблоны писем для синдиката и главного аккаунта
	/*data, err := ioutil.ReadFile("/var/www/ratuscrm/files/example.html")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	_, err = mAcc.CreateEmailTemplate(models.EmailTemplate{Name: "example", Code: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = acc357.CreateEmailTemplate(models.EmailTemplate{Name: "example", Code: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = accSyndicAd.CreateEmailTemplate(models.EmailTemplate{Name: "example", Code: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = brouser.CreateEmailTemplate(models.EmailTemplate{Name: "example", Code: string(data)})
	if err != nil {
		log.Fatal(err)
	}*/

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
		models.RoleClient,
	)
	if err != nil || owner == nil {
		log.Fatal("Не удалось создать korotaev'a: ", err)
	}

	airoClimat, err := korotaev.CreateAccount(models.Account{
		Name:                                "AIRO Climate",
		Website:                             "http://airoclimate.ru",
		Type:                                "internet-shop",
		ApiEnabled:                          true,
		UiApiEnabled:                        true,
		VisibleToClients:                    false,
	})
	if err != nil || airoClimat == nil {
		log.Fatal("Не удалось создать аккаунт Brouser")
		return
	}

	// 2. добавляем меня как админа
	_, err = airoClimat.AppendUser(*owner, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2.2. Добавляем mex388 как админа
	_, err = airoClimat.AppendUser(*mex388, models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя mex388 in brouser")
		return
	}

	// 2. Создаем домен для airoClimat
	domainAiroClimate, err := airoClimat.CreateDomain(models.Domain{
		Hostname: "airoclimate.ru",
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

	// 3. Добавляем почтовые ящики
	_, err = domainAiroClimate.AddMailBox(models.EmailBox{Default: true, Allowed: true, Name: "AIRO Climate", Box: "info"})
	if err != nil {
		log.Fatal("Не удалось создать MailBoxes для Brouser: ", err)
	}

	// 4. !!! Создаем магазин
	airoShop, err := airoClimat.CreateShop(models.Shop{Name: "airoclimate.ru", Address: "г. Москва, р-н Текстильщики", Email: "info@airoclimate.ru", Phone: "+7 (4832) 77-03-73"})
	if err != nil {
		log.Fatal("Не удалось создать Shop для airoClimat: ", err)
	}

	/*err = airoClimat.CreateBaseEavAttributes()
	if err != nil {
		log.Fatal("Не удалось создать Атрибуты для airoClimat: ", err)
	}*/

	// 5. Создаем 3 категории товаров
	groupAiroRoot, err := airoShop.CreateProductGroup(models.ProductGroup{Name: "Бактерицидные облучатели", URL: ""});
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	groupAiro1, err := groupAiroRoot.CreateChild(models.ProductGroup{Name: "Бактерицидные рециркуляторы", URL: "bactericidal-recirculators"})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	groupAiro2, err := groupAiroRoot.CreateChild(models.ProductGroup{Name: "Бактерицидные камеры", URL: "bactericidal-chambers"})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}

	// 6. Создаем карточки товара
	cards := []models.ProductCard{
		{ID: 0, URL: "airo-dez-adjustable-black", 	Name:"Рециркулятор AIRO-DEZ черный с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ черный с регулировкой"},
		{ID: 0, URL: "airo-dez-black", 				Name:"Рециркулятор AIRO-DEZ черный", MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ черный"},
		{ID: 0, URL: "airo-dez-adjustable-white", 	Name:"Рециркулятор AIRO-DEZ белый с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ белый с регулировкой"},
		{ID: 0, URL: "airo-dez-white", 				Name:"Рециркулятор AIRO-DEZ белый", MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ белый"},
		{ID: 0, URL: "airo-dez-compact", 			Name: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", MetaTitle: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", },

		{ID: 0, URL: "airo-dezpuf", Name: "Бактерицидная камера пуф AIRO-DEZPUF", MetaTitle: "Бактерицидная камера пуф AIRO-DEZPUF"},
		{ID: 0, URL: "airo-dezpuf-wenge", Name: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге", MetaTitle: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"},
		
		{ID: 0, URL: "airo-dezbox", Name: "Бактерицидная камера AIRO-DEZBOX", MetaTitle: "Бактерицидная камера AIRO-DEZBOX", },
		{ID: 0, URL: "airo-dezbox-white", Name: "Бактерицидная камера AIRO-DEZBOX белая", MetaTitle: "Бактерицидная камера AIRO-DEZBOX белая", },
		{ID: 0, URL: "airo-deztumb", Name: "Тумба облучатель бактерицидный AIRO-DEZTUMB", MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB", },
		{ID: 0, URL: "airo-deztumb-big", Name: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", },

		{ID: 0, URL: "airo-deztumb-pine", Name: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", MetaTitle: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", },

	}

	//metadata := json.RawMessage(`{"color": "white", "bodyMaterial": "металл", "filterType": "угольно-фотокаталитический"}`)


	// 7. Создаем список товаров
	products := []models.Product{
		{
			SKU:"", Model: "AIRO-DEZ с регулировкой черный",
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности черный", ShortName: "Рециркулятор AIRO-DEZ черный",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 19500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZ черный",
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ черный", ShortName: "Рециркулятор AIRO-DEZ черный",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 17500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZ с регулировкой белый",
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ с регулировкой мощности белый",  ShortName: "Рециркулятор AIRO-DEZ белый",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 19500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZ белый",
			Name:"Рециркулятор воздуха бактерицидный AIRO-DEZ",  ShortName: "Рециркулятор AIRO-DEZ белый",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 17500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZ COMPACT",
			Name:"Мобильный аиродезинфектор AIRO-DEZ COMPACT",  ShortName: "Аиродезинфектор AIRO-DEZ COMPACT",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 39000.00, RetailDiscount: 3000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZPUF",
			Name:"Бактерицидная камера пуф AIRO-DEZPUF",  ShortName: "Камера пуф AIRO-DEZPUF",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 11000.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZPUF венге",
			Name:"Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге",  ShortName: "Камера AIRO-DEZBOX",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 12000.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZBOX",
			Name:"Бактерицидная камера AIRO-DEZBOX",  ShortName: "Камера AIRO-DEZBOX",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 7800.00, RetailDiscount: 800,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZBOX белая",
			Name:"Бактерицидная камера AIRO-DEZBOX белая",  ShortName: "Камера AIRO-DEZBOX белая",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 7800.00, RetailDiscount: 800,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZTUMB",
			Name:"Тумба облучатель бактерицидный AIRO-DEZTUMB",  ShortName: "Бактерицидная тумба AIRO-DEZTUMB",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIROTUMB big",
			Name:"Тумба облучатель бактерицидный AIRO-DEZTUMB big",  ShortName: "Облучатель AIROTUMB big",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
			SKU:"", Model: "AIRO-DEZTUMB касцина",
			Name:"Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина",  ShortName: "Бактерицидная тумба AIRO-DEZTUMB",
			ProductType: models.ProductTypeCommodity, UnitMeasurementID: 1,
			RetailPrice: 11500.00, RetailDiscount: 1000,
			ShortDescription: "",
			Description: "",
			Attributes: postgres.Jsonb{
				RawMessage: MapToRawJson(map[string]interface{}{
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
		_, err = airoShop.CreateProductWithCardAndGroup(products[i], cards[i], &group.ID)
		if err != nil {
			log.Fatal("Не удалось создать Product для airoClimat: ", err)
		}
	}
	
	return

/*	shops := [] *models.Shop{
		{AccountID:3, Name:"Магазин на Маяковке", Address:"Москва, ул. Долгоруковская, дом 9, стр. 3"},
	}

	product_groups := [] *models.ProductGroup{
		{ ShopID:1, Code:"root", URL:"/", Name:"Главная", Breadcrumb: "Главная", Description:""},

		{ParentID:1, ShopID:1, Code:"tea", URL:"tea", Name:"Чай", Breadcrumb: "Чай", Description:""},
		{ParentID:1, ShopID:1, Code:"coffee", URL:"coffee", Name:"Кофе", Breadcrumb: "Кофе", Description:""},
		{ParentID:1, ShopID:1, Code:"gift", URL:"gift", Name:"Подарки", Breadcrumb: "Подарки", Description:""},
		{ParentID:1, ShopID:1, Code:"accessories", URL:"accessories", Name:"Посуда и аксессуары", Breadcrumb: "Посуда и аксессуары", Description:""},

		{ParentID:2, ShopID:1, Code:"tea.puer", 	URL:"puer", 	Name:"Пуэр", Breadcrumb: "Пуэр", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.oolong",	URL:"oolong", 	Name:"Улунский чай", Breadcrumb: "Улунский чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.red", 	URL:"red", 		Name:"Красный чай", Breadcrumb: "Красный чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.green", 	URL:"green", 	Name:"Зеленый чай", Breadcrumb: "Зеленый чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.white", 	URL:"white", 	Name:"Белый чай", Breadcrumb: "Белый чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.yellow",	URL:"yellow", 	Name:"Желтый чай", Breadcrumb: "Желтый чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.herbal", 	URL:"herbal", 	Name:"Травяной чай", Breadcrumb: "Травяной чай", Description:""},
		{ParentID:2, ShopID:1, Code:"tea.additives",URL:"additives",Name:"Чайные добавки", Breadcrumb: "Чайные добавки", Description:""},

		{ParentID:2, ShopID:1, Code:"tea.china", URL:"china", Name:"Китайский чай", Breadcrumb: "Китайский чай", Description:""}, // country = china & type = tea
		{ParentID:2, ShopID:1, Code:"tea.taiwan", URL:"taiwan", Name:"Тайваньский чай", Breadcrumb: "Тайваньский чай", Description:""}, // country = taiwan & type = tea

		{ParentID:5, ShopID:1, Code:"accessories.tableware.brewing", URL:"tableware-for-brewing", Name:"Посуда для заварки китайского чая", Breadcrumb: "Посуда для заварки китайского чая", Description:""}, // country = taiwan & type = tea

		{ParentID:16, ShopID:1, Code:"accessories.tableware.brewing.gunfu", URL:"gunfu", Name:"Типоды (Гунфу)", Breadcrumb: "Типоды (Гунфу чайники)", Description:""}, // country = taiwan & type = tea

	}

	products := [] *models.Product{
		{AccountID:3, ProductGroupID: 8, SKU:"1017", URL:"er-e-i-ya-dyan-hun", Name:"Эр Е И Я Дянь Хун", },
		{AccountID:3, ProductGroupID: 8, SKU:"1133", URL:"hun-ta", 			Name:"Хун Та", },
		{AccountID:3, ProductGroupID: 8, SKU:"579", URL:"dyan-hun-mao-fen", Name:"Дянь Хун Мао Фэн", },
		{AccountID:3, ProductGroupID: 8, SKU:"910", URL:"syao-chzhun", Name:"Сяо Чжун", },
		{AccountID:3, ProductGroupID: 8, SKU:"587", URL:"dyan-hun-tszin-hao", Name:"Дянь Хун Цзинь Хао", },
		{AccountID:3, ProductGroupID: 8, SKU:"865", URL:"hun-sun-chjen", Name:"Хун Сун Чжень", },
		{AccountID:3, ProductGroupID: 8, SKU:"300", URL:"tszin-tszyun-mey", Name:"Цзинь Цзюнь Мэй", },
		{AccountID:3, ProductGroupID: 8, SKU:"940", URL:"e-shen-hun-cha", Name:"Е Шен Хун Ча", ShortDescription:"Дикорастущий красный чай"},
		{AccountID:3, ProductGroupID: 8, SKU:"1018", URL:"chzhun-go-hun", Name:"Чжун Го Хун"},
		{AccountID:3, ProductGroupID: 8, SKU:"859", URL:"dyan-hun-sosnovye-igly", Name:"Дянь Хун \"Сосновые иглы\""},
		{AccountID:3, ProductGroupID: 8, SKU:"965", URL:"li-chzhi-hun-cha", Name:"Ли Чжи Хун Ча"},

		{AccountID:3, ProductGroupID: 17, SKU:"80", URL:"samadoyo-b-06", Name:"SAMADOYO B-06 (600 мл)", ShortDescription:"Чайник с кнопкой (типод)"}, // 12
	}

	offers := [] *models.Offer{
		{AccountID:3, Name:"25гр (пробник)", Price:350.00, Discount:0},
		{AccountID:3, Name:"50гр", Price:550.00, Discount:0},
		{AccountID:3, Name:"100гр", Price:1100.00, Discount:150},
		{AccountID:3, Name:"100гр + типод", Price:2200.00, Discount:400},
	}

	fmt.Println(offers, products, shops, product_groups)*/


}

func MapToRawJson(input map[string]interface{}) json.RawMessage {

	b, err := json.Marshal(input)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
}
