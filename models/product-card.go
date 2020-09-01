package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	// Id     				uint 	`json:"id" gorm:"primaryKey"`
	Id        			uint 	`json:"id" gorm:"primaryKey"`
	PublicId			uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 			uint 	`json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null
	WebSiteId 			uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится
	WebPageId 			uint 	`json:"web_page_id" gorm:"type:int;index;"` // группа товаров, категория товаров

	Enabled 			bool 	`json:"enabled" gorm:"type:bool;default:true"` // активна ли карточка товара

	Label	 			*string 	`json:"label" gorm:"type:varchar(255);"` // что выводить в список товаров
	Path 				*string 	`json:"path" gorm:"type:varchar(255);"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb 			*string 	`json:"breadcrumb" gorm:"type:varchar(255);"`

	MetaTitle 			*string 	`json:"meta_title" gorm:"type:varchar(255);"`
	MetaKeywords 		*string 	`json:"meta_keywords" gorm:"type:varchar(255);"`
	MetaDescription 	*string 	`json:"meta_description" gorm:"type:varchar(255);"`

	// Full description нет т.к. в карточке описание берется от офера
	ShortDescription 	*string 	`json:"short_description" gorm:"type:varchar(255);"` // для превью карточки товара
	Description 		*string 	`json:"description" gorm:"type:text;"` // фулл описание товара

	// Хелперы карточки: переключение по цветам, размерам и т.д.
	SwitchProducts	 	datatypes.JSON `json:"switch_products"` // {color, size} Параметры переключения среди предложений

	// Preview Images - небольшие пережатые изображения товара(ов)
	Image 				[]Storage	`json:"images" gorm:"polymorphic:Owner;"`

	WebPages 			[]ProductCard 	`json:"web_pages" gorm:"many2many:web_page_product_card;"`
	// WebSite		 		WebSite 	`json:"-" gorm:"-"`
	Products 			[]Product 	`json:"products" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
}

func (ProductCard) PgSqlCreate() {
	if err := db.Migrator().AutoMigrate(&ProductCard{}); err != nil { log.Fatal(err) }
	err := db.Exec("ALTER TABLE product_cards ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductCard{}, "Products", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
}

// ############# Entity interface #############
func (productCard ProductCard) GetId() uint { return productCard.Id }
func (productCard *ProductCard) setId(id uint) { productCard.Id = id }
func (productCard *ProductCard) setPublicId(publicId uint) { productCard.PublicId = publicId }
func (productCard ProductCard) GetAccountId() uint { return productCard.AccountId }
func (productCard *ProductCard) setAccountId(id uint) { productCard.AccountId = id }
func (ProductCard) SystemEntity() bool { return false }
// ############# End Entity interface #############

func (productCard *ProductCard) BeforeCreate(tx *gorm.DB) error {
	productCard.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductCard{}).Where("account_id = ?",  productCard.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productCard.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productCard *ProductCard) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
func (productCard *ProductCard) AfterCreate(tx *gorm.DB) error {
	event.AsyncFire(Event{}.ProductCardCreated(productCard.AccountId, productCard.Id))
	return nil
}
func (productCard *ProductCard) AfterUpdate(tx *gorm.DB) error {
	event.AsyncFire(Event{}.ProductCardUpdated(productCard.AccountId, productCard.Id))
	return nil
}
func (productCard *ProductCard) AfterDelete(tx *gorm.DB) error {
	event.AsyncFire(Event{}.DeliveryOrderDeleted(productCard.AccountId, productCard.Id))
	return nil
}
// ######### CRUD Functions ############
func (productCard ProductCard) create() (Entity, error)  {

	en := productCard

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false,false, true).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}
func (ProductCard) get(id uint) (Entity, error) {

	var productCard ProductCard

	err := productCard.GetPreloadDb(false,false,true).First(&productCard, id).Error
	if err != nil {
		return nil, err
	}
	return &productCard, nil
}
func (productCard *ProductCard) load() error {

	err := productCard.GetPreloadDb(false,false,true).First(productCard, productCard.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productCard *ProductCard) loadByPublicId() error {

	if productCard.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductCard - не указан  Id"}
	}
	if err := productCard.GetPreloadDb(false,false, true).First(productCard, "account_id = ? AND public_id = ?", productCard.AccountId, productCard.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductCard) getList(accountId uint, sortBy string) ([]Entity, int64, error) {
	return ProductCard{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (ProductCard) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, int64, error) {

	emailCampaigns := make([]ProductCard,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductCard{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns, "label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductCard{}).
			Where("label ILIKE ? OR path ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&ProductCard{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductCard{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(emailCampaigns))
	for i := range emailCampaigns {
		entities[i] = &emailCampaigns[i]
	}

	return entities, total, nil
}
func (productCard *ProductCard) update(input map[string]interface{}) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","order","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productCard.GetPreloadDb(true,false,false).Where(" id = ?", productCard.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productCard.GetPreloadDb(true,false,false).First(productCard, productCard.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productCard *ProductCard) delete () error {
	return productCard.GetPreloadDb(true,true,false).Where("id = ?", productCard.Id).Delete(productCard).Error
}
// ######### END CRUD Functions ############

func (productCard *ProductCard) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(productCard)
	} else {
		_db = _db.Model(&ProductCard{})
	}

	if preload {
		// return _db.Preload("EmailTemplate").Preload("EmailBox").Preload("UsersSegment")
		// return _db
		return _db.Preload("Image", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})
	} else {
		return _db
	}
}

