package middleware

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strconv"
)

//Тут собраны посредники определяющие issuerAccountId и добавляющие его в контекст

// Вставляет в контекст issuerAccountId = 1
func ContextMainAccount(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// тут могла быть специальная функция, берущая данные из памяти (чекая по времени)
		issuerAccount, err := models.GetMainAccount() // RatusCRM
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}

		// For future
		r = r.WithContext(context.WithValue(r.Context(), "issuer", "app ui/api"))
		r = r.WithContext(context.WithValue(r.Context(), "issuerAccountId", issuerAccount.ID))
		r = r.WithContext(context.WithValue(r.Context(), "issuerAccount", issuerAccount))

		next.ServeHTTP(w, r)
	})

}

// Вставляет в контекст issuerAccountId из hashId (раскрытие issuer account)
func ContextMuxVarAccountHashId(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountHashId := mux.Vars(r)["accountHashId"] // защищаемся от парсинга / спама

		if len(accountHashId) != 12 {
			u.Respond(w, u.MessageError(nil, "The hashID length must be 12 symbols"))
			return
		}

		issuerAccount,err := models.GetAccountByHash(accountHashId)
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}


		r = r.WithContext(context.WithValue(r.Context(), "auth", "ui/api"))
		r = r.WithContext(context.WithValue(r.Context(), "issuerAccountId", issuerAccount.ID))
		r = r.WithContext(context.WithValue(r.Context(), "issuerAccount", issuerAccount))

		next.ServeHTTP(w, r)
	})

}

// Вставляет в контекст issuerAccountId из accountId (раскрытие issuer account)
func ContextMuxVarIssuerAccountId(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		accountStr := mux.Vars(r)["accountId"]

		if accountStr == "" {
			u.Respond(w, u.MessageError(u.Error{Message:"Account id not found!"}))
			return
		}

		accountIdParse, err :=  strconv.ParseUint(accountStr, 10, 64)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(u.Error{Message:"Account id не корректно"}))
			return
		}

		if accountIdParse < 1 {
			u.Respond(w, u.MessageError(nil, "The hashID length must be 12 symbols"))
			return
		}

		accountId := uint(accountIdParse)

		account,err := models.GetAccount(accountId)
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "auth", "ui/api"))
		//r = r.WithContext(context.WithValue(r.Context(), "issuerAccountId", issuerAccount.ID))
		r = r.WithContext(context.WithValue(r.Context(), "accountId", account.ID))
		//r = r.WithContext(context.WithValue(r.Context(), "issuerAccount", issuerAccount))
		r = r.WithContext(context.WithValue(r.Context(), "account", account)) // адрес переменной!!!

		next.ServeHTTP(w, r)
	})

}
