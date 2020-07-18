package appCr

import (
	"encoding/json"
	"github.com/nkokorev/crm-go/controllers/utilsCr"
	"github.com/nkokorev/crm-go/models"
	u "github.com/nkokorev/crm-go/utils"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func UserCreate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil {
		return
	}

	var input struct {
		models.User
		RoleId uint `json:"roleId"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var role models.Role
	if err = account.LoadEntity(&role, input.RoleId); err != nil {
		u.Respond(w, u.MessageError(err, "Роль пользователя не найдена!"))
		return
	}

	if role.IsOwner() {
		u.Respond(w, u.MessageError(err, "Нельзя создать пользователя с ролью владельца аккаунта"))
		return
	}

	// Т.к. пароль не передается, читаем и назначем отдельно json -
	input.User.Password = input.Password

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

	account, err := utilsCr.GetWorkAccount(w, r)
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

	account, err := utilsCr.GetWorkAccount(w, r)
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
		sortBy = ""
	}
	if sortDesc {
		sortBy += " desc"
	}

	var total uint
	users := make([]models.UserAndRole,0)

	var list []uint
	listStr := r.URL.Query().Get("list")
	if listStr != "" && listStr != "all" {
		listArr := strings.Split(listStr, ",")
		for _, v := range listArr {
			i, err := strconv.ParseUint(v, 10, 64)
			if err == nil {
				list = append(list, uint(i))
			}
		}

		// I. получаем выборку пользователей

		users, total, err = account.GetUsersByListID(list, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список пользователей"))
			return
		}


	} else {

		// II. Получаем pagination list
		var roles []uint
		rolesStr := r.URL.Query().Get("roles")
		if rolesStr == "" || rolesStr == "all" {
			_roles, err := account.GetRoleList()
			if err != nil {
				u.Respond(w, u.MessageError(err, "Не удалось получить список ролей пользователей"))
				return
			}
			for i := range _roles {
				roles = append(roles, _roles[i].ID)
			}
		} else {
			rolesArr := strings.Split(rolesStr, ",")
			for _, v := range rolesArr {
				i, err := strconv.ParseUint(v, 10, 64)
				if err == nil {
					roles = append(roles, uint(i))
				}
			}
		}

		search, ok := utilsCr.GetQuerySTRVarFromGET(r, "search")
		if !ok {
			search = ""
		}

		users, total, err = account.GetUserListPagination(offset, limit, sortBy, search, roles)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список пользователей"))
			return
		}

	}





	resp := u.Message(true, "GET Account User List")
	resp["total"] = total
	resp["users"] = users
	u.Respond(w, resp)
}

func UserUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
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

	var roleId float64
	var role models.Role

	if roleIdStr, ok := input["roleId"]; ok {
		roleId, ok = roleIdStr.(float64)
		if !ok {
			roleId = 0
		} else {
			// 1. Получаем роль, которую надо назначить
			rolePtr, err := account.GetRole(uint(roleId))
			if err != nil {
				u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
				return
			}

			if rolePtr.IsOwner() {
				u.Respond(w, u.MessageError(err, "Нельзя назначить пользователя с ролью владельца аккаунта"))
				return
			}

			role = *rolePtr
		}
	}

	delete(input, "roleId")

	user, err := account.UpdateUser(userId, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Пользователь не найден"))
		return
	}


	// Обновляем роль пользователя
	currentRole, err := account.GetUserRole(*user)
	if err == nil && roleId > 0 && (currentRole.ID != uint(roleId)){



		err = account.UpdateUserRole(*user, role)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
			return
		}

		user, err = account.GetUserWithRoleId(user.ID)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Ошибка при поиске пользователя"))
			return
		}

	}

	resp := u.Message(true, "PATCH User Update")
	resp["user"] = user
	u.Respond(w, resp)
}

// Удаляет пользователя из аккаунта
// Если issuerId = accountId, то может быть применен запрос на удаление пользователя 
func UserRemoveFromAccount(w http.ResponseWriter, r *http.Request) {
	account, err := utilsCr.GetWorkAccount(w, r)
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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось прочитать тело сообщения"))
		return
	}

	// Узнаем доп. данные
	var input struct {
		SoftDelete bool `json:"softDelete,omitempty"`
	}

	if len(string(body)) >= 0 {
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			input.SoftDelete = false
		}

	} else {
		input.SoftDelete = false
	}

	if input.SoftDelete {
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
