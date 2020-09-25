package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

func ProductCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get JSON-request
	var input struct{
		models.Product
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	product, err := account.CreateEntity(&input.Product)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка во время создания"}))
		return
	}

	resp := u.Message(true, "POST Product Created")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке product Id"))
		return
	}

	var product models.Product

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// 2. Узнаем, какой id учитывается нужен
	publicOk := utilsCr.GetQueryBoolVarFromGET(r, "public_id")

	if publicOk  {
		err = account.LoadEntityByPublicId(&product, productId,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить объект"))
			return
		}
	} else {
		err = account.LoadEntity(&product, productId,preloads)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(err, "Не удалось загрузить магазин"))
			return
		}
	}

	resp := u.Message(true, "GET Product ")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductListPaginationGet(w http.ResponseWriter, r *http.Request) {

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

	var total int64 = 0
	webSites := make([]models.Entity,0)

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if all {
		webSites, total, err = account.GetListEntity(&models.Product{}, sortBy,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	} else {
		// webHooks, total, err = account.GetWebHooksPaginationList(offset, limit, search)
		webSites, total, err = account.GetPaginationListEntity(&models.Product{}, offset, limit, sortBy, search, nil,preloads)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список страниц"))
			return
		}
	}

	resp := u.Message(true, "GET Product PaginationList")
	resp["products"] = webSites
	resp["total"] = total
	u.Respond(w, resp)
}

func ProductUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var product models.Product
	err = account.LoadEntity(&product, productId, preloads)
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

	// product, err := account.UpdateProduct(productId, &input.Product)
	err = account.UpdateEntity(&product, input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH Product Update")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductDelete(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id шаблона"))
		return
	}

	var product models.Product
	err = account.LoadEntity(&product, productId,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить магазин"))
		return
	}

	if err = account.DeleteEntity(&product); err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при удалении продукта"))
		return
	}

	resp := u.Message(true, "DELETE Product Successful")
	u.Respond(w, resp)
}

func ProductSyncProductCategories(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	var input struct{
		ProductCategories []models.ProductCategory `json:"product_categories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	if err = product.SyncProductCategoriesByIds(input.ProductCategories); err !=nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "PATCH Product sync Product Categories")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductSyncProductTags(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	var input struct{
		ProductTags []models.ProductTag `json:"product_categories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	if err = product.SyncProductTagsByIds(input.ProductTags); err !=nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "PATCH Product Sync Product Tags")
	resp["product"] = product
	u.Respond(w, resp)
}


/*func ProductRemoveSourceItem(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCategoryId, err := utilsCr.GetUINTVarFromRequest(r, "productCategoryId")
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

	var productCategory models.ProductCategory
	if err =account.LoadEntity(&productCategory, productCategoryId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productCategory.RemoveProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из категории товара"))
		return
	}

	if err = account.LoadEntity(&productCategory, productCategoryId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductCategory Products Remove")
	resp["product_category"] = productCategory
	u.Respond(w, resp)
}

func ProductAppendSourceItem(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productCategoryId, err := utilsCr.GetUINTVarFromRequest(r, "productCategoryId")
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

	var productCategory models.ProductCategory
	if err =account.LoadEntity(&productCategory, productCategoryId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке категории товара"}))
		return
	}

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = productCategory.AppendProduct(&product); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления продукта из категории товара"))
		return
	}

	var _productCategory models.ProductCategory
	if err = account.LoadEntity(&_productCategory, productCategoryId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка загрузки категории товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductCategory Append Product")
	resp["product_category"] = _productCategory
	u.Respond(w, resp)
}*/

// Добавляет продукт в POST сообщение по source_id
func ProductAppendSourceItem(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке product Id"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var product models.Product
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}
	
	var input struct{
		SourceId		uint		`json:"source_id"`
		AmountUnits 	float64 	`json:"amount_units"`
		EnableViewing	bool 		`json:"enable_viewing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	// Получаем продукт source
	productSource := models.Product{}
	if err =account.LoadEntity(&productSource, input.SourceId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	if err = product.AppendSourceItem(&productSource,input.AmountUnits, input.EnableViewing, true); err !=nil {
		u.Respond(w, u.MessageError(err, "Ошибка добавления товара как источник"))
		return
	}

	// еще раз загружаем товар
	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "PATCH ProductTagGroup Append Product")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductRemoveSourceItem(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	var product models.Product
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	// Получаем id товара, которого нужно удалить из source items
	productSourceId, err := utilsCr.GetUINTVarFromRequest(r, "productSourceId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productSourceId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	if err = product.RemoveSourceItem(productSourceId); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка удаления товара из source"))
		return
	}

	if err = account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "DELETE Product remove source item")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductUpdateSourceItem(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productId"))
		return
	}

	var product models.Product
	if err =account.LoadEntity(&product, productId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	// Получаем id товара, которого нужно удалить из source items
	productSourceId, err := utilsCr.GetUINTVarFromRequest(r, "productSourceId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке productSourceId"))
		return
	}

	// Загружаем связь, которую нужно изменить
	var productSource models.Product
	if err = account.LoadEntity(&productSource, productSourceId, nil); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	// Данные по обновлению
	var input struct {
		AmountUnits 	float64 `json:"amount_units"`
		EnableViewing	bool	`json:"enable_viewing"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	maps := structs.Map(&input)

	if err = product.UpdateSourceItem(productSourceId, maps); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Ошибка обновления товара из source"))
		return
	}

	if err = account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "DELETE Product remove source item")
	resp["product"] = product
	u.Respond(w, resp)
}

func ProductSyncSourceItems(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	productId, err := utilsCr.GetUINTVarFromRequest(r, "productId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке Id productId"))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	product := models.Product{}
	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	var input struct{
		ProductSources []models.ProductSource `json:"product_sources"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 1"))
		return
	}

	if err = product.SyncSourceItems(input.ProductSources); err !=nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	if err =account.LoadEntity(&product, productId, preloads); err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка в загрузке товара"}))
		return
	}

	resp := u.Message(true, "PATCH Product Sync Source Items")
	resp["product"] = product
	u.Respond(w, resp)
}