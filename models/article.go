package models

import (
	"database/sql"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

type Article struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	HashId 		string 	`json:"hash_id" gorm:"type:varchar(12);uniqueIndex;not null;"` // публичный Id для защиты от спама/парсинга
	WebSiteId 	*uint 	`json:"web_site_id" gorm:"type:int;index;"`

	Public	 	bool 	`json:"public" gorm:"type:bool;default:false"` // Опубликована ли статья
	Shared	 	bool 	`json:"shared" gorm:"type:bool;default:false"` // Расшарена ли статья

	Path 		*string `json:"path" gorm:"type:varchar(255);"` // идентификатор страницы url
	Breadcrumb 	*string `json:"breadcrumb" gorm:"type:varchar(255);"`
	
	Name 		*string `json:"name" gorm:"type:varchar(255);"` // Полное имя Имя статьи
	ShortName 	*string `json:"short_name" gorm:"type:varchar(255);"` // Короткое имя статьи

	Body 		*string `json:"body" gorm:"type:text;"` // pgsql: text
	Description *string `json:"description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)

	MetaTitle 			*string `json:"meta_title" gorm:"type:varchar(255);"`
	MetaKeywords 		*string `json:"meta_keywords" gorm:"type:varchar(255);"`
	MetaDescription 	*string `json:"meta_description" gorm:"type:varchar(255);"`

	// Обновлять только через AppendImage, превью изображение
	Image 				*Storage	`json:"image" gorm:"polymorphic:Owner;"` //association_autoupdate:false;

	//Attributes 		postgres.Jsonb `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	// Reviews []Review // Product reviews (отзывы на статью)
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Article) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().CreateTable(&Article{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&Article{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE articles \n    ADD CONSTRAINT articles_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT articles_web_page_id_fkey FOREIGN KEY (web_page_id) REFERENCES web_pages(id) ON DELETE SET NULL ON UPDATE CASCADE;\n    ").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (article *Article) BeforeCreate(tx *gorm.DB) error {
	article.Id = 0

	article.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Article{}).Where("account_id = ?",  article.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	article.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (article *Article) AfterFind(tx *gorm.DB) (err error) {
	return nil
}

// ############# Entity interface #############
func (article Article) GetId() uint { return article.Id }
func (article *Article) setId(id uint) { article.Id = id }
func (article *Article) setPublicId(publicId uint) { article.PublicId = publicId }
func (article Article) GetAccountId() uint { return article.AccountId }
func (article *Article) setAccountId(id uint) { article.AccountId = id }
func (Article) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (article Article) create() (Entity, error)  {

	en := article

	if err := db.Create(&en).Error; err != nil {
		return nil, err
	}

	err := en.GetPreloadDb(false,false, nil).First(&en, en.Id).Error
	if err != nil {
		return nil, err
	}

	var newItem Entity = &en

	return newItem, nil
}

func (Article) get(id uint, preloads []string) (Entity, error) {

	var article Article

	err := (&Article{}).GetPreloadDb(false,false, preloads).First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}
func (Article) getByHashId(hashId string) (*Article, error) {
	article := Article{}

	err := (&Article{}).GetPreloadDb(false,true, nil).First(&article, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}
func (article *Article) load(preloads []string) error {

	err := article.GetPreloadDb(false,false, preloads).First(article, article.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (article *Article) loadByPublicId(preloads []string) error {

	if article.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Article - не указан  Id"}
	}
	if err := article.GetPreloadDb(false,false, preloads).First(article, "account_id = ? AND public_id = ?", article.AccountId, article.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (Article) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Article{}.getPaginationList(accountId, 0, 25, sortBy, "",nil, preload)
}
func (Article) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	articles := make([]Article,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		// jsearch := search
		search = "%"+search+"%"

		err := (&Article{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).
			Where( "account_id = ?", accountId).Where(filter).
			Find(&articles, "name ILIKE ? OR short_name ILIKE ? OR body ILIKE ? OR description ILIKE ?",search, search,search,search).Error

		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Article{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR description ILIKE ? ", accountId, search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		// fmt.Println("Articles filter: ",filter)

		err := (&Article{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).
			Where( "account_id = ?", accountId).Where(filter).Find(&articles).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Article{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(articles))
	for i := range articles {
		entities[i] = &articles[i]
	}

	return entities, total, nil
}

func (article *Article) update(input map[string]interface{}, preloads []string) error {

	delete(input,"image")
	// delete(input,"preloads")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := article.GetPreloadDb(false,false,nil).Where(" id = ?", article.Id).
		Omit("id", "account_id","created_at","public_id").Updates(input).Error; err != nil {
		return err
	}

	err := article.GetPreloadDb(false,false,preloads).First(article, article.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (article *Article) delete () error {
	return article.GetPreloadDb(true,false,nil).Where("id = ?", article.Id).Delete(article).Error
}
// ######### END CRUD Functions ############

func (article *Article) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(article)
	} else {
		_db = _db.Model(&Article{})
	}

	if autoPreload {
		return db.Preload("Image", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		})

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

// ########## SELF FUNCTIONAL ############


func (account Account) GetArticleSharedByHashId(hashId string) (*Article, error) {

	article, err := Article{}.getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if !article.Shared {
		return nil, utils.Error{Message: "Article not shared"}
	}

	return article, nil
}