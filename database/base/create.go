package base

import (
	"fmt"
	"github.com/nkokorev/crm-go/models"
	"log"
)

func RefreshTables() {

	// пересоздаем БД
	//ReplaceDataBase(os.Getenv("db_name"))

	var err error
	pool := models.GetPool()

	// дропаем системные таблицы
	err = pool.Exec("drop table if exists eav_attributes, eav_attr_type, api_keys, user_accounts, product_card_offers, offers, offer_compositions, product_cards, product_groups, stock_products, stocks, shops, products, accounts, users").Error
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// Таблица типов атрибутов EAV-модели. В зависимости от типа атрибута и его параметров он соответствующем образом обрабатывается во фронтенде и бэкенде.
	err = pool.Exec("create table  users (\n id SERIAL PRIMARY KEY UNIQUE,\n username varchar(32) NOT NULL UNIQUE,\n email varchar(60) NOT NULL UNIQUE,\n password varchar(255) NOT NULL UNIQUE,\n \n name varchar(32) DEFAULT '',\n surname varchar(32) DEFAULT '',\n patronymic varchar(32) DEFAULT '',\n \n default_account_id INT DEFAULT NULL,\n created_at timestamp DEFAULT NOW(),\n updated_at timestamp DEFAULT CURRENT_TIMESTAMP,\n deleted_at timestamp DEFAULT NULL\n);\n").Error
	if err != nil {
		fmt.Println("Cant create table users", err)
	}

	// Таблица аккаунтов.
	err = pool.Exec("create table  accounts (\n id SERIAL PRIMARY KEY UNIQUE,\n name varchar(32),\n created_at timestamp DEFAULT NOW(),\n updated_at timestamp DEFAULT CURRENT_TIMESTAMP,\n deleted_at timestamp DEFAULT null\n);\n").Error
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// Магазины (Shops).
	err = pool.Exec("create table shops (\n  id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    name VARCHAR(255) NOT NULL, -- имя магазина    \n    address VARCHAR(255) -- потом можно более детально сделать адрес\n \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table shops", err)
	}

	err = pool.Exec("create table product_groups (\n    id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    parent_id INT REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по продукту\n    shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по магазину\n\n    url VARCHAR(255), -- url \"red-tea\" \"china-tea\" \"components\"\n    code VARCHAR(255), -- tea, coffee, china, ... какой-то уникальный (!!) код категории\n    name VARCHAR(255), -- имя каталога (тут можно добавить много других имен, в навигационном меню, например)\n    breadcrumb VARCHAR(255), -- текст в навигационной тепочке    \n    short_description VARCHAR(255), -- карткое описание раздела\n    description text, -- описание раздела\n    \n    meta_title VARCHAR(255), -- описание группы\n    meta_keywords VARCHAR(255), -- описание группы\n    meta_description VARCHAR(255), -- описание группы\n          \n     constraint uix_product_group_account_shop_parent_url_id UNIQUE (account_id,shop_id, parent_id, url),\n     constraint uix_product_group_account_shop_code_id UNIQUE (account_id,shop_id,code)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table product_group", err)
	}

	// Таблица продуктов
	err = pool.Exec("create table products (\n  id SERIAL PRIMARY KEY UNIQUE,\n     account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     product_group_id INT REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    -- article VARCHAR(32) NOT NULL, -- публичный артикул\n     sku VARCHAR(32), -- Stock Keeping Unit («складская учётная единица»)\n     url VARCHAR(255) NOT NULL, -- URL страницы\n     \n     name VARCHAR(255),\n  short_description VARCHAR(255), -- карткое описание раздела\n     description text, -- описание товара (32000 знаков)\n    -- constraint uix_products_article_account_id UNIQUE (article, account_id)\n     constraint uix_products_sku_account_id UNIQUE (sku, account_id)\n     \n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	// Физически склады (Stocks). Объект принимает товары (приходы), списывает и т.д.
	err = pool.Exec("create table stocks (\n  id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    code VARCHAR(255), -- уникальный код склада\n    name VARCHAR(255) NOT NULL,    \n    address VARCHAR(255), -- потом можно более детально сделать адрес\n\n    -- created_at timestamp DEFAULT NOW(),\n    -- updated_at timestamp DEFAULT CURRENT_TIMESTAMP,\n    -- deleted_at timestamp DEFAULT null,\n        constraint uix_stocks_account_code_id UNIQUE (account_id, code)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}



	// M:M Stock <> Product  Таблица продуктов
	err = pool.Exec("create table stock_products (\n    id SERIAL PRIMARY KEY UNIQUE,\n    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для ускорения поиска\n    stock_id INT NOT NULL REFERENCES stocks(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по складу\n    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, -- поиск по продукту\n\n    in_stock DECIMAL (13,3) DEFAULT 0.0, -- запас\n    stockpile DECIMAL (13,3) DEFAULT 0.0, -- резерв (зарезервировано) тут может быть условие, можно ли резеревировать больше остатка\n    \n    -- allow_reserve_out_of_stock BOOLEAN DEFAULT FALSE, -- можно ли резервировать больше реального запаса\n      \n     constraint uix_stock_products_account_store_product_id UNIQUE (account_id, stock_id, product_id)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}





	// склад (stock)
	// интернет-магазин (store)
	// витрина (store_view)
	// карточка товара (product_card)


	//// Таблица оферов (чуть шире, чем товарные предложения, т.к. может быть несколько продуктов (2 по цене 1, наборы))
	err = pool.Exec("create table offers (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, \n\n  name VARCHAR(255), -- метка товарного предложения (\"в подрочной упаковке\", \"в разломе\", ...)\n  price DECIMAL (20,2) CONSTRAINT positive_price CHECK (price > 0), -- 2 знака после запятой\n  discount DECIMAL (20,2) DEFAULT 0.0 CONSTRAINT positive_discount CHECK ( discount <= price ) -- 2 знака после запятой \n\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}


	err = pool.Exec("create table product_cards (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  shop_id INT NOT NULL REFERENCES shops(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  \n--   offers integer[][2],\n  \n  url VARCHAR(255),\n  breadcrumb VARCHAR(255),\n  short_description VARCHAR(255),\n  description text,\n  \n  -- meta group \n  meta_title VARCHAR (255),\n  meta_description VARCHAR (255),\n  meta_keywords VARCHAR (255)\n     -- constraint uix_products_article_account_id UNIQUE (article, account_id)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	// Таблица APIKey
	err = pool.Exec("create table api_keys (\n  token VARCHAR(255) PRIMARY KEY UNIQUE,\n  account_id int NOT NULL REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE,\n  name VARCHAR(255) default '',\n  status BOOLEAN NOT NULL DEFAULT TRUE,\n  created_at timestamp default NOW(),\n  updated_at timestamp default CURRENT_TIMESTAMP,\n    constraint uix_api_keys_token_account_id UNIQUE (token, account_id)\n     -- foreign key (account_id) references accounts(id) on delete cascade\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table api_keys", err)
	}

	// ##### EAV

	// [EAV_ATTR_TYPE] Таблица типов атрибутов EAV-модели. В зависимости от типа атрибута и его параметров он соответствующем образом обрабатывается во фронтенде и бэкенде.
	err = pool.Exec("create table  eav_attr_type (\n -- id int unsigned auto_increment,\n code varchar(32) primary key unique, -- json: text_field, text_area, date, Multiple Select...\n name varchar(32), -- label: Text Field, Text Area, Date, Multiple Select...\n \n    \n -- todo: добавить системные атрибуты типа, такие как: максимальная длина поля, минимальная длина поля, проверка при сохранении поля и т.д.\n -- min_len int default null,\n -- max_len int default null,\n table_name varchar(32) not null, -- имя таблицы, содержащие данные данного типа\n description varchar(255) -- описание типа.\n);\n").Error
	if err != nil {
		log.Fatal("Cant create table eav_attr_type", err)
	}

	// [eav_attributes] Таблица атрибутов EAV-модели.
	err = pool.Exec("create table  eav_attributes (\n id serial primary key unique,\n account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE , -- системные нафиг, все привязаны к аккаунту\n code VARCHAR(32), -- json: color, price, description. Уникальные значения только в рамках одного аккаунта!\n attr_type_code varchar(32) REFERENCES eav_attr_type(code) ON DELETE CASCADE ON UPDATE CASCADE, -- index !!!\n multiple BOOLEAN DEFAULT FALSE, -- множественный выбор (first() / findAll())\n label VARCHAR(32), -- label: Цвет, цена, описание\n required BOOLEAN DEFAULT FALSE,\n CONSTRAINT uix_eav_attributes_code_account_id UNIQUE (code, account_id) -- уникальные значения в рамках одного аккаунта\n \n);").Error
	if err != nil {
		log.Fatal("Cant create table eav_attributes: ", err)
	}

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
	err = pool.Exec("create table user_accounts (\n    user_id INT REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n    account_id INT REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n    constraint uix_user_accounts_user_account_id UNIQUE (user_id, account_id)\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// M:M Offer <> Product
	err = pool.Exec("create table offer_compositions (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  offer_id INT NOT NULL REFERENCES offers(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n  \n  volume DECIMAL(13,3) NOT NULL DEFAULT 0.0 -- какой объем входит в оффер (шт, литры, граммы, кг и т.д.) \n\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	// M:M ProductCard <> Offer
	err = pool.Exec("create table product_card_offers (\n  id SERIAL PRIMARY KEY UNIQUE,\n  account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  product_card_id INT NOT NULL REFERENCES product_cards(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  offer_id INT NOT NULL REFERENCES offers(id) ON DELETE CASCADE ON UPDATE CASCADE, -- для скорост выборки\n  -- product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE, \n  \n  \"order\" int,\n     constraint uix_product_card_offers_account_product_card_offer_id UNIQUE (account_id, product_card_id, offer_id)\n);\n\n").Error
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	//// M:M Products <> Attributes
	//err = pool.Exec("create table eav_product_offer_attributes (\n     account_id INT REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- обязательно, чтобы ускорить выборку\n     product_offer_id INT REFERENCES product_offers(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     eav_attributes_id INT REFERENCES eav_attributes(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     constraint uix_eav_product_attributes_account_product_eav_attributes_id unique (account_id, product_offer_id, eav_attributes_id)\n);\n\n").Error
	//if err != nil {
	//	fmt.Println("Cant create table accounts", err)
	//}

	// M:M Products <> Varchar values
	// err = pool.Exec("create table eav_product_values_varchar (\n     product_id INT REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n     eav_varchar_value_id INT REFERENCES eav_varchar_values(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     constraint uix_eav_product_values_varchar_product_value_id unique (product_id, eav_varchar_value_id)\n     \n);\n\n").Error
	//if err != nil {
	//	fmt.Println("Cant create table accounts", err)
	//}

	// загружаем стоковые данные для EAV таблиц
	UploadEavData()

	// аккаунты и тестовые продукты
	UploadTestData()


}

func ReplaceDataBase(name string)  {

	var err error
	pool := models.GetPool()

	err = pool.Exec("DROP DATABASE IF EXISTS " + name + ";").Error
	if err != nil {
		log.Fatalf("Cant drop database: %s, err: %s", name, err)
	}
	// пересоздаем создаем базу данных
	err = pool.Exec("CREATE DATABASE " + name + " OWNER crm ENCODING UTF8;").Error
	if err != nil {
		log.Fatalf("Cant create database: %s, err: %s", name, err)
	}

	// Выбираем тестовую базу и в нее заносим все изменения
	/*if _, err = pool.Exec("select " + name); err != nil {
		log.Fatal("Cant set test data base: ", err)
	}*/
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

	// 1. Создаем пользователей
	users := [] *models.User{
		{Username:"admin", Email:"kokorevn@gmail.com", Password:"qwerty109#QW", Name:"Никита", Surname:"Кокорев", Patronymic:"Романович"},
		{Username:"nkokorev", Email:"mex388@gmail.com", Password:"qwerty109#QW", Name:"Никита", Surname:"Кокорев", Patronymic:"Романович"},
		{Username:"vpopov", Email:"vp@357gr.ru", Password:"qwerty109#QW", Name:"Василий", Surname:"Попов", Patronymic:"Николаевич"},
	}

	// 2. Аккаунты
	accounts := [] *models.Account{
		{Name:"RatusMedia"},
		{Name:"Rus Marketing"},
		{Name:"357 грамм"},
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

	for i,_ := range users {

		if err := users[i].Create(); err != nil {
			log.Fatalf("Неудалось создать базового пользователя: %v, Error: %s", users[i], err)
			return
		}

		if err := users[i].CreateAccount(accounts[i]); err != nil {
			log.Fatalf("Неудалось создать аккаунт: %v, Error: %s", accounts[i], err)
			return
		}

		if err := accounts[i].AppendUser(users[0]); err != nil {
			log.Fatalf("Неудалось добавить админа в аккаунт: %v, Error: %s", accounts[i], err)
			return
		}

		apiKey := &models.ApiKey{Name:"Key for site", Status:true}
		if err := accounts[i].CreateApiToken(apiKey); err != nil {
			log.Fatalf("Неудалось создать API ключ для аккаунта: %v, Error: %s", accounts[i], err)
			return
		}

	}

	for _, r := range shops {
		if err := r.Create(); err != nil {
			log.Fatalf("Неудалось создать магазин для 357 грамм", r.Name, err)
			return
		}
	}

	for _, r := range product_groups {
		if err := r.Create(); err != nil {
			log.Fatalf("Неудалось группу для магазина 357 грамм", r.Name, err)
			return
		}
	}

	for _, r := range products {
		if err := r.Create(); err != nil {
			log.Fatalf("Неудалось создать продукт для 357 грамм", r.Name, err)
			return
		}
	}



	for _, r := range offers {
		if err := r.Create(); err != nil {
			log.Fatalf("Неудалось создать offer для 357 грамм", r.Name, err)
			return
		}

	}



	if err := offers[0].ProductAppend(*products[10], 25.0); err != nil {
		log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[1].ProductAppend(*products[10], 50.0); err != nil {
		log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[2].ProductAppend(*products[10], 100.0); err != nil {
		log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[3].ProductAppend(*products[10], 100.0); err != nil {
		log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
		return
	}
	if err := offers[3].ProductAppend(*products[11], 1.0); err != nil {
		log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
		return
	}

	for _, r := range pcs {
		if err := r.Create(); err != nil {
			log.Fatalf("Неудалось создать pcs для 357 грамм", r.URL, err)
			return
		}
	}

	for i,_ := range offers {
		if err := pcs[0].OfferAppend(*offers[i], i); err != nil {
			log.Fatalf("Неудалось добавить продукт в оффер, Error: %s", err)
			return
		}
	}




	for _, r := range attributes {
		if err := accounts[2].CreateEavAttribute(r); err != nil {
			log.Fatalf("Неудалось добавить атрибут %v в аккаунт, Error: %s", r.Label, err)
			return
		}
	}

/*	// 2. Создаем аккаунты (RatusMedia, Rus-Marketing, 357gr,... )
	err = pool.Exec("insert into accounts\n    (name, created_at)\nvalues\n    ('RatusMedia', NOW()),\n    ('Rus Marketing', NOW()),\n    ('357 грамм', NOW())\n").Error
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_type: ", err)
	}*/


}

