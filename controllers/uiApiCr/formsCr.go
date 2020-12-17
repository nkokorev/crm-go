package uiApiCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"net"
	"net/http"
)

func UiApiForm(w http.ResponseWriter, r *http.Request) {
	
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Проверяем запрос по IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке запроса"))
		return
	}

	// Проверяем на спам запросов
	if (models.Question{AccountId: account.Id}).IsSpam(ip) {
		u.Respond(w, u.MessageError(err, "Слишком много запросов за последние 24 часа"))
		return
	}

	// 1. Читаем вход
	var input = struct {

		// ID магазина, от имени которого создается заявка

		Channel *struct {
			WebSiteId 	*uint	`json:"web_site_id"`
			FormName 	*string	`json:"form_name"`
		} `json:"channel"`

		Customer *struct {
			// for auth user
			Id 		*uint 	`json:"id"`
			HashId 	*string	`json:"hash_id"`

			// For none
			Email	*string	`json:"email"`
			Phone	*string	`json:"phone"`

			Name 		*string `json:"name"`
			Surname 	*string `json:"surname"`
			Patronymic 	*string `json:"patronymic"`

			Subscribing	*bool `json:"subscribing"` // Если подписка
			SubscriptionReason 	*string `json:"subscription_reason"` // Если подписка
		} `json:"customer"`

		// for Subscribing == nil
		Question *struct {
			Message	*string	`json:"message"`

			ExpectAnAnswer *bool `json:"expect_an_answer"`
			ExpectACallback *bool `json:"expect_a_callback"`

			// Default: simple
			QuestionType	*models.QuestionType `json:"question_type"`

		} `json:"question"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте обязательные поля и типы переменных"))
		return
	}

	// Отсекаем не рабочие варианты
	if input.Channel == nil {
		u.Respond(w, u.MessageError(
			u.Error {Message: "Ошибка регистрации вопроса", Errors: map[string]interface{}{"channel":"Укажите канал сообщения"}}))
		return
	}
	if input.Customer == nil {
		u.Respond(w, u.MessageError(
			u.Error{Message: "Ошибка регистрации вопроса", Errors: map[string]interface{}{"customer":"Укажите данные отправителя"}}))
		return
	}
	if input.Question == nil && input.Customer.Subscribing == nil{
		u.Respond(w, u.MessageError(
			u.Error{Message: "Ошибка регистрации вопроса", Errors: map[string]interface{}{"question":"Укажите данные вопроса"}}))
		return
	}

	// 2. Получаем магазин из контекста
	var webSite models.WebSite
	if input.Channel.WebSiteId != nil {
		if err := account.LoadEntity(&webSite, *input.Channel.WebSiteId,nil); err != nil {

			u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте id магазина"))
			return
		}
	}

	// 3 Get User New or No
	var user models.User
	if input.Customer.HashId != nil {

		if len([]rune(*input.Customer.HashId)) > 32 {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка заполнения формы", Errors: map[string]interface{}{"customer.hash_id": "Необходимо указать валидный hash id"}}))
			return
		}
		// to get Customer by Hash Id
		_user, err := account.GetUserByHashId(*input.Customer.HashId)
		if err != nil || _user == nil{
			u.Respond(w, u.MessageError(u.Error{Message: "Не удалось найти пользователя", Errors: map[string]interface{}{"customer.hash_id": "Пользователь не найден"}}))
			return
		}

		user = *_user

	} else {

		// Create new User
		var e u.Error

		if input.Customer.Email == nil && input.Customer.Phone == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка заполнения формы",
				Errors: map[string]interface{}{"email": "Одно из полей должно быть заполнено","phone":"Одно из полей должно быть заполнено"}}))
			return
		}

		if input.Customer.Email == nil && input.Customer.Subscribing != nil && *input.Customer.Subscribing{
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка заполнения формы",
				Errors: map[string]interface{}{"email": "Укажите свой email"}}))
			return
		}

		if input.Customer.Email != nil {
			if  err := u.EmailValidation(*input.Customer.Email); err != nil {
				e.AddErrors("email", err.Error())
			}
		}
		if input.Customer.Phone != nil {
			if u.VerifyPhone(*input.Customer.Phone, "ru") != nil {
				e.AddErrors("phone", "Указан не верный формат номера")
			}

			if len([]rune(*input.Customer.Phone)) > 32 {
				e.AddErrors("phone", "Слишком длинный номер телефона, максимум 32")
			}
		}

		if input.Customer.Name != nil && len([]rune(*input.Customer.Name)) > 64 {
			e.AddErrors("name", "Слишком длинное имя" )
		}
		if input.Customer.Surname != nil && len([]rune(*input.Customer.Surname)) > 64 {
			e.AddErrors("surname", "Фамилия слишком длинная")
		}
		if input.Customer.Patronymic != nil && len([]rune(*input.Customer.Patronymic)) > 64 {
			e.AddErrors("patronymic", "Отчество слишком длинное" )
		}
		if input.Customer.SubscriptionReason == nil {
			input.Customer.SubscriptionReason = u.STRp("ui.api from web")
		}
		if input.Customer.SubscriptionReason != nil && len([]rune(*input.Customer.SubscriptionReason)) > 32 {
			e.AddErrors("subscription_reason", "Слишком длинное поле, max 32 символа")
		}

		if err := e.GetErrors();err != nil {
			e.Message = "Проверьте правильность заполнения формы"
			u.Respond(w, u.MessageError(e))
			return
		}

		// 3. Если есть такой клиент по email - проверяем его статус подписки
		if account.ExistUserByEmail(input.Customer.Email) {

			// Находим пользователя
			user, err := account.GetUserByEmail(*input.Customer.Email)
			if err != nil || user == nil {
				e.Message = "Ошибка заполнения формы"
				u.Respond(w, u.MessageError(e))
				return
			}

			// Проверяем статус отподписки
			if !user.Subscribed {
				u.Respond(w, u.MessageError(u.Error{Message: "Ошибка заполнения формы", Errors: map[string]interface{}{"email": "Этот email уже отписан от всех рассылок"}}))
				return
			}

			// Уже подписан => все ок!
			u.Respond(w, u.Message(true, "Запрос успешно отправлен!"))
			return

		}

		// 4. Если есть такой клиент по phone - проверяем его статус подписки
		if account.ExistUserByPhone(input.Customer.Phone) {
			user, err := account.GetUserByPhone(*input.Customer.Phone)
			if (err != nil || user == nil) && err != gorm.ErrRecordNotFound {
				e.Message = "Ошибка заполнения формы"
				u.Respond(w, u.MessageError(e))
				return
			}

			// Проверяем статус подписки
			if !user.Subscribed {
				u.Respond(w, u.MessageError(u.Error{Message: "Ошибка заполнения формы", Errors: map[string]interface{}{"phone": "Этот адрес уже уже отписан от всех рассылок"}}))
				return
			}
		}

		// 5. Создаем нового пользователя
		// var user models.User

		user.Email = input.Customer.Email
		user.Phone = input.Customer.Phone

		user.Name 		= input.Customer.Name
		user.Surname 	= input.Customer.Surname
		user.Patronymic = input.Customer.Patronymic

		if input.Customer.Subscribing != nil && *input.Customer.Subscribing {
			user.SubscriptionReason = input.Customer.SubscriptionReason
			user.Subscribed = true
		} else {
			user.Subscribed = false
		}

		// Получаем роль клиента
		role, err := account.GetRoleByTag(models.RoleClient)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления вопроса"}))
			return
		}

		// Создаем пользователя
		newUser, err := account.CreateUser(user, *role)
		if err != nil || newUser == nil{
			u.Respond(w, u.MessageError(err))
			return
		}
		user = *newUser
	}

	// Если это запрос от пользователя / подписка
	if input.Question == nil {
		u.Respond(w, u.Message(true, "Ваш запрос успешно обработан!"))
		return
	}

	// ================================================ //
	
	// 4. Проверяем сам вопрос по логике нахождения элементов
	if input.Question.ExpectACallback == nil && input.Question.ExpectAnAnswer == nil {
		u.Respond(w, u.MessageError(
			u.Error{Message: "Необходимо указать запрос вопроса или ответного звонка",
				Errors: map[string]interface{}{"question.expect_an_answer": "Необходимо указать запрос вопроса или ответного звонка",
					"question.expect_a_callback": "Необходимо указать запрос вопроса или ответного звонка"}}))
		return
	}
	
	// 4.1 - CallBack && !Phone  = false => реджект
	if input.Question.ExpectACallback != nil && *input.Question.ExpectACallback && user.Phone == nil{
		u.Respond(w, u.MessageError(
			u.Error{Message: "Ошибка в обработке запроса",
				Errors: map[string]interface{}{"customer.phone": "Укажите номер телефона"}}))
		return
	}

	// 4.2 CallBack && !Phone => таки куда звонить?
	if input.Question.ExpectAnAnswer != nil && *input.Question.ExpectAnAnswer && user.Email == nil {
		u.Respond(w, u.MessageError(
			u.Error{Message: "Ошибка в обработке запроса",
				Errors: map[string]interface{}{"customer.email": "Укажите свой email"}}))
		return
	}

	// 4.2 CallBack && !Phone => таки куда звонить?
	if input.Question.Message == nil && input.Question.ExpectAnAnswer != nil && *input.Question.ExpectAnAnswer {
		u.Respond(w, u.MessageError(
			u.Error{Message: "Ошибка в обработке запроса",
				Errors: map[string]interface{}{"question.message": "Укажите свой вопрос"}}))
		return
	}
	
	// fix QuestionType
	if input.Question.QuestionType == nil {
		input.Question.QuestionType = u.STRp("simple")
	}

	// ############################# ## ## //

	var e u.Error
	if input.Question.Message != nil && len([]rune(*input.Question.Message)) > 255 {
		e.AddErrors("question.message", "Длина сообщения слишком большая. Максимум 255 символов")
	}
	if input.Question.QuestionType != nil &&  len([]rune(*input.Question.QuestionType)) > 64 {
		e.AddErrors("question.question_type", "Слишком большая длина указанного типа. Максимум 64 символа")
	}
	if input.Channel.FormName != nil && len([]rune(*input.Channel.FormName)) > 64 {
		e.AddErrors("channel.form_name", "Слишком длинное имя формы. Максимум 64 символа")
	}

	// 6.



	// 5. Создаем вопрос в зависимости от его типа
	var question models.Question

	question.AccountId 		= 	account.Id
	question.QuestionType 	= 	*input.Question.QuestionType
	question.UserId 		= 	user.Id
	if input.Question.Message != nil {
		question.Message = input.Question.Message
	}
	if input.Question.QuestionType != nil {
		question.QuestionType = *input.Question.QuestionType
	}
	if input.Question.ExpectAnAnswer != nil {
		question.ExpectAnAnswer	= *input.Question.ExpectAnAnswer
	}
	if input.Question.ExpectACallback != nil {
		question.ExpectACallback = *input.Question.ExpectACallback
	}
	if input.Channel.FormName != nil {
		question.FormName = input.Channel.FormName
	}
	question.Ipv4 = &ip

	_, err = account.CreateEntity(&question)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Не удалось создать запрос"}))
		// u.Respond(w, u.MessageError(err))
		return
	}


	// Уже подписан => все ок!
	u.Respond(w, u.Message(true, "Ваш вопрос успешно отправлен!"))
	return
}

func UiApiSubscribe(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// 1. Читаем вход
	var input = struct {
		// ID магазина, от имени которого создается заявка
		Channel struct {
			WebSiteId 	*uint	`json:"web_site_id"`
		}
		Customer struct {
			Email	*string	`json:"email"`
			Phone	*string	`json:"phone"`

			Name 		*string `json:"name"`
			Surname 	*string `json:"surname"`
			Patronymic 	*string `json:"patronymic"`

			SubscriptionReason 	*string `json:"subscription_reason"`
		} `json:"customer"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте обязательные поля и типы переменных"))
		return
	}

	// 2. Получаем магазин из контекста
	var webSite models.WebSite
	if input.Channel.WebSiteId != nil {
		if err := account.LoadEntity(&webSite, *input.Channel.WebSiteId,nil); err != nil {

			u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте id магазина"))
			return
		}
	}

	// 3. Проверяем наличие полей
	var e u.Error
	if input.Customer.Email == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка оформления подписки", Errors: map[string]interface{}{"email": "Необходимо указать email"}}))
		return
	}

	// Проверяем содержание полей
	if  err := u.EmailValidation(*input.Customer.Email); err != nil {
		e.AddErrors("email", err.Error())
	}
	if input.Customer.Phone != nil {
		if u.VerifyPhone(*input.Customer.Phone, "ru") != nil {
			e.AddErrors("phone", "Указан не верный формат номера")
		}

		if len([]rune(*input.Customer.Phone)) > 32 {
			e.AddErrors("phone", "Слишком длинный номер телефона, максимум 32")
		}
	}

	if input.Customer.Name != nil && len([]rune(*input.Customer.Name)) > 64 {
		e.AddErrors("name", "Слишком длинное имя" )
	}
	if input.Customer.Surname != nil && len([]rune(*input.Customer.Surname)) > 64 {
		e.AddErrors("surname", "Фамилия слишком длинная")
	}
	if input.Customer.Patronymic != nil && len([]rune(*input.Customer.Patronymic)) > 64 {
		e.AddErrors("patronymic", "Отчество слишком длинное" )
	}
	if input.Customer.SubscriptionReason == nil {
		input.Customer.SubscriptionReason = u.STRp("ui.api from web")
	}
	if input.Customer.SubscriptionReason != nil && len([]rune(*input.Customer.SubscriptionReason)) > 32 {
		e.AddErrors("subscription_reason", "Причина подписки слишком длинная, max 32 символа")
	}

	// 4. Проверяем ошибки - если есть, то возвращаем
	if err := e.GetErrors();err != nil {
		e.Message = "Проверьте правильность заполнения формы"
		u.Respond(w, u.MessageError(e))
		return
	}

	// 3. Если есть такой клиент по email - проверяем его статус подписки
	if account.ExistUserByEmail(input.Customer.Email) {

		// Находим пользователя
		user, err := account.GetUserByEmail(*input.Customer.Email)
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
	if account.ExistUserByPhone(input.Customer.Phone) {
		user, err := account.GetUserByPhone(*input.Customer.Phone)
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

	user.Email = input.Customer.Email
	user.Phone = input.Customer.Phone

	user.Name 		= input.Customer.Name
	user.Surname 	= input.Customer.Surname
	user.Patronymic = input.Customer.Patronymic
	user.SubscriptionReason = input.Customer.SubscriptionReason

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