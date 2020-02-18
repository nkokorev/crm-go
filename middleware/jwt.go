package middleware

import (
	"context"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

/*
* ### Auth by JWT ###

	Любой jwt-токен имеет в своем составе информацию о пользователе и аккаунте, выдавшим ключ (issuer).
	userId - id пользователя, на имя которого выписан ключ
	accountId - id аккаунта, в котором пользователь авторизован


 */


// проверяет валидность 32-символного api-ключа
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
			resp := u.Message(false, "Invalid/Malformed auth token")
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, resp)
			return
		}

		// Собираем вторую часть строки "Bearer <...>"
		token := splitted[1] //Grab the token part, what we are truly interested in

		// ищем ApiKey с указанным токеном
		key, err := models.GetApiKey(token)
		if err != nil {
			resp := u.Message(false, "Invalid/Malformed auth token")
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
		account, err := models.GetAccount(key.AccountID)
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
		r = r.WithContext(context.WithValue(r.Context(), "accountId", account.ID))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Требует авторизации по User
func JwtUserAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
		tk := &models.JWT{}
		//tk, err := models.ParseAndDecryptToken(tokenPart)
		if err := tk.ParseAndDecryptToken(tokenPart);err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Неудалось прочитать ключ авторизации"))
			return
		}

		if tk.UserID < 1 {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не авторизован"))
			return
		}

		// Подружаем аккаунт, к которому привязан пользователь
		if r.Context().Value("issuerAccount") == nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"account":"not load"}}))
			return
		}
		account := r.Context().Value("issuerAccount").(models.Account)

		// Загружаем пользователя в рамках аккаунта
		user, err := account.GetUserById(tk.UserID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не найден"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userId", tk.UserID))
		r = r.WithContext(context.WithValue(r.Context(), "user", user))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}

// Требует User и Account авторизации, Проверяет доступ пользователя к аккаунту.
func JwtFullAuthentication(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := r.Header.Get("Authorization") //Grab the token from the header

		if tokenHeader == "" { //Token is missing, returns with error code 403 Unauthorized
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Отсутствует ключ авторизации"))
			return
		}

		splitted := strings.Split(tokenHeader, " ") //The token normally comes in format `Bearer {token-body}`, we check if the retrieved token matched this requirement
		if len(splitted) != 2 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Некорректный ключ авторизации"))
			return
		}

		// Собираем вторую часть строки "Bearer kSDkfslfds390d2w...."
		tokenPart := splitted[1] //Grab the token part, what we are truly interested in
		tk := &models.JWT{}
		//tk, err := models.ParseAndDecryptToken(tokenPart)
		if err := tk.ParseAndDecryptToken(tokenPart);err != nil {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Неудалось прочитать ключ авторизации"))
			return
		}

		if tk.UserID < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Пользователь не авторизован"))
			return
		}

		if tk.AccountID < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Авторизуйтесь в аккаунте"))
			return
		}

		account, err := models.GetAccount(tk.AccountID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Аккаунт в котором вы авторизованы - не найден"))
			return
		}

		user, err := account.GetUserById(tk.UserID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message(false, "Пользователь не найден"))
			return
		}

		// 1. todo: Проверим, есть ли связь аккаунт и пользователя m <> m
		if false {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в обработке запроса", Errors: map[string]interface{}{"user":"not load"}}))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userId", tk.UserID))
		r = r.WithContext(context.WithValue(r.Context(), "user", user))
		r = r.WithContext(context.WithValue(r.Context(), "accountId", tk.AccountID))
		r = r.WithContext(context.WithValue(r.Context(), "account", account))

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
