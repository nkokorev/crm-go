# crm-go
Ядро Ratus CRM, включая GUI/API интерфейсы. 


## API - интерфейсы

### App/Public ui-api
app-url: `http://ratuscrm.com/api/`
public-url: `http://ui.api.ratuscrm.com`

**Auth in ui-interface**

You must choose auth settings: 
1. auth by email & pwd (default)
2. auth by username & pwd
3. auth by phone & once code*

We are recommended #1. If you whant hidden user's email - choose #2.

Strong reccomended not change this.

### Methods ui-api

**Create user**

This method create user account in your account of RatusCRM.

see also: *CreateOrUpdate*

**Schema url**

`[POST] <api-url>/accounts/{account_id}/users`

<table>
<tr>
<th>parametr</th>
<th>type</th>
<th>required</th>
</tr>
<tr>
<td>username</td>
<td>string</td>
<td>if auth</td>
</tr>
<tr>
<td>email</td>
<td>string</td>
<td>if auth</td>
</tr>
<tr>
<td>phone</td>
<td>string</td>
<td>no</td>
</tr>
<tr>
<td>password</td>
<td>string</td>
<td>no</td>
</tr>
<tr>
<td>name</td>
<td>string</td>
<td>no</td>
</tr>
<tr>
<td>soname</td>
<td>string</td>
<td>no</td>
</tr>
</table>


Attention: 
- one of {username,email,phone} must be not null.
- username, email, phone must be unique | account.
- if username not null, email not null too.
- username required email

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
