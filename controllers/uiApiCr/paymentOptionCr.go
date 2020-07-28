package uiApiCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func PaymentOptionGetList(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	webSiteId, err := utilsCr.GetUINTVarFromRequest(r, "webSiteId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id магазина"))
		return
	}

	var webSite models.WebSite
	err = account.LoadEntity(&webSite, webSiteId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	resp := u.Message(true, "GET PaymentOption List")
	resp["paymentOptions"] = webSite.PaymentOptions
	u.Respond(w, resp)
}

// Уведомление от Я.Кассы
func PaymentYandexWebHook(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	paymentYandexHashId, ok := utilsCr.GetSTRVarFromRequest(r, "paymentYandexHashId")
	if !ok {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка - не найден paymentYandexHashId"}))
		return
	}

	paymentYandex, err := account.GetPaymentYandexByHashId(paymentYandexHashId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти способ оплаты"))
		return
	}

	fmt.Println("Способ оплаты найден: ", paymentYandex.Name)

	//////////////////////

	type ObjectPayment struct {
		Id string `json:"id"`
		Status string `json:"status"`
		Paid bool `json:"paid"`
	}
	// Читаем вход
	var input struct {
		Event string `json:"event"`
		Object ObjectPayment `json:"object"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка в запросе: проверьте обязательные поля и типы переменных"))
		return
	}

	fmt.Println("Input from Yandex Payment: ", input)
	fmt.Println("ObjectPayment Event: ", input.Event)
	fmt.Println("ObjectPayment ID: ", input.Object.Id)
	fmt.Println("ObjectPayment Status: ", input.Object.Status)
	fmt.Println("ObjectPayment Paid: ", input.Object.Paid)

	resp := u.Message(true, "Payment Option Yandex")
	u.Respond(w, resp)
}

