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

	// обрабатываем все запросы со слешем и без
	rBase := mux.NewRouter().StrictSlash(true)

	switch os.Getenv("APP_ENV") {
	case "local":
		crmHost = "crm-local.me"
		rBase.Use(middleware.CorsAccessControl)
		//crmHost = "localhost:8090"
	case "public":
		crmHost = "ratuscrm.com"
	default:
		crmHost = "ratuscrm.com"
	}



	// ### Монтируем все три точки входа для API ###
	rApi := rBase.Host("api." + crmHost).Subrouter() // api.ratuscrm.com
	rApp := rBase.Host("app." + crmHost).PathPrefix("/ui-api").Subrouter() // app.ratuscrm.com/ui-api
	rUiApi := rBase.Host("ui.api." + crmHost).Subrouter() // ui.api.ratuscrm.com


	// ### Перемещаем точку монтирования для ui/api интерфейсов + отсекаем функции проверки роутов ###
	rUiApi = rUiApi.PathPrefix("/accounts/{accountHashId:[a-z0-9]+}").Subrouter()

	// Дополнительные псевдо-точки для навешивания middleware
	rAppAuthUser := rApp.PathPrefix("").Subrouter()
	rAppAuthFull := rApp.PathPrefix("").Subrouter()
	rUiApiAuthFull := rUiApi.PathPrefix("").Subrouter()
	// #### Подключаем Middleware ####

/**

	## Посредники проверяющие флаги настройки системы
	middleware.CheckApiStatus - проверяет статус API для всех аккаунтов
	middleware.CheckAppUiApiStatus - проверяет статус App UI/API в настройках GUI RatusCRM
	middleware.CheckUiApiStatus - проверяет статус Public UI/API для всех аккаунтов

	## Посредники определяющие signedAccount
	middleware.ContextMuxVarAccount - Вставляет в контекст issuerAccountId из hashAccountId (раскрытие issuer account) && issuerAccount
	middleware.ContextMainAccount - устанавливает в контекст issuerAccountId = 1 && issuerAccount

	## Посредники работающие с авторизацией пользователя
  	middleware.BearerAuthentication - читает с проверкой JWT, проверяет статус API в аккаунте. Дополняет контекст accountId && account
	middleware.JwtUserAuthentication - проверяет JWT и устанавливает в контекст userId & user
	middleware.JwtFullAuthentication - проверяет JWT и устанавливает в контекст userId & user, accountId && account

	--------------------------


	Посредники добавляют в контексте следующую информацию:

  	issuer			- откуда пришел запрос: "app ui/api", "ui-api", "api"
	issuerAccountId - id root аккаунта, от имени которого происходит запрос. В App RatusCRM всегда = 1.
	accountId 		- id аккаунта, в котором пользователь авторизован. Равен 0, если не авторизован. В UI/API совпадает с issuerAccountId.
	userId 			- id авторизованного пользователя. Равен 0, если не авторизован.

  	issuerAccount	- аккаунт откуда пришел запрос
	account			- аккаунт, в котором авторизован пользователь
	user			- пользователь, прошедший авторизацию
*/

	// Все запросы API имеют в контексте accountId, account. У них нет userId.
	rApi.Use(			middleware.CheckApiStatus, 			middleware.BearerAuthentication)

	// Все запросы App UI/API имеют в контексте accountId, account. Могут иметь userId и userId + accountId + account
	rApp.Use(			middleware.CheckAppUiApiStatus, middleware.ContextMainAccount)
	rAppAuthUser.Use(	middleware.CheckAppUiApiStatus, middleware.ContextMainAccount, middleware.JwtUserAuthentication) // set userId
	rAppAuthFull.Use(	middleware.CheckAppUiApiStatus,	middleware.JwtFullAuthentication) // set userId,accountId,account

	// Через UI/API запросы всегда идут в контексте аккаунта
	rUiApi.Use(			middleware.CheckUiApiStatus,	middleware.ContextMuxVarAccount)
	rUiApiAuthFull.Use(	middleware.CheckUiApiStatus, 	middleware.ContextMuxVarAccount, middleware.JwtFullAuthentication) // set userId,accountId,account


	// ### Передаем запросы в обработку ###
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

	// Три способа регистрации
	rBase.HandleFunc("sign-up", controllers.UserSignUp).Methods(http.MethodPost, http.MethodOptions) // проверяет UiApiUserRegistrationRequiredFields

	rBase.HandleFunc("", controllers.UserRegistration).Methods(http.MethodPost, http.MethodOptions) // deprecated
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