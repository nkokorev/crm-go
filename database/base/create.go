package base

import (
	"fmt"
	"log"

)

func RefreshTables() {

	// пересоздаем БД
	//ReplaceDataBase(os.Getenv("db_name"))

	var err error
	pool := GetPool()

	// дропаем системные таблицы
	_, err = pool.Exec("drop table if exists eav_product_attributes, eav_product_values_varchar, eav_varchar_values, eav_attributes, eav_attr_type, api_keys, products, accounts")
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}


	// Таблица типов атрибутов EAV-модели. В зависимости от типа атрибута и его параметров он соответствующем образом обрабатывается во фронтенде и бэкенде.
	_, err = pool.Exec("create table  accounts (\n id SERIAL PRIMARY KEY UNIQUE,\n name varchar(32),\n created_at timestamp DEFAULT NOW(),\n updated_at timestamp DEFAULT CURRENT_TIMESTAMP,\n deleted_at timestamp DEFAULT null\n);\n")
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// Таблица продуктов
	_, err = pool.Exec("create table products (\n  id SERIAL PRIMARY KEY UNIQUE,\n     account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     sku varchar(32) default '',\n     name varchar(255) default '',\n     constraint uix_products_sku_account_id unique (sku, account_id)\n     -- foreign key (account_id) references accounts(id) ON DELETE CASCADE \n);\n\n")
	if err != nil {
		fmt.Println("Cant create table products", err)
	}

	// Таблица APIKey
	_, err = pool.Exec("create table api_keys (\n  id SERIAL PRIMARY KEY UNIQUE,\n     token varchar(32) not null unique,\n     account_id int NOT NULL REFERENCES accounts (id) ON DELETE CASCADE ON UPDATE CASCADE,\n     name varchar(255) default '',\n     status boolean default true,\n     created_at timestamp default NOW(),\n     updated_at timestamp default CURRENT_TIMESTAMP,\n     constraint uix_api_key_token_account_id unique (token, account_id)\n     -- foreign key (account_id) references accounts(id) on delete cascade\n);\n\n")
	if err != nil {
		fmt.Println("Cant create table api_keys", err)
	}

	// ##### EAV

	// [EAV_ATTR_TYPE] Таблица типов атрибутов EAV-модели. В зависимости от типа атрибута и его параметров он соответствующем образом обрабатывается во фронтенде и бэкенде.
	_, err = pool.Exec("create table  eav_attr_type (\n -- id int unsigned auto_increment,\n code varchar(32) primary key unique, -- json: text_field, text_area, date, Multiple Select...\n name varchar(32), -- label: Text Field, Text Area, Date, Multiple Select...\n \n    \n -- todo: добавить системные атрибуты типа, такие как: максимальная длина поля, минимальная длина поля, проверка при сохранении поля и т.д.\n -- min_len int default null,\n -- max_len int default null,\n table_name varchar(32) not null, -- имя таблицы, содержащие данные данного типа\n description varchar(255) -- описание типа.\n);\n")
	if err != nil {
		log.Fatal("Cant create table eav_attr_type", err)
	}

	// [eav_attributes] Таблица атрибутов EAV-модели.
	_, err = pool.Exec("create table  eav_attributes (\n id serial primary key unique,\n -- account_id INT DEFAULT 0 REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE, -- index !!! null == system\n code varchar(32), -- json: color, price, description. Уникальные значения только в рамках одного аккаунта!\n eav_attr_type_code varchar(32) REFERENCES eav_attr_type(code) ON DELETE CASCADE ON UPDATE CASCADE, -- index !!!\n multiple boolean default false, -- множественный выбор (first() / findAll())\n label varchar(32), -- label: Цвет, цена, описание\n required BOOLEAN DEFAULT FALSE\n -- constraint uix_eav_attributes_code_account_id unique (code, account_id) -- уникальные значения в рамках одного аккаунта\n \n);")
	if err != nil {
		log.Fatal("Cant create table eav_attributes: ", err)
	}

	// ### Создание таблиц для хранения значений атрибутов [VARCHAR, TEXT, DATE, BOOLEAN, INT, DECIMAL]

	// 1. [eav_attr_values_varchar] хранение значений атрибутов EAV-модели типа VARCHAR.
	_, err = pool.Exec("create table  eav_varchar_values (\n id SERIAL PRIMARY KEY UNIQUE,\n eav_attr_id INT REFERENCES eav_attributes(id) ON DELETE CASCADE ON UPDATE CASCADE, -- внешний ключ, указывающий на владельца\n -- eav_attr_type_code varchar(32), # внешний ключ, указывающий на тип атрибута \n value varchar(255) default ''\n);")
	if err != nil {
		log.Fatal("Cant create table eav_attr_varchar", err)
	}

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

	// M:M Products <> Attributes
	_, err = pool.Exec("create table eav_product_attributes (\n     product_id INT REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n     eav_attributes_id INT REFERENCES eav_attributes(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n     constraint uix_eav_product_attributes_product_account_id unique (product_id, eav_attributes_id)\n);\n\n")
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// M:M Products <> Varchar values
	_, err = pool.Exec("create table eav_product_values_varchar (\n     product_id INT REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE ,\n     eav_varchar_value_id INT REFERENCES eav_varchar_values(id) ON DELETE CASCADE ON UPDATE CASCADE,\n     constraint uix_eav_product_values_varchar_product_value_id unique (product_id, eav_varchar_value_id)\n     \n);\n\n")
	if err != nil {
		fmt.Println("Cant create table accounts", err)
	}

	// загружаем стоковые данные для EAV таблиц
	UploadEavData()

	// аккаунты и тестовые продукты
	UploadTestData()


}

func ReplaceDataBase(name string)  {

	var err error
	pool := GetPool()

	_, err = pool.Exec("DROP DATABASE IF EXISTS " + name + ";")
	if err != nil {
		log.Fatalf("Cant drop database: %s, err: %s", name, err)
	}
	// пересоздаем создаем базу данных
	_, err = pool.Exec("CREATE DATABASE " + name + " OWNER crm ENCODING UTF8;")
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
	pool := GetPool()

	// Добавляем в таблицу типов атрибутов EAV-моделей используемы типы (7). Частные типы - не предпологаются.
	_, err = pool.Exec("insert into eav_attr_type\n    (name, code, table_name, description)\nvalues\n    ('Текстовое поле', 'text_field', 'eav_attr_varchar', 'Текстовое поле для хранения кратких текстовых данных до 255 символов.'),\n    ('Текстовая область', 'text_area', 'eav_attr_varchar', 'Текстовая область для хранения кратких текстовых данных до 255 символов.'),\n    ('Текстовый редактор', 'text_editor', 'eav_attr_text', 'Редактируемый wysiwyg-редактором текст до 16383 символов.'),\n    ('Дата', 'date', 'eav_attr_date', 'Дата в формате UTC.'),\n    ('Да / Нет', 'bool', 'eav_attr_boolean', 'Логический формат данных, который может принимать значение ИСТИНА (1) и ЛОЖЬ (0).'),\n    ('Целое число', 'int', 'eav_attr_int', 'Целое число от -2147483648 до 2147483648.'),\n    ('Десятичное число', 'decimal', 'eav_attr_decimal', 'Знаковое десятичное число, 10 знаков до и после запятой.');")
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_type: ", err)
	}

	// Добавляем в таблицу атрибутов EAV-моделей системные атрибуты.
	_, err = pool.Exec("insert into eav_attributes\n    (eav_attr_type_code, label, code, required, multiple)\nvalues\n    ('text_field', 'Имя продукта', 'name', false, false),\n    ('text_field', 'Производитель', 'manufactures', false, false),\n    ('text_editor', 'Описание', 'description', false, false),\n    ('decimal', 'Цена', 'price', false, false),\n    ('date', 'Дата производства', 'manufacture_date', false, false),\n    ('text_field', 'Размер одежды', 'clothing_size', false, true),\n    ('text_field', 'Тип упаковки', 'pkg_type', false, true),\n    ('text_field', 'Состав', 'composition', false, false)\n    ")
	if err != nil {
		log.Fatal("Cant insert into table eav_attributes: ", err)
	}

	// загружаем значения varchar
	_, err = pool.Exec("insert into eav_varchar_values\n    (eav_attr_id, value)\nvalues\n    (6, 'S'), -- Размер одежды\n    (6, 'M'), -- Размер одежды\n    (6, 'L'), -- Размер одежды\n    (7, 'Без упаковки (без упаковки)'), -- Тип упаковки\n    (7, 'Подарочный пакет'), -- Тип упаковки\n    (8, 'хлопок 90%, эластан 10%'),-- Состав\n    (8, 'вискоза 89%, эластан 11%'), -- Состав\n    (8, 'вискоза 89%, эластан 11%'), -- Состав\n    (8, 'хлопок 100%') -- Состав\n    ")
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_values_varchar: ", err)
	}

}

func UploadTestData() {
	var err error
	pool := GetPool()

	// Добавляем основные аккаунты
	_, err = pool.Exec("insert into accounts\n    (name, created_at)\nvalues\n    ('RatusMedia', NOW()),\n    ('Rus Marketing', NOW()),\n    ('357 грамм', NOW())\n")
	if err != nil {
		log.Fatal("Cant insert into table eav_attr_type: ", err)
	}
}

