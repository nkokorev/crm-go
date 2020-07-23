package appCr

import (
	"bytes"
	"encoding/json"
	"fmt"
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
		fmt.Println(err)
		u.Respond(w, u.MessageError(u.Error{Message:"Слишком большой файл. Максимум 32 Mb."}))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
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
	/*var productID uint64
	productSTR := r.FormValue("productID")
	if productSTR != "" {
		productID, err = strconv.ParseUint(productSTR, 10, 64)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка чтения данных о файле productID"}))
			return
		}
	}

	var emailID uint64
	emailSTR := r.FormValue("emailID")
	if emailSTR != "" {
		emailID, err = strconv.ParseUint(emailSTR, 10, 64)
		if err != nil {
			fmt.Println(err)
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка чтения данных о файле emailID"}))
			return
		}
	}*/

	// нужна отдельная привязка к файлам
	ownerID, ok := utilsCr.GetQueryUINTVarFromGET(r, "ownerID")
	if !ok || ownerID < 1 {
		ownerID = 0
	}

	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "ownerType")
	if !ok {
		ownerType = ""
	}

	// check accoount entity
	/*switch ownerType {
	case "products":
		_, err = account.GetProduct(ownerID)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска товара для загрузки изображения"}))
			return
		}
	case "articles":
		_, err = account.GetArticle(ownerID)
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

	fs := models.Storage{
		Name: strings.ToLower(header.Filename),
		Data: buf.Bytes(),
		MIME: header.Header.Get("Content-Type"),
		Size: uint(header.Size),
		// OwnerID: ownerID, // нуу хз
		// OwnerType: ownerType,
	}

	fileEntity, err := account.CreateEntity(&fs)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Сервер не может обработать запрос")) // что это?)
		return
	}

	_fl, ok := fileEntity.(*models.Storage)

	switch ownerType {
	case "products":
		err = (models.Product{ID: ownerID}).AppendAssociationImage(_fl)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска продуктадля загрузки изображения"}))
			return
		}
	case "articles":
		err = (models.Article{ID: ownerID}).AppendAssociationImage(_fl)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска статьи для загрузки изображения"}))
			return
		}
	}



	diskSpaceUsed, err := account.StorageDiskSpaceUsed()
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка подсчета объема файлов"}))
		return
	}

	resp := u.Message(true, "File is save!")
	resp["file"] = _fl
	resp["diskSpaceUsed"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageGetFile(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	fileID, err := utilsCr.GetUINTVarFromRequest(r,"fileID")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	var file models.Storage
	err = account.LoadEntity(&file, fileID)
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
	resp["diskSpaceUsed"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageGet(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	fileID, err := utilsCr.GetUINTVarFromRequest(r,"fileID")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	var file models.Storage
	err = account.LoadEntity(&file,fileID)
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
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	// personal type
	ownerID, ok := utilsCr.GetQueryUINTVarFromGET(r, "ownerID")
	if !ok || ownerID < 1 {
		ownerID = 0
	}
	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "ownerType")
	if !ok {
		ownerType = ""
	}
	
	// without Data (body of file)
	// todo тут надо тип файлов дописать
	var total uint = 0
	files := make([]models.Entity,0)

	files, total, err = account.GetStoragePaginationListByOwner(offset, limit, sortBy, search, ownerID, ownerType)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список ВебХуков"))
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
	resp["diskSpaceUsed"] = diskSpaceUsed
	u.Respond(w, resp)
}

func StorageUpdateFile(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get file in Data base
	fileID, err := utilsCr.GetUINTVarFromRequest(r,"fileID")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"ID файла не найден"}))
		return
	}

	var file models.Storage
	err = account.LoadEntity(&file,fileID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
		return
	}
	
	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.UpdateEntity(&file,input)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка обновления файла"}))
		return
	}


	resp := u.Message(true, "Storage file saved")
	resp["file"] = file
	u.Respond(w, resp)
}

func StorageDeleteFile(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// Get file in Data base
	fileID, err := utilsCr.GetUINTVarFromRequest(r,"fileID")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	var file models.Storage
	err = account.LoadEntity(&file,fileID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
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
	resp["diskSpaceUsed"] = diskSpaceUsed
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
	resp["diskSpaceUsed"] = diskSpaceUsed
	u.Respond(w, resp)
}




// ### FOR CDN ###
func StorageCDNGet(w http.ResponseWriter, r *http.Request) {

	hashID, ok := utilsCr.GetSTRVarFromRequest(r, "hashID")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	fs, err := (models.Account{}).StorageGetPublicByHashID(hashID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	// w.Header("Content-Type", writer.FormDataContentType())
	// w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Type", fs.MIME)
	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",fs.Name))
	// w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s",fs.Name))
	w.Write(fs.Data)
}

func ArticleRawPreviewCDNGet(w http.ResponseWriter, r *http.Request) {


	hashID, ok := utilsCr.GetSTRVarFromRequest(r, "hashID")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	article, err := (models.Account{}).GetArticleSharedByHashID(hashID)
	if err != nil {
		fmt.Println("fdsfds", err)
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(article.Body))
}

func ArticleCompilePreviewCDNGet(w http.ResponseWriter, r *http.Request) {

	hashID, ok := utilsCr.GetSTRVarFromRequest(r, "hashID")
	if !ok  {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка id file"}))
		return
	}

	article, err := (models.Account{}).GetArticleSharedByHashID(hashID)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(article.Body))
}