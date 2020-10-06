package models

import (
	"database/sql"
	"errors"
	"github.com/fatih/structs"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"time"
)

type Storage struct {

	Id     		uint   	`json:"id" gorm:"primaryKey"`
	HashId 		string 	`json:"hash_id" gorm:"type:varchar(12);uniqueIndex;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	OwnerID   	uint	`json:"owner_id" `   // ?? gorm:"association_foreignkey:Id"
	// OwnerID   	uint	`json:"-" `   // ?? gorm:"association_foreignkey:Id"
	OwnerType	string	`json:"owner_type" gorm:"type:varchar(80);column:owner_type"`

	Priority 	int		`json:"priority" gorm:"type:int;"` // Порядок отображения (часто нужно файлам)
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // выводить ли где-то это изображение или нет

	Name 				string `json:"name" gorm:"type:varchar(255);"` // имя файла (оно же при отдаче)
	Alt 				*string `json:"alt" gorm:"type:varchar(255);"` // alt для изображений
	ShortDescription 	*string `json:"short_description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)
	Description 		*string `json:"description" gorm:"type:text;"` // pgsql: text // большое описание изображения (не, ну мало ли фанаты фото)

	// MetaData
	MIME 		string 	`json:"mime" gorm:"type:varchar(90);"` // мета тип файла
	Size 		int64 	`json:"size" gorm:"type:int;"` // Kb

	Data 		[]byte `json:"data" gorm:"type:bytea;"` // тело файла

	// ** Hidden **
	URL 		string 	`json:"_url" gorm:"-"` // see AfterFind

	// Время жизни файла, по умолчанию - null (без ограничений)
	ExpiredAt 	*time.Time  `json:"expired_at"`

	CreatedAt 	time.Time  `json:"created_at"`
	UpdatedAt 	time.Time  `json:"updated_at"`
}

func (Storage) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.Migrator().CreateTable(&Storage{})
	// db.Model(&Storage{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE storage ADD CONSTRAINT storages_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}
func (Storage) TableName() string {
	return "storage"
}

func (fs *Storage) BeforeCreate(tx *gorm.DB) error {
	fs.Id = 0
	fs.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	fs.CreatedAt = time.Now().UTC()
	return nil
}
func (fs *Storage) AfterCreate(tx *gorm.DB) error {
	switch fs.OwnerType {
	case "products":
		// AsyncFire(*Event{}.ProductUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ProductUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_id":fs.OwnerID}))
	case "articles":
		// AsyncFire(*Event{}.ArticleUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ArticleUpdated", map[string]interface{}{"account_id":fs.AccountId, "article_id":fs.OwnerID}))
	case "web_pages":
		// AsyncFire(*Event{}.WebPageUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("WebPageUpdated", map[string]interface{}{"account_id":fs.AccountId, "web_page_id":fs.OwnerID}))
	case "product_cards":
		// AsyncFire(*Event{}.ProductCardUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ProductCardUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_card_id":fs.OwnerID}))
	case "manufactures":
		// AsyncFire(*Event{}.ManufacturerUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ManufacturerUpdated", map[string]interface{}{"account_id":fs.AccountId, "manufacturer_id":fs.OwnerID}))
	default:
		// AsyncFire(*Event{}.StorageCreated(fs.AccountId, fs.Id))
		AsyncFire(NewEvent("StorageCreated", map[string]interface{}{"account_id":fs.AccountId, "storage_id":fs.Id}))
	}
	return nil
}
func (fs *Storage) AfterDelete(tx *gorm.DB) (err error) {
	switch fs.OwnerType {
	case "products":
		// AsyncFire(*Event{}.ProductUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ProductUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_id":fs.OwnerID}))
	case "articles":
		// AsyncFire(*Event{}.ArticleUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ArticleUpdated", map[string]interface{}{"account_id":fs.AccountId, "article_id":fs.OwnerID}))
	case "web_pages":
		// AsyncFire(*Event{}.WebPageUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("WebPageUpdated", map[string]interface{}{"account_id":fs.AccountId, "web_page_id":fs.OwnerID}))
	case "product_cards":
		// AsyncFire(*Event{}.ProductCardUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ProductCardUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_card_id":fs.OwnerID}))
	case "manufactures":
		// AsyncFire(*Event{}.ManufacturerUpdated(fs.AccountId, uint(fs.OwnerID)))
		AsyncFire(NewEvent("ManufacturerUpdated", map[string]interface{}{"account_id":fs.AccountId, "manufacturer_id":fs.OwnerID}))
	default:
		// AsyncFire(*Event{}.StorageDeleted(fs.AccountId, fs.Id))
		AsyncFire(NewEvent("StorageDeleted", map[string]interface{}{"account_id":fs.AccountId, "storage_id":fs.Id}))
	}
	return nil
}

func (fs *Storage) AfterFind(tx *gorm.DB) (err error) {

	fs.LoadURL()
	/*
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
		fs.URL = crmHost + "/products/images/" + fs.HashId
	case "EmailTemplate":
		fs.URL = crmHost + "/emails/images/" + fs.HashId
		//...
	case "Article":
		// fs.URL = crmHost + "/emails/images/" + fs.HashId
	default:
		fs.URL = crmHost + "/public/" + fs.HashId
	}*/

	return nil
}
func (fs *Storage) LoadURL() {
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
		fs.URL = crmHost + "/products/images/" + fs.HashId
	case "EmailTemplate":
		fs.URL = crmHost + "/emails/images/" + fs.HashId
	case "Article":
		fs.URL = crmHost + "/articles/images/" + fs.HashId
	case "WebPage":
		fs.URL = crmHost + "/web-pages/" + fs.HashId
	default:
		fs.URL = crmHost + "/public/" + fs.HashId
	}
}

// ############# Entity interface #############
func (fs Storage) GetId() uint { return fs.Id }
func (fs *Storage) setId(id uint) { fs.Id = id }
func (fs *Storage) setPublicId(id uint) { }
func (fs Storage) GetAccountId() uint { return fs.AccountId }
func (fs *Storage) setAccountId(id uint) { fs.AccountId = id }
func (Storage) SystemEntity() bool { return false }
// ############# Entity interface #############


// ######### CRUD Functions ############
func (fs Storage) create() (Entity, error)  {

	// 1. Получаем Аккаунт
	account, err := GetAccount(fs.AccountId); if err != nil {
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
	file.LoadURL()

	var entity Entity = &file

	return entity, nil
}
func (Storage) get(id uint, preloads []string) (Entity, error) {

	var fs Storage

	err := fs.GetPreloadDb(false,false,preloads, false).First(&fs, id).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
func (Storage) getByHashId(hashId string) (*Storage, error)  {

	fs := Storage{}

	err := fs.GetPreloadDb(false,false,nil, false).First(&fs, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
func (fs *Storage) load(preloads []string) error {
	if fs.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Storage - не указан  Id"}
	}

	err := fs.GetPreloadDb(false,false,preloads, false).First(fs,fs.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*Storage) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (Storage) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Storage{}.getPaginationList(accountId,0,100,sortBy,"",nil,preload)
}

func (Storage) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	files := make([]Storage,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Storage{}).GetPreloadDb(false,false,preloads,true).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).Where(filter).
			Find(&files, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Storage{}).GetPreloadDb(false,false,nil,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Storage{}).GetPreloadDb(false,false,preloads, true).
			Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).Where(filter).
			Find(&files).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Storage{}).GetPreloadDb(false,false,nil, true).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(files))
	for i := range files {
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

func (fs *Storage) update(input map[string]interface{}, preloads []string) error {

	delete(input,"size")
	// delete(input,"owner_id")
	// delete(input,"owner_type")
	// Обновляем без data, т.к. изображение не изменяется
	delete(input,"data")
	delete(input,"mime")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"owner_id","priority"}); err != nil {
		return err
	}

	err := db.Model(fs).Omit("id", "hashId", "account_id","created_at").Updates(input).Error
	if err != nil {
		return err
	}

	switch fs.OwnerType {
	case "products":
		AsyncFire(NewEvent("ProductUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_id":fs.OwnerID}))
	case "articles":
		AsyncFire(NewEvent("ArticleUpdated", map[string]interface{}{"account_id":fs.AccountId, "article_id":fs.OwnerID}))
	case "web_pages":
		AsyncFire(NewEvent("WebPageUpdated", map[string]interface{}{"account_id":fs.AccountId, "web_page_id":fs.OwnerID}))
	case "product_cards":
		AsyncFire(NewEvent("ProductCardUpdated", map[string]interface{}{"account_id":fs.AccountId, "product_card_id":fs.OwnerID}))
	case "manufactures":
		AsyncFire(NewEvent("ManufacturerUpdated", map[string]interface{}{"account_id":fs.AccountId, "manufacturer_id":fs.OwnerID}))
	default:
		// AsyncFire(*Event{}.StorageUpdated(fs.AccountId, fs.Id))
		AsyncFire(NewEvent("StorageUpdated", map[string]interface{}{"account_id":fs.AccountId, "storage_id":fs.Id}))
	}

	err = fs.GetPreloadDb(false,false, preloads, true).First(fs, fs.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (Storage) UpdatePriority(input []Storage) error {
	for _,v := range input {
		// priority, ok = v.Priority.(int)
		if err := (&Storage{Id: v.Id }).update(map[string]interface{}{"priority":v.Priority},nil); err != nil {
			return err
		}
	}
	return nil
}
func (fs *Storage) SetAutoPriority() error {

	var lastIdx sql.NullInt64
	err := db.Table("storage").Where("account_id = ? AND owner_id = ? AND owner_type = ?",  fs.AccountId, fs.OwnerID, fs.OwnerType).
		Select("max(priority)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }

	// fmt.Println(fs.AccountId, fs.OwnerID, fs.OwnerType)
	// fmt.Println("lastIdx.Int64: ", lastIdx.Int64)

	fs.Priority = 1 + int(lastIdx.Int64)

	return nil

}
func (fs Storage) GetAutoPriority() int {

	var lastIdx sql.NullInt64
	err := db.Table("storage").Where("account_id = ? AND owner_id = ? AND owner_type = ?",  fs.AccountId, fs.OwnerID, fs.OwnerType).
		Select("max(priority)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return 0 }

	// fmt.Println(fs.AccountId, fs.OwnerID, fs.OwnerType)
	// fmt.Println("lastIdx.Int64: ", lastIdx.Int64)

	return 1 + int(lastIdx.Int64)

}
func (fs *Storage) delete () error {
	if err := db.Model(Storage{}).Where("id = ?", fs.Id).Delete(fs).Error; err != nil {
		return err
	}
	return nil
}
func (fs *Storage) GetPreloadDb(getModel bool, autoPreload bool, preloads []string, skipData bool) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(fs)
	} else {
		_db = _db.Model(&Storage{})
	}

	if skipData {
		_db.Select(Storage{}.SelectArrayWithoutDataURL())
	}
	
	if autoPreload {
		return _db
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}
// ######### END CRUD Functions ############


// ########### ACCOUNT FUNCTIONAL ###########

// Возвращает объем использованного диска в КБ
func (account Account) StorageDiskSpaceUsed() (int64, error) {
	var sum,count int64

	// sum = 0
	// count := int64(0)
	err := db.Table("storage").Where("account_id = ?", account.Id).Count(&count).Error
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	err = db.Table("storage").Where("account_id = ?", account.Id).Select("sum(size)").Row().Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, nil
}

func (account Account) StorageGetByHashId(hashId string) (*Storage, error) {

	fs, err := (Storage{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if fs.AccountId != account.Id {
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
func (account Account) GetStoragePaginationListByOwner(offset, limit int, sortBy, search string, ownerId uint, ownerType string) ([]Entity, int64, error) {

	if ownerType == "" {
		return  Storage{}.getPaginationList(account.Id, offset, limit, sortBy, search,nil,nil)
	}


	files := make([]Storage,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", account.Id).
			Find(&files, "name ILIKE ? OR code ILIKE ? OR description ILIKE ? AND owner_id = ? AND owner_type = ?", search,search,search, ownerId, ownerType).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ? AND owner_id = ? AND owner_type = ?", account.Id, search,search,search, ownerId, ownerType).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Storage{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ? AND owner_id = ? AND owner_type = ?", account.Id, ownerId, ownerType).
			Find(&files).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Storage{}).Where("account_id = ? AND owner_id = ? AND owner_type = ?", account.Id, ownerId, ownerType).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(files))
	for i := range files {
		entities[i] = &files[i]
	}

	return entities, total, nil
}

// ########### END OF ACCOUNT FUNCTIONAL ###########

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

	// Устанавливаем приоритет
	var priority = Storage{AccountId: product.AccountId, OwnerID: product.Id, OwnerType: "products"}.GetAutoPriority()

	if file.Id > 0 {
		 if err := fs.update(map[string]interface{}{"owner_id":product.Id,"owner_type":"products","priority":priority},nil); err != nil {
		 	return err
		 }
	} else {
		file.Priority = priority
		if err := db.Model(&product).Association("Images").Append(file); err != nil {
			return err
		}
	}



	return nil
}
func (productCard ProductCard) AppendAssociationImage(fs Entity) error {
	file, ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение"}
	}

	// Устанавливаем приоритет
	var priority = Storage{AccountId: productCard.AccountId, OwnerID: productCard.Id, OwnerType: "product_cards"}.GetAutoPriority()
	
	if file.Id > 0 {
		if err := fs.update(map[string]interface{}{"owner_id":productCard.Id,"owner_type":"product_cards","priority":priority},nil); err != nil {
			return err
		}
	} else {
		file.Priority = priority
		if err := db.Model(&productCard).Association("Images").Append(file); err != nil {
			return err
		}
	}

	return nil
}
func (article Article) AppendAssociationImage(fs Entity) error {
	file, ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение продукту"}
	}

	if file.Id > 0 {
		if err := fs.update(map[string]interface{}{"owner_id":article.Id,"owner_type":"articles"},nil); err != nil {
			return err
		}
	} else {
		if err := db.Model(&article).Association("Image").Append(file); err != nil {
			return err
		}
	}

	return nil
}
func (webPage WebPage) AppendAssociationImage(fs Entity) error {
	file, ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение"}
	}

	if file.Id > 0 {
		if err := fs.update(map[string]interface{}{"owner_id":webPage.Id,"owner_type":"web_pages"},nil); err != nil {
			return err
		}
	} else {
		if err := db.Model(&webPage).Association("Image").Append(file); err != nil {
			return err
		}
	}

	return nil
}
func (manufacturer Manufacturer) AppendAssociationImage(fs Entity) error {
	file, ok := fs.(*Storage)
	if !ok {
		return utils.Error{Message: "Не возможно добавить изображение"}
	}

	if file.Id > 0 {
		if err := fs.update(map[string]interface{}{"owner_id":manufacturer.Id,"owner_type":"manufactures"},nil); err != nil {
			return err
		}
	} else {
		if err := db.Model(&manufacturer).Association("Image").Append(file); err != nil {
			return err
		}
	}

	return nil
}

