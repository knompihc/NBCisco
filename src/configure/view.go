/********************************************************************
 * FileName:     view.go
 * Project:      Havells StreetComm
 * Module:       view
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure
import (
	"net/http"
	"io"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	//"fmt"
)
func View(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	var script string
	pid := r.URL.Query().Get("pid")
	rows,err :=db.Query("Select idschedule,ScheduleStartTime,ScheduleEndTime,pwm from schedule limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)

	cn:=0
	script+="<table class='table table-bordered table-hover'><thead><tr><th>Schedule ID</th><th>Start Date</th><th>End Date</th><th>Start Time</th><th>End Time</th><th>Brightness Level</th></tr></thead><tbody>"
	for rows.Next(){
		var shid,pwm int
		var from,to string
		err=rows.Scan(&shid,&from,&to,&pwm)
		fd:=strings.Split(from, " ")
		td:=strings.Split(to, " ")
		chkErr(err,&w)
		if cn<=9 {
			script += "<tr><td>" + strconv.Itoa(shid) + "</td><td>" + fd[0] + "</td><td>" + td[0] +"</td><td>" + fd[1] +"</td><td>" + td[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}
		cn++;
	}
	script+="</tbody></table>"
	if cn>10{
		script+="y"
	}
	io.WriteString(w, script)
}