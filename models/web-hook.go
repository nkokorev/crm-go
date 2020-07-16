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
func (webHook WebHook) getId() uint { return webHook.ID }
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
func (WebHook) TableName() string {
	return "web_hooks"
}

// ######### CRUD Functions ############
/*func (webHook WebHook) create() (*WebHook, error) {
	var whNew = webHook
	err := db.Create(&whNew).First(&whNew).Error
	return &whNew, err
}*/
func (webHook WebHook) create() (Entity, error)  {
	var newItem Entity = &webHook

	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
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

func (WebHook) getListByAccount(accountId uint) ([]WebHook, error) {

	whs := make([]WebHook,0)

	err := db.Find(&whs, "accountId = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return whs, nil

}

/*func (webHook *WebHook) update(input interface{}) error {
	return db.Model(webHook).Omit("id", "account_id").Update(input).Error

}*/
func (webHook *WebHook) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(webHook).Omit("id", "account_id").Update(input).Error
}

func (webHook WebHook) delete () error {
	return db.Model(WebHook{}).Where("id = ?", webHook.ID).Delete(webHook).Error
}
// ######### END CRUD Functions ############

/*func (account Account) CreateWebHook(input WebHook) (*WebHook, error) {
	input.AccountID = account.ID
	return input.create()
}

func (account Account) GetWebHook(id uint) (*WebHook, error) {

	wh, err := WebHook{}.get(id)
	if err != nil {
		return nil, err
	}

	if wh.AccountID != account.ID {
		return nil, utils.Error{Message: "WebHook принадлежит другому аккаунту"}
	}

	return wh, nil
}

func (account Account) GetWebHookByEvent(eventType WebHookType) (*WebHook, error) {

	wh, err := WebHook{}.getByEvent(eventType)
	if err != nil {
		return nil, err
	}

	if wh.AccountID != account.ID {
		return nil, utils.Error{Message: "WebHook принадлежит другому аккаунту"}
	}

	return wh, nil
}

func (account Account) CallWebHookIfExist(eventType WebHookType, object WebHookEventObject) bool {

	webHook, err := account.GetWebHookByEvent(eventType)
	if err != nil {
		return false
	}

	return webHook.Call(object)
}

func (account Account) GetWebHooks() ([]WebHook, error) {
	return WebHook{}.getListByAccount(account.ID)
}
func (account Account) GetWebHooksPaginationList(offset, limit int, search string) ([]WebHook, int, error) {

	webHooks := make([]WebHook,0)
	//groups := []ProductGroup{}

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&WebHook{}).
			Limit(limit).
			Offset(offset).
			Where("account_id = ?", account.ID).
			Where("url ILIKE ? OR name ILIKE ? OR description ILIKE ?" , search,search,search).
			Order("id").
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&WebHook{}).
			Limit(limit).
			Offset(offset).
			Where("account_id = ?", account.ID).
			Order("id").
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}
	var total int
	if err := db.Model(&WebHook{}).Where("account_id = ?", account.ID).Count(&total).Error; err != nil {
		return nil, 0, utils.Error{Message: "Ошибка получения числа вебхуков"}
	}

	return webHooks, total, nil
}

func (account Account) UpdateWebHook(webHookId uint, input interface{}) (*WebHook, error) {

	webHook, err := account.GetWebHook(webHookId)
	if err != nil {
		return nil, err
	}

	err = webHook.update(input)
	if err != nil {
		return nil, err
	}

	return webHook, nil

}

func (account Account) DeleteWebHook(webHookId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	webHook, err := account.GetWebHook(webHookId)
	if err != nil {
		return err
	}

	return webHook.delete()
}*/

// ##################

func (webHook WebHook) Call(e event.Event) error {

	tplUrl, err := template.New("url").Parse(webHook.URL)
	if err != nil {
		// log.Println("Error parse URL: ", err)
		return utils.Error{Message: fmt.Sprintf("Error parse URL: %v", err)}
	}

	urlB := new(bytes.Buffer)
	var data interface{}

	if e != nil && e.Data() != nil {
		data = e.Data()
	}

	err = tplUrl.Execute(urlB, data)
	if err != nil {
		// log.Println("Error Execute URL: ", err)
		return utils.Error{Message: fmt.Sprintf("Error execute URL: %v", err)}
	}


	url := urlB.String()

	// fmt.Println("URL: ", url)

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

/*func (webHook WebHook) Call(entity WebHookEventObject) bool {

	tplUrl, err := template.New("url").Parse(webHook.URL)
	if err != nil {
		//fmt.Println("Error parse URL: ", err)
		return false
	}

	urlB := new(bytes.Buffer)
	err = tplUrl.Execute(urlB, entity)
	if err != nil {
		return false
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
		//fmt.Println(err)
		return false
	}
	defer response.Body.Close()

	return true
}*/