package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

// Вид номенклатуры - ассортиментные группы продаваемых товаров.
type ProductGroup struct {
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
	// ProductCards 	[]ProductCard 	`json:"product_cards"`

	// Страницы, на которых выводятся карточки товаров этой товарной группы
	// WebPages 		[]WebPage 	`json:"web_pages" gorm:"many2many:web_page_product_groups;"`

	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (productGroup ProductGroup) GetId() uint { return productGroup.Id }
func (productGroup *ProductGroup) setId(id uint) { productGroup.Id = id }
func (productGroup *ProductGroup) setPublicId(publicId uint) { productGroup.PublicId = publicId }
func (productGroup ProductGroup) GetAccountId() uint { return productGroup.AccountId }
func (productGroup *ProductGroup) setAccountId(id uint) { productGroup.AccountId = id }
func (ProductGroup) SystemEntity() bool { return false }
// ############# End Entity interface #############
func (ProductGroup) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&ProductGroup{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&ProductGroup{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&ProductGroup{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")
	// db.Model(&ProductGroup{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
	// db.Model(&ProductGroup{}).AddForeignKey("users_segment_id", "users_segments(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE web_pages " +
		"ADD CONSTRAINT web_pages_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		// "ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

/*	err = db.SetupJoinTable(&ProductGroup{}, "ProductCards", &WebPageProductCard{})
	if err != nil {
		log.Fatal(err)
	}*/
}
func (productGroup *ProductGroup) BeforeCreate(tx *gorm.DB) error {
	productGroup.Id = 0
	
	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&ProductGroup{}).Where("account_id = ?",  productGroup.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	productGroup.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (productGroup *ProductGroup) AfterFind(tx *gorm.DB) (err error) {

	return nil
}
// ######### CRUD Functions ############
func (productGroup ProductGroup) create() (Entity, error)  {

	en := productGroup

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
func (ProductGroup) get(id uint, preloads []string) (Entity, error) {

	var productGroup ProductGroup

	err := productGroup.GetPreloadDb(false,false,preloads).First(&productGroup, id).Error
	if err != nil {
		return nil, err
	}
	return &productGroup, nil
}
func (productGroup *ProductGroup) load(preloads []string) error {

	err := productGroup.GetPreloadDb(false,false,preloads).First(productGroup, productGroup.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (productGroup *ProductGroup) loadByPublicId(preloads []string) error {
	
	if productGroup.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить ProductGroup - не указан  Id"}
	}
	if err := productGroup.GetPreloadDb(false,false, preloads).First(productGroup, "account_id = ? AND public_id = ?", productGroup.AccountId, productGroup.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (ProductGroup) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return ProductGroup{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (ProductGroup) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	productGroups := make([]ProductGroup,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&ProductGroup{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductGroup{}).
			Where("label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&ProductGroup{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&productGroups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&ProductGroup{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(productGroups))
	for i := range productGroups {
		entities[i] = &productGroups[i]
	}

	return entities, total, nil
}
func (productGroup *ProductGroup) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := productGroup.GetPreloadDb(false,false,nil).Where(" id = ?", productGroup.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := productGroup.GetPreloadDb(false,false,preloads).First(productGroup, productGroup.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (productGroup *ProductGroup) delete () error {
	return productGroup.GetPreloadDb(true,false,nil).Where("id = ?", productGroup.Id).Delete(productGroup).Error
}
// ######### END CRUD Functions ############

func (productGroup *ProductGroup) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(productGroup)
	} else {
		_db = _db.Model(&ProductGroup{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Image"})

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
func (productGroup ProductGroup) CreateChild(wp ProductGroup) (Entity, error){
	wp.ParentId = productGroup.Id

	_webPage, err := wp.create()
	if err != nil {return nil, err}

	return _webPage, nil
}

