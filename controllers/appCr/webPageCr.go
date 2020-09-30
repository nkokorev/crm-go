package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func WebPageCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.WebPage
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	webPage, err := account.CreateEntity(&input.WebPage)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST WebPage Created")
	resp["web_page"] = webPage
	u.Respond(w, resp)
}

func WebPageGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	webPageId, err := utilsCr.GetUINTVarFromRequest(r, "webPageId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке webPage Id"))
		return
	}

	var webPage models.WebPage

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if publicOk  {
		err = account.LoadEntityByPublicId(&webPage, webPageId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&webPage, webPageId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET WebPage ")
	resp["web_page"] = webPage
	u.Respond(w, resp)
}

func WebPageListPaginationGet(w http.ResponseWriter, r *http.Request) {

	var account *models.Account
	var err error
	// 1. Получаем рабочий аккаунт в зависимости от источника (автома. сверка с {hashId}.)

	account, err = utilsCr.GetWorkAccount(w,r)
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
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var total int64 = 0
	webPages := make([]models.Entity,0)

	// Узнаем нужен ли фильтр
	filter := map[string]interface{}{}
	webSiteId, _filterWebSite := utilsCr.GetQueryUINTVarFromGET(r, "webSiteId")
	if _filterWebSite {
		filter["web_site_id"] = webSiteId
	}

	if all && len(filter) < 1 {
		webPages, total, err = account.GetListEntity(&models.WebPage{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		webPages, total, err = account.GetPaginationListEntity(&models.WebPage{}, offset, limit, sortBy, search, filter,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}
	
	resp := u.Message(true, "GET Web Pages Pagination List")
	resp["web_pages"] = webPages
	resp["total"] = total
	u.Respond(w, resp)
}

func WebPageUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webPageId, err := utilsCr.GetUINTVarFromRequest(r, "webPageId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}
	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var webPage models.WebPage
	err = account.LoadEntity(&webPage, webPageId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось загрузить данные"))
		return
	}

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// fix variables
	/*if err := u.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}*/

	// webPage, err := account.UpdateWebPage(webPageId, &input.WebPage)
	err = account.UpdateEntity(&webPage, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH WebPage Update")
	resp["web_page"] = webPage
	u.Respond(w, resp)
}

func WebPageDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webPageId, err := utilsCr.GetUINTVarFromRequest(r, "webPageId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}


	var webPage models.WebPage
	err = account.LoadEntity(&webPage, webPageId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&webPage); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении web-страницы"))
		return
	}

	resp := u.Message(true, "DELETE WebPage Successful")
	u.Respond(w, resp)
}

func WebPageSyncProductCategories(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webPageId, err := utilsCr.GetUINTVarFromRequest(r, "webPageId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	webPage := models.WebPage{}
	if err =account.LoadEntity(&webPage, webPageId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Карточка товара"}))
		return
	}

	var input struct{
		Items []models.WebPageProductCategories `json:"web_page_product_categories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	if err = webPage.SyncProductCategoriesByIds(input.Items); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	// Загружаем еще раз
	if err = account.LoadEntity(&webPage, webPageId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Карточка товара"}))
		return
	}

	resp := u.Message(true, "PATCH Product Card MassUpdates")
	resp["web_page"] = webPage
	u.Respond(w, resp)
}

func WebPageRemoveCategory(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	webPageId, err := utilsCr.GetUINTVarFromRequest(r, "webPageId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productCategoryId, err := utilsCr.GetUINTVarFromRequest(r, "productCategoryId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var webPage models.WebPage
	if err =account.LoadEntity(&webPage, webPageId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	// Загружаем, чтобы убедить, что она принадлежит этому аккаунту
	productCategory := models.ProductCategory{}
	if err =account.LoadEntity(&productCategory, productCategoryId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = webPage.RemoveProductCategory(productCategory); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления категории со страницы"))
		return
	}

	// Загружаем еще раз
	if err = account.LoadEntity(&webPage, webPageId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Карточка товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductCard Products Remove")
	resp["web_page"] = webPage
	u.Respond(w, resp)
}

func WebPageAppendCategory(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCardId, err := utilsCr.GetUINTVarFromRequest(r, "productCardId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id emailQueueId"))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var productCard models.ProductCard
	if err =account.LoadEntity(&productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке карточки товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productCard.AppendProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из карточки товара"))
		return
	}

	var _productCard models.ProductCard
	if err = account.LoadEntity(&_productCard, productCardId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки карточки товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductCard Append Product")
	resp["product_card"] = _productCard
	u.Respond(w, resp)
}