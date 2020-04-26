package controllers

import (
	"errors"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func GetIssuerAccount(w http.ResponseWriter, r *http.Request) (error, *models.Account) {

	// Получаем аккаунт, в котором авторизуется пользователь
	if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return errors.New("Issuer account is null!"), nil
	}

	issuerAccount := r.Context().Value("issuerAccount").(*models.Account)

	if issuerAccount.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return errors.New("Id of issuer account is zero!"), nil
	}

	return nil, issuerAccount
}

//
func GetWorkAccount(w http.ResponseWriter, r *http.Request) (error, *models.Account) {

	// Получаем аккаунт, в котором авторизуется пользователь
	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return errors.New("Issuer account is null!"), nil
	}

	account := r.Context().Value("account").(*models.Account)

	if account.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return errors.New("Id of issuer account is zero!"), nil
	}

	return nil, account
}
