package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func EmailNotificationCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	var input struct{
		models.EmailNotification
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	entity, err := account.CreateEntity(&input.EmailNotification)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания EmailNotification"}))
		return
	}

	resp := u.Message(true, "POST EmailNotification Create")
	resp["email_notification"] = entity
	u.Respond(w, resp)
}

func EmailNotificationGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	emailNotificationId, err := utilsCr.GetUINTVarFromRequest(r, "emailNotificationId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailNotification"))
		return
	}

	var emailNotification models.EmailNotification

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&emailNotification, emailNotificationId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект 1"))
			return
		}
	} else {
		err = account.LoadEntity(&emailNotification, emailNotificationId)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось найти объект 2"))
			return
		}
	}
	
	resp := u.Message(true, "GET Email Notification")
	resp["email_notification"] = emailNotification
	u.Respond(w, resp)
}

func EmailNotificationGetListPagination(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	limit, ok := utilsCr.GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 25
	}
	offset, ok := utilsCr.GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	sortDesc := utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = "id"
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}
	// 2. Узнаем, какой список нужен
	all := utilsCr.GetQueryBoolVarFromGET(r, "all")


	var total int64 = 0
	emailNotifications := make([]models.Entity,0)

	if all {
		emailNotifications, total, err = account.GetListEntity(&models.EmailNotification{}, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить данные"))
			return
		}
	} else {
		// emailNotifications, total, err = account.GetEmailNotificationsPaginationList(offset, limit, search)
		emailNotifications, total, err = account.GetPaginationListEntity(&models.EmailNotification{}, offset, limit, sortBy, search, nil)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить данные"))
			return
		}
	}


	resp := u.Message(true, "GET Email Notification PaginationList")
	resp["email_notifications"] = emailNotifications
	resp["total"] = total
	u.Respond(w, resp)
}

func EmailNotificationUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	emailNotificationId, err := utilsCr.GetUINTVarFromRequest(r, "emailNotificationId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var emailNotification models.EmailNotification
	if err = account.LoadEntity(&emailNotification, emailNotificationId); err != nil {
		u.Respond(w, u.MessageError(err, "Уведомление не найдено"))
		return
	}

	if err = account.UpdateEntity(&emailNotification, input); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Email Notification Update")
	resp["email_notification"] = emailNotification
	u.Respond(w, resp)
}

func EmailNotificationDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	emailNotificationId, err := utilsCr.GetUINTVarFromRequest(r, "emailNotificationId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id"))
		return
	}

	var emailNotification models.EmailNotification
	if err = account.LoadEntity(&emailNotification, emailNotificationId); err != nil {
		u.Respond(w, u.MessageError(err, "Уведомление не найдено"))
		return
	}

	if err = account.DeleteEntity(&emailNotification); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении"))
		return
	}

	resp := u.Message(true, "DELETE EmailNotification Successful")
	u.Respond(w, resp)
}
