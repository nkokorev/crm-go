# crm-go
Ядро Ratus CRM, включая GUI/API интерфейсы. 

All parameters use lowerCamelCase style:
```json
{"username": "", "mobilePhone": ""}
```

## Methods UI-API | Interface

All methods call by scheme: `<api-url>/<method-scheme>`

- app-url: `http://ratuscrm.com/api/`
- public-url: `http://ui.api.ratuscrm.com`

## Account interfaces

Schema of account data:

| Json  | Type | Required | Default |Description |
| --- | :---: | :---: |:---: | --- |
| `id`  | uint | true | - | Уникальный идентификатор учетной записи аккаунта |
| `name`  | string | true | - | Имя аккаунта, видно другим пользователям |
| `website`  | string | no | - | Основной вебсайт компании |
| `type`  | string | no | - | Основной вебсайт компании |
| `uiApiEnabled` | bool | yes | false | Включен ли ui-api для аккаунта |
 


## User interfaces

Пользователь в системе идентифицируется по:
 - ID аккаунта, через которого пользователь был зарегистрирован
 - именю учетной записи / email'у / мобильному телефону

Вы можете выбрать доступные варианты авторизации в настройках аккаунта.

### CreateUser

This method create user account in your account of RatusCRM. 

see also: *CreateOrUpdate*

[POST] `/accounts/{account_id}/users`

| Parameters  | Type | Required | Description |
| --- | :---: | :---: | --- |
| `username`  | string  | no | Имя учетной записи пользователя |
| `email`  | string  | no | Контактный email для системных уведомлений | 
| `mobilePhone`  | string  | no | Мобильный телефон для SMS-уведомлений |
| `password`  | string  | no | Минимум одна цифра, строчная, прописная буква и спецсимвол, мин. 8 символов. |
| `name`  | string  | no | Имя пользователя |
| `surname`  | string  | no | Фамилия пользователя |
| `patronymic`  | string  | no | Отчество пользователя |


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

We are recommended #1. If you whant hidden user's email - choose #2.

Strong reccomended not change this.

### Public api (Bearer Authentication)
url: `http://api.ratuscrm.com`.<br>
Create api-token & set role in your account.

# 1. UI API

- APP: `http://ratuscrm.com/api/ui/`
- Public: `http://ui.api.ratuscrm.com`

Методы:
- CreateUser
- DeleteUser
- AuthUser

### Create user
`[POST] http://<api:schema>/accounts/{account_id}/users`

Parametrs:
- [required, string] username
- [required, string] email
- [required, string] password
- [required, string] phone
- [required, string] name
- [required, string] soname
- ...

### 

Внутренее api для графического интерфейса (vue-cli):
http://ratuscrm.com/ui-api/


**Микроконтроллеры должны использовать функции в контексте аккаунта**
