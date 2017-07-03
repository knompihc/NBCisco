/********************************************************************
 * FileName:     sessionManagement.go
 * Project:      Havells StreetComm
 * Module:       sguUtils
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/

//Session Management of user is handled here.

package userMgmt

import (
	"log"
	"net/http"

	"github.com/sessions"
)

var (
	store  = sessions.NewCookieStore([]byte("something-very-secret"))
	logger *log.Logger
)

//Checks validity of user session currently logged in.
func IsSessionValid(w http.ResponseWriter, r *http.Request) bool {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return false
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return false
	}
	return true
}
