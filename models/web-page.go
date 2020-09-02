package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
	"time"
)

type WebPage struct {
	Id     		uint	`json:"id" gorm:"primaryKey"`
	PublicId	uint	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId 	uint 	`json:"web_site_id" gorm:"type:int;index;not null;"`
	ParentId 	uint	`json:"parent_id"`
	// Children 	*WebPage `json:"_children" gorm:"-"`

	// code for scope routes (группы, категории....)
	Code 		*string `json:"code" gorm:"type:varchar(255);"`

	// Routing
	Path 		*string `json:"path" gorm:"type:varchar(255);"`			// Имя пути - catalog, cat, /, ..
	Label 		*string `json:"label" gorm:"type:varchar(255);"` 		// menu label - Чай, кофе, ..
	RouteName 	*string `json:"route_name" gorm:"type:varchar(50);"` 	// route name: delivery, info.index, cart
	IconName 	*string `json:"icon_name" gorm:"type:varchar(50);"` 	// icon name

	Order 				int		`json:"order" gorm:"type:int;default:10;"` // Порядок отображения в текущей иерархии категории
	Breadcrumb 			*string `json:"breadcrumb" gorm:"type:varchar(255);"`
	ShortDescription 	*string `json:"short_description" gorm:"type:varchar(255);"`
	Description 		*string `json:"description" gorm:"type:text;"`

	MetaTitle 		*string 	`json:"meta_title" gorm:"type:varchar(255);"`
	MetaKeywords 	*string 	`json:"meta_keywords" gorm:"type:varchar(255);"`
	MetaDescription *string 	`json:"meta_description" gorm:"type:varchar(255);"`
	Image 			*Storage	`json:"image" gorm:"polymorphic:Owner;"`

	// Если страница временная (ну мало ли!)
	ExpiredAt 		*time.Time  `json:"expired_at"`
	
	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}

// ############# Entity interface #############
func (webPage WebPage) GetId() uint { return webPage.Id }
func (webPage *WebPage) setId(id uint) { webPage.Id = id }
func (webPage *WebPage) setPublicId(publicId uint) { webPage.PublicId = publicId }
func (webPage WebPage) GetAccountId() uint { return webPage.AccountId }
func (webPage *WebPage) setAccountId(id uint) { webPage.AccountId = id }
func (WebPage) SystemEntity() bool { return false }
// ############# End Entity interface #############
func (WebPage) PgSqlCreate() {
	db.Migrator().CreateTable(&WebPage{})
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

	return nil
}
// ######### CRUD Functions ############
func (webPage WebPage) create() (Entity, error)  {

	en := webPage

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
func (WebPage) get(id uint) (Entity, error) {

	var webPage WebPage

	err := webPage.GetPreloadDb(false,false,true).First(&webPage, id).Error
	if err != nil {
		return nil, err
	}
	return &webPage, nil
}
func (webPage *WebPage) load() error {

	err := webPage.GetPreloadDb(false,false,true).First(webPage, webPage.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (webPage *WebPage) loadByPublicId() error {
	
	if webPage.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить WebPage - не указан  Id"}
	}
	if err := webPage.GetPreloadDb(false,false, true).First(webPage, "account_id = ? AND public_id = ?", webPage.AccountId, webPage.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (WebPage) getList(accountId uint, sortBy string) ([]Entity, int64, error) {
	return WebPage{}.getPaginationList(accountId, 0, 100, sortBy, "",nil)
}
func (WebPage) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{}) ([]Entity, int64, error) {

	emailCampaigns := make([]WebPage,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		search = "%"+search+"%"

		err := (&WebPage{}).GetPreloadDb(true,false,true).Limit(limit).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns, "label ILIKE ? OR code ILIKE ? OR route_name ILIKE ? OR icon_name ILIKE ? OR meta_title ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ? OR description ILIKE ?", search,search,search,search,search,search,search,search).Error

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
		
		err := (&WebPage{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&emailCampaigns).Error
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
	entities := make([]Entity,len(emailCampaigns))
	for i := range emailCampaigns {
		entities[i] = &emailCampaigns[i]
	}

	return entities, total, nil
}
func (webPage *WebPage) update(input map[string]interface{}) error {

	delete(input,"image")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"parent_id","order","web_site_id"}); err != nil {
		return err
	}
	input = utils.FixInputDataTimeVars(input,[]string{"expired_at"})

	if err := webPage.GetPreloadDb(true,false,false).Where(" id = ?", webPage.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := webPage.GetPreloadDb(true,false,false).First(webPage, webPage.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (webPage *WebPage) delete () error {
	return webPage.GetPreloadDb(true,true,false).Where("id = ?", webPage.Id).Delete(webPage).Error
}
// ######### END CRUD Functions ############

func (webPage *WebPage) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(webPage)
	} else {
		_db = _db.Model(&WebPage{})
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

func (webPage WebPage) CreateChild(wp WebPage) (Entity, error){
	wp.ParentId = webPage.Id
	wp.WebSiteId = webPage.WebSiteId

	_webPage, err := wp.create()
	if err != nil {return nil, err}

	return _webPage, nil
}
