package appCr

import (
	"bytes"
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"io"
	"net/http"
	"strings"
)

// ### NON PUBLIC Function ### //
// todo: дописать вские мелочи
func StorageCreateFile(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	err = r.ParseMultipartForm(32 << 20) // 32Mb
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Слишком большой файл. Максимум 32 Mb."}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка парсинга"}))
		return
	}

	var buf bytes.Buffer
	defer file.Close()


	// 12 Kb = 12022 Байт
	// size := float64(0)	// Mb
	// size =  float64(header.Size)/float64(1024)
	// fmt.Printf("Size: %d bytes\n", header.Size)
	// fmt.Println("Header: ", header.Header)
	// fmt.Println("purpose: ", r.Body)
	// fmt.Println("Content-Type: ", header.Header.Get("Content-Type"))
	// fmt.Println("File name: ", header.Filename)

	// Собираем всякие мета данные загружаемого файла для его предназначения
	/*var productId uint64
	productSTR := r.FormValue("productId")
	if productSTR != "" {
		productId, err = strconv.ParseUint(productSTR, 10, 64)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка чтения данных о файле productId"}))
			return
		}
	}

	var emailId uint64
	emailSTR := r.FormValue("emailId")
	if emailSTR != "" {
		emailId, err = strconv.ParseUint(emailSTR, 10, 64)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка чтения данных о файле emailId"}))
			return
		}
	}*/

	// нужна отдельная привязка к файлам
	ownerId, ok := utilsCr.GetQueryUINTVarFromGET(r, "owner_id")
	if !ok || ownerId < 1 {
		ownerId = 0
	}

	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "owner_type")
	if !ok {
		ownerType = ""
	}

	// check accoount entity
	/*switch ownerType {
	case "products":
		_, err = account.GetProduct(ownerId)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска товара для загрузки изображения"}))
			return
		}
	case "articles":
		_, err = account.GetArticle(ownerId)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска статьи для загрузки изображения"}))
			return
		}
	}*/


	_, err = io.Copy(&buf, file)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка сохранения файла"}))
		return
	}

	/*priority, _, err := r.ParseForm("priority")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка парсинга"}))
		return
	}*/

	// Определяем приоритет


	var fs models.Storage
	if  ownerId > 0 {
		fs = models.Storage{
			AccountId: account.Id,
			Name: strings.ToLower(header.Filename),
			Data: buf.Bytes(),
			MIME: header.Header.Get("Content-Type"),
			Size: int64(header.Size),
			OwnerID: ownerId, // нуу хз
			OwnerType: ownerType,
		}
	} else {
		fs = models.Storage{
			AccountId: account.Id,
			Name: strings.ToLower(header.Filename),
			Data: buf.Bytes(),
			MIME: header.Header.Get("Content-Type"),
			Size: int64(header.Size),
		}
	}


	if err := fs.SetAutoPriority(); err != nil {
		u.Respond(w, u.MessageError(err, "Сервер не может установить приоритет изображения")) // что это?)
		return
	}

	fileEntity, err := account.CreateEntity(&fs)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Сервер не может обработать запрос")) // что это?)
		return
	}

	_fl, ok := fileEntity.(*models.Storage)

	/*switch ownerType {
	case "products":
		err = (models.Product{Id: ownerId}).AppendAssociationImage(_fl)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска продукта для загрузки изображения"}))
			return
		}
	case "articles":
		err = (models.Article{Id: ownerId}).AppendAssociationImage(_fl)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска статьи для загрузки изображения"}))
			return
		}
	}*/



	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "File is save!")
	resp["file"] = _fl
	resp["disk_space_used"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageGetFile(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var file models.Storage
	err = account.LoadEntity(&file, fileId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
		return
	}

	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "Storage get file")
	resp["file"] = file
	resp["disk_space_used"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageGet(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var file models.Storage
	err = account.LoadEntity(&file,fileId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
		return
	}

	resp := u.Message(true, "Storage get file")
	resp["file"] = file
	u.Respond(w, resp)
}

func StorageGetListPagination(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
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

	// personal type
	ownerId, ok := utilsCr.GetQueryUINTVarFromGET(r, "owner_id")
	if !ok || ownerId < 1 {
		ownerId = 0
	}
	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "owner_type")
	if !ok {
		ownerType = ""
	}
	
	// without Data (body of file)
	// todo тут надо тип файлов дописать
	var total int64 = 0
	files := make([]models.Entity,0)

	files, total, err = account.GetStoragePaginationListByOwner(offset, limit, sortBy, search, ownerId, ownerType)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список файлов"))
		return
	}

	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "Storage get file")
	resp["files"] = files
	resp["total"] = total
	resp["disk_space_used"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageUpdateFile(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get file in Data base
	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Id файла не найден"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var file models.Storage
	err = account.LoadEntity(&file,fileId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
		return
	}
	
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}
	delete(input, "data")
	err = account.UpdateEntity(&file,input,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка обновления файла"}))
		return
	}


	resp := u.Message(true, "Storage file saved")
	resp["file"] = file
	u.Respond(w, resp)
}

func StorageMassUpdates(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	var input struct{
		Files []models.Storage `json:"files"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if err = (models.Storage{}).UpdatePriority(input.Files); err !=nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе 2"))
		return
	}

	// Проверяем переданные файлы
	// fmt.Println("input.Files[0].OwnerType: ",	input.Files[0].OwnerType)
	// fmt.Println("input.Files[0].OwnerID: ",		input.Files[0].OwnerID)

	if len(input.Files) < 1 {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе: необходимо указать []files"))
		return
	}

	if len(input.Files[0].OwnerType) < 1 || input.Files[0].OwnerID < 1 {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе: необходимо указать OwnerType & OwnerID"))
		return
	}

	filter := make(map[string]interface{},0)
	if len(input.Files) > 0 {
		filter["owner_id"] = input.Files[0].OwnerID
		filter["owner_type"] = input.Files[0].OwnerType
	}

	files := make([]models.Entity,0)

	var total int64
	files, total, err = account.GetPaginationListEntity(&models.Storage{}, 0, 25, "priority", "", filter,nil)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список"))
		return
	}

	resp := u.Message(true, "PATCH Storage MassUpdates")
	resp["files"] = files
	resp["total"] = total
	u.Respond(w, resp)
}

func StorageDeleteFile(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get file in Data base
	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	preloads := utilsCr.GetQueryStringArrayFromGET(r, "preloads")

	var file models.Storage
	err = account.LoadEntity(&file,fileId,preloads)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Не удалось найти файл"}))
		return
	}
	
	err = account.DeleteEntity(&file)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка удаления файла"))
		return
	}

	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "Email templates created")
	resp["disk_space_used"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageDiskSpaceUsed(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "Storage get file")
	resp["disk_space_used"] = diskSpaceUsed
	u.Respond(w, resp)
}


// ### FOR CDN ###
func StorageCDNGet(w http.ResponseWriter, r *http.Request) {

	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "hashId")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	fs, err := (models.Account{}).StorageGetPublicByHashId(hashId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	// w.Header("Content-Type", writer.FormDataContentType())
	// w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Type", fs.MIME)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Max-Age", "86400") // max 600
	// w.Header().Set("X-Content-Type-Options", "nosniff")
	// w.Header().Set("Access-Control-Allow-Methods", "GET,HEAD,OPTIONS,POST,PUT,PATCH,DELETE")
	// w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept-Language,Cache-Control,Content-Type,Content-Length,Accept,Origin,X-Requested-With,Access-Control-Request-Headers,Access-Control-Request-Method,Access-Control-Allow-Credentials,Host, Origin, User-Agent, Referer")

	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",fs.Name))
	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",fs.Name))
	
	w.Write(fs.Data)
}

func ArticleRawPreviewCDNGet(w http.ResponseWriter, r *http.Request) {


	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "hashId")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	article, err := (models.Account{}).GetArticleSharedByHashId(hashId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	if article.Body == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка: отсутствует тело статьи"}))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(*article.Body))
}

func ArticleCompilePreviewCDNGet(w http.ResponseWriter, r *http.Request) {

	hashId, ok := utilsCr.GetSTRVarFromRequest(r, "hashId")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	article, err := (models.Account{}).GetArticleSharedByHashId(hashId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	if article.Body == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка: отсутствует тело статьи"}))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(*article.Body))
}