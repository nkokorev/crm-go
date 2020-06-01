package controllers

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"reflect"
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

	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}

	// получаем объект типа Аккаунт
	accountI := r.Context().Value("account")
	if reflect.TypeOf(accountI).Elem().String() != "models.Account" {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}
	account := accountI.(*models.Account)

	if account == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Account nil pointer")
	}

	if account.ID < 1 {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Id of issuer account is zero!")
	}

	return account, nil
}

func GetWorkAccountCheckHashId(w http.ResponseWriter, r *http.Request) (*models.Account, error) {

	if r.Context().Value("account") == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}

	// получаем объект типа Аккаунт
	accountI := r.Context().Value("account")
	if reflect.TypeOf(accountI).Elem().String() != "models.Account" {
		u.Respond(w, u.MessageError(u.Error{Message: "Account is not valid"}))
		return nil, errors.New("Issuer account is null!")
	}
	account := accountI.(*models.Account)

	if account == nil {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Account nil pointer")
	}

	// получаем переменную из строки запрос URL: {hashId}
	hashId, ok := GetSTRVarFromRequest(r,"hashId")
	if !ok {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка hash id code of account"}))
		return nil, errors.New("Ошибка hash id code of account")
	}

	if account.HashID != hashId {
		u.Respond(w, u.MessageError(u.Error{Message: "Ошибка авторизации"}))
		return nil, errors.New("Вы авторизованы в другом аккаунте")
	}

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

func GetSTRVarFromRequest(r *http.Request, name string) (string, bool) {

	strVar := mux.Vars(r)[name]

	if strVar == "" {
		return "", false
	}

	return string(strVar), true
}

// FOR GET Requests!!!!
func GetQueryINTVarFromGET(r *http.Request, key string) (int, bool) {

	strVar := r.URL.Query().Get(key)

	if strVar == "" {
		return 0, false
	}

	intVar, err := strconv.ParseInt(strVar, 10, 64)
	if err != nil {
		return 0, false
	}

	return int(intVar), true
}

// FOR GET Requests!!!!
func GetQuerySTRVarFromGET(r *http.Request, key string) (string, bool) {

	strVar := r.URL.Query().Get(key)

	if strVar == "" {
		return "", false
	}

	return string(strVar), true
}