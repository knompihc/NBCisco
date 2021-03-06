/********************************************************************
 * FileName:     subscribe.go
 * Project:      Havells StreetComm
 * Module:       subscribe
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure
import (
	"net/http"
	"io"
	//"fmt"
	_ "github.com/go-sql-driver/mysql"
)
func Subscribe(w http.ResponseWriter, r *http.Request){
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	mobile := "91"+r.URL.Query().Get("mobile")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows,err:=db.Query("Select id from admin where mobile_num='"+mobile+"'")
	defer rows.Close()
	chkErr(err,&w)
	for rows.Next(){
		logger.Println("Already subscribed!!")
		io.WriteString(w,"already")
		return
	}
	stmt, err := db.Prepare("INSERT admin SET name=?,email_id=?,mobile_num=?")
	defer stmt.Close()
	chkErr(err,&w)
	_, eorr:=stmt.Exec(name,email,mobile)
	chkErr(eorr,&w)
	if(eorr==nil){
		io.WriteString(w,"done")
	}
	defer stmt.Close()
}