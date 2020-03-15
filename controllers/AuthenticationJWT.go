package controllers

import (
	"fmt"
	u "github.com/nkokorev/crm-go/utils"
	"net/http"
)

// под защитой jwt-middleware
func AuthenticationJWTCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Println("AuthenticationJWTCheck!")
	resp := u.Message(true, "Auth token is valid")
	u.Respond(w, resp)
}
