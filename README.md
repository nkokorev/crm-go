# crm-go
Ядро Ratus CRM, включая GUI/API интерфейсы. 


## API - интерфейсы

### App ui-api
URL: `http://ratuscrm.com/api/`

### Public ui-api
URL: `http://ui.api.ratuscrm.com`

### Public api (Bearer Authentication)
URL: `http://api.ratuscrm.com`

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