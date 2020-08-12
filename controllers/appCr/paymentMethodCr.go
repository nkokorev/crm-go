package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func PaymentMethodCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct {
		models.PaymentMethod
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// хз хз
	paymentMethod, err := account.CreateEntity(input.PaymentMethod)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания метода"}))
		return
	}

	resp := u.Message(true, "POST PaymentMethod Created")
	resp["paymentMethod"] = paymentMethod
	u.Respond(w, resp)
}

func PaymentMethodGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// узнаем ID элемента
	paymentMethodId, err := utilsCr.GetUINTVarFromRequest(r, "paymentMethodId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	// 2. Узнаем, какой список нужен
	code, ok := utilsCr.GetQuerySTRVarFromGET(r, "code")
	if !ok {
		u.Respond(w, u.MessageError(err, "Необходимо указать тип метода"))
		return
	}

	paymentMethod, err := account.GetPaymentMethodByCode(code, paymentMethodId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "GET PaymentMethod")
	resp["paymentMethod"] = paymentMethod
	u.Respond(w, resp)
}

func PaymentMethodGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	// todo: get webSiteid: ?webSite=1,2,3

	paymentMethods, err := account.GetPaymentMethods()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при получении списка методов оплаты"))
		return
	}

	resp := u.Message(true, "GET PaymentMethod List")
	resp["paymentMethods"] =  paymentMethods
	u.Respond(w, resp)
}

func PaymentMethodUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// узнаем ID элемента
	paymentMethodId, err := utilsCr.GetUINTVarFromRequest(r, "paymentMethodId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	// 2. Узнаем, какой список нужен
	code, ok := utilsCr.GetQuerySTRVarFromGET(r, "code")
	if !ok {
		u.Respond(w, u.MessageError(err, "Необходимо указать тип метода"))
		return
	}

	paymentMethod, err := account.GetPaymentMethodByCode(code, paymentMethodId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(paymentMethod, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH PaymentMethod Update")
	resp["paymentMethod"] = paymentMethod
	u.Respond(w, resp)
}

func PaymentMethodDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// узнаем ID элемента
	paymentMethodId, err := utilsCr.GetUINTVarFromRequest(r, "paymentMethodId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке web site Id"))
		return
	}

	// 2. Узнаем, какой список нужен
	code, ok := utilsCr.GetQuerySTRVarFromGET(r, "code")
	if !ok {
		u.Respond(w, u.MessageError(err, "Необходимо указать тип метода"))
		return
	}

	paymentMethod, err := account.GetPaymentMethodByCode(code, paymentMethodId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	if err = account.DeleteEntity(paymentMethod); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE PaymentMethod Successful")
	u.Respond(w, resp)
}
