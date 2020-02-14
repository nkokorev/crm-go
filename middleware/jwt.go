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

	Любой jwt-токен имеет в своем составе информацию о пользователе и аккаунте, выдавшим ключ (signedAccountId).
	userId - id пользователя, на имя которого выписан ключ
	accountId - id аккаунта, в котором пользователь авторизован
	signedAccountId - ВСЕГДА id аккаунта, выдавшего ключ (root account у каждого пользователя).

 */

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

		ctx1 := context.WithValue(r.Context(), "userId", tk.UserID)
		r = r.WithContext(ctx1)

		/*ctx3 := context.WithValue(r.Context(), "signedAccountId", tk.SignedAccountID)
		r = r.WithContext(ctx3)*/

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

		ctx1 := context.WithValue(r.Context(), "userId", tk.UserID)
		r = r.WithContext(ctx1)

		ctx2 := context.WithValue(r.Context(), "accountId", tk.AccountID)
		r = r.WithContext(ctx2)

		ctx3 := context.WithValue(r.Context(), "account", account)
		r = r.WithContext(ctx3)



		// Проверяем доступ к аккаунту


		next.ServeHTTP(w, r) //proceed in the middleware chain!
	})
}
