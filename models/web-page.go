package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type WebPage struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId 	*uint 	`json:"web_site_id" gorm:"type:int;index;"`
	ParentId 	uint	`json:"parent_id"`
	// Children 	*WebPage `json:"_children" gorm:"-"`

	// code for scope routes (группы, категории....)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Routing
	Path 		*string `json:"path" gorm:"type:varchar(255);"`			// Имя пути - catalog, cat, /, ..
	Label 		*string `json:"label" gorm:"type:varchar(255);"` 		// menu label - Чай, кофе, ..
	RouteName 	*string `json:"route_name" gorm:"type:varchar(50);"` 	// route name: delivery, info.index, cart
	IconName 	*string `json:"icon_name" gorm:"type:varchar(50);"` 	// icon name

	// Порядок отображения в текущей иерархии категории
	Priority 			int		`json:"priority" gorm:"type:int;default:10;"`

	Breadcrumb 			*string `json:"breadcrumb" gorm:"type:varchar(255);"`
	ShortDescription 	*string `json:"short_description" gorm:"type:varchar(255);"`
	Description 		*string `json:"description" gorm:"type:text;"`

	MetaTitle 		*string 	`json:"meta_title" gorm:"type:varchar(255);"`
	MetaKeywords 	*string 	`json:"meta_keywords" gorm:"type:varchar(255);"`
	MetaDescription *string 	`json:"meta_description" gorm:"type:varchar(255);"`

	// Hidden data
	ProductsCategoriesCount		int64 	`json:"_product_categories_count" gorm:"-"`

	// У страницы может быть картинка превью.. - например, для раздела услуг
	Image 			*Storage	`json:"image" gorm:"polymorphic:Owner;"`
	ProductCategories	[]ProductCategory 	`json:"product_categories" gorm:"many2many:web_page_product_categories;"`

	// Если страница временная (ну мало ли!)
	ExpiredAt 		*time.Time  `json:"expired_at"`
	
	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

func (WebPage) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WebPage{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&WebPage{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&WebPage{}).AddForeignKey("email_template_id", "email_templates(id)", "SET NULL", "CASCADE")
	// db.Model(&WebPage{}).AddForeignKey("email_box_id", "email_boxes(id)", "SET NULL", "CASCADE")
	// db.Model(&WebPage{}).AddForeignKey("users_segment_id", "users_segments(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE web_pages " +
		"ADD CONSTRAINT web_pages_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		// "ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE," +
		"ADD CONSTRAINT web_pages_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE SET NULL ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&WebPage{}, "ProductCategories", &WebPageProductCategory{})
	if err != nil {
		log.Fatal(err)
	}

}

func (webPage *WebPage) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(webPage)
	} else {
		_db = _db.Model(&WebPage{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Image","ProductCategories"})

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
func (webPage *WebPage) BeforeCreate(tx *gorm.DB) error {
	webPage.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&WebPage{}).Where("account_id = ?",  webPage.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	webPage.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (webPage *WebPage) AfterFind(tx *gorm.DB) (err error) {
	webPage.ProductsCategoriesCount =  db.Model(webPage).Association("ProductCategories").Count()
	return nil
}
func (webPage *WebPage) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.WebSiteCreated(webSite.AccountId, webSite.Id))
	// AsyncFire(NewEvent("WebPageCreated", map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id}))
	return nil
}
// ############# Entity interface #############
func (webPage WebPage) GetId() uint { return webPage.Id }
func (webPage *WebPage) setId(id uint) { webPage.Id = id }
func (webPage *WebPage) setPublicId(publicId uint) { webPage.PublicId = publicId }
func (webPage WebPage) GetAccountId() uint { return webPage.AccountId }
func (webPage *WebPage) setAccountId(id uint) { webPage.AccountId = id }
func (WebPage) SystemEntity() bool { return false }
// ############# End Entity interface #############

// ######### CRUD Functions ############
func (webPage WebPage) create() (Entity, error)  {

	item := webPage

	if err := db.Create(&item).Error; err != nil {
		return nil, err
	}

	err := item.GetPreloadDb(false, false, nil).First(&item, item.Id).Error
	if err != nil {
		return nil, err
	}

	AsyncFire(NewEvent("WebPageCreated", map[string]interface{}{"account_id":item.AccountId, "web_page_id":item.Id}))

	var newItem Entity = &item

	return newItem, nil
}
func (WebPage) get(id uint, preloads []string) (Entity, error) {

	var webPage WebPage

	err := webPage.GetPreloadDb(false,false,preloads).First(&webPage, id).Error
	if err != nil {
		return nil, err
	}
	return &webPage, nil
}
func (webPage *WebPage) load(preloads []string) error {

	err := webPage.GetPreloadDb(false,false,preloads).First(webPage, webPage.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (webPage *WebPage) loadByPublicId(preloads []string) error {
	
	if webPage.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить WebPage - не указан  Id"}
	}
	if err := webPage.GetPreloadDb(false,false, preloads).First(webPage, "account_id = ? AND public_id = ?", webPage.AccountId, webPage.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (WebPage) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return WebPage{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (WebPage) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webPages := make([]WebPage,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&WebPage{}).GetPreloadDb(false,false,preloads).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&webPages, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebPage{}).
			Where("label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {
		
		err := (&WebPage{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Where(filter).Find(&webPages).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebPage{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webPages))
	for i := range webPages {
		entities[i] = &webPages[i]
	}

	return entities, total, nil
}
func (webPage *WebPage) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	delete(input,"product_categories")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","priority","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := webPage.GetPreloadDb(false,false,nil).Where(" id = ?", webPage.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := webPage.GetPreloadDb(false,false,preloads).First(webPage, webPage.Id).Error
	if err != nil {
		return err
	}

	// fmt.Println(map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id})
	AsyncFire(NewEvent("WebPageUpdated", map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id}))
	
	return nil
}
func (webPage *WebPage) delete () error {
	if err := webPage.GetPreloadDb(true,false,nil).Where("id = ?", webPage.Id).Delete(webPage).Error; err != nil {return err}
	AsyncFire(NewEvent("WebPageDeleted", map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id}))
	return nil
}
// ######### END CRUD Functions ############

func (webPage WebPage) CreateChild(wp WebPage) (Entity, error){
	wp.ParentId = webPage.Id
	wp.WebSiteId = webPage.WebSiteId

	_webPage, err := wp.create()
	if err != nil {return nil, err}

	return _webPage, nil
}
func (webPage WebPage) AppendProductCategory(productCategory *ProductCategory, strict bool, priority *int) error {

	// 1. Загружаем продукт еще раз
	if err := productCategory.load(nil); err != nil {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, она не найдена"}
	}

	if webPage.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: нельзя добавить tag, т.к. продукта не загружен"}
	}
	if productCategory.Id < 1 {
		return utils.Error{Message: "Не создана категория продуктов"}
	}

	// 2. Проверяем есть ли уже в этой категории этот продукт
	if webPage.ExistProductCategoryById(productCategory.Id) {
		if strict {
			return utils.Error{Message: "Категория уже числиться за товаром"}
		} else {
			return nil
		}
	}

	if err := db.Model(&WebPageProductCategory{}).Create(
		&WebPageProductCategory{ProductCategoryId: productCategory.Id, WebPageId: webPage.Id, Priority: priority}).Error; err != nil {
		return err
	}

	AsyncFire(NewEvent("WebPageAppendedProductCategory",
		map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id,"product_category_id":productCategory.Id}))

	return nil
}
func (webPage WebPage) RemoveProductCategory(productCategory ProductCategory) error {

	// Загружаем еще раз
	if err := productCategory.load(nil); err != nil {
		return err
	}

	if webPage.AccountId < 1 || webPage.Id < 1  || productCategory.Id < 1 {
		return utils.Error{Message: "Техническая ошибка: account id || product card id || product category id == nil"}
	}

	if err := db.Where("web_page_id = ? AND product_category_id = ?", webPage.Id, productCategory.Id).
		Delete(&WebPageProductCategory{}).Error; err != nil {
		return err
	}

	AsyncFire(NewEvent("WebPageRemovedProductCategory",
		map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id,"product_category_id":productCategory.Id}))

	account, err := GetAccount(webPage.AccountId)
	if err == nil && account != nil {
		// AsyncFire(*Event{}.WebPageUpdated(account.Id, webPage.Id))
		// AsyncFire(*Event{}.ProductCategoryUpdated(account.Id, productCategory.Id))

		AsyncFire(NewEvent("WebPageUpdated", map[string]interface{}{"account_id":account.Id, "web_page_id":webPage.Id}))
		AsyncFire(NewEvent("ProductCategoryUpdated", map[string]interface{}{"account_id":account.Id, "product_category_id":productCategory.Id}))
	}

	return nil
}
func (webPage WebPage) SyncProductCategoriesByIds(items []WebPageProductCategory) error {

	// 1. Загружаем продукт еще раз
	if webPage.Id < 1 {
		return utils.Error{Message: "WebPage не найден"}
	}

	// очищаем список категорий
	if err := db.Model(&webPage).Association("ProductCategories").Clear(); err != nil {
		return err
	}

	for _,item := range items {
		if err := webPage.AppendProductCategory(&ProductCategory{Id: item.ProductCategoryId}, false, item.Priority); err != nil {
			return err
		}
	}

	AsyncFire(NewEvent("WebPageSyncProductCategories", map[string]interface{}{"account_id":webPage.AccountId, "web_page_id":webPage.Id}))

	return nil
}
func (webPage WebPage) ExistProductCategoryById(productCategoryId uint) bool {

	var el WebPageProductCategory

	err := db.Model(&WebPageProductCategory{}).Where("web_page_id = ? AND product_category_id = ?",webPage.Id, productCategoryId).First(&el).Error
	if err != nil {
		return false
	}

	return true
}

func (webPage WebPage) ManyToManyProductCategoryById(productCategoryId uint) (*WebPageProductCategory, error) {

	var el WebPageProductCategory

	err := db.Model(&WebPageProductCategory{}).Where("web_page_id = ? AND product_category_id = ?",webPage.Id, productCategoryId).First(&el).Error
	if err != nil {
		return nil, utils.Error{Message: "Объект не найден в на странице"}
	}

	return &el, nil
}