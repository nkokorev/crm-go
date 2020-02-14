# crm-go
Ядро Ratus CRM, включая GUI/API интерфейсы. 

All parameters use lowerCamelCase style:
```json
{"username": "", "mobilePhone": ""}
```

## API Interface

|  | ignored | required | Auth type | description |
| --- | :---: |:---: | :---: | --- |
| ratuscrm.com/api | `uiApiEnabled` | `appApiEnabled` | JWT (AES) | JSON UI-API for app.ratuscrm.com |
| ui.api.ratuscrm.com |  | `uiApiEnabled` | JWT (AES) | JSON UI-API for company websites |
| api.ratuscrm.com |  | `apiEnabled` | Bearer token |Standard Rest JSON API   |


## CRM Settings

| Json name | Type | Default |Description |
| --- | :---: |:---: | --- |
| `apiEnabled` | bool | true | Принимать запросы по API |
| `appUiApiEnabled` | bool | true | Принимать запросы по APP UI-API |
| `uiApiEnabled` | bool | true | Принимтаь ли запросы по публичному UI-API |
| `apiDisableMessage` | string | "API is unavailable..." | Ответ при отключенном API |
| `uiApiDisabledMessage` | string | "UI-API is unavailable..." | Ответ при отключенном публичном UI-API | 
| `appUiApiDisableMessage` | string | "Из-за работ на сервере..." | Ответ при отключенном APP UI-API | 

При отключенном APP UI-API GUI должен выводить не предложение логина, а специальную заставку.

## Account interfaces

Переменные с приставкой `uiApi` применяются для работы через UI/API.

DB Schema of account data:

| Json name | Type | Default |Description |
| --- | :---: |:---: | --- |
| `id`  | uint | `gen` | Уникальный ID аккаунта |
| `name`  | string | - | Имя аккаунта, виден другим пользователям |
| `website`  | string | - | Основной вебсайт компании |
| `type`  | string | - | Основной вебсайт компании |
| `apiEnabled` | bool | true | Принимать ли запросы через API |
| `uiApiEnabled` | bool | false | Принимать ли запросы через публичный UI-API |
| `uiApiAesEnabled` | bool | true | Включение AES-128/CFB шифрования |
| `uiApiAesKey` | string | `gen` | 16 символный UTF-8 ключ шифрования AES-128 |
| `uiApiJwtKey` | string | `gen` | 32 символный UTF-8 Ключ подписи JWT/HS256 |
| `uiApiEnabledUserRegistration` | bool | false | Разрешить регистрацию новых пользователей |
| `uiApiUserRegistrationInvitationOnly` | bool | false | Регистрация только по персональным приглашеним |
| `uiApiAuthMethods` | array | `{email}` | Доступные способы авторизации | 
| `uiApiUserRegistrationRequiredFields` | array | `{email}` | Список полей обязательных при регистрации | 
| `uiApiUserEmailDeepValidation` | bool | false | Запрос на сервер с проверкой email-адреса | 

## User interfaces

Пользователь в системе идентифицируется по:
 - ID аккаунта, через которого пользователь был зарегистрирован
 - именю учетной записи / email'у / мобильному телефону (одно из всегда есть)

Вы можете настроить обязательные поля при регистрации нового пользователя через UI/API (имя, телефон и т.д.).
Вы можете выбрать доступные варианты авторизации в настройках аккаунта.

### CreateUser

This method create user account in your account of RatusCRM. 

see also: *CreateOrUpdate*

[POST] `/accounts/{account_id}/users`

| Parameters  | Type | Default | Description |
| --- | :---: | :---: | --- |
| `username`  | string  | null | Имя учетной записи пользователя |
| `email`  | string  | null | Контактный email для системных уведомлений | 
| `phone`  | string  | null | Мобильный телефон для SMS-уведомлений указывается в международном формат + или дополнительное поле региона|
| `phoneRegion`  | string  | RU | Регион мобильного оператора (US,RU,KZ ...)|
| `password`  | string  | null | Минимум одна цифра, строчная, прописная буква и спецсимвол, мин. 8 символов. |
| `name`  | string  | null | Имя пользователя |
| `surname`  | string  | null | Фамилия пользователя |
| `patronymic`  | string  | null | Отчество пользователя |


Attention: 
- one of {username,email,phone} must be not null.
- username, email, phone must be unique | account.
- if username not null, email not null too.
- username required email

### AuthUser

You must choose auth settings: 
1. auth by email & pwd (default)
2. auth by username & pwd
3. auth by phone & once code*

We are recommended #1. If you want hidden user's email - choose #2.
