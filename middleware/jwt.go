package middleware

import (
	"context"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strings"
)

// Требует авторизации по User
func JwtUserAuthentication(next http.Handler) http.Handler {

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

		if tk.UserId < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Пользователь не авторизован"))
			return
		}

		ctx1 := context.WithValue(r.Context(), "user_id", tk.UserId)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "account_id", tk.AccountId)
		r = r.WithContext(ctx2)

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	});
}

// Требует авторизации по Аккаунту
func JwtAccountAuthentication(next http.Handler) http.Handler {

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

		if tk.AccountId < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Авторизуйтесь в аккаунте"))
			return
		}

		ctx1 := context.WithValue(r.Context(), "user_id", tk.UserId)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "account_id", tk.AccountId)
		r = r.WithContext(ctx2)

		next.ServeHTTP(w, r) //proceed in the middleware chain!
	});
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

		ctx1 := context.WithValue(r.Context(), "user_id", tk.UserId)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "account_id", tk.AccountId)
		r = r.WithContext(ctx2)

		if tk.UserId < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Пользователь не авторизован"))
			return
		}

		if tk.AccountId < 1 {
			w.WriteHeader(http.StatusForbidden)
			u.Respond(w, u.Message(false, "Авторизуйтесь в аккаунте"))
			return
		}

		// Проверяем доступ к аккаунту


		next.ServeHTTP(w, r) //proceed in the middleware chain!
	});
}
