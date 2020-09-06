package models

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type WebHookType = string

const (
	EventShopCreated 	WebHookType = "ShopCreated"
	EventShopUpdated 	WebHookType = "ShopUpdated"
	EventShopDeleted 	WebHookType = "ShopDeleted"
	EventShopsUpdate 	WebHookType = "ShopsUpdate"

	EventProductCreated 	WebHookType = "ProductCreated"
	EventProductUpdated 	WebHookType = "ProductUpdated"
	EventProductDeleted 	WebHookType = "ProductDeleted"
	EventProductsUpdate 		WebHookType = "ProductsUpdate"

	EventProductCardCreated 	WebHookType = "ProductCardCreated"
	EventProductCardUpdated 	WebHookType = "ProductCardUpdated"
	EventProductCardDeleted 	WebHookType = "ProductCardDeleted"
	EventProductCardsUpdate 	WebHookType = "ProductCardsUpdate"

	EventWebPageCreated 	WebHookType = "WebPageCreated"
	EventWebPageUpdated 	WebHookType = "WebPageUpdated"
	EventWebPageDeleted 	WebHookType = "WebPageDeleted"
	EventWebPagesUpdate 	WebHookType = "WebPagesUpdate"

	EventArticleCreated 	WebHookType = "ArticleCreated"
	EventArticleUpdated 	WebHookType = "ArticleUpdated"
	EventArticleDeleted 	WebHookType = "ArticleDeleted"
	EventArticlesUpdate 	WebHookType = "ArticlesUpdate"

	EventUpdateAllShopData 	WebHookType = "UpdateAllShopData"
)

type WebHook struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя вебхука
	
	Code		WebHookType `json:"code" gorm:"type:varchar(128);default:''"` // Имя события

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	
	Description string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)
	URL 		string 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
	HttpMethod	string `json:"http_method" gorm:"type:varchar(15);default:'get';"` // Тип вызова (GET, POST, PUT, puth и т.д.)
}

// ############# Entity interface #############
func (webHook WebHook) GetId() uint { return webHook.Id }
func (webHook *WebHook) setId(id uint) { webHook.Id = id }
func (webHook *WebHook) setPublicId(id uint) {webHook.PublicId = id}
func (webHook WebHook) GetAccountId() uint { return webHook.AccountId }
func (webHook *WebHook) setAccountId(id uint) { webHook.AccountId = id }
func (WebHook) SystemEntity() bool { return false }

// ############# Entity interface #############

func (WebHook) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&WebHook{}); err != nil {log.Fatal(err)}
	// db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE web_hooks ADD CONSTRAINT web_hooks_conditions_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (webHook *WebHook) BeforeCreate(tx *gorm.DB) error {
	webHook.Id = 0
	var lastIdx sql.NullInt64

	row := db.Model(&WebHook{}).Where("account_id = ?",  webHook.AccountId).
		Select("max(public_id)").Row()
	if row != nil {
		err := row.Scan(&lastIdx)
		if err != nil && err != gorm.ErrRecordNotFound { return err }
	}

	webHook.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}

// ######### CRUD Functions ############
func (webHook WebHook) create() (Entity, error)  {

	_item := webHook
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,true, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}
func (WebHook) get(id uint, preloads []string) (Entity, error) {

	var item CartItem

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (webHook *WebHook) load(preloads []string) error {
	if webHook.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := webHook.GetPreloadDb(false, false, preloads).First(webHook, webHook.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (webHook *WebHook) loadByPublicId(preloads []string) error {

	if webHook.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}

	if err := webHook.GetPreloadDb(false,false, preloads).
		First(webHook, "account_id = ? AND public_id = ?", webHook.AccountId, webHook.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (WebHook) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return WebHook{}.getPaginationList(accountId,0,25,sortBy, "", nil, preload)
}
func (WebHook) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webHooks := make([]WebHook,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&WebHook{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&WebHook{}).GetPreloadDb(false, false, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&WebHook{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebHook{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}
func (WebHook) getByEvent(eventName string) (*WebHook, error) {

	wh := WebHook{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}
func (webHook *WebHook) update(input map[string]interface{}, preloads []string) error {

	// delete(input,"product")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := webHook.GetPreloadDb(false, false, nil).Where("id = ?", webHook.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := webHook.GetPreloadDb(false,false, preloads).First(webHook, webHook.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (webHook *WebHook) delete () error {
	return db.Model(WebHook{}).Where("id = ?", webHook.Id).Delete(webHook).Error
}
// ######### END CRUD Functions ############

func (webHook WebHook) Execute(data map[string]interface{}) error {

	// проверка
	if !webHook.Enabled || webHook.URL == "" {
		return utils.Error{Message: "Не корректные данные ВебХука"}
	}

	// Создаем шаблон для вычисления URL
	tplUrl, err := template.New("url").Parse(webHook.URL)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Error parse URL: %v", err)}
	}

	urlB := new(bytes.Buffer)
	/*var data interface{}

	if e != nil && e.Data() != nil {
		data = e.Data()
	}*/

	err = tplUrl.Execute(urlB, data)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Error execute URL: %v", err)}
	}

	url := urlB.String()

	var response *http.Response
	var request *http.Request

	switch webHook.HttpMethod {

		case http.MethodPost:
			response, err = http.Post(url, "application/json", nil)

		case http.MethodGet:
			response, err = http.Get(url)

		case http.MethodPatch, http.MethodPut:
			client := &http.Client{}
			request, err = http.NewRequest("PATCH", url, strings.NewReader(""))
			if err != nil {
				break
			}
			response, err = client.Do(request)

		case http.MethodDelete:
			client := &http.Client{}
			request, err = http.NewRequest("DELETE", url, nil)
			response, err = client.Do(request)
	}

	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Error execute URL: %v", err)}
	}
	defer response.Body.Close()

	return nil
}

func (webHook WebHook) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&webHook)
	} else {
		_db = _db.Model(&WebHook{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{""})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}
