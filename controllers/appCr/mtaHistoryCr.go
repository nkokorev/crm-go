package appCr

import (
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func MTAHistoryGetListPagination(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil { return }

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

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// Узнаем нужен ли фильтр
	filter := map[string]interface{}{}
	ownerType, _filterOwnerType := utilsCr.GetQuerySTRVarFromGET(r, "ownerType")
	ownerId, _filterOwnerId := utilsCr.GetQueryUINTVarFromGET(r, "ownerId")
	if _filterOwnerType && _filterOwnerId{
		filter["owner_type"] = ownerType
		filter["owner_id"] = ownerId
	}
	var total int64 = 0
	mtaHistories := make([]models.Entity,0)

	if all {
		mtaHistories, total, err = account.GetListEntity(&models.MTAHistory{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список истории"))
			return
		}
	} else {

		mtaHistories, total, err = account.GetPaginationListEntity(&models.MTAHistory{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список mta-history"))
			return
		}
	}

	resp := u.Message(true, "GET MTA History Pagination List")
	resp["total"] = total
	resp["mta_histories"] = mtaHistories
	u.Respond(w, resp)
}
