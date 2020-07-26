package uiApiCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

/**
	1. Определяем target: Customer User. Нужно создавать нового или ищем среди существующих по userHashId.
		
	2.
 */

type CreateOrderForm struct {
	// Факт существующего пользователя
	UserHashId		string	`json:"userHashId"`

	// Object для создания нового пользователя (альтернатива)
	// Если окажется, что персональные данные заняты - заказ будет приписан существующему (?)
	// todo: - создать идеологию обработки заказов и вывести настройки в модель. Можно запретить создавать новых пользователей....
	Customer	Customer	`json:"customer"`

	// Компании - это еще будет доработано.
	// todo: если заказчик юр.лицо, будет требоваться CustomerCompany для создания заказа или поиск модели Company - hashId
	CompanyHashId			string	`json:"companyHashId"`
	
	// hashId персонального менеджера (если такой есть)
	PersonalManagerHashId	string	`json:"personalManagerHashId"`

	// ID магазина, от имени которого создается заявка
	WebSiteId 		uint	`json:"webSiteId"`

	// true = частное лицо, false - юрлицо
	Individual		bool 	`json:"individual"`

	// подписка на новости
	SubscribeNews		bool 	`json:"subscribeNews"`

	CustomerComment	string	`json:"customerComment"`

	// передается по Code, т.к. ID может поменяться.
	// todo: !!!	сделать настройки для каждого канала по принятым заявкам	!!!
	OrderChannel 	string 	`json:"orderChannel"`

	// Выбирается канал доставки, а также все необходимые данные
	// todo: для каждого канала сделать доступным метод доставки (по умолчанию: все)
	Delivery 	DeliveryData	`json:"delivery"`
	
	// Собственно, сама корзина
	Cart []CartData `json:"cart"`
	/*Cart map[string]struct{
		ProductId 	uint `json:"productId"` // id product
		Count 		uint `json:"count"`      // число позиций
	} `json:"cart"`*/

}

// todo: список обязательных полей - дело настроек OrderSettings
type Customer struct {
	Email 		string `json:"email" ` // email должен проходить deepValidate()
	Phone		string `json:"phone" ` // форматируем сами телефон

	Name 		string `json:"name"`
	Surname 	string `json:"surname" `
	Patronymic 	string `json:"patronymic"`
}
type DeliveryData struct {
	Id 		uint 	`json:"id"` 	// id доставки в ее таблице
	Code 	string 	`json:"code"`
	
	Address		string 	`json:"address"` 		// адрес доставки
	PostalCode	string 	`json:"postalCode"`
}
type CartData struct {
	ProductId 	uint	`json:"productId"`	// id product
	Quantity	uint	`json:"quantity"`	// число позиций
}


func UiApiOrderCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	/*var input struct{
		models.Order
	}*/
	var input CreateOrderForm

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

/*	order, err := account.CreateEntity(&input.Order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания ключа"}))
		return
	}*/

	resp := u.Message(true, "POST Order Created")
	resp["order"] = input
	u.Respond(w, resp)
}