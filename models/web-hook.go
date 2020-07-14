package models

import (
	"bytes"
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
	"text/template"
)

type EventTypeOld = string

type WebHookEventObject interface {
	getId() uint
}

const (
	EventShopCreated 	EventTypeOld = "ShopCreated"
	EventShopUpdated 	EventTypeOld = "ShopUpdated"
	EventShopDeleted 	EventTypeOld = "ShopDeleted"
	EventShopsUpdate 	EventTypeOld = "ShopsUpdate"

	EventProductCreated 	EventTypeOld = "ProductCreated"
	EventProductUpdated 	EventTypeOld = "ProductUpdated"
	EventProductDeleted 	EventTypeOld = "ProductDeleted"
	EventProductsUpdate 		EventTypeOld = "ProductsUpdate"

	EventProductCardCreated 	EventTypeOld = "ProductCardCreated"
	EventProductCardUpdated 	EventTypeOld = "ProductCardUpdated"
	EventProductCardDeleted 	EventTypeOld = "ProductCardDeleted"
	EventProductCardsUpdate 	EventTypeOld = "ProductCardsUpdate"

	EventProductGroupCreated 	EventTypeOld = "ProductGroupCreated"
	EventProductGroupUpdated 	EventTypeOld = "ProductGroupUpdated"
	EventProductGroupDeleted 	EventTypeOld = "ProductGroupDeleted"
	EventProductGroupsUpdate 	EventTypeOld = "ProductGroupsUpdate"

	EventArticleCreated 	EventTypeOld = "ArticleCreated"
	EventArticleUpdated 	EventTypeOld = "ArticleUpdated"
	EventArticleDeleted 	EventTypeOld = "ArticleDeleted"
	EventArticlesUpdate 	EventTypeOld = "ArticlesUpdate"

	EventUpdateAllShopData 	EventTypeOld = "UpdateAllShopData"
)

type WebHook struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	EventType	EventTypeOld 	`json:"eventType" gorm:"type:varchar(128);default:''"` // Имя события

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя вебхука
	Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)
	URL 		string 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
	HttpMethod		string `json:"httpMethod" gorm:"type:varchar(15);default:'get';"` // Тип вызова (GET, POST, PUT, puth и т.д.)
	//URLTemplate 		template.Template 	`json:"url" gorm:"type:varchar(255);"` // вызов, который совершается
}

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
func (webHook WebHook) create() (*WebHook, error) {
	var whNew = webHook
	err := db.Create(&whNew).First(&whNew).Error
	return &whNew, err
}

func (WebHook) get(id uint) (*WebHook, error) {

	wh := WebHook{}

	if err := db.First(&wh, id).Error; err != nil {
		return nil, err
	}

	return &wh, nil
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

func (webHook *WebHook) update(input interface{}) error {
	return db.Model(webHook).Omit("id", "account_id").Update(input).Error

}

func (webHook WebHook) delete () error {
	return db.Model(WebHook{}).Where("id = ?", webHook.ID).Delete(webHook).Error
}
// ######### END CRUD Functions ############

func (account Account) CreateWebHook(input WebHook) (*WebHook, error) {
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

func (account Account) GetWebHookByEvent(eventType EventTypeOld) (*WebHook, error) {

	wh, err := WebHook{}.getByEvent(eventType)
	if err != nil {
		return nil, err
	}

	if wh.AccountID != account.ID {
		return nil, utils.Error{Message: "WebHook принадлежит другому аккаунту"}
	}

	return wh, nil
}

func (account Account) CallWebHookIfExist(eventType EventTypeOld, object WebHookEventObject) bool {

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
}

// ##################

func (webHook WebHook) Call(entity WebHookEventObject) bool {

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

	// fmt.Println(url)

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
}