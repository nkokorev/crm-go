package base

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/models"
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

	// дропаем системные таблицы
	// err = pool.Exec("drop table if exists product_card_products, unit_measurements, product_cards, products, product_groups, shops").Error
	/*err = pool.Exec("drop table if exists eav_attribute_types, eav_attributes_varchar, eav_attributes_int, eav_attributes").Error
	if err != nil {
		fmt.Println("Cant create tables 0: ", err)
		return
	}*/

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

	models.WebHook{}.PgSqlCreate()
	models.Article{}.PgSqlCreate()

	UploadTestData()
}

func RefreshTablesPart_II() {

	pool := models.GetPool()


	pool.DropTableIfExists(models.Lead{}, models.Delivery{})

	models.Lead{}.PgSqlCreate()
	models.Delivery{}.PgSqlCreate()

	UploadTestDataPart_II()
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

	_, err = acc357.ApiKeyCreate(models.ApiKey{Name:"Для сайта"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
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

	_, err = airoClimat.ApiKeyCreate(models.ApiKey{Name:"Для сайта"})
	if err != nil {
		log.Fatalf("Не удалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}

	// 4. !!! Создаем магазин
	airoShop, err := airoClimat.CreateShop(models.Shop{Name: "airoclimate.ru", Address: "г. Москва, р-н Текстильщики", Email: "info@airoclimate.ru", Phone: "+7 (4832) 77-03-73"})
	if err != nil {
		log.Fatal("Не удалось создать Shop для airoClimat: ", err)
	}
	
	groupAiroRoot, err := airoShop.CreateProductGroup(
		models.ProductGroup{
			Code: "root", Name: "", URL: "/", IconName: "far fa-home", RouteName: "info.index",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}

	groupAiroCatalogRoot, err := groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog", Name: "Весь каталог", URL: "catalog", IconName: "far fa-th-large", RouteName: "catalog",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	groupAiro1, err := groupAiroCatalogRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog",Name: "Бактерицидные рециркуляторы", URL: "bactericidal-recirculators", IconName: "far fa-fan-table", RouteName: "catalog.recirculators",
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	groupAiro2, err := groupAiroCatalogRoot.CreateChild(
		models.ProductGroup{
			Code: "catalog",Name: "Бактерицидные камеры", URL: "bactericidal-chambers", IconName: "far fa-box-full", RouteName: "catalog.chambers",
			})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}

	//////////////

	_, err = groupAiroRoot.CreateChild(

		models.ProductGroup{
			Code: "info", Name: "Статьи", URL: "articles", IconName: "far fa-books", RouteName: "info.articles", Order: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
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
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "О компании", URL: "about", IconName: "far fa-home-heart", RouteName: "info.about",      Order: 5,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "Политика конфиденциальности", URL: "privacy-policy", IconName: "far fa-home-heart", RouteName: "info.privacy-policy",      Order: 6,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "info", Name: "Контакты", URL: "contacts", IconName: "far fa-address-book", RouteName: "info.contacts",  Order: 10,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}

	////////
	_, err = groupAiroRoot.CreateChild(
		models.ProductGroup{
			Code: "cart", Name: "Корзина", URL: "cart", IconName: "far fa-cart-arrow-down", RouteName: "cart", Order: 1,
		})
	if err != nil {
		log.Fatal("Не удалось создать ProductGroup для airoClimat shop: ", err)
	}

	// 6. Создаем карточки товара
	cards := []models.ProductCard{
		{ID: 0, URL: "airo-dez-adjustable-black", 	Label:"Рециркулятор AIRO-DEZ черный с регулировкой", Breadcrumb: "Рециркулятор AIRO-DEZ черный с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ черный с регулировкой",
			SwitchProducts: postgres.Jsonb{
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
			}},
		{ID: 0, URL: "airo-dez-black", 				Label:"Рециркулятор AIRO-DEZ черный", Breadcrumb: "Рециркулятор AIRO-DEZ черный", MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ черный"},
		{ID: 0, URL: "airo-dez-adjustable-white", 	Label:"Рециркулятор AIRO-DEZ белый с регулировкой", Breadcrumb: "Рециркулятор AIRO-DEZ белый с регулировкой", MetaTitle: "Рециркулятор AIRO-DEZ белый с регулировкой"},
		{ID: 0, URL: "airo-dez-white", 				Label:"Рециркулятор AIRO-DEZ белый", Breadcrumb: "Рециркулятор AIRO-DEZ белый",MetaTitle: "Рециркулятор воздуха бактерицидный AIRO-DEZ белый"},
		{ID: 0, URL: "airo-dez-compact", 			Label: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", Breadcrumb: "Мобильный аиродезинфектор AIRO-DEZ COMPACT",MetaTitle: "Мобильный аиродезинфектор AIRO-DEZ COMPACT", },

		{ID: 0, URL: "airo-dezpuf",			Label: "Бактерицидная камера пуф AIRO-DEZPUF", Breadcrumb: "Бактерицидная камера пуф AIRO-DEZPUF",MetaTitle: "Бактерицидная камера пуф AIRO-DEZPUF"},
		{ID: 0, URL: "airo-dezpuf-wenge", 	Label: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге", Breadcrumb: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге",MetaTitle: "Бактерицидная тумба пуф AIRO-DEZPUF цвет дуб венге"},
		
		{ID: 0, URL: "airo-dezbox", 		Label: "Бактерицидная камера AIRO-DEZBOX", Breadcrumb: "Бактерицидная камера AIRO-DEZBOX",MetaTitle: "Бактерицидная камера AIRO-DEZBOX", },
		{ID: 0, URL: "airo-dezbox-white", 	Label: "Бактерицидная камера AIRO-DEZBOX белая",Breadcrumb: "Бактерицидная камера AIRO-DEZBOX белая", MetaTitle: "Бактерицидная камера AIRO-DEZBOX белая", },
		{ID: 0, URL: "airo-deztumb", 		Label: "Тумба облучатель бактерицидный AIRO-DEZTUMB", Breadcrumb: "Тумба облучатель бактерицидный AIRO-DEZTUMB",MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB", },
		{ID: 0, URL: "airo-deztumb-big", 	Label: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", Breadcrumb: "Тумба облучатель бактерицидный AIRO-DEZTUMB big",MetaTitle: "Тумба облучатель бактерицидный AIRO-DEZTUMB big", },

		{ID: 0, URL: "airo-deztumb-pine", 	Label: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина",Breadcrumb: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", MetaTitle: "Бактерицидная тумба AIRO-DEZTUMB цвет сосна касцина", },

	}

	//metadata := json.RawMessage(`{"color": "white", "bodyMaterial": "металл", "filterType": "угольно-фотокаталитический"}`)

	// 7. Создаем список товаров
	products := []models.Product{
		{
			Model: ToStringPointer("AIRO-DEZ с регулировкой черный"), //,
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
			Model: ToStringPointer("AIRO-DEZ черный"),
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
			Model: ToStringPointer("AIRO-DEZ с регулировкой белый"),
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
			Model: ToStringPointer("AIRO-DEZ белый"),
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
			Model: ToStringPointer("AIRO-DEZ COMPACT"),
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
			Model: ToStringPointer("AIRO-DEZPUF"),
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
			Model: ToStringPointer("AIRO-DEZPUF венге"),
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
			Model: ToStringPointer("AIRO-DEZBOX"),
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
			Model: ToStringPointer("AIRO-DEZBOX белая"),
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
			Model: ToStringPointer("AIRO-DEZTUMB"),
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
			Model: ToStringPointer("AIROTUMB big"),
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
			Model: ToStringPointer("AIRO-DEZTUMB касцина"),
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

	// 9. Создаем вебхкуи
	domainAiroSite := ""
	AppEnv := os.Getenv("APP_ENV")

	switch AppEnv {
	case "local":
		domainAiroSite = "http://airoclimate.me"
	case "public":
		domainAiroSite = "http://airoclimate.ratus-dev.ru"
	default:
		domainAiroSite = "http://airoclimate.ratus-dev.ru"
	}

	webHooks := []models.WebHook {
		{Name: "Upload all shop data", EventType: models.EventUpdateAllShopData, URL: domainAiroSite + "/ratuscrm/webhooks/upload/all", HttpMethod: http.MethodGet},

		// Upload all
		{Name: "Upload all shop", EventType: models.EventShopsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/shops", HttpMethod: http.MethodGet},
		{Name: "Upload all products", EventType: models.EventProductsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/products", HttpMethod: http.MethodGet},
		{Name: "Upload all product cards", EventType: models.EventProductCardsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards", HttpMethod: http.MethodGet},
		{Name: "Upload all product groups", EventType: models.EventProductGroupsUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups", HttpMethod: http.MethodGet},
		{Name: "Upload all articles", EventType: models.EventArticlesUpdate, URL: domainAiroSite + "/ratuscrm/webhooks/articles", HttpMethod: http.MethodGet},

		// Create entity
		{Name: "Create shop", EventType: models.EventShopCreated, URL: domainAiroSite + "/ratuscrm/webhooks/shops/{{.ID}}", HttpMethod: http.MethodPost},
		{Name: "create product", EventType: models.EventProductCreated, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.ID}}", HttpMethod: http.MethodPost},
		{Name: "Create product card", EventType: models.EventProductCardCreated, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.ID}}", HttpMethod: http.MethodPost},
		{Name: "Create product group", EventType: models.EventProductGroupCreated, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.ID}}", HttpMethod: http.MethodPost},
		{Name: "Create article", EventType: models.EventArticleCreated, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.ID}}", HttpMethod: http.MethodPost},

		// Update entity
		{Name: "Update shop", EventType: models.EventShopUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/shops", HttpMethod: http.MethodPatch},
		{Name: "Update product", EventType: models.EventProductUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.ID}}", HttpMethod: http.MethodPatch},
		{Name: "Update product card", EventType: models.EventProductCardUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.ID}}", HttpMethod: http.MethodPatch},
		{Name: "Update product group", EventType: models.EventProductGroupUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.ID}}", HttpMethod: http.MethodPatch},
		{Name: "Update article", EventType: models.EventArticleUpdated, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.ID}}", HttpMethod: http.MethodPatch},

		// Delete
		{Name: "Delete shop", EventType: models.EventShopDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/shops", HttpMethod: http.MethodDelete},
		{Name: "Delete product", EventType: models.EventProductDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/products/{{.ID}}", HttpMethod: http.MethodDelete},
		{Name: "Delete product card", EventType: models.EventProductCardDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/product-cards/{{.ID}}", HttpMethod: http.MethodDelete},
		{Name: "Delete product group", EventType: models.EventProductGroupDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/product-groups/{{.ID}}", HttpMethod: http.MethodDelete},
		{Name: "Delete article", EventType: models.EventArticleDeleted, URL: domainAiroSite + "/ratuscrm/webhooks/articles/{{.ID}}", HttpMethod: http.MethodDelete},
	}
	for i,_ := range webHooks {
		 _, err = airoClimat.CreateWebHook(webHooks[i])
		if err != nil {
			log.Fatal("Не удалось создать webHook: ", err)
		}

	}
	
	return

}

func UploadTestDataPart_II() {

	// 1. Получаем главный аккаунт
	account, err := models.GetMainAccount()
	if err != nil {
		log.Fatalf("Не удалось найти главный аккаунт: %v", err)
	}

	/*_l := models.Lead{Name: "Test Entity"}

	_, err = account.CreateEntity(&_l)
	if err != nil {
		log.Fatalf("Не удалось получить Emtity: %v", err)
	}*/

	deliveries := []models.Delivery{
		{Name: "Самовывоз", Enabled: true, ShopID: 1},
		{Name: "Курьерская доставка по г. Москва", Enabled: true, ShopID: 1},
		{Name: "Почта России", Enabled: true, ShopID: 1},
	}

	for i,_ := range deliveries {
		_, err := account.CreateEntity(&deliveries[i])
		if err != nil {
			log.Fatalf("Не удалось получить entDeliveries: %v", err)
		}
	}

	var delivery models.Delivery
	if err = account.GetEntity(&delivery, 2); err != nil {
		log.Fatal(err)
	}


	fmt.Printf("Type: %T \n: ", delivery)
	fmt.Println("delivery name: ", delivery.Name)

	/*var d models.Lead
	lead, err := account.GetEntity(1, &d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("lead: ", lead)*/

}

func MapToRawJson(input map[string]interface{}) json.RawMessage {

	b, err := json.Marshal(input)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return b
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
				//OwnerID: uint(index),
				//EmailId: uint(emailId),
			}
			file, err := account.StorageCreateFile(&fs)
			if err != nil {
				log.Fatalf("unable to create file: %v", err)
			}

			err = (models.Product{ID: uint(index)}).AppendAssociationImage(*file)
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