package middleware

import (
	"context"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

/*
* ### Auth by JWT ###

	Любой jwt имеет в своем составе информацию о пользователе и аккаунте, выдавшим ключ (issuer).
	userId - id пользователя, на имя которого выписан ключ
	accountId - id аккаунта, в котором пользователь авторизован
*/

// проверяет валидность 32-символьного api-ключа. Вставляет в контекст accountId && account
func BearerAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Получаем заголовок авторизации
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Header Authorization not found"))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			resp := u.Message(false, "Invalid/Malformed auth token: wrong format")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Собираем вторую часть строки "Bearer <...>"
		token := splitted[1] //Grab the token part, what we are truly interested in

		// ищем ApiKey с указанным токеном
		key, err := models.GetApiKeyByToken(token)
		if err != nil {

			resp := u.Message(false, "Invalid/Malformed auth token: key not found")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		if !key.Enabled {
			resp := u.Message(false, "Auth token is disabled")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// ищем аккаунт, связанный с apiKey
		account, err := models.GetAccount(key.AccountId)
		if err != nil || account == nil {
			resp := u.Message(false, "Account not found, plz refresh auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Проверяем статус включения API в аккаунте
		if !account.ApiEnabled {
			resp := u.Message(false, "Sorry, API is disabled for current account")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// имеет смысл, т.к. все запросы быдут связаны с аккаунтом
		r = r.WithContext(context.WithValue(r.Context(), "issuer", "api"))
		r = r.WithContext(context.WithValue(r.Context(), "account", account))
		r = r.WithContext(context.WithValue(r.Context(), "accountId", account.Id))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Требует авторизации по User, а также вставляет в контекст userId && user
func JwtCheckUserAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {


		// Проверяем контекст на выпуск аккаунта, чтобы из-под него проверить подпись JWT
		if r.Context().Value("issuerAccount") == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
			return
		}
		issuerAccount := r.Context().Value("issuerAccount").(*models.Account)
		if issuerAccount.Id < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.MessageError(u.Error{Message: "Error: account is not valid"}))
			return
		}

		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Отсутствует ключ авторизации"))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Некорректный ключ авторизации"))
			return
		}

		// Собираем вторую часть строки "Bearer kSDkfslfds390d2w...."
		tokenPart := splitted[1] //Grab the token part, what we are truly interested in

		tk, err := issuerAccount.ParseAndDecryptToken(tokenPart)
		if err != nil || tk == nil {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Не удалось прочитать ключ авторизации"))
			return
		}

		if tk.UserId < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не авторизован 1"))
			return
		}

		// Подгружаем аккаунт, к которому привязан пользователь
		if r.Context().Value("issuerAccount") == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка в обработке запроса", Errors: map[string]interface{}{"account": "not load"}}))
			return
		}

		account, err := utilsCr.GetAccountByHashId(w,r)
		if err != nil {
			return
		}
		// Загружаем пользователя в рамках аккаунта
		user, err := account.GetUser(tk.UserId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не найден"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userId", tk.UserId))
		r = r.WithContext(context.WithValue(r.Context(), "user", user))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Полная проверка User & Account авторизация + проверка доступа. Вставляет userId && user, accountId && account
func JwtCheckFullAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Проверяем контекст на выпуск аккаунта, чтобы из-под него проверить подпись JWT
		if r.Context().Value("issuerAccount") == nil {
			u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
			return
		}
		issuerAccount := r.Context().Value("issuerAccount").(*models.Account)
		if issuerAccount.Id < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.MessageError(u.Error{Message: "Error: account is not valid"}))
			return
		}

		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Отсутствует ключ авторизации"))
			return
		}

		//The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		splitted := strings.Split(tokenHeader, " ")
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Некорректный ключ авторизации"))
			return
		}

		// Собираем вторую часть строки "Bearer kSDkfslfds390d2w...."
		tokenPart := splitted[1] //Grab the token part, what we are truly interested in

		// AccountId? UserId? - собираем из токена
		//tk := &models.JWT{}

		// Парсим в tk токен со всеми данными
		tk, err := issuerAccount.ParseAndDecryptToken(tokenPart)
		if err != nil || tk == nil {
			w.WriteHeader(http.StatusUnauthorized)
			// fmt.Println(tokenPart)
			// fmt.Println(tk)
			// fmt.Println(err)
			u.Respond(w, u.Message(false, "Не удалось прочитать ключ авторизации"))
			return
		}

		if tk.UserId < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не авторизован 2"))
			return
		}

		if tk.AccountId < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Авторизуйтесь в аккаунте"))
			return
		}

		account, err := models.GetAccount(tk.AccountId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Аккаунт в котором вы авторизованы - не найден"))
			return
		}

		user, err := account.GetUser(tk.UserId)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не найден"))
			return
		}

		// 1. todo: Проверим, есть ли связь аккаунт и пользователя m <> m
		if false {
			u.Respond(w, u.MessageError(u.Error{Message: "Ошибка в обработке запроса", Errors: map[string]interface{}{"user": "not load"}}))
			return
		}

		// проверка на роль клиента и доступ в CRM
		if r.Context().Value("issuer") == "app" {
			// todo role & account.AllowClientLoginCrm
		}

		// fmt.Println("JwtFullAuthentication")
		// fmt.Printf("Context Account: %v\n",  account.Name)
		// fmt.Printf("userId: %v\n", user.Name)

		r = r.WithContext(context.WithValue(r.Context(), "issuer", "ui-api"))
		r = r.WithContext(context.WithValue(r.Context(), "userId", tk.UserId))
		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		r = r.WithContext(context.WithValue(r.Context(), "accountId", tk.AccountId))
		r = r.WithContext(context.WithValue(r.Context(), "account", account))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
