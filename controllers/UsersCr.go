package controllers

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// limit & offset OR search
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