package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"time"
)

type Storage struct {

	ID     uint   `json:"-" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`
	
	Name string `json:"name" gorm:"type:varchar(255);"`
	Data []byte `json:"data" gorm:"type:bytea;"`

	// MetaData
	MIME 	string 	`json:"mime" gorm:"type:varchar(90);"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (Storage) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Storage{})
	db.Exec("ALTER TABLE storage \n ADD CONSTRAINT file_storage_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")

}

func (Storage) TableName() string {
	return "storage"
}

func (fs *Storage) BeforeCreate(scope *gorm.Scope) error {
	fs.ID = 0
	fs.HashID = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	fs.CreatedAt = time.Now().UTC()

	return nil
}

func (fs Storage) create() (*Storage, error)  {
	err := db.Create(&fs).Error
	return &fs, err
}

func (Storage) get(id uint) (*Storage, error)  {

	fs := Storage{}

	err := db.First(&fs, id).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

func (Storage) getByHashId(hashId string) (*Storage, error)  {

	fs := Storage{}

	err := db.First(&fs, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}

// ########### ACCOUNT FUNCTIONAL ###########

func (account Account) StorageCreate(fs *Storage) (*Storage, error) {

	fs.AccountID = account.ID
	return fs.create()
}

func (account Account) StorageGet(id uint) (*Storage, error) {

	fs, err := (Storage{}).get(id)
	if err != nil {
		return nil, err
	}

	if fs.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return fs, nil
}

func (account Account) StorageGetByHashId(hashId string) (*Storage, error) {

	fs, err := (Storage{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if fs.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return fs, nil
}

func (Account) StorageGetPublicByHashId(hashId string) (*Storage, error) {

	fs, err := (Storage{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}
	
	return fs, nil
}

// ########### END OF ACCOUNT FUNCTIONAL ###########