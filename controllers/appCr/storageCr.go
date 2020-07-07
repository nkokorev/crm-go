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
	ownerId, ok := utilsCr.GetQueryUINTVarFromGET(r, "ownerId")
	if !ok || ownerId < 1 {
		ownerId = 0
	}

	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "ownerType")
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

	fs := models.Storage{
		Name: strings.ToLower(header.Filename),
		Data: buf.Bytes(),
		MIME: header.Header.Get("Content-Type"),
		Size: uint(header.Size),
		// OwnerID: ownerId, // нуу хз
		// OwnerType: ownerType,
	}

	_fl, err := account.StorageCreateFile(&fs)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Сервер не может обработать запрос")) // что это?)
		return
	}

	switch ownerType {
	case "products":
		err = (models.Product{ID: ownerId}).AppendAssociationImage(*_fl)
		if err != nil {
			u.Respond(w, u.MessageError(u.Error{Message:"Ошибка поиска продуктадля загрузки изображения"}))
			return
		}
	case "articles":
		err = (models.Article{ID: ownerId}).AppendAssociationImage(*_fl)
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

	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	fs, err := account.StorageGet(fileId)
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
	resp["file"] = *fs
	resp["diskSpaceUsed"] = diskSpaceUsed
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

	fs, err := account.StorageGet(fileId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Получения списка"}))
		return
	}

	resp := u.Message(true, "Storage get file")
	resp["file"] = *fs
	u.Respond(w, resp)
}

func StorageGetListPagination(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := utilsCr.GetQueryUINTVarFromGET(r, "limit")
	if !ok || limit < 1 {
		limit = 100
	}
	offset, ok := utilsCr.GetQueryUINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	///
	ownerId, ok := utilsCr.GetQueryUINTVarFromGET(r, "ownerId")
	if !ok || ownerId < 1 {
		ownerId = 0
	}
	ownerType, ok := utilsCr.GetQuerySTRVarFromGET(r, "ownerType")
	if !ok {
		ownerType = ""
	}
	
	// without Data (body of file)
	// todo тут надо тип файлов дописать
	files, total, err := account.StorageGetList(offset, limit, search, &ownerId, &ownerType)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка получения списка файлов"}))
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
	fileId, err := utilsCr.GetUINTVarFromRequest(r,"fileId")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"ID файла не найден"}))
		return
	}

	fs, err := account.StorageGet(fileId)
	if err != nil || fs == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}

	// 2. Get JSON-request
	/*input := struct {
		models.Storage
	}{}*/

	var input map[string]interface{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	err = account.StorageUpdateFile(fs, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка во время обновления данных файла"))
		return
	}

	resp := u.Message(true, "Storage file saved")
	resp["file"] = *fs
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

	err = account.StorageDeleteFile(fileId)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Произошла ошибка во время удаления"}))
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




// Example OLD  function
func StorageStore(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// r.ParseMultipartForm(4096)
	// v := r.FormValue("file")
	// r.ParseMultipartForm(32 << 20) // limit your max input length!

	file, header, err := r.FormFile("file")
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка парсинга"}))
		return
	}

	var buf bytes.Buffer
	defer file.Close()

	// 12 Kb = 12022
	// size := float64(0)	// Mb
	// size =  float64(header.Size)/float64(1024)
	fmt.Printf("Size: %d bytes\n", header.Size)
	fmt.Println("Header: ", header.Header)
	fmt.Println("Content-Type: ", header.Header.Get("Content-Type"))
	fmt.Println("File name: ", header.Filename)

	// name := strings.Split(header.Filename, ".")
	// fmt.Printf("File name: %s\n", name[0])
	// Copy the file data to my buffer
	_, err = io.Copy(&buf, file);
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка сохранения файла"}))
		return
	}

	fs := models.Storage{
		Name: header.Filename,
		Data: buf.Bytes(),
		MIME: header.Header.Get("Content-Type"),
	}
	_, err = account.StorageCreateFile(&fs)
	if err != nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка создания файла"}))
		return
	}


	resp := u.Message(true, "File is save!")
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
		fmt.Println("fdsfds", err)
		u.Respond(w, u.MessageError(u.Error{Message:"Файл не найден"}))
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(article.Body))
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(article.Body))
}