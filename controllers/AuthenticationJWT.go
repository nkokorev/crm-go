package controllers

import (
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// под защитой jwt-middleware
func AuthenticationJWTCheck(w http.ResponseWriter, r *http.Request) {
	resp := u.Message(true, "Auth token is valid")
	u.Respond(w, resp)
}
