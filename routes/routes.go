package routes

import (
	"github.com/gorilla/mux"
	"github.com/nkokorev/crm-go/middleware"
	"os"
)

func Handlers() *mux.Router {

	var crmHost string

	// root route - handle all request
	rBase := mux.NewRouter().StrictSlash(true)

	// Environment variable: local, public etc..
	AppEnv := os.Getenv("APP_ENV")

	// Set AppEnv variable
	switch AppEnv {
	case "local":
		crmHost = "crm.local"
		//crmHost = "127.0.0.1:8090"
		//crmHost = "localhost:8090"
	case "public":
		crmHost = "ratuscrm.com"
	default:
		crmHost = "ratuscrm.com"
	}

	// Set CORS politic for local development
	if AppEnv == "local" {
		rBase.Use(middleware.CorsAccessControl)
	}

	// Mount all root point of routes
	rApi := rBase.Host("api." + crmHost).Subrouter()                                                        // API [api.ratuscrm.com]
	rApp := rBase.Host("app." + crmHost).PathPrefix("/ui-api").Subrouter()                                  // APP [app.ratuscrm.com/ui-api]
	rUiApi := rBase.Host("ui.api." + crmHost).PathPrefix("/accounts/{accountHashId:[a-z0-9]+}").Subrouter() // UI/API [ui.api.ratuscrm.com]

	// ### Перемещаем точку монтирования для ui/api интерфейсов + отсекаем функции проверки роутов ###
	//rUiApi = rUiApi.PathPrefix("/accounts/{accountHashId:[a-z0-9]+}").Subrouter()

	/*
		Important! Use ApiHash AES or JWT key:
					1. API have't (clearly)
					2. App use only RatusCRM AES/JWT key
					3. UI/API use special AES/JWT key for him account (by accountId)

		JWT/AES key:
					issuerAccountId - id аккаунта, который выдает авторизацию. В App UI RatusCRM всегда = 1.
					accountId 		- id аккаунта, в котором пользователь работает (в UI/API совпадает с issuerAccountId).
					userId 			- id авторизованного пользователя.

		  			issuerAccount	- аккаунт откуда пришел запрос
					account			- аккаунт, в котором авторизован пользователь
					user			- пользователь, прошедший авторизацию

		authAccount - авторизующий аккаунт (используются его AES/JWT ключи).
		Context: [issuer]: 'app' / 'api' / 'ui-api'.

		# APP
			url: app.ratuscrm.com/ui-api
			auth: Bearer token

		# UI/API
			url: ui.api.ratuscrm.com.
			auth: Bearer token

		# API
			url: api.ratuscrm.com
			auth: md5 hash


	*/

	/**
		1. The request can have 4 statuses: guest, authUser, authAccount, fullAuth (user + account)
		2. API has only authAccount
		3. APP has 3 status: guest, authUser and fullAuth
		4. UI/API has 2 status: guest, fullAuth.

	 	Context statuses:
			1. guest - has't userId or accountId
			2. authUser: have userId & user (type of models.User)
			3. authAccount: have accountId & account (type of models.Account)
			4. fullAuth: have userId, accountId, user, account

		Important notice: about use hash AES or JWT key:
			1. API has't (clearly)
			2. App use only RatusCRM AES/JWT key
			3. UI/API use special AES/JWT key for him account (by accountId)
	*/

	// New route point for using middleware
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

	  	issuer			- откуда пришел запрос: "app", "ui-api", "api"
		issuerAccountId - id root аккаунта, от имени которого происходит запрос. В App RatusCRM всегда = 1.
		accountId 		- id аккаунта, в котором пользователь авторизован. Равен 0, если не авторизован. В UI/API совпадает с issuerAccountId.
		userId 			- id авторизованного пользователя. Равен 0, если не авторизован.

	  	issuerAccount	- аккаунт откуда пришел запрос
		account			- аккаунт, в котором авторизован пользователь
		user			- пользователь, прошедший авторизацию
	*/

	// Все запросы API имеют в контексте accountId, account. У них нет userId.
	rApi.Use(middleware.CorsAccessControl, middleware.CheckApiStatus, middleware.BearerAuthentication)

	// Все запросы App UI/API имеют в контексте accountId, account. Могут иметь userId и userId + accountId + account
	rApp.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount)
	rAppAuthUser.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtUserAuthentication)
	rAppAuthFull.Use(middleware.CheckAppUiApiStatus, middleware.AddContextMainAccount, middleware.JwtFullAuthentication)

	// Через UI/API запросы всегда идут в контексте аккаунта
	rUiApi.Use(middleware.CorsAccessControl, middleware.CheckUiApiStatus, middleware.ContextMuxVarAccountHashId)
	rUiApiAuthFull.Use(middleware.CorsAccessControl, middleware.CheckUiApiStatus, middleware.ContextMuxVarAccountHashId, middleware.JwtFullAuthentication) // set userId,accountId,account

	// ### Передаем запросы в обработку ###
	ApiRoutes(rApi)
	AppRoutes(rApp, rAppAuthUser, rAppAuthFull)
	UiApiRoutes(rUiApi, rUiApiAuthFull)

	rBase.NotFoundHandler = middleware.NotFoundHandler()

	return rBase
}
