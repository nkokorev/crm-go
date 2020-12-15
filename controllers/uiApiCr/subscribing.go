package uiApiCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"net/http"
)

func UiApiSubscribe(w http.ResponseWriter, r *http.Request) {

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

	// 1. Читаем вход
	var input = struct {
		// ID магазина, от имени которого создается заявка
		WebSiteId 	*uint	`json:"web_site_id"`
		Subscriber struct {
			Email	*string	`json:"email"`
			Phone	*string	`json:"phone"`

			Name 		*string `json:"name"`
			Surname 	*string `json:"surname"`
			Patronymic 	*string `json:"patronymic"`

			SubscriptionReason 	*string `json:"subscription_reason"`
		} `json:"subscriber"`
	}{}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте обязательные поля и типы переменных"))
		return
	}
	
	// 2. Получаем магазин из контекста
	var webSite models.WebSite
	if input.WebSiteId != nil {
		if err := account.LoadEntity(&webSite, *input.WebSiteId,[]string{"EmailBoxes"}); err != nil {

			u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте id магазина"))
			return
		}
	}

	// 3. Проверяем наличие полей
	var e u.Error 
	if input.Subscriber.Email == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления подписки", Errors: map[string]interface{}{"email": "Необходимо указать email"}}))
		return
	}

	// Проверяем содержание полей
	if  err := u.EmailValidation(*input.Subscriber.Email); err != nil {
		e.AddErrors("email", err.Error())
	}
	if input.Subscriber.Phone != nil {
		if u.VerifyPhone(*input.Subscriber.Phone, "ru") != nil {
			e.AddErrors("phone", "Указан не верный формат номера")
		}

		if len([]rune(*input.Subscriber.Phone)) > 32 {
			e.AddErrors("phone", "Слишком длинный номер телефона, максимум 32")
		}
	}
	
	if input.Subscriber.Name != nil && len([]rune(*input.Subscriber.Name)) > 64 {
		e.AddErrors("name", "Слишком длинное имя" )
	}
	if input.Subscriber.Surname != nil && len([]rune(*input.Subscriber.Surname)) > 64 {
		e.AddErrors("surname", "Фамилия слишком длинная")
	}
	if input.Subscriber.Patronymic != nil && len([]rune(*input.Subscriber.Patronymic)) > 64 {
		e.AddErrors("patronymic", "Отчество слишком длинное" )
	}
	if input.Subscriber.SubscriptionReason == nil {
		input.Subscriber.SubscriptionReason = u.STRp("ui.api from web")
	}
	if input.Subscriber.SubscriptionReason != nil && len([]rune(*input.Subscriber.SubscriptionReason)) > 32 {
		e.AddErrors("subscription_reason", "Причина подписки слишком длинная, max 32 символа")
	}

	// 4. Проверяем ошибки - если есть, то возвращаем
	if err := e.GetErrors();err != nil {
		e.Message = "Проверьте правильность заполнения формы"
		u.Respond(w, u.MessageError(e))
		return
	}
	
	// 3. Если есть такой клиент по email - проверяем его статус подписки
	if account.ExistUserByEmail(input.Subscriber.Email) {

		// Находим пользователя
		user, err := account.GetUserByEmail(*input.Subscriber.Email)
		if err != nil || user == nil {
			e.Message = "Ошибка оформления подписки"
			u.Respond(w, u.MessageError(e))
			return
		}
		
		// Проверяем статус отподписки
		if !user.Subscribed {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления подписки", Errors: map[string]interface{}{"email": "Этот email уже отписан от всех рассылок"}}))
			return
		}

		// Уже подписан => все ок!
		u.Respond(w, u.Message(true, "Пользователь успешно подписан!"))
		return

	}

	// 4. Если есть такой клиент по phone - проверяем его статус подписки
	if account.ExistUserByPhone(input.Subscriber.Phone) {
		user, err := account.GetUserByPhone(*input.Subscriber.Phone)
		if (err != nil || user == nil) && err != gorm.ErrRecordNotFound {
			e.Message = "Ошибка оформления подписки"
			u.Respond(w, u.MessageError(e))
			return
		}

		// Проверяем статус подписки
		if !user.Subscribed {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления подписки", Errors: map[string]interface{}{"phone": "Этот адрес уже уже отписан от всех рассылок"}}))
			return
		}
	}

	// 5. Создаем нового пользователя
	var user models.User
	
	user.Email = input.Subscriber.Email
	user.Phone = input.Subscriber.Phone

	user.Name 		= input.Subscriber.Name
	user.Surname 	= input.Subscriber.Surname
	user.Patronymic = input.Subscriber.Patronymic
	user.SubscriptionReason = input.Subscriber.SubscriptionReason

	// Получаем роль клиента
	role, err := account.GetRoleByTag(models.RoleClient)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления подписки"}))
		return
	}

	// Создаем пользователя
	_, err = account.CreateUser(user, *role)
	if err != nil {
		u.Respond(w, u.MessageError(err))
		return
	}
	
	// Уже подписан => все ок!
	u.Respond(w, u.Message(true, "Пользователь успешно подписан!"))
	return
}