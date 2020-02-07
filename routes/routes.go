package routes

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"

	"github.com/nkokorev/crm-go/controllers"
	//"github.com/nkokorev/gui-server/controllers/old"
	"github.com/nkokorev/crm-go/middleware"
)

func Handlers() *mux.Router {

	var crmHost string

	switch os.Getenv("APP_ENV") {
	case "local":
		crmHost = "crm-local.me"
	case "public":
		crmHost = "ratuscrm.com"
	default:
		crmHost = "ratuscrm.com"
	}

	// обрабатываем все запросы со слешем и без
	//rBase := mux.NewRouter().StrictSlash(false)
	rBase := mux.NewRouter().StrictSlash(true)

	// монтируем все три точки входа для API
	rApi := rBase.Host("api." + crmHost).Subrouter() // api.ratuscrm.com
	rApp := rBase.Host("app." + crmHost).PathPrefix("/ui-api").Subrouter() // app.ratuscrm.com/ui-api
	rUiApi := rBase.Host("ui.api." + crmHost).Subrouter() // ui.api.ratuscrm.com

	// Функции проверки роутов API
	rApi.HandleFunc("/", controllers.CheckApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)
	rApp.HandleFunc("/", controllers.CheckAppUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions) // +
	rUiApi.HandleFunc("/", controllers.CheckUiApi).Methods(http.MethodGet, http.MethodPost, http.MethodOptions)

	// перемещаем точку монтирования для ui/api интерфейсов + отсекаем функции проверки роутов
	rApi = rApi.PathPrefix("").Subrouter()
	rApp = rApp.PathPrefix("").Subrouter()
	rUiApi = rUiApi.PathPrefix("/accounts/{accountId:[0-9]+}").Subrouter()

	// Дополнительные псевдо-точки для навешивания middleware
	rAppAuthUser := rApp.PathPrefix("").Subrouter()
	rAppAuthFull := rApp.PathPrefix("").Subrouter()
	rUiApiAuthFull := rUiApi.PathPrefix("").Subrouter()

	// Подключаем middleware
	rApi.Use(middleware.ApiEnabled, middleware.BearerAuthentication)

	rApp.Use(middleware.AppUiApiEnabled)
	rAppAuthUser.Use(middleware.JwtUserAuthentication)
	rAppAuthFull.Use(middleware.JwtFullAuthentication)


	rUiApi.Use(middleware.UiApiEnabled,middleware.CheckAccountId, middleware.CheckUiApiEnabled)
	//rUiApiAuthFull.Use(middleware.CorsAccessControl, middleware.CheckAccountId, middleware.JwtFullAuthentication)
	rUiApiAuthFull.Use(middleware.CheckAccountId, middleware.CheckUiApiEnabled)

	ApiRoutes(rApi)
	AppRoutes(rApp, rAppAuthUser, rAppAuthFull)
	UiApiRoutes(rUiApi, rUiApiAuthFull)

	rBase.NotFoundHandler = middleware.NotFoundHandler()

	return rBase
}


/**
* [/user] - группа роутов отвечающих за создание пользователей
 */
var UserRoutes = func (rBase, rUser, r_acc, r_full *mux.Router) {

	// hidden: code for auth: account_name: <>
	rBase.HandleFunc("/auth", controllers.UserAuthorization).Methods(http.MethodPost, http.MethodOptions)

	rBase.HandleFunc("", controllers.UserCreate).Methods(http.MethodPost, http.MethodOptions)
	rUser.HandleFunc("", controllers.UserGetProfile).Methods(http.MethodGet, http.MethodOptions)

	rBase.HandleFunc("/email-verification", controllers.UserEmailVerificationConfirm).Methods(http.MethodPost, http.MethodOptions)

	//rUser.HandleFunc("/email-verification", crm.UserSendEmailVerification).Methods(http.MethodGet, http.MethodOptions)
	rUser.HandleFunc("/email-verification/invite", controllers.UserSendEmailInviteVerification).Methods(http.MethodGet, http.MethodOptions)

	rBase.HandleFunc("/recovery/username", controllers.UserRecoveryUsername).Methods(http.MethodPost, http.MethodOptions)

	// ### Password -Routes- ###

	// Устанавливает новый пароль {"password":"...", "password_old":"..."}. Если password_reset = true, то старый пароль не требуется
	rUser.HandleFunc("/password", controllers.UserSetPassword).Methods(http.MethodPost, http.MethodOptions)

	// отправляет email с инструцией по сбросу пароля
	rBase.HandleFunc("/password/reset/send-email", controllers.UserRecoveryPasswordSendMail).Methods(http.MethodPost, http.MethodOptions)

	// одноразовая авторизация по {"token":"<...>"}: сбрасывает пароль и совершает авторизацию.
	rBase.HandleFunc("/password/reset/confirm", controllers.UserPasswordResetConfirm).Methods(http.MethodPost, http.MethodOptions)


	// отмена сброса пароля. Важная функция для антивзлома. Можно использовать тот же токен, но лучше создавать отдельный...
	//rBase.HandleFunc("/password/reset/report", crm.UserRecoveryPassword).Methods(http.MethodDelete, http.MethodOptions)

	// почему бы функции ниже не перенести в [AccountRoutes]?
	rUser.HandleFunc("/accounts", controllers.UserGetAccounts).Methods(http.MethodGet, http.MethodOptions)
	rUser.HandleFunc("/accounts/{account_id:[0-9]+}/auth", controllers.UserLoginInAccount).Methods(http.MethodGet, http.MethodOptions)

}

/**
* [/accounts] - группа роутов отвечающих за авторизацию и аутентификацию пользователя и аккаунта
 */
var AccountRoutes = func (rBase, rUser, r_acc, r_full *mux.Router) {

	rUser.HandleFunc("", controllers.AccountCreate).Methods(http.MethodPost, http.MethodOptions)
	r_acc.HandleFunc("", controllers.AccountGetProfile).Methods(http.MethodGet, http.MethodOptions)

}