package models

import (
	"errors"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"os"
	"strings"
	"time"
)

type Storage struct {

	ID     		uint   	`json:"id" gorm:"primary_key"`
	HashID 		string 	`json:"hashID" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	OwnerID   	uint	//`json:"-"`   // ?? gorm:"association_foreignkey:ID"
	OwnerType	string	//`json:"ownerType" gorm:"type:varchar(80);column:owner_type"`

	//Product		Product	`json:"-" gorm:"polymorphic:Owner;"`

	//ProductID 	uint	`json:"productID" gorm:"type:int;default:null;"` // id of products
	//EmailID 	uint	`json:"emailID" gorm:"type:int;default:null;"` // id of email template

	Priority 	int		`json:"priority" gorm:"type:int;default:null;"` // Порядок отображения (часто нужно файлам)
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

/*func (fs *Storage) AfterUpdate(tx *gorm.DB) (err error) {

	if len(fs.OwnerType) > 0 {
		account, err := GetAccount(fs.AccountID)
		if err == nil {
			account.CallWebHookUpdated(*fs)
		}
	}
	return
}*/

/*func (fs *Storage) AfterCreate(tx *gorm.DB) (err error) {

	if fs.OwnerType != "" {

		account, err := GetAccount(fs.AccountID)
		if err == nil {
			account.CallEventByStorageCreated(*fs)
		}
	}
	return
}*/

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

	switch fs.OwnerType {
	case "Product":
		fs.URL = crmHost + "/products/images/" + fs.HashID
	case "EmailTemplate":
		fs.URL = crmHost + "/emails/images/" + fs.HashID
		//...
	case "Article":
		//..
	default:
		fs.URL = crmHost + "/public/" + fs.HashID
	}

	return nil
}

// ############# Entity interface #############
func (fs Storage) GetID() uint { return fs.ID }
func (fs *Storage) setID(id uint) { fs.ID = id }
func (fs Storage) GetAccountID() uint { return fs.AccountID }
func (fs *Storage) setAccountID(id uint) { fs.AccountID = id }
func (Storage) SystemEntity() bool { return false }
// ############# Entity interface #############


// ######### CRUD Functions ############
func (fs Storage) create() (Entity, error)  {

	// 1. Получаем Аккаунт
	account, err := GetAccount(fs.AccountID); if err != nil {
		return nil, err
	}

	used, err := account.StorageDiskSpaceUsed()
	if err != nil {
		return nil, err
	}

	if (account.DiskSpaceAvailable - used) < fs.Size {
		return nil, utils.Error{Message: "Нехватка дискового пространства"}
	}

	file := fs
	if err := db.Create(&file).Error; err != nil {
		return nil, err
	}

	account.CallEventByStorageCreated(file)

	var entity Entity = &file

	return entity, nil
}
func (Storage) get(id uint) (Entity, error) {

	var fs Storage

	err := db.First(&fs, id).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
func (Storage) getByHashID(hashID string) (*Storage, error)  {

	fs := Storage{}

	err := db.First(&fs, "hash_id = ?", hashID).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
func (fs *Storage) load() error {
	if fs.ID < 1 {
		return utils.Error{Message: "Невозможно загрузить Storage - не указан  ID"}
	}

	err := db.First(fs,fs.ID).Error
	if err != nil {
		return err
	}
	return nil
}

func (Storage) getList(accountID uint, sortBy string) ([]Entity, uint, error) {
	return Storage{}.getPaginationList(accountID,0,100,sortBy,"")
}

func (Storage) getPaginationList(accountID uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	files := make([]Storage,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&files, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountID, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountID).
			Find(&files).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).Where("account_id = ?", accountID).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(files))
	for i,_ := range files {
		entities[i] = &files[i]
	}

	return entities, total, nil
}

func (Storage) getByEvent(eventName string) (*Storage, error) {

	wh := Storage{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (fs *Storage) update(input map[string]interface{}) error {
	err := db.Set("gorm:association_autoupdate", false).
		Model(fs).Omit("id", "hashID", "account_id","created_at").Updates(input).Error;
	if err != nil {
		return err
	}

	account, err := GetAccount(fs.AccountID); if err == nil {
		account.CallEventByStorageUpdated(*fs)
	}

	return nil
}

func (fs Storage) delete () error {
	if err := db.Model(Storage{}).Where("id = ?", fs.ID).Delete(fs).Error; err != nil {
		return err
	}

	// Вызываем обновление emit events
	account, err := GetAccount(fs.AccountID); if err == nil {
		account.CallEventByStorageDeletes(fs)
	}

	return nil
}
// ######### END CRUD Functions ############


// ########### ACCOUNT FUNCTIONAL ###########

// Возвращает объем использованного диска в КБ
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

func (account Account) StorageGetByHashID(hashID string) (*Storage, error) {

	fs, err := (Storage{}).getByHashID(hashID)
	if err != nil {
		return nil, err
	}

	if fs.AccountID != account.ID {
		return nil, errors.New("Шаблон принадлежит другому аккаунту")
	}

	return fs, nil
}

func (Account) StorageGetPublicByHashID(hashID string) (*Storage, error) {

	fs, err := (Storage{}).getByHashID(hashID)
	if err != nil {
		return nil, err
	}
	
	return fs, nil
}
func (account Account) GetStoragePaginationListByOwner(offset, limit int, sortBy, search string, ownerID uint, ownerType string) ([]Entity, uint, error) {

	if ownerType == "" {
		return  Storage{}.getPaginationList(account.ID, offset, limit, sortBy, search)
	}


	files := make([]Storage,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", account.ID).
			Find(&files, "name ILIKE ? OR code ILIKE ? OR description ILIKE ? AND owner_id = ? AND owner_type = ?", search,search,search, ownerID, ownerType).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ? AND owner_id = ? AND owner_type = ?", account.ID, search,search,search, ownerID, ownerType).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ? AND owner_id = ? AND owner_type = ?", account.ID, ownerID, ownerType).
			Find(&files).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).Where("account_id = ? AND owner_id = ? AND owner_type = ?", account.ID, ownerID, ownerType).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(files))
	for i,_ := range files {
		entities[i] = &files[i]
	}

	return entities, total, nil
}

// ########### END OF ACCOUNT FUNCTIONAL ###########

func (account Account) CallEventByStorageCreated(fs Storage) {
	switch fs.OwnerType {
	case "products":
		event.AsyncFire(Event{}.ProductUpdated(account.ID, fs.OwnerID))
	case "articles":
		event.AsyncFire(Event{}.ArticleUpdated(account.ID, fs.OwnerID))
	default:
		event.AsyncFire(Event{}.StorageCreated(account.ID, fs.OwnerID))
	}
}
func (account Account) CallEventByStorageUpdated(fs Storage) {
	switch fs.OwnerType {
	case "products":
		event.AsyncFire(Event{}.ProductUpdated(account.ID, fs.OwnerID))
	case "articles":
		event.AsyncFire(Event{}.ArticleUpdated(account.ID, fs.OwnerID))
	default:
		event.AsyncFire(Event{}.StorageUpdated(account.ID, fs.OwnerID))
	}
}
func (account Account) CallEventByStorageDeletes(fs Storage) {
	switch fs.OwnerType {
	case "products":
		event.AsyncFire(Event{}.ProductDeleted(account.ID, fs.OwnerID))
		// go account.CallWebHookIfExist(EventArticleUpdated, Product{ID: fs.OwnerID})
	case "articles":
		event.AsyncFire(Event{}.ArticleDeleted(account.ID, fs.OwnerID))
		// go account.CallWebHookIfExist(EventArticleUpdated, Article{ID: fs.OwnerID})
	default:
		event.AsyncFire(Event{}.StorageDeleted(account.ID, fs.OwnerID))
	}
}

func (Storage) SelectArrayWithoutDataURL() []string {
	fields := structs.Names(&Storage{}) //.(map[string]string)
	fields = utils.RemoveKey(fields, "URL")
	fields = utils.RemoveKey(fields, "Data")
	return utils.ToLowerSnakeCaseArr(fields)
}

// Для работы со связанными моделями
/*func (product Product) AppendAssociationImage(fs Storage) error {
	if err := db.Model(&product).Association("Images").Append(fs).Error; err != nil {
		return err
	}
	return nil
}*/
func (product Product) AppendAssociationImage(fs Entity) error {
	file,ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение продукту"}
	}
	if err := db.Model(&product).Association("Images").Append(file).Error; err != nil {
		return err
	}
	return nil
}

func (article Article) AppendAssociationImage(fs Entity) error {
	file, ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение продукту"}
	}

	if err := db.Model(&article).Association("Image").Append(file).Error; err != nil {
		return err
	}
	return nil
}