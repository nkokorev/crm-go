package controllers

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strconv"
)

func GetIssuerAccount(w http.ResponseWriter, r *http.Request) (*models.Account, error) {

	// Получаем аккаунт, в котором авторизуется пользователь
	if r.Context().Value("issuerAccount") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}

	issuerAccount := r.Context().Value("issuerAccount").(*models.Account)

	if issuerAccount.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Id of issuer account is zero!")
	}

	return issuerAccount,nil
}

// Возвращает рабочий контроллер
// todo ???
func GetWorkAccount(w http.ResponseWriter, r *http.Request) (*models.Account, error) {

	// Получаем аккаунт, в котором авторизуется пользователь
	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}

	account := r.Context().Value("account").(*models.Account)

	if account.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Id of issuer account is zero!")
	}

	return account, nil
}

func GetUINTVarFromRequest(r *http.Request, key string) (uint, error) {

	strVar := mux.Vars(r)[key]

	if strVar == "" {
		return 0, errors.New("Не верно указан account ID")
	}

	accountIdParse, err := strconv.ParseUint(strVar, 10, 64)
	if err != nil {
		return 0, errors.New("Не верно указан account ID")
	}

	if accountIdParse < 1 {
		return 0, errors.New("Не верно указан account ID")
	}

	return uint(accountIdParse), nil
}

func GetSTRVarFromRequest(r *http.Request, name string) (string, error) {

	strVar := mux.Vars(r)[name]

	if strVar == "" {
		return "", errors.New("Не верно указан account ID")
	}

	return string(strVar), nil
}