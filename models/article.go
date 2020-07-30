package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"strings"
	"time"
)

type Article struct {
	Id     		uint   `json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	HashId 		string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга

	Public	 	bool 	`json:"public" gorm:"type:bool;default:false"` // Опубликована ли статья
	Shared	 	bool 	`json:"shared" gorm:"type:bool;default:false"` // Расшарена ли статья

	URL 		string `json:"url" gorm:"type:varchar(255);"` // идентификатор страницы url
	Breadcrumb 	string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	
	Name 		string `json:"name" gorm:"type:varchar(255);"` // Полное имя Имя статьи
	ShortName 	string `json:"shortName" gorm:"type:varchar(255);default:NULL"` // Короткое имя статьи


	Body 		string `json:"body" gorm:"type:text;"` // pgsql: text
	Description string `json:"description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)

	MetaTitle 			string `json:"metaTitle" gorm:"type:varchar(255);default:null;"`
	MetaKeywords 		string `json:"metaKeywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription 	string `json:"metaDescription" gorm:"type:varchar(255);default:null;"`

	// Обновлять только через AppendImage
	Image 				*Storage	`json:"image" gorm:"polymorphic:Owner;"` //association_autoupdate:false;

	//Attributes 		postgres.Jsonb `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	// Reviews []Review // Product reviews (отзывы на статью)
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (Article) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Article{})
	db.Model(&Article{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (article *Article) BeforeCreate(scope *gorm.Scope) error {
	article.Id = 0
	article.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}
func (article *Article) AfterCreate(scope *gorm.Scope) (error) {
	event.AsyncFire(Event{}.ArticleCreated(article.AccountId, article.Id))
	return nil
}
func (article *Article) AfterUpdate(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.ArticleUpdated(article.AccountId, article.Id))

	return nil
}
func (article *Article) AfterDelete(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.ArticleDeleted(article.AccountId, article.Id))
	return nil
}

// ######### INTERFACE EVENT Functions ############
func (article Article) getId() uint {
	return article.Id
}
func (article Article) getAccountId() uint {
	return article.AccountId
}
func (article Article) setAccountId(id uint) {
	article.AccountId = id
}
func (article Article) getEntityName() string {
	return "Article"
}
// ######### END OF INTERFAe Functions ############

// ######### CRUD Functions ############
func (article Article) create() (*Article, error)  {
	var newArticle = article
	err := db.Create(&newArticle).Error
	if err != nil {return nil, err}
	return &newArticle, err
}
func (Article) get(id uint) (*Article, error) {

	article := Article{}

	if err := db.Model(&article).Preload("Image", func(db *gorm.DB) *gorm.DB {
		return db.Select(Storage{}.SelectArrayWithoutDataURL())
	}).First(&article, id).Error; err != nil {
		return nil, err
	}

	return &article, nil
}
func (Article) getByHashId(hashId string) (*Article, error) {

	article := Article{}

	if err := db.Model(&article).Preload("Image", func(db *gorm.DB) *gorm.DB {
		return db.Select(Storage{}.SelectArrayWithoutDataURL())
	}).First(&article, "hash_id = ?", hashId).Error; err != nil {
		return nil, err
	}

	return &article, nil
}
func (Article) getList(accountId uint) ([]Article, error) {

	articles := make([]Article,0)

	err := db.Model(&Article{}).Find(&articles, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return articles, nil
}
func (article *Article) update(input map[string]interface{}) error {
	// err := db.Set("gorm:association_autoupdate", false).Model(article).Omit("id", "account_id").Update(input).Error
	err := db.Set("gorm:association_autoupdate", false).Model(article).Omit("id", "account_id").Updates(input).Error

	// err := db.Debug().Model(&Article{}).Omit("accountHashId").Select("name", "shortName").Where("id = ?", article.Id).Update(input).Error
	if err != nil {
		return err
	}

	return nil
}
  func (article *Article) delete () error {
	return db.Model(Article{}).Where("id = ?", article.Id).Delete(article).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateArticle(input Article) (*Article, error) {
	input.AccountId = account.Id

	article, err := input.create()
	if err != nil {
		return nil, err
	}

	// event.AsyncFire(Event{}.ArticleCreated(account.Id, article.Id))

	return article, nil
}

func (account Account) GetArticle(articleId uint) (*Article, error) {

	article, err := Article{}.get(articleId)
	if err != nil {
		return nil, err
	}

	if account.Id != article.AccountId {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	return article, nil
}

func (account Account) GetArticleByHashId(hashId string) (*Article, error) {

	article, err := Article{}.getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if account.Id != article.AccountId {
		return nil, utils.Error{Message: "С принадлежит другому аккаунту"}
	}

	return article, nil
}

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

func (account Account) GetArticleListPagination(offset, limit int, search string) ([]Article, uint, error) {

	articles := make([]Article,0)

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&Article{}).
			Preload("Image", func(db *gorm.DB) *gorm.DB {
				return db.Select(Storage{}.SelectArrayWithoutDataURL())
			}).
			Limit(limit).
			Offset(offset).
			Order("id").
			Where("account_id = ?", account.Id).
			Find(&articles, "name ILIKE ? OR short_name ILIKE ? OR body ILIKE ? OR description ILIKE ?" , search, search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&Article{}).Preload("Image", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		}).Limit(limit).Offset(offset).Order("id").Find(&articles, "account_id = ?", account.Id).Error


		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}

	// len(cards) != всему списку!
	var total uint
	err := db.Model(&Article{}).Where("account_id = ?", account.Id).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема"}
	}

	return articles, total, nil
}

func (account Account) UpdateArticle(articleId uint, input map[string]interface{}) (*Article, error) {

	article, err := account.GetArticle(articleId)
	if err != nil {
		return nil, err
	}

	if account.Id != article.AccountId {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	err = article.update(input)
	if err != nil {
		return nil, err
	}

	// event.AsyncFire(Event{}.ArticleUpdated(account.Id, article.Id))

	return article, err

}

func (account Account) DeleteArticle(articleId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	article, err := account.GetArticle(articleId)
	if err != nil {
		return err
	}

	err = article.delete()
	if err !=nil { return err }

	// event.AsyncFire(Event{}.ArticleDeleted(account.Id, article.Id))

	return nil
}
// ######### END OF ACCOUNT Functions ############


// ########## SELF FUNCTIONAL ############
