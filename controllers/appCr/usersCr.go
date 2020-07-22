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
		RoleID uint `json:"roleID"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	var role models.Role
	if err = account.LoadEntity(&role, input.RoleID); err != nil {
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

	userID, err := utilsCr.GetUINTVarFromRequest(r, "userID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	user, err := account.GetUser(userID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Не удалось найти пользователя"))
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
	// users := make([]models.UserAndRole,0)
	users := make([]models.User,0)

	// При наличии "list=1,2,3" делается выборка по указанным ID
	var list []uint
	listStr := r.URL.Query().Get("list")

	if listStr != "" {
		
		listArr := strings.Split(listStr, ",")
		for _, v := range listArr {
			i, err := strconv.ParseUint(v, 10, 64)
			if err == nil {
				list = append(list, uint(i))
			}
		}

		// I. получаем выборку пользователей

		users, total, err = account.GetUsersByList(list, sortBy)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось получить список пользователей"))
			return
		}


	} else {
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

		resp := u.Message(true, "GET Account User List")
		resp["total"] = total
		resp["users"] = users
		u.Respond(w, resp)
	}


}

func UserUpdate(w http.ResponseWriter, r *http.Request) {

	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	userID, err := utilsCr.GetUINTVarFromRequest(r, "userID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		u.Respond(w, u.MessageError(err, "Техническая ошибка в запросе"))
		return
	}

	// Если обновляются роли, удаляем из общего массива input и потом отдельно обновляем
	var roleID float64
	var apiRoleID64 float64 = -1
	var role models.Role

	/*if roleIDStr, ok := input["roleID"]; ok {
		roleID, ok = roleIDStr.(float64)
		if !ok {
			roleID = 0
		} else {
			// 1. Получаем роль, которую надо назначить. Если роль вне аккаунта и не системная, получим ошибку
			rolePtr, err := account.GetRole(uint(roleID))
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

	delete(input, "roleID")*/

	if _role, ok := input["_role"]; ok {

		mRole, ok := _role.(map[string]interface{})
		if !ok {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя: роль не опознана"))
			return
		}

		roleIDVar, ok := mRole["id"]
		if !ok {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя: id роли не опознана"))
			return
		}

		roleID64, ok := roleIDVar.(float64)
		if !ok {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя: id роли не читается"))
			return
		}

		// 1. Получаем роль, которую надо назначить. Если роль вне аккаунта и не системная, получим ошибку
		rolePtr, err := account.GetRole(uint(roleID64))
		if err != nil {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
			return
		}

		if rolePtr.IsOwner() {
			u.Respond(w, u.MessageError(err, "Нельзя назначить пользователя с ролью владельца аккаунта"))
			return
		}

		role = *rolePtr
		roleID = roleID64
	}


	if apiRoleIDI, ok := input["roleID"]; ok {
		if _apiRoleID64, ok := apiRoleIDI.(float64); ok {
			apiRoleID64 = _apiRoleID64
		}
	}

	delete(input, "_role")
	delete(input, "roleID")
	delete(input, "accountUser")

	// Обновляем данные пользователя

	user, err := account.UpdateUser(userID, input)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Пользователь не найден"))
		return
	}


	// Обновляем роль пользователя, если она изменилась
	currentRole, err := account.GetUserRole(*user)
	if err == nil && roleID > 0 && (currentRole.ID != uint(roleID)){

		_, err = account.SetUserRole(user, role)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
			return
		}
	}  else {

		 if apiRoleID64 > 0 {
			 currentRole2, err := account.GetUserRole(*user)
			 if err == nil && (currentRole2.ID != uint(apiRoleID64)){

			 	// получаем роль 2
				 rolePtr, err := account.GetRole(uint(apiRoleID64))
				 if err != nil {
					 u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
					 return
				 }

				 if rolePtr.IsOwner() {
					 u.Respond(w, u.MessageError(err, "Нельзя назначить пользователя с ролью владельца аккаунта"))
					 return
				 }
				 
				 _, err = account.SetUserRole(user, *rolePtr)
				 if err != nil {
					 u.Respond(w, u.MessageError(err, "Ошибка в обновлении роли пользователя"))
					 return
				 }
			 }
		 }
	}


	user, err = account.GetUserWithAUser(user.ID)
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка при поиске пользователя"))
		return
	}

	resp := u.Message(true, "PATCH User Update")
	resp["user"] = user
	u.Respond(w, resp)
}

// Удаляет пользователя из аккаунта
// Если issuerID = accountID, то может быть применен запрос на удаление пользователя
func UserRemoveFromAccount(w http.ResponseWriter, r *http.Request) {
	
	account, err := utilsCr.GetWorkAccount(w, r)
	if err != nil || account == nil {
		return
	}

	userID, err := utilsCr.GetUINTVarFromRequest(r, "userID")
	if err != nil {
		u.Respond(w, u.MessageError(err, "Ошибка в обработке ID"))
		return
	}

	user, err := account.GetUser(userID)
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
		// Убирает аккаунт пользователя или удаляет, если он из этого аккаунта
		err = account.RemoveUser(user)
		if err != nil {
			u.Respond(w, u.MessageError(err, "Не удалось исключить пользователя"))
			return
		}
	}

	resp := u.Message(true, "DELETE User from Account")
	u.Respond(w, resp)
}
