/********************************************************************
 * FileName:     graph.go
 * Project:      Havells StreetComm
 * Module:       graph
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
	//"fmt"
)
func Graph(w http.ResponseWriter, r *http.Request) {
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
	rows,err :=db.Query("Select sgu_id,timestamp,Vr,Vy,Vb,Ir,Iy,Ib,Pf,KW,KVA,KWH,KVAH,rKVAH,Run_Hours,freq from parameters a where id=(select id from parameters b where b.sgu_id=a.sgu_id order by timestamp desc limit 1) limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)

	cn:=0
	script+="<table class='table table-bordered table-hover'><thead><tr><th>Select</th><th>SGU ID</th><th>Timestamp</th><th>Vr</th><th>Vy</th><th>Vb</th><th>Ir</th><th>Iy</th><th>Ib</th><th>Pf</th><th>KW</th><th>KVA</th><th>KWH</th><th>KVAH</th><th>rKVAH</th><th>Run Hours</th><th>Freq</th></tr></thead><tbody>"
	for rows.Next(){
		var sgid int
		var ti string
		var Vr,Vy,Vb,Ir,Iy,Ib,Pf,KW,KVA,KWH,KVAH,rKVAH,Run_Hours,freq string
		err=rows.Scan(&sgid,&ti,&Vr,&Vy,&Vb,&Ir,&Iy,&Ib,&Pf,&KW,&KVA,&KWH,&KVAH,&rKVAH,&Run_Hours,&freq)
		chkErr(err,&w)
		if cn<=9 {
			if cn==0{
				script += "<tr><td><input type='radio' name='plot' checked value='" + strconv.Itoa(sgid) + "'></td><td>"+ strconv.Itoa(sgid) + "</td><td>" +ti+"</td><td>"+ Vr + "</td><td>" + Vy +"</td><td>"+Vb+"</td><td>"+ Ir +"</td><td>"+ Iy +"</td><td>"+ Ib +"</td><td>"+  Pf +"</td><td>"+ KW +"</td><td>"+ KVA +"</td><td>"+ KWH +"</td><td>"+KVAH+"</td><td>"+rKVAH+"</td><td>"+Run_Hours+"</td><td>"+freq+"</td></tr>"
			}else{
				script += "<tr><td><input type='radio' name='plot' value='" + strconv.Itoa(sgid) + "'></td><td>"+ strconv.Itoa(sgid) + "</td><td>" +ti+"</td><td>"+ Vr + "</td><td>" + Vy +"</td><td>"+Vb+"</td><td>"+ Ir +"</td><td>"+ Iy +"</td><td>"+ Ib +"</td><td>"+  Pf +"</td><td>"+ KW +"</td><td>"+ KVA +"</td><td>"+ KWH +"</td><td>"+KVAH+"</td><td>"+rKVAH+"</td><td>"+Run_Hours+"</td><td>"+freq+"</td></tr>"
			}
		}
		cn++
	}
	script+="</tbody></table>"
	if cn>10{
		script+="y"
	}
	io.WriteString(w, script)
}