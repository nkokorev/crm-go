package appCr

import (
	"encoding/json"
	"fmt"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
	"strconv"
	"strings"
)

func UserCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil {
		return
	}

	var input struct{
		models.User
		RoleId uint `json:"roleId"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	fmt.Println("RoleId: ", input.RoleId)

	var role models.Role
	if err = account.LoadEntity(&role, input.RoleId); err != nil {
		fmt.Println(err)
		u.Respond(w, u.MessageError(err, "Роль пользователя не найдена!"))
		return
	}

	if role.IsOwner() {
		u.Respond(w, u.MessageError(err, "Нельзя создать пользователя с ролью владельца аккаунта"))
		return
	}

	user, err := account.CreateUser(input.User, role)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось создать пользователя"))
		return
	}
	
	resp := u.Message(true, "CREATE User in Account")
	resp["user"] = user
	u.Respond(w, resp)
}

func UserGet(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	userId, err := utilsCr.GetUINTVarFromRequest(r, "userId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	user, err := account.GetUser(userId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти магазин"))
		return
	}

	resp := u.Message(true, "GET User")
	resp["user"] = user
	u.Respond(w, resp)
}

func UsersGetListPagination(w http.ResponseWriter, r *http.Request) {

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
	sortDesc:= utilsCr.GetQueryBoolVarFromGET(r, "sortDesc") // обратный или нет порядок
	sortBy, ok := utilsCr.GetQuerySTRVarFromGET(r, "sortBy")
	if !ok {
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	rolesStr := r.URL.Query().Get("roles")
	rolesArr := strings.Split(rolesStr, ",")
	var roles []uint
	for _, v := range rolesArr {
		i, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			roles = append(roles, uint(i))
		}
	}

	search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
	if !ok {
		search = ""
	}

	users, total, err := account.GetUserListPagination(offset, limit, sortBy, search, roles)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось получить список пользователей"))
		return
	}

	resp := u.Message(true, "GET Account User List")
	resp["total"] = total
	resp["users"] = users
	u.Respond(w, resp)
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	userId, err := utilsCr.GetUINTVarFromRequest(r, "userId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	user, err := account.UpdateUser(userId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Пользователь не найден"))
		return
	}

	resp := u.Message(true, "PATCH User Update")
	resp["user"] = user
	u.Respond(w, resp)
}

// Удаляет пользователя из аккаунта
// Если issuerId = accountId, то может быть применен запрос на удаление пользователя 
func UserRemoveFromAccount(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w,r)
	if err != nil || account == nil {
		return
	}

	userId, err := utilsCr.GetUINTVarFromRequest(r, "userId")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	user, err := account.GetUser(userId)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Пользователь не найден"))
		return
	}

	// Узнаем доп. данные
	var input struct{
		DeleteUser bool `json:"deleteUser"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	if input.DeleteUser {
		err = account.DeleteUser(user)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось удалить пользователя"))
			return
		}
	} else {
		err = account.RemoveUser(user)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось исключить пользователя"))
			return
		}
	}


	resp := u.Message(true, "DELETE User from Account")
	u.Respond(w, resp)
}


