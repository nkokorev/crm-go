package controllers

import (
	"bytes"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"io"
	"net/http"
)

func StorageStore(w http.ResponseWriter, r *http.Request) {

	account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}

	// r.ParseMultipartForm(4096)
	// v := r.FormValue("file")
	// r.ParseMultipartForm(32 << 20) // limit your max input length!

	file, header, err := r.FormFile("file")
	if err != nil {
		fmt.Println("Error of parsing: ", err)
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
	_, err = account.StorageCreate(&fs)
	if err != nil {
		fmt.Println("Ошибка создания файла: ", err)
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка создания файла"}))
		return
	}


	resp := u.Message(true, "File is save!")
	u.Respond(w, resp)
}

func StorageGet(w http.ResponseWriter, r *http.Request) {

	/*account, err := GetWorkAccount(w,r)
	if err != nil || account == nil {
		u.Respond(w, u.MessageError(u.Error{Message:"Ошибка авторизации"}))
		return
	}*/

	hashId, err := GetSTRVarFromRequest(r, "hashId")
	if err != nil  {
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
	w.Write(fs.Data)
}