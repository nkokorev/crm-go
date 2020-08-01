package uiApiCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"github.com/ttacon/libphonenumber"
	"log"
	"net/http"
)

/**
	1. Определяем target: Customer User. Нужно создавать нового или ищем среди существующих по userHashId.
		
	2.
 */

type CreateOrderForm struct {
	// Факт существующего пользователя
	CustomerHashId		string	`json:"customerHashId"`

	// Object для создания нового пользователя (альтернатива)
	// Если окажется, что персональные данные заняты - заказ будет приписан существующему (?)
	// todo: - создать идеологию обработки заказов и вывести настройки в модель. Можно запретить создавать новых пользователей....
	Customer	Customer	`json:"customer"`

	// Компании - это еще будет доработано.
	// todo: если заказчик юр.лицо, будет требоваться CustomerCompany для создания заказа или поиск модели Company - hashId
	CompanyHashId			string	`json:"companyHashId"`
	
	// hashId персонального менеджера (если такой есть)
	ManagerHashId	string	`json:"managerHashId"`

	// ID магазина, от имени которого создается заявка
	WebSiteId 		uint	`json:"webSiteId"`

	// true = частное лицо, false - юрлицо
	Individual		bool 	`json:"individual"`

	// подписка на новости
	SubscribeNews		bool 	`json:"subscribeNews"`

	CustomerComment	string	`json:"customerComment"`

	// передается по Code, т.к. ID может поменяться.
	// todo: !!!	сделать настройки для каждого канала по принятым заявкам	!!!
	OrderChannelCode	string 	`json:"orderChannel"`

	// Выбирается канал доставки, а также все необходимые данные
	// todo: для каждого канала сделать доступным метод доставки (по умолчанию: все)
	Delivery 	models.DeliveryData	`json:"delivery"`
	
	// Собственно, сама корзина
	Cart []models.CartData `json:"cart"`

	// Способ оплаты PaymentOptions          // online, cashe..
	PaymentMethodCode string `json:"paymentMethodCode"`
	PaymentMethodId uint `json:"paymentMethodId"`
}

// todo: список обязательных полей - дело настроек OrderSettings
type Customer struct {
	Email 		string `json:"email" ` // email должен проходить deepValidate()
	Phone		string `json:"phone" ` // форматируем сами телефон

	Name 		string `json:"name"`
	Surname 	string `json:"surname" `
	Patronymic 	string `json:"patronymic"`
}

func UiApiOrderCreate(w http.ResponseWriter, r *http.Request) {

	/*u.Respond(w, u.MessageError(u.Error{Message:"Указанный тип оплаты не поддерживает данный тип доставки",
		Errors: map[string]interface{}{
		"phone":"Указанный номер телефона занят",
		"email":"Обязательно нужно указать",
		"surname":"А где фамилия?",
		"name":"Укажите имя",
		}}))
	return*/

	// Итоговая стоимость заказа
	var totalCost float64
	totalCurrency := "RUB"
	totalCost = 0
	
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Читаем вход
	var input CreateOrderForm
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте обязательные поля и типы переменных"))
		return
	}

	if err := validateOrder(input); err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}

	// 1. Получаем магазин из контекста
	var webSite models.WebSite
	if err := account.LoadEntity(&webSite, input.WebSiteId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте id магазина"))
		return
	}

	// 6. Находим способ оплаты
	if input.PaymentMethodCode == "" || input.PaymentMethodId < 1 {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска способа оплаты", Errors: map[string]interface{}{"paymentMethodCode":"Необходимо указать способ оплаты"}}))
		return
	}
	/*paymentOption, err := account.GetPaymentOptionByCode(input.PaymentOptionsCode)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска способа оплаты", Errors: map[string]interface{}{"orderChannel":"Способ оплаты не найден"}}))
		return
	}*/
	paymentMethod, err := account.GetPaymentMethod(input.PaymentMethodCode, input.PaymentMethodId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска способа оплаты", Errors: map[string]interface{}{"orderChannel":"Способ оплаты не найден"}}))
		return
	}

	// 3. Находим тип доставки
	if  input.Delivery.Code == "" ||  input.Delivery.Id < 1 {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в определении типа доставки", Errors: map[string]interface{}{"delivery":"не указан тип доставки или id"}}))
		return
	}
	delivery, err := webSite.GetDelivery(input.Delivery.Code, input.Delivery.Id)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в определении типа доставки", Errors: map[string]interface{}{"delivery":"тип доставки или id указаны не верно"}}))
		return
	}

	// Проверяем тип доставки и способ оплаты
	if !delivery.ExistPaymentMethod(paymentMethod) {
		u.Respond(w, u.MessageError(u.Error{Message:"Указанный тип оплаты не поддерживает данный тип доставки",
			Errors: map[string]interface{}{"paymentMethod":"Указанный тип оплаты не поддерживает данный тип доставки"}}))
		return
	}
	
	// 2. Создаем список продуктов, считаем стоимость каждого
	var cartItems []models.CartItem
	for _,v := range input.Cart {
		
		// 1.1 Получаем продукт
		product, err := account.GetProduct(v.ProductId)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска ответственного менеджера", Errors: map[string]interface{}{"managerHashId":"Не удалось найти менеджера"}}))
			return
		}

		// 1.2 Если продукт недоступен к заказу
		if !product.Enabled {
			u.Respond(w, u.MessageError(u.Error{Message:fmt.Sprintf("Заказ содержит продукты недоступные к заказу: %v", product.Name), Errors: map[string]interface{}{"cart":"Не корректный список продуктов"}}))
			return
		}

		// 1.3 Считаем цену товара с учетом скидок
		ProductCost := product.RetailPrice - product.RetailDiscount

		// 1.4 Формируем и добавляем Cart Item в общий список
		cartItems = append(cartItems, models.CartItem{
			AccountId: account.Id,
			ProductId: product.Id,
			Description: product.Name,
			Quantity: v.Quantity,
			Amount: models.PaymentAmount{Value: ProductCost, Currency: "RUB"},
			VatCode: product.VatCodeId,
			// OrderId: order.Id,
		})

		// 1.5 Считаем общую стоимость заказа
		totalCost += ProductCost * float64(v.Quantity)
	}

	// 4. Определяем стоимость доставки
	deliveryCost, _, err := webSite.CalculateDelivery(models.DeliveryRequest{
		Cart: input.Cart,
		DeliveryData: models.DeliveryData {
			Id: input.Delivery.Id,
			Code: input.Delivery.Code,
			PostalCode: input.Delivery.PostalCode,
			Address: input.Delivery.Address,
		}})
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка расчета стоимости доставки", Errors: map[string]interface{}{"delivery":err.Error()}}))
		// u.Respond(w, u.MessageError(err, "Ошибка расчета стоимости доставки"))
		return
	}

	// 4.1. Добавляем в список заказа - доставку
	deliveryAmount := models.PaymentAmount{AccountId: account.Id, Value: deliveryCost, Currency: "RUB"}
	cartItems = append(cartItems, models.CartItem{
		AccountId: account.Id,
		Description: delivery.GetName(),
		Quantity: 1,
		Amount: deliveryAmount,
		VatCode: delivery.GetVatCode().YandexCode,
	})

	// 4.2 Добавляем стоимость доставки к общей стоимости
	totalCost += deliveryCost

	// 5. Находим канал заявки
	if input.OrderChannelCode == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска источника заявки", Errors: map[string]interface{}{"orderChannel":"Необходимо указать канал заявки"}}))
		return
	}
	orderChannel, err := account.GetOrderChannelByCode(input.OrderChannelCode)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска источника заявки", Errors: map[string]interface{}{"orderChannel":"канал не найден"}}))
		return
	}

	// 6. Создаем / находим пользователя
	var customer *models.User
	if input.CustomerHashId != "" {
		// Если ищем пользователя среди существующих
		user, err := account.GetUserByHashId(input.CustomerHashId)
		if err != nil || user == nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска пользователя", Errors: map[string]interface{}{"customerHashId":"Не удалось найти пользователя"}}))
			return
		}
		customer = user
	} else {
		// Если необходимо создать пользователя

		// 6.1 Проверяем, есть ли пользователь с такими контактными данными в существующем аккаунте
		userByEmail, errEmail := account.GetUserByEmail(input.Customer.Email)
		userByPhone, errPhone := account.GetUserByPhone(input.Customer.Phone, "RU")

		// todo: подсовывать ли данные существующим клиентам или нет ? - вывести в настройки
		if errPhone == nil {
			customer = userByPhone
		}
		if errEmail == nil {
			customer = userByEmail
		}
		if customer == nil {
			var _customer models.User
			_customer.Email = input.Customer.Email
			_customer.Phone = input.Customer.Phone

			_customer.Name = input.Customer.Name
			_customer.Surname = input.Customer.Surname
			_customer.Patronymic = input.Customer.Patronymic

			// 1.3 Роль - клиент
			role, err := account.GetRoleByTag(models.RoleClient)
			if err != nil {
				log.Fatalf("Не удалось найти аккаунт: %v", err)
			}

			// 2. Создаем пользователя
			user, err := account.CreateUser(_customer, *role)
			if err != nil {
				u.Respond(w, u.MessageError(u.Error{Message:"Ошибка создания пользователя"}))
				return
			}

			customer = user
		}

	}

	// 7. Создаем / находим менеджера
	var manager models.User
	if input.ManagerHashId != "" {
		// Если ищем пользователя среди существующих
		_manager, err := account.GetUserByHashId(input.ManagerHashId)
		if err != nil || _manager == nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска ответственного менеджера", Errors: map[string]interface{}{"managerHashId":"Не удалось найти менеджера"}}))
			return

		}
		manager = *_manager
	}

	// 8. Создаем / находим компанию
	if input.CompanyHashId != "" {
		// todo создание / поиск компании
		fmt.Println("Поиск компании..")
	}

	//////////////////////

	// 9. Создаем заказ
	var _order models.Order

	_order.CustomerComment = input.CustomerComment
	_order.ManagerId = manager.Id
	_order.Individual = input.Individual
	_order.WebSiteId = webSite.Id
	_order.CustomerId = customer.Id
	// order.CompanyId = CompanyId.Id
	_order.OrderChannelId = orderChannel.Id
	_order.Amount = models.PaymentAmount{Value: totalCost, Currency: totalCurrency, AccountId: account.Id}
	_order.CartItems = cartItems
	// _order.PaymentOptionId = paymentOption.Id
	_order.PaymentMethodId = paymentMethod.GetId()
	_order.PaymentMethodType = paymentMethod.GetType()

	// Создаем order
	orderEntity, err := account.CreateEntity(&_order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	order, ok := orderEntity.(*models.Order)
	if !ok {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время конвертации заказа"}))
		return
	}

	// Создаем платеж на основании заказа 
	payment, err := paymentMethod.CreatePaymentByOrder(*order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания платежа", Errors: map[string]interface{}{"payment":err.Error()}}))
		return
	}

	// Создаем доставку на основании заказа
	_, err = delivery.CreateDeliveryOrder(input.Delivery, deliveryAmount, *order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания доставки", Errors: map[string]interface{}{"delivery":err.Error()}}))
		return
	}

	resp := u.Message(true, "POST Order Created")
	resp["order"] = order
	resp["payment"] = payment
	u.Respond(w, resp)
}

func validateOrder(input CreateOrderForm) error {
	var e u.Error
	// Проверяем данные
	if input.Customer.Phone == "" {
		e.AddErrors("phone", "Укажите свой контактный телефон")
	} else {
		_, err := libphonenumber.Parse(input.Customer.Phone , "RU")
		if err != nil {
			e.AddErrors("phone", err.Error())
		}
	}

	if input.Customer.Email == "" {
		e.AddErrors("email", "Необходимо указать email")
	} else {
		if err := u.EmailDeepValidation(input.Customer.Email); err != nil {
			e.AddErrors("email", err.Error())
		}
	}

	if input.Customer.Name == "" {
		e.AddErrors("name", "Укажите имя получателя товара")
	}

	if e.HasErrors() {
		e.Message = "Проверьте правильность заполнения формы"
		return e
	}

	return nil
	
}