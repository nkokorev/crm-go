package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Вид номенклатуры - ассортиментные группы продаваемых товаров.
type WarehouseProduct struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	ParentId 	uint	`json:"parent_id"`

	// code for scope routes (группы, категории....)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Routing
	Path 		*string `json:"path" gorm:"type:varchar(255);"`			// Имя пути - catalog, cat, /, ..
	Label 		*string `json:"label" gorm:"type:varchar(255);"` 		// menu label - Чай, кофе, ..
	RouteName 	*string `json:"route_name" gorm:"type:varchar(50);"` 	// route name: delivery, info.index, cart
	IconName 	*string `json:"icon_name" gorm:"type:varchar(50);"` 	// icon name

	// Порядок отображения в текущей иерархии категории
	Priority 			int		`json:"priority" gorm:"type:int;default:10;"`

	// HasMany ...
	ProductCards 	[]ProductCard 	`json:"product_cards"`

	// Страницы, на которых выводятся карточки товаров этой товарной группы
	WebPages 		[]WebPage 	`json:"web_pages" gorm:"many2many:web_page_product_groups;"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (warehouseProduct WarehouseProduct) GetId() uint { return warehouseProduct.Id }
func (warehouseProduct *WarehouseProduct) setId(id uint) { warehouseProduct.Id = id }
func (warehouseProduct *WarehouseProduct) setPublicId(publicId uint) { warehouseProduct.PublicId = publicId }
func (warehouseProduct WarehouseProduct) GetAccountId() uint { return warehouseProduct.AccountId }
func (warehouseProduct *WarehouseProduct) setAccountId(id uint) { warehouseProduct.AccountId = id }
func (WarehouseProduct) SystemEntity() bool { return false }
// ############# End Entity interface #############
func (WarehouseProduct) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WarehouseProduct{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&WarehouseProduct{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&WarehouseProduct{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")
	// db.Model(&WarehouseProduct{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
	// db.Model(&WarehouseProduct{}).AddForeignKey("users_segment_id", "users_segments(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE web_pages " +
		"ADD CONSTRAINT web_pages_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		// "ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&WarehouseProduct{}, "ProductCards", &WebPageProductCard{})
	if err != nil {
		log.Fatal(err)
	}
}
func (warehouseProduct *WarehouseProduct) BeforeCreate(tx *gorm.DB) error {
	warehouseProduct.Id = 0
	
	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&WarehouseProduct{}).Where("account_id = ?",  warehouseProduct.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	warehouseProduct.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (warehouseProduct *WarehouseProduct) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
// ######### CRUD Functions ############
func (warehouseProduct WarehouseProduct) create() (Entity, error)  {

	en := warehouseProduct

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false, false, nil).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}
func (WarehouseProduct) get(id uint, preloads []string) (Entity, error) {

	var warehouseProduct WarehouseProduct

	err := warehouseProduct.GetPreloadDb(false,false,preloads).First(&warehouseProduct, id).Error
	if err != nil {
		return nil, err
	}
	return &warehouseProduct, nil
}
func (warehouseProduct *WarehouseProduct) load(preloads []string) error {

	err := warehouseProduct.GetPreloadDb(false,false,preloads).First(warehouseProduct, warehouseProduct.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (warehouseProduct *WarehouseProduct) loadByPublicId(preloads []string) error {
	
	if warehouseProduct.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить WarehouseProduct - не указан  Id"}
	}
	if err := warehouseProduct.GetPreloadDb(false,false, preloads).First(warehouseProduct, "account_id = ? AND public_id = ?", warehouseProduct.AccountId, warehouseProduct.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (WarehouseProduct) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return WarehouseProduct{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (WarehouseProduct) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	warehouseProducts := make([]WarehouseProduct,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&WarehouseProduct{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&warehouseProducts, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WarehouseProduct{}).
			Where("label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&WarehouseProduct{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&warehouseProducts).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WarehouseProduct{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(warehouseProducts))
	for i := range warehouseProducts {
		entities[i] = &warehouseProducts[i]
	}

	return entities, total, nil
}
func (warehouseProduct *WarehouseProduct) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := warehouseProduct.GetPreloadDb(false,false,nil).Where(" id = ?", warehouseProduct.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := warehouseProduct.GetPreloadDb(false,false,preloads).First(warehouseProduct, warehouseProduct.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (warehouseProduct *WarehouseProduct) delete () error {
	return warehouseProduct.GetPreloadDb(true,false,nil).Where("id = ?", warehouseProduct.Id).Delete(warehouseProduct).Error
}
// ######### END CRUD Functions ############

func (warehouseProduct *WarehouseProduct) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(warehouseProduct)
	} else {
		_db = _db.Model(&WarehouseProduct{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Image","ProductCards"})

		for _,v := range allowed {
			if v == "Image" {
				_db.Preload("Image", func(db *gorm.DB) *gorm.DB {
					return db.Select(Storage{}.SelectArrayWithoutDataURL())
				})
			} else {
				_db.Preload(v)
			}
		}
		return _db
	}
}
func (warehouseProduct WarehouseProduct) CreateChild(wp WarehouseProduct) (Entity, error){
	wp.ParentId = warehouseProduct.Id

	_webPage, err := wp.create()
	if err != nil {return nil, err}

	return _webPage, nil
}
func (warehouseProduct WarehouseProduct) AppendProductCard(input *ProductCard, optPriority... int) error {

	priority := 10
	if len(optPriority) > 0 {
		priority = optPriority[0]
	}
	var productCard *ProductCard
	if input.Id < 1 {
		proPtr, err := input.create()
		if err != nil {
			return err
		}
		_productCard, ok := proPtr.(*ProductCard)
		if !ok {
			return utils.Error{Message: "Ошибка преобразования продуктовой карточки"}
		}
		productCard = _productCard
	} else {
		productCard = input
	}
	if err := db.Model(&WebPageProductCard{}).Create(
		&WebPageProductCard{WebPageId: warehouseProduct.Id, ProductCardId: productCard.Id, Priority: priority}).Error; err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductCardUpdated(account.Id, productCard.Id))
	}

	return nil
}
