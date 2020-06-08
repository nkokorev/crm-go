package models

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"time"
)

type Storage struct {

	ID     uint   `json:"-" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID uint `json:"-" gorm:"type:int;index;not_null;"`
	
	Name string `json:"name" gorm:"type:varchar(255);"` // имя файла
	Data []byte `json:"data" gorm:"type:bytea;"` // тело файла

	// MetaData
	MIME 	string 	`json:"mime" gorm:"type:varchar(90);"` // мета тип файла
	Size 	uint 	`json:"size" gorm:"type:int;"` // Kb

	// Назначение файла
	Purpose	uint 	`json:"purpose"` // 1 - free, 2 - products, 3 - emails,

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

func (fs *Storage) update(input interface{}) error {
	return db.Model(fs).Omit("id", "hashId", "account_id","created_at", "updated_at").Update(structs.Map(input)).Error
}

func (fs Storage) Delete () error {
	return db.Model(Storage{}).Where("id = ?", fs.ID).Delete(fs).Error
}

// ########### ACCOUNT FUNCTIONAL ###########
func (account Account) StorageCreateFile(fs *Storage) (*Storage, error) {
	// check disk space
	used, err := account.StorageDiskSpaceUsed()
	if err != nil {
		return nil, err
	}

	if (account.DiskSpaceAvailable - used) < fs.Size {
		return nil, utils.Error{Message: "Нехватка дискового пространства"}
	}
	
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

func (account Account) StorageGetFiles(limit, offset int) ([]Storage, error) {

	var files []Storage

	err := db.Limit(limit).Offset(offset).
		Find(&files, "account_id = ?", account.ID).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Println("Ошибка получения списка файлов")
		return nil, err
	}

	return files, nil
}

func (account Account) StorageGetList() ([]Storage, error) {

	var files []Storage

	// without data
	err := db.Select([]string{"id", "hash_id", "account_id", "name", "mime", "size", "updated_at", "created_at"}).Find(&files, "account_id = ?", account.ID).Error
	// err := db.Limit(limit).Offset(offset).Find(&files, "account_id = ?", account.ID).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Println("Ошибка получения списка файлов")
		return nil, err
	}

	return files, nil
}

func (account Account) StorageUpdateFile(fs *Storage, input interface{}) error {

	if fs.AccountID != account.ID {
		return errors.New("Файл принадлежит другому аккаунту")
	}

	return fs.update(input)
}

func (account Account) StorageDiskSpaceUsed() (uint, error) {
	var sum,count uint

	sum = 0
	count = 0
	err := db.Table("storage").Where("account_id = ?", account.ID).Count(&count).Error
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	err = db.Table("storage").Where("account_id = ?", account.ID).Select("sum(size)").Row().Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, nil
}

func (Account) StorageGetPublicByHashId(hashId string) (*Storage, error) {

	fs, err := (Storage{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}
	
	return fs, nil
}

// ########### END OF ACCOUNT FUNCTIONAL ###########