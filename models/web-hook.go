package models

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
	"text/template"
)

type WebHookType = string

type WebHookEventObject interface {
	getId() uint
}

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

	EventProductGroupCreated 	WebHookType = "ProductGroupCreated"
	EventProductGroupUpdated 	WebHookType = "ProductGroupUpdated"
	EventProductGroupDeleted 	WebHookType = "ProductGroupDeleted"
	EventProductGroupsUpdate 	WebHookType = "ProductGroupsUpdate"

	EventArticleCreated 	WebHookType = "ArticleCreated"
	EventArticleUpdated 	WebHookType = "ArticleUpdated"
	EventArticleDeleted 	WebHookType = "ArticleDeleted"
	EventArticlesUpdate 	WebHookType = "ArticlesUpdate"

	EventUpdateAllShopData 	WebHookType = "UpdateAllShopData"
)

type WebHook struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя вебхука
	
	Code		WebHookType `json:"code" gorm:"type:varchar(128);default:''"` // Имя события

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	
	Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)
	URL 		string 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
	HttpMethod		string `json:"httpMethod" gorm:"type:varchar(15);default:'get';"` // Тип вызова (GET, POST, PUT, puth и т.д.)
	//URLTemplate 		template.Template 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
}

// ############# Entity interface #############
func (webHook WebHook) GetId() uint { return webHook.ID }
func (webHook *WebHook) setId(id uint) { webHook.ID = id }
func (webHook WebHook) GetAccountId() uint { return webHook.AccountID }
func (webHook *WebHook) setAccountId(id uint) { webHook.AccountID = id }
func (WebHook) systemEntity() bool { return false }

// ############# Entity interface #############

func (WebHook) PgSqlCreate() {
	db.CreateTable(&WebHook{})
	db.Model(&WebHook{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (webHook *WebHook) BeforeCreate(scope *gorm.Scope) error {
	webHook.ID = 0
	return nil
}

// ######### CRUD Functions ############
func (webHook WebHook) create() (Entity, error)  {
	// if err := db.Create(&webHook).Find(&webHook, webHook.ID).Error; err != nil {
	wb := webHook
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (WebHook) get(id uint) (Entity, error) {

	var webHook WebHook

	err := db.First(&webHook, id).Error
	if err != nil {
		return nil, err
	}
	return &webHook, nil
}
func (webHook *WebHook) load() error {
	if webHook.ID < 1 {
		return utils.Error{Message: "Невозможно загрузить WebHook - не указан  ID"}
	}

	err := db.First(webHook).Error
	if err != nil {
		return err
	}
	return nil
}

func (WebHook) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webHooks := make([]WebHook,0)
	var total uint

	err := db.Model(&WebHook{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&webHooks).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&WebHook{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (WebHook) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]WebHook,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&WebHook{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebHook{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&WebHook{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	for i,_ := range webHooks {
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

func (webHook *WebHook) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(webHook).Omit("id", "account_id").Updates(input).Error
}

func (webHook WebHook) delete () error {
	return db.Model(WebHook{}).Where("id = ?", webHook.ID).Delete(webHook).Error
}
// ######### END CRUD Functions ############

func (webHook WebHook) Execute(e event.Event) error {

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
	var data interface{}

	// Данные для расчета url вебхука
	if e != nil && e.Data() != nil {
		data = e.Data()
	}

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
