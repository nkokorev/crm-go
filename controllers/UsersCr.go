package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// limit & offset OR search

func CreateUser(w http.ResponseWriter, r *http.Request) {
	
	account, err := GetWorkAccount(w,r)
	if err != nil {
		return
	}

	var input struct{
		models.User
		Role string `json:"role"`
	}

	if err := GetInputInterface(w,r, &input); err != nil {
		return
	}

	/*fmt.Println(input["role"])
	if input["fdsfsd"] == nil {
		fmt.Println("Nil!")
	}*/

	// role := input.Role.(models.AccessRole)

	user, err := account.CreateUser(input.User, input.Role)
	if err != nil {
		fmt.Println("Error: ", err)
		u.Respond(w, u.MessageError(err, "Не удалось создать пользователя"))
		return
	}
	// user := input.UserData
	                              

	resp := u.Message(true, "CREATE User IN Account")
	resp["account"] = account
	// resp["input"] = input
	resp["user"] = user
	u.Respond(w, resp)
}

func GetUserList(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}

	// 2. Узнаем, какой список нужен
	limit, ok := GetQueryINTVarFromGET(r, "limit")
	if !ok {
		limit = 100
	}
	offset, ok := GetQueryINTVarFromGET(r, "offset")
	if !ok || offset < 0 {
		offset = 0
	}
	search, ok := GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	// тут должны быть только утвержденные типы
	// userTypes := strings.Split(r.URL.Query().Get("types"),",")


	// fmt.Printf("Limit %d\n", limit)
	// fmt.Printf("Offset %d\n", offset)
	// fmt.Printf("Users %s\n", userTypes)

	users, total, err := account.GetUserList(offset, limit, search)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список пользователей"))
		return
	}



	resp := u.Message(true, "GET Account User List")
	resp["total"] = total
	resp["users"] = users
	u.Respond(w, resp)
}

func RoleList(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}

	roles, err := account.GetRoleList()
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список ролей"))
		return
	}



	resp := u.Message(true, "GET Account Role List")
	resp["roles"] = roles
	u.Respond(w, resp)
}

func RemoveUserFromAccount(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}

	userId, ok := GetSTRVarFromRequest(r, "userHashId")
	if !ok {
		u.Respond(w, u.MessageError(nil, "Не удалось ID пользователя"))
		return
	}

	err = account.RemoveUserByHashId(userId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось удалить пользователя"))
		return
	}

	resp := u.Message(true, "DELETE User from Account")
	u.Respond(w, resp)
}

func UpdateUserData(w http.ResponseWriter, r *http.Request) {
	// 1. Получаем рабочий аккаунт (автома. сверка с {hashId}.)
	account, err := GetWorkAccountCheckHashId(w,r)
	if err != nil || account == nil {
		return
	}

	userHashId, ok := GetSTRVarFromRequest(r, "userHashId")
	if !ok {
		u.Respond(w, u.MessageError(nil, "Не удалось ID пользователя"))
		return
	}

	/*input := struct {
		models.User
		// NativePwd   string `json:"password"`    // потому что пароль из User{} не читается т.к. json -
		// InviteToken string `json:"inviteToken"` // может присутствовать
	}{}*/

	input := map[string]interface{}{}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user, err := account.UserUpdateByHashId(userHashId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при обновлении"))
		return
	}

	resp := u.Message(true, "PATCH USER Update")
	resp["user"] = user
	u.Respond(w, resp)
}