package middleware

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

//Тут собраны посредники определяющие issuerAccountId и добавляющие его в контекст

// Add to Context(r) issuerAccountId (= 1, Ratus CRM)
func AddContextMainAccount(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		issuerAccount, err := models.GetMainAccount() // RatusCRM
		if err != nil {
			u.Respond(w, u.MessageError(nil, "An account with the specified hash ID was not found"))
			return
		}

		// For future
		r = r.WithContext(context.WithValue(r.Context(), "issuer", "app"))
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

		issuerAccount, err := models.GetAccountByHash(accountHashId)
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
