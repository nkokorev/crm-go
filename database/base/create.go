package base

import (
	"fmt"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/models"
	"io/ioutil"
	"log"
	"time"
)

func RefreshTables() {

	var err error
	pool := models.GetPool()

	// дропаем системные таблицы
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

	// Таблица пользователей
	models.User{}.PgSqlCreate()

	models.AccountUser{}.PgSqlCreate()

	/*// в этой таблице хранятся пользовательские email-уведомления
	err = pool.Exec("create table  email_access_tokens (\ntoken varchar(255) PRIMARY KEY UNIQUE, -- сам уникальный ключ\naction_type VARCHAR(255) NOT NULL DEFAULT 'verification', -- verification, recover (username, password, email), join to account, [invite-one], [invite-unlimited], [invite-free] - свободный инвайт...\ndestination_email varchar(255) NOT NULL, -- куда фактически был отправлен token (для безопасности) или для кого предназначается данный инвайт (например, строго по емейлу)\nowner_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE, -- ID пользователя, кто создавал этот ключ (может быть self) \nnotification_count INT DEFAULT 0, -- число уведомлений\nnotification_at TIMESTAMP DEFAULT NULL, -- время уведомления\ncreated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n--  expired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP\n);\n").Error
	if err != nil {
		log.Fatal("Cant create table user_email_send", err)
	}

	// Магазины (Shops).
	err = pool.Exec("create table shops (\n  id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    name VARCHAR(255) NOT NULL, -- имя магазина    \n    address VARCHAR(255) -- потом можно более детально сделать адрес\n \n);\n\n").Error
	if err != nil {
		log.Fatal("Cant create table shops", err)
	}

	err = pool.Exec("create table product_groups (\n    id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    parent_id INT REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по продукту\n    shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по магазину\n\n    url VARCHAR(255), -- url \"red-tea\" \"china-tea\" \"components\"\n    code VARCHAR(255), -- tea, coffee, china, ... какой-то уникальный (!!) код категории\n    name VARCHAR(255), -- имя каталога (тут можно добавить много других имен, в навигационном меню, например)\n    breadcrumb VARCHAR(255), -- текст в навигационной тепочке    \n    short_description VARCHAR(255), -- карткое описание раздела\n    description text, -- описание раздела\n    \n    meta_title VARCHAR(255), -- описание группы\n    meta_keywords VARCHAR(255), -- описание группы\n    meta_description VARCHAR(255), -- описание группы\n          \n     constraint uix_product_group_account_shop_parent_url_id UNIQUE (account_id,shop_id, parent_id, url),\n     constraint uix_product_group_account_shop_code_id UNIQUE (account_id,shop_id,code)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		log.Fatal("Cant create table product_group", err)
	}

	// Таблица продуктов
	err = pool.Exec("create table products (\n  id SERIAL PRIMARY KEY UNIQUE,\n     account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     product_group_id INT REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    -- article VARCHAR(32) NOT NULL, -- публичный артикул\n     sku VARCHAR(32), -- Stock Keeping Unit («складская учётная единица»)\n     url VARCHAR(255) NOT NULL, -- URL страницы\n     \n     name VARCHAR(255),\n  short_description VARCHAR(255), -- карткое описание раздела\n     description text, -- описание товара (32000 знаков)\n    -- constraint uix_products_article_account_id UNIQUE (article, account_id)\n     constraint uix_products_sku_account_id UNIQUE (sku, account_id)\n     \n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		log.Fatal("Cant create table products", err)
	}

	// Физически склады (Stocks). Объект принимает товары (приходы), списывает и т.д.
	err = pool.Exec("create table stocks (\n  id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    code VARCHAR(255), -- уникальный код склада\n    name VARCHAR(255) NOT NULL,    \n    address VARCHAR(255), -- потом можно более детально сделать адрес\n\n    -- created_at timestamp DEFAULT NOW(),\n    -- updated_at timestamp DEFAULT CURRENT_TIMESTAMP,\n    -- deleted_at timestamp DEFAULT null,\n        constraint uix_stocks_account_code_id UNIQUE (account_id, code)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		log.Fatal("Cant create table products", err)
	}

	// M:M Stock <> Product  Таблица продуктов
	err = pool.Exec("create table stock_products (\n    id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для ускорения поиска\n    stock_id INT NOT NULL REFERENCES stocks(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по складу\n    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по продукту\n\n    in_stock DECIMAL (13,3) DEFAULT 0.0, -- запас\n    stockpile DECIMAL (13,3) DEFAULT 0.0, -- резерв (зарезервировано) тут может быть условие, можно ли резеревировать больше остатка\n    \n    -- allow_reserve_out_of_stock BOOLEAN DEFAULT FALSE, -- можно ли резервировать больше реального запаса\n      \n     constraint uix_stock_products_account_store_product_id UNIQUE (account_id, stock_id, product_id)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		log.Fatal("Cant create table products", err)
	}


	//// Таблица оферов (чуть шире, чем товарные предложения, т.к. может быть несколько продуктов (2 по цене 1, наборы))
	err = pool.Exec("create table offers (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, \n\n  name VARCHAR(255), -- метка товарного предложения (\"в подрочной упаковке\", \"в разломе\", ...)\n  price DECIMAL (20,2) CONSTRAINT positive_price CHECK (price > 0), -- 2 знака после запятой\n  discount DECIMAL (20,2) DEFAULT 0.0 CONSTRAINT positive_discount CHECK ( discount <= price ) -- 2 знака после запятой \n\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	err = pool.Exec("create table product_cards (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  \n--   offers integer[][2],\n  \n  url VARCHAR(255),\n  breadcrumb VARCHAR(255),\n  short_description VARCHAR(255),\n  description text,\n  \n  -- meta group \n  meta_title VARCHAR (255),\n  meta_description VARCHAR (255),\n  meta_keywords VARCHAR (255)\n     -- constraint uix_products_article_account_id UNIQUE (article, account_id)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}*/

	// Таблица APIKey
	pool.CreateTable(&models.ApiKey{})
	pool.Exec("ALTER TABLE api_keys \n    ADD CONSTRAINT api_keys_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\ncreate unique index uix_api_keys_token_account_id ON api_keys (token,account_id);")

	// ##### EAV

	// [EAV_ATTR_TYPE] Таблица типов атрибутов EAV-модели. В зависимости от типа атрибута и его параметров он соответствующем образом обрабатывается во фронтенде и бэкенде.
	/*err = pool.Exec("create table  eav_attr_type (\n -- id int unsigned auto_increment,\n code varchar(32) primary key unique, -- json: text_field, text_area, date, Multiple Select...\n name varchar(32), -- label: Text Field, Text Area, Date, Multiple Select...\n \n    \n -- todo: добавить системные атрибуты типа, такие как: максимальная длина поля, минимальная длина поля, проверка при сохранении поля и т.д.\n -- min_len int default null,\n -- max_len int default null,\n table_name varchar(32) not null, -- имя таблицы, содержащие данные данного типа\n description varchar(255) -- описание типа.\n);\n").Error
	if err != nil {
		log.Fatal("Cant create table eav_attr_type", err)
	}

	// [eav_attributes] Таблица атрибутов EAV-модели.
	err = pool.Exec("create table  eav_attributes (\n id serial primary key unique,\n account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE , -- системные нафиг, все привязаны к аккаунту\n code VARCHAR(32), -- json: color, price, description. Уникальные значения только в рамках одного аккаунта!\n attr_type_code varchar(32) REFERENCES eav_attr_type(code) ON DELETE CASCADE ON UPDATE CASCADE, -- index !!!\n multiple BOOLEAN DEFAULT FALSE, -- множественный выбор (first() / findAll())\n label VARCHAR(32), -- label: Цвет, цена, описание\n required BOOLEAN DEFAULT FALSE,\n CONSTRAINT uix_eav_attributes_code_account_id UNIQUE (code, account_id) -- уникальные значения в рамках одного аккаунта\n \n);").Error
	if err != nil {
		log.Fatal("Cant create table eav_attributes: ", err)
	}*/

	// SMTP Settings
	models.Domain{}.PgSqlCreate()
	models.EmailBox{}.PgSqlCreate()
	models.EmailTemplate{}.PgSqlCreate()
	// models.EnvelopePublished{}.PgSqlCreate()

	// ### Создание таблиц для хранения значений атрибутов [VARCHAR, TEXT, DATE, BOOLEAN, INT, DECIMAL]

	// 1. [eav_attr_values_varchar] хранение значений атрибутов EAV-модели типа VARCHAR.
	//err = pool.Exec("create table  eav_varchar_values (\n id SERIAL PRIMARY KEY UNIQUE,\n eav_attr_id INT REFERENCES eav_attributes(id) ON DELETE CASCADE ON UPDATE CASCADE, -- внешний ключ, указывающий на владельца\n -- eav_attr_type_code varchar(32), # внешний ключ, указывающий на тип атрибута \n value varchar(255) default ''\n);").Error
	//if err != nil {
	//	log.Fatal("Cant create table eav_attr_varchar", err)
	//}

	/*// 2. [eav_attr_values_text] хранение значений атрибутов EAV-модели типа TEXT.
	_, err = pool.Exec("create or replace table  eav_attr_values_text (\n id int unsigned auto_increment primary key,\n eav_attr_id int unsigned, # внешний ключ, указывающий на владельца\n eav_attr_type_id int unsigned, # внешний ключ, указывающий на тип атрибута \n value text default '',\n foreign key (eav_attr_id) references eav_attributes(id) on delete cascade\n # foreign key (eav_attr_type_id) references eav_attr_type(id) on delete cascade\n);")
	if err != nil {
		fmt.Println("Cant create table eav_attr_text", err)
	}

	// 3. [eav_attr_values_date] хранение значений атрибутов EAV-модели типа DATE.
	_, err = pool.Exec("create or replace table  eav_attr_values_date (\n id int unsigned auto_increment primary key,\n eav_attr_id int unsigned, # внешний ключ, указывающий на владельца\n eav_attr_type_id int unsigned, # внешний ключ, указывающий на тип атрибута \n value date,\n foreign key (eav_attr_id) references eav_attributes(id) on delete cascade\n # foreign key (eav_attr_type_id) references eav_attr_type(id) on delete cascade\n);")
	if err != nil {
		fmt.Println("Cant create table eav_attr_date", err)
	}

	// 4. [eav_attr_values_boolean] хранение значений атрибутов EAV-модели типа BOOLEAN.
	_, err = pool.Exec("create or replace table  eav_attr_values_boolean (\n id int unsigned auto_increment primary key,\n eav_attr_id int unsigned, # внешний ключ, указывающий на владельца\n eav_attr_type_id int unsigned, # внешний ключ, указывающий на тип атрибута \n value boolean,\n foreign key (eav_attr_id) references eav_attributes(id) on delete cascade\n # foreign key (eav_attr_type_id) references eav_attr_type(id) on delete cascade\n);")
	if err != nil {
		fmt.Println("Cant create table eav_attr_boolean", err)
	}

	// 5. [eav_attr_values_int] хранение значений атрибутов EAV-модели типа INT.
	_, err = pool.Exec("create or replace table  eav_attr_values_int (\n id int unsigned auto_increment primary key,\n eav_attr_id int unsigned, # внешний ключ, указывающий на владельца\n eav_attr_type_id int unsigned, # внешний ключ, указывающий на тип атрибута \n value int,\n foreign key (eav_attr_id) references eav_attributes(id) on delete cascade\n # foreign key (eav_attr_type_id) references eav_attr_type(id) on delete cascade\n);")
	if err != nil {
		fmt.Println("Cant create table eav_attr_values_int", err)
	}

	// 6. [eav_attr_values_decimal] хранение значений атрибутов EAV-модели типа DECIMAL.
	_, err = pool.Exec("create or replace table  eav_attr_values_decimal (\n id int unsigned auto_increment primary key,\n eav_attr_id int unsigned, # внешний ключ, указывающий на владельца\n eav_attr_type_id int unsigned, # внешний ключ, указывающий на тип атрибута \n value decimal(20,10),\n foreign key (eav_attr_id) references eav_attributes(id) on delete cascade\n # foreign key (eav_attr_type_code) references eav_attr_type(code) on delete cascade\n);")
	if err != nil {
		fmt.Println("Cant create table decimal", err)
	}*/

	// ## ВНешние таблицы связи

	// M:M User <> Account

	// M:M Offer <> Product
	/*err = pool.Exec("create table offer_compositions (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  offer_id INT NOT NULL REFERENCES offers(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  \n  volume DECIMAL(13,3) NOT NULL DEFAULT 0.0 -- какой объем входит в оффер (шт, литры, граммы, кг и т.д.) \n\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	// M:M ProductCard <> Offer
	err = pool.Exec("create table product_card_offers (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  product_card_id INT NOT NULL REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  offer_id INT NOT NULL REFERENCES offers(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, \n  \n  \"order\" int,\n     constraint uix_product_card_offers_account_product_card_offer_id UNIQUE (account_id, product_card_id, offer_id)\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	models.Order{}.PgSqlCreate()*/

	// Загружаем стоковые данные для EAV таблиц
	// UploadEavData()

	// Аккаунты и тестовые продукты
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
	domainMain, err := mAcc.CreateDomain(models.Domain {
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



	////////////////////////////////////

	// 357 Грамм

	// 1. Создаем Василий
	vpopov, err := mAcc.CreateUser(
		models.User{
			Username:"antiglot",
			// Email:"vp@357gr.ru",
			Email:"mail-test@ratus-dev.ru",
			PhoneRegion: "RU",
			Phone: "89055294696",
			Password:"qwerty109#QW",
			Name:"Висилий",
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
	_, err = acc357.AppendUser(*owner,models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 4. Создаем домен для 357gr
	domain357gr, err := acc357.CreateDomain(models.Domain {
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
	_, err = accSyndicAd.AppendUser(*owner,models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2. Создаем домен для синдиката
	domainSynd, err := accSyndicAd.CreateDomain(models.Domain {
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
	_, err = brouser.AppendUser(*owner,models.RoleAdmin)
	if err != nil {
		log.Fatal("Не удалось добавить пользователя admin in 357gr")
		return
	}

	// 2. Создаем домен для синдиката
	domainBrouser, err := brouser.CreateDomain(models.Domain {
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
	data, err := ioutil.ReadFile("/var/www/ratuscrm/files/example.html")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}

	_, err = mAcc.CreateEmailTemplate(models.EmailTemplate{Name: "example", Body: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = acc357.CreateEmailTemplate(models.EmailTemplate{Name: "example", Body: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = accSyndicAd.CreateEmailTemplate(models.EmailTemplate{Name: "example", Body: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	_, err = brouser.CreateEmailTemplate(models.EmailTemplate{Name: "example", Body: string(data)})
	if err != nil {
		log.Fatal(err)
	}

	
	return

	// 1. Создаем пользователей
	//users := [] *models.User{
	//	{SignedAccountID:1, Username:"admin", Email:"kokorevn@gmail.com", Password:"qwerty109#QW", Name:"Никита", Surname:"Кокорев", Patronymic:"Романович", EmailVerifiedAt:&timeNow},
		//{SignedAccountID:1, Username:"nkokorev", Email:"mex388@gmail.com", Password:"qwerty109#QW", Name:"Никита", Surname:"Кокорев", Patronymic:"Романович", EmailVerifiedAt: &timeNow, InvitedUserID:1,},
		//{SignedAccountID:1, Username:"vpopov", Email:"vp@357gr.ru", Password:"qwerty109#QW", Name:"Василий", Surname:"Попов", Patronymic:"Николаевич", EmailVerifiedAt: &timeNow, InvitedUserID:1, },
		//{SignedAccountID:2, Username:"vpopov", Email:"vp@357gr.ru", Password:"qwerty109#QW", Name:"Василий", Surname:"Попов", Patronymic:"Николаевич", EmailVerifiedAt: &timeNow, InvitedUserID:1, },
	//}

	// 2. Аккаунты
	accounts := [] *models.Account{
		{Name:"RatusMedia"},
		{Name:"Rus Marketing"},
		{Name:"Stan-Prof"},
		{Name:"Vtvent"},
		{Name:"CS-Garant"},
		{Name:"357 грамм"},
		{Name:"SyndicAd"},
	}

	shops := [] *models.Shop{
		{AccountID:3, Name:"Магазин на Маяковке", Address:"Москва, ул. Долгоруковская, дом 9, стр. 3"},
	}

	product_groups := [] *models.ProductGroup{
		{AccountID:3, ShopID:1, Code:"root", URL:"/", Name:"Главная", Breadcrumb: "Главная", Description:""},

		{AccountID:3, ParentID:1, ShopID:1, Code:"tea", URL:"tea", Name:"Чай", Breadcrumb: "Чай", Description:""},
		{AccountID:3, ParentID:1, ShopID:1, Code:"coffee", URL:"coffee", Name:"Кофе", Breadcrumb: "Кофе", Description:""},
		{AccountID:3, ParentID:1, ShopID:1, Code:"gift", URL:"gift", Name:"Подарки", Breadcrumb: "Подарки", Description:""},
		{AccountID:3, ParentID:1, ShopID:1, Code:"accessories", URL:"accessories", Name:"Посуда и аксессуары", Breadcrumb: "Посуда и аксессуары", Description:""},

		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.puer", 	URL:"puer", 	Name:"Пуэр", Breadcrumb: "Пуэр", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.oolong",	URL:"oolong", 	Name:"Улунский чай", Breadcrumb: "Улунский чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.red", 	URL:"red", 		Name:"Красный чай", Breadcrumb: "Красный чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.green", 	URL:"green", 	Name:"Зеленый чай", Breadcrumb: "Зеленый чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.white", 	URL:"white", 	Name:"Белый чай", Breadcrumb: "Белый чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.yellow",	URL:"yellow", 	Name:"Желтый чай", Breadcrumb: "Желтый чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.herbal", 	URL:"herbal", 	Name:"Травяной чай", Breadcrumb: "Травяной чай", Description:""},
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.additives",URL:"additives",Name:"Чайные добавки", Breadcrumb: "Чайные добавки", Description:""},


		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.china", URL:"china", Name:"Китайский чай", Breadcrumb: "Китайский чай", Description:""}, // country = china & type = tea
		{AccountID:3, ParentID:2, ShopID:1, Code:"tea.taiwan", URL:"taiwan", Name:"Тайваньский чай", Breadcrumb: "Тайваньский чай", Description:""}, // country = taiwan & type = tea

		{AccountID:3, ParentID:5, ShopID:1, Code:"accessories.tableware.brewing", URL:"tableware-for-brewing", Name:"Посуда для заварки китайского чая", Breadcrumb: "Посуда для заварки китайского чая", Description:""}, // country = taiwan & type = tea

		{AccountID:3, ParentID:16, ShopID:1, Code:"accessories.tableware.brewing.gunfu", URL:"gunfu", Name:"Типоды (Гунфу)", Breadcrumb: "Типоды (Гунфу чайники)", Description:""}, // country = taiwan & type = tea

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

	pcs := [] *models.ProductCard{
		{AccountID:3,ShopID:1,URL:"teguanin"},
	}

	attributes := [] *models.EavAttribute{
		{Code:"size", Label:"Размер одежды", Multiple:false, Required:false, AttrTypeCode: "text_field"},
	}

	// 3. Стоковые атрибуты продуктов в аккаунте



	/*for i, _ := range accounts {
		if err := accounts[i].Create();err != nil {
			log.Fatal("Не удалось создать аккаунт", err)
		}
	}*/


	// Регистрируем пользователей в аккаунте RatusCRM
	// Как будто они регистрируются через общий вход: [POST]: ui.api.ratuscrm.com/accounts/{account_id = 1}/users
	// Нужна коллективная антиспам-защита. Храним большой список ip-адресов, с которых приходит спам. Возможно, проверяем HOST или еще что-то или может выдавать код.
	


	for _, r := range shops {
		if err := r.Create(); err != nil {
			log.Fatalf("Не удалось создать магазин для 357 грамм %v %v", r.Name, err)
			return
		}
	}

	for _, r := range product_groups {
		if err := r.Create(); err != nil {
			log.Fatalf("Не удалось группу для магазина 357 грамм %v %v", r.Name, err)
			return
		}
	}

	for _, r := range products {
		if err := r.Create(); err != nil {
			log.Fatalf("Не удалось создать продукт для 357 грамм %v %v", r.Name, err)
			return
		}
	}

	for _, r := range offers {
		if err := r.Create(); err != nil {
			log.Fatalf("Не удалось создать offer для 357 грамм %v %v ", r.Name, err)
			return
		}

	}

	if err := offers[0].ProductAppend(*products[10], 25.0); err != nil {
		log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[1].ProductAppend(*products[10], 50.0); err != nil {
		log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[2].ProductAppend(*products[10], 100.0); err != nil {
		log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[3].ProductAppend(*products[10], 100.0); err != nil {
		log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[3].ProductAppend(*products[11], 1.0); err != nil {
		log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
		return
	}

	for _, r := range pcs {
		if err := r.Create(); err != nil {
			log.Fatalf("Не удалось создать pcs для 357 грамм %v %v", r.URL, err)
			return
		}
	}

	for i,_ := range offers {
		if err := pcs[0].OfferAppend(*offers[i], i); err != nil {
			log.Fatalf("Не удалось добавить продукт в оффер, Error: %s", err)
			return
		}
	}

	for _, r := range attributes {
		if err := accounts[2].CreateEavAttribute(r); err != nil {
			log.Fatalf("Не удалось добавить атрибут %v в аккаунт, Error: %s", r.Label, err)
			return
		}
	}

/*	// 2. Создаем аккаунты (RatusMedia, Rus-Marketing, 357gr,... )
	err = pool.Exec("insert into accounts\n    (name, created_at)\nvalues\n    ('RatusMedia', NOW()),\n    ('Rus Marketing', NOW()),\n    ('357 грамм', NOW())\n").Error
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_type: ", err)
	}*/


}
