package models

import (
	"encoding/json"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"log"
)

type Article struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	Public	 	bool 	`json:"public" gorm:"type:bool;default:true"` // Опубликована ли статья
	Name 		string `json:"name" gorm:"type:varchar(128);default:''"` // Имя статьи


	Body 	string `json:"body" gorm:"type:text;"` // pgsql: text
	Description string `json:"description" gorm:"type:varchar(255);"` // pgsql: varchar - это зачем?)

	Images 			[]Storage 	`json:"images" gorm:"PRELOAD:true;association_autoupdate:false;polymorphic:Owner;"`  // ?gorm:""

	Attributes 		postgres.Jsonb `json:"attributes" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	// Reviews []Review // Product reviews (отзывы на статью)
	// Questions []question // вопросы по товару
	// Video []Video // видеообзоры по товару на ютубе

	Account Account `json:"-"`
}

func (Article) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Article{})
	db.Model(&Article{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (article *Article) BeforeCreate(scope *gorm.Scope) error {
	article.ID = 0
	return nil
}

// ######### INTERFACE EVENT Functions ############
func (article Article) getId() uint {
	return article.ID
}
func (article Article) getAccountId() uint {
	return article.AccountID
}
func (article Article) setAccountId(id uint) {
	article.AccountID = id
}

func (article Article) getEntityName() string {
	return "Article"
}
// ######### END OF INTERFAe Functions ############

// ######### CRUD Functions ############
func (article Article) create() (*Article, error)  {
	var newArticle = article
	err := db.Create(&newArticle).First(&newArticle).Error
	return &newArticle, err
}

func (Article) get(id uint) (*Article, error) {

	article := Article{}

	if err := db.Model(&article).Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Select(Storage{}.SelectArrayWithoutDataURL())
	}).First(&article, id).Error; err != nil {
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
	err := db.Model(article).Omit("id", "account_id").Update(input).Error
	if err != nil {
		return err
	}

	return nil
}

func (article Article) delete () error {
	return db.Model(Article{}).Where("id = ?", article.ID).Delete(article).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############
func (account Account) CreateArticle(input Article) (*Article, error) {
	input.AccountID = account.ID

	article, err := input.create()
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	go account.CallWebHookIfExist(EventArticleCreated, article)

	return article, nil
}

func (account Account) GetArticle(articleId uint) (*Article, error) {

	article, err := Article{}.get(articleId)
	if err != nil {
		return nil, err
	}

	if account.ID != article.AccountID {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
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
			Preload("Images", func(db *gorm.DB) *gorm.DB {
				return db.Select(Storage{}.SelectArrayWithoutDataURL())
			}).
			Limit(limit).
			Offset(offset).
			Where("account_id = ?", account.ID).
			Find(&articles, "id ILIKE ? OR name ILIKE ? OR description ILIKE ? OR short_description ILIKE ?" , search, search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&Article{}).Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Select(Storage{}.SelectArrayWithoutDataURL())
		}).Limit(limit).Offset(offset).Find(&articles, "account_id = ?", account.ID).Error


		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}

	// len(cards) != всему списку!
	var total uint
	err := db.Model(&Article{}).Where("account_id = ?", account.ID).Count(&total).Error
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

	if account.ID != article.AccountID {
		return nil, utils.Error{Message: "Товар принадлежит другому аккаунту"}
	}

	// parse attrs
	jsonInput, err := json.Marshal(input["attributes"])
	if err != nil {
		log.Fatal("Eroror json: ", err)
	}
	article.Attributes = postgres.Jsonb{RawMessage: jsonInput}

	err = article.update(input)
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	go account.CallWebHookIfExist(EventArticlesUpdate, article)

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

	go account.CallWebHookIfExist(EventArticleDeleted, article)

	return nil
}
// ######### END OF ACCOUNT Functions ############


// ########## SELF FUNCTIONAL ############
