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

	PaymentMethod struct{
		Id uint `json:"id"`
		Code string `json:"code"`
		Type string `json:"type"`

	}

}

// todo: список обязательных полей - дело настроек OrderSettings
type Customer struct {
	Email 		string `json:"email" ` // email должен проходить deepValidate()
	Phone		string `json:"phone" ` // форматируем сами телефон

	Name 		string `json:"name"`
	Surname 	string `json:"surname" `
	Patronymic 	string `json:"patronymic"`
}

// Отложенная выдача товара!
func UiApiOrderCreate(w http.ResponseWriter, r *http.Request) {

	/*u.Respond(w, u.MessageError(u.Error{Message:"Указанный тип оплаты не поддерживает данный тип доставки",
		Errors: map[string]interface{}{
		"phone":"Указанный номер телефона занят",
		"email":"Обязательно нужно указать",
		"surname":"А где фамилия?",
		"name":"Укажите имя",
		}}))
	return*/

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

	// 1. Получаем магазин из контекста
	var webSite models.WebSite
	if err := account.LoadEntity(&webSite, input.WebSiteId); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте id магазина"))
		return
	}
	
	// 0. Находим канал заявки
	if input.OrderChannelCode == "" {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска источника заявки", Errors: map[string]interface{}{"orderChannel":"Необходимо указать канал заявки"}}))
		return
	}
	channel, err := account.GetOrderChannelByCode(input.OrderChannelCode)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска источника заявки", Errors: map[string]interface{}{"orderChannel":"канал не найден"}}))
		return
	}

	switch channel.Code {
	case "through_the_basket":
		createOrderFromBasket(w, input, *account, webSite, *channel)
		return
	case "callback_phone":
		createOrderFromCallbackPhone(w, input, *account, webSite, *channel)
		return
	case "callback_form":
		createOrderFromCallbackForm(w, input, *account, webSite, *channel)
		return
	default:
		u.Respond(w, u.MessageError(u.Error{Message:"Указанный канал не поддерживается",
			Errors: map[string]interface{}{
				"orderChannel":"Не поддерживаемый канал заявок",
			}}))
		return
	}
}

func validateCustomerOrder(input CreateOrderForm) error {
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

func getCustomerFromInput(account models.Account, userHashId, email, phone, name, surname, patronymic string) (*models.User, error) {

	var user models.User

	if userHashId != "" {
		// Если ищем пользователя среди существующих
		_user, err := account.GetUserByHashId(userHashId)
		if err != nil || _user == nil {
			return nil, u.Error{Message:"Ошибка поиска пользователя", Errors: map[string]interface{}{"hashId":"Не удалось найти пользователя"}}
		}
		user = *_user
	} else {
		// Если необходимо создать пользователя

		// 6.1 Проверяем, есть ли пользователь с такими контактными данными в существующем аккаунте
		userByEmail, errEmail := account.GetUserByEmail(email)
		userByPhone, errPhone := account.GetUserByPhone(phone, "RU")


		// ! Оба пользователя найдены = В случае, если у обоих пользователей указаны телефон и емейл, но при этом емейл разные
		if errEmail == nil && errPhone == nil && userByEmail.Email != "" && userByPhone.Email != "" && (userByEmail.Email != userByPhone.Email) {
			return nil, u.Error{Message:"Ошибка идентификации пользователя",
				Errors: map[string]interface{}{
				"email":"Пользователь с таким email'ом имеет другой телефон",
				"phone":"Пользователь с таким телефоном имеет другой email",
				}}
		}

		// ! Оба пользователя найдены = В случае, если у обоих пользователей указаны телефон и емейл, но при этом телефоны разные
		if errEmail == nil && errPhone == nil && userByEmail.Phone != "" && userByPhone.Phone != "" && (userByEmail.Phone != userByPhone.Phone) {
			return nil, u.Error{Message:"Ошибка идентификации пользователя",
				Errors: map[string]interface{}{
					"email":"Пользователь с таким email'ом имеет другой телефон",
					"phone":"Пользователь с таким телефоном имеет другой email",
				}}
		}

		// Найден только один пользователь = и у него указан телефон и в заявке у него другой телефон
		if errEmail == nil && errPhone != nil &&
			userByEmail.Phone != "" && phone != "" &&
			(userByEmail.Phone != phone) {
			return nil, u.Error{Message:"Ошибка идентификации пользователя",
				Errors: map[string]interface{}{
					"phone":"Пользователь с указанным email имеет другой номер телефона",
				}}
		}
		// Найден только один пользователь = и у него указан телефон и в заявке у него другой телефон
		if errPhone == nil && errEmail != nil &&
			userByPhone.Email != "" && email != "" &&
			(userByPhone.Email != email) {
			return nil, u.Error{Message:"Ошибка идентификации пользователя",
				Errors: map[string]interface{}{
					"email":"Пользователь с указанным телефоном имеет другой email",
				}}
		}

		if errPhone == nil {
			user = *userByPhone
		}
		if errEmail == nil {
			user = *userByEmail
		}
		if user.Id < 1 {
			var _user models.User
			_user.Email = email
			_user.Phone = phone

			_user.Name = name
			_user.Surname = surname
			_user.Patronymic = patronymic

			// 1.3 Роль - клиент
			role, err := account.GetRoleByTag(models.RoleClient)
			if err != nil {
				return nil, u.Error{Message:"Ошибка создания пользователя"}
			}

			// 2. Создаем пользователя
			__user, err := account.CreateUser(_user, *role)
			if err != nil {
				fmt.Println(err)
				return nil, u.Error{Message:"Ошибка создания пользователя"}
			}

			user = *__user
		}
	}

	return &user, nil
}

func createOrderFromBasket(w http.ResponseWriter, input CreateOrderForm, account models.Account, webSite models.WebSite, channel models.OrderChannel) {

	// Итоговая стоимость заказа
	var totalCost float64
	totalCurrency := "RUB"
	totalCost = 0

	
	if err := validateCustomerOrder(input); err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}

	// 6. Находим способ оплаты
	if input.PaymentMethod.Code == "" || input.PaymentMethod.Id < 1 {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска способа оплаты", Errors: map[string]interface{}{"paymentMethodCode":"Необходимо указать способ оплаты"}}))
		return
	}

	paymentMethod, err := account.GetPaymentMethodByCode(input.PaymentMethod.Code, input.PaymentMethod.Id)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска способа оплаты", Errors: map[string]interface{}{"paymentMethod":"Способ оплаты не найден"}}))
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

		if product.PaymentSubject.Id < 1 {
			log.Printf("product.PaymentSubject.Id < 1")
			u.Respond(w, u.MessageError(u.Error{Message:fmt.Sprintf("Ошибка во время создания заказа")}))
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
			PaymentSubjectId: product.PaymentSubjectId, // признак предмета расчета
			PaymentSubjectYandex: product.PaymentSubject.Code,
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
		PaymentSubjectId: delivery.GetPaymentSubject().Id, // признак предмета расчета
		PaymentSubjectYandex: delivery.GetPaymentSubject().Code,
	})

	// 4.2 Добавляем стоимость доставки к общей стоимости
	totalCost += deliveryCost

	// 6. Создаем / находим пользователя
	// 6. Создаем / находим пользователя
	customer, err := getCustomerFromInput(account, input.CustomerHashId, input.Customer.Email, input.Customer.Phone, input.Customer.Name, input.Customer.Surname, input.Customer.Patronymic)
	if err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}

	// 7. Находим менеджера
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
	_order.OrderChannelId = channel.Id
	_order.Amount = models.PaymentAmount{Value: totalCost, Currency: totalCurrency, AccountId: account.Id}
	_order.CartItems = cartItems
	_order.PaymentMethodId = paymentMethod.GetId()
	_order.PaymentMethodType = paymentMethod.GetType()

	// Создаем order
	orderEntity, err := account.CreateEntity(&_order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}
	
	var order models.Order
	if err := account.LoadEntity(&order, orderEntity.GetId()); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	// Моментальная доставкатовара или нет
	var mode models.PaymentMode
	if paymentMethod.IsInstantDelivery() {
		mode, err = models.PaymentMode{}.GetFullPaymentMode()
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
			return
		}
	}   else {
		mode, err = models.PaymentMode{}.GetFullPrepaymentMode()
		if err != nil {
			log.Printf("Не удалось получить полную предоплату: %v", err)
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
			return
		}
	}

	// Создаем платеж на основании заказа
	payment, err := paymentMethod.CreatePaymentByOrder(order, mode)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания платежа", Errors: map[string]interface{}{"payment":err.Error()}}))
		return
	}

	// Создаем заказ на доставку на основании заказа. Даже если это моментальная выдача товара (должен быть соответствующий способ).
	_, err = delivery.CreateDeliveryOrder(input.Delivery, deliveryAmount, order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания доставки", Errors: map[string]interface{}{"delivery":err.Error()}}))
		return
	}

	resp := u.Message(true, "POST Order Created")
	resp["order"] = order
	resp["payment"] = payment
	u.Respond(w, resp)
}

func createOrderFromCallbackPhone(w http.ResponseWriter, input CreateOrderForm, account models.Account, webSite models.WebSite, channel models.OrderChannel) {

	if input.Customer.Phone == "" {
		u.Respond(w, u.MessageError(
			u.Error{Message:"Ошибка создания заявки",
				Errors: map[string]interface{}{"phone":"Необходимо указать номер телефона"}}))
		return
	} else {
		_, err := libphonenumber.Parse(input.Customer.Phone , "RU")
		if err != nil {
			u.Respond(w, u.MessageError(
				u.Error{Message:"Ошибка создания заявки",
					Errors: map[string]interface{}{"phone":"Формат телефона указан не верно"}}))
			return
		}
	}

	// 6. Создаем / находим пользователя
	customer, err := getCustomerFromInput(account, input.CustomerHashId, input.Customer.Email, input.Customer.Phone, input.Customer.Name, input.Customer.Surname, input.Customer.Patronymic)
	if err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}

	// 7. Находим менеджера
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
	// _order.CompanyId = CompanyId.Id
	_order.OrderChannelId = channel.Id

	// Создаем order
	orderEntity, err := account.CreateEntity(&_order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	var order models.Order
	if err := account.LoadEntity(&order, orderEntity.GetId()); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	resp := u.Message(true, "POST Order Created")
	resp["order"] = order
	u.Respond(w, resp)
}

func createOrderFromCallbackForm(w http.ResponseWriter, input CreateOrderForm, account models.Account, webSite models.WebSite, channel models.OrderChannel) {

	if input.CustomerComment == "" {
		u.Respond(w, u.MessageError(
			u.Error{Message:"Ошибка создания заявки",
				Errors: map[string]interface{}{"customerComment":"Укажите свой вопрос"}}))
		return
	}
	if input.Customer.Phone == "" && input.Customer.Email == "" {
		u.Respond(w, u.MessageError(
			u.Error{Message:"Ошибка создания заявки",
				Errors: map[string]interface{}{"phone":"Необходимо указать номер телефона или email"}}))
		return
	}
	if input.Customer.Phone != "" {
		_, err := libphonenumber.Parse(input.Customer.Phone , "RU")
		if err != nil {
			u.Respond(w, u.MessageError(
				u.Error{Message:"Ошибка создания заявки",
					Errors: map[string]interface{}{"phone":"Формат телефона указан не верно"}}))
			return
		}
	}
	if input.Customer.Email != "" {
		if err := u.EmailValidation(input.Customer.Email); err != nil {
			u.Respond(w, u.MessageError(
				u.Error{Message:"Ошибка создания заявки",
					Errors: map[string]interface{}{"email":"Проверьте правильность написания email'а"}}))
			return
		}
	}

	// 6. Создаем / находим пользователя
	customer, err := getCustomerFromInput(account,
		input.CustomerHashId, input.Customer.Email, input.Customer.Phone,
		input.Customer.Name, input.Customer.Surname, input.Customer.Patronymic)

	if err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}

	// 7. Находим менеджера
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
	// _order.CompanyId = CompanyId.Id
	_order.OrderChannelId = channel.Id

	// Создаем order
	orderEntity, err := account.CreateEntity(&_order)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	var order models.Order
	if err := account.LoadEntity(&order, orderEntity.GetId()); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания заказа"}))
		return
	}

	resp := u.Message(true, "POST Order Created")
	resp["order"] = order
	u.Respond(w, resp)
}

