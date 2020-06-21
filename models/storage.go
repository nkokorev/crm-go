package models

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"os"
	"strings"
	"time"
)

type Storage struct {

	ID     		uint   `json:"id" gorm:"primary_key"`
	HashID 		string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID 	uint `json:"-" gorm:"type:int;index;not null;"`
	ProductId 	uint	`json:"productId" gorm:"type:int;default:null;"` // id of products
	EmailId 	uint	`json:"emailId" gorm:"type:int;default:null;"` // id of email template

	Priority 	uint		`json:"priority" gorm:"type:int;default:null;"` // Порядок отображения (часто нужно файлам)
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // выводить ли где-то это изображение или нет

	Name 				string `json:"name" gorm:"type:varchar(255);"` // имя файла (оно же при отдаче)
	Alt 				string `json:"alt" gorm:"type:varchar(255);"` // alt для изображений
	ShortDescription 	string `json:"shortDescription" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 		string `json:"description" gorm:"type:text;"` // pgsql: text // большое описание изображения (не, ну мало ли фанаты фото)

	// MetaData
	MIME 		string 	`json:"mime" gorm:"type:varchar(90);"` // мета тип файла
	Size 		uint 	`json:"size" gorm:"type:int;"` // Kb

	Data 		[]byte `json:"data" gorm:"type:bytea;"` // тело файла
	URL 		string 	`json:"url" sql:"-"` // see AfterFind

	CreatedAt 	time.Time  `json:"createdAt"`
	UpdatedAt 	time.Time  `json:"updatedAt"`
}

func (Storage) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Storage{})
	db.Exec("ALTER TABLE storage \n ADD CONSTRAINT storage_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--  ADD CONSTRAINT storage_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE ON UPDATE CASCADE,\n--  ADD CONSTRAINT storage_email_template_id_fkey FOREIGN KEY (email_id) REFERENCES email_templates(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\n-- create unique index uix_storage_product_id ON storage (account_id,product_id) WHERE product_id IS NOT NULL;\n-- create unique index uix_storage_email_template_id ON storage (account_id,email_id) WHERE email_id IS NOT NULL;\n")

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

func (fs *Storage) AfterFind() (err error) {

	// todo: дописать формирование url
	// 1. Добавлям URL в зависимости от типа файла:''
	AppEnv := os.Getenv("APP_ENV")
	crmHost := ""
	switch AppEnv {
		case "local":
			crmHost = "http://cdn.crm.local"
		case "public":
			crmHost = "https://cdn.ratuscrm.com"
		default:
			crmHost = "https://cdn.ratuscrm.com"
	}
	

	if fs.ProductId > 0 {
		fs.URL = crmHost + "/products/images/" + fs.HashID
	} else {
		if fs.EmailId > 0 {
			fs.URL = crmHost + "/emails/images/" + fs.HashID
		} else {
			fs.URL = crmHost + "/public/" + fs.HashID
		}
	}

	return nil
}

func (fs Storage) create() (*Storage, error)  {
	err := db.Create(&fs).First(&fs).Error
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
	// fmt.Println(input)
	// fmt.Println("Input: ", input)
	// return db.Model(fs).Omit("id", "hashId", "account_id","created_at", "updated_at").Updates(structs.Map(input)).Error
	return db.Model(fs).Omit("id", "hashId", "account_id","created_at", "updated_at").Updates(input).Error
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
	file, err := fs.create()
	if err != nil {
		return nil, err
	}

	if file.ProductId > 0 {
		go account.CallWebHookIfExist(EventProductUpdated, Product{ID: file.ProductId})
	}

	return file, nil
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

func (account Account) StorageGetList(offset, limit uint, search string, productId, emailId *uint) ([]Storage, uint, error) {

	files := make([]Storage,0)

	var err error
	
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"


		// Выборку по файлам
		if *productId > 0 {
			err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
				Where("account_id = ? AND product_id = ?", account.ID, productId).
				Find(&files, "name ILIKE ? OR short_description ILIKE ? OR description ILIKE ?" , search,search,search).Error
		} else {
			if *emailId > 0 {
				err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
					Where("account_id = ? AND email_id = ?", account.ID, emailId).
					Find(&files, "name ILIKE ? OR short_description ILIKE ? OR description ILIKE ?" , search,search,search).Error
			} else {
				err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
					Where("account_id = ?", account.ID).
					Find(&files, "name ILIKE ? OR short_description ILIKE ? OR description ILIKE ?" , search,search,search).Error
			}
		}

		// correction not found res
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		// Выборку по файлам
		if *productId > 0 {
			err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
				Find(&files, "account_id = ? AND product_id = ?", account.ID, productId).Error
		} else {
			if *emailId > 0 {
				err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
					Find(&files, "account_id = ? AND email_id = ?", account.ID, emailId).Error
			} else {
				err = db.Model(&Storage{}).Limit(limit).Offset(offset).Select(Storage{}.SelectArrayWithoutData()).
					Find(&files, "account_id = ?", account.ID).Error
			}
		}

		// correction not found res
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}
	
	var total uint
	if *productId > 0 {
		err = db.Model(&Storage{}).Select(Storage{}.SelectArrayWithoutData()).Where("account_id = ? AND product_id = ?", account.ID, productId).Count(&total).Error
	} else {
		if *emailId > 0 {
			err = db.Model(&Storage{}).Select(Storage{}.SelectArrayWithoutData()).Where("account_id = ? AND email_id = ?", account.ID, emailId).Count(&total).Error
		} else {
			err = db.Model(&Storage{}).Select(Storage{}.SelectArrayWithoutData()).Where("account_id = ?", account.ID).Count(&total).Error
		}
	}
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема"}
	}

	return files, total, nil
}

func (account Account) StorageUpdateFile(file *Storage, input interface{}) error {

	if file.AccountID != account.ID {
		return errors.New("Файл принадлежит другому аккаунту")
	}

	if err := file.update(input); err != nil {
		return err
	}

	if file.ProductId > 0 {
		go account.CallWebHookIfExist(EventProductUpdated, Product{ID: file.ProductId})
	}

	return nil
}

func (account Account) StorageDeleteFile(fileId uint) error {

	file, err := account.StorageGet(fileId)
	if err != nil {
		return err
	}

	if err := file.Delete(); err != nil {
		return err
	}

	if file.ProductId > 0 {
		go account.CallWebHookIfExist(EventProductUpdated, Product{ID: file.ProductId})
	}

	return nil
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

func (Storage) SelectArrayWithoutData() []string {
	fields := structs.Names(&Storage{}) //.(map[string]string)
	fields = utils.RemoveKey(fields, "Data")
	fields = utils.RemoveKey(fields, "URL")
	return utils.ToLowerSnakeCaseArr(fields)
}