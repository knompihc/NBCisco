/********************************************************************
 * FileName:     zone.go
 * Project:      Havells StreetComm
 * Module:       zone
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure
import (
	"net/http"
	"io"
	"strconv"
//"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strings"
)
func Addzone(w http.ResponseWriter, r *http.Request){
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
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows,err:=db.Query("Select id from zone where name='"+name+"'")
	defer rows.Close()
	chkErr(err,&w)
	for rows.Next(){
		logger.Println("Already Present!!")
		io.WriteString(w,"already")
		return
	}
	stmt, err := db.Prepare("INSERT zone SET name=?")
	defer stmt.Close()
	chkErr(err,&w)
	_, eorr:=stmt.Exec(name)
	chkErr(eorr,&w)
	if(eorr==nil){
		io.WriteString(w,"done")
	}
	defer stmt.Close()
}

func Zoneconfigure(w http.ResponseWriter, r *http.Request){
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	pid := r.URL.Query().Get("pid")
	var script string

	rows,err :=db.Query("Select id,name from zone limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)

	cn:=0
	script+="<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;' >Zone Name</th><th style='text-align: center;'>Zone ID</th><th style='text-align: center;'>View Added SGUS</th><th style='text-align: center;'>Add SGUS</th><th style='text-align: center;'>Remove SGUS</th></tr></thead><tbody>"
	for rows.Next(){
		var name string
		var zid int
		err=rows.Scan(&zid,&name)
		chkErr(err,&w)
		if cn<=9 {
			script += "<tr><td style='text-align: center;'>" + name + "</td><td style='text-align: center;'>" + strconv.Itoa(zid) + "</td><td style='text-align: center;'><a href='#' id='vc_" + strconv.Itoa(zid) + "' class='btn btn-info btn-sm viewc'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td style='text-align: center;'>  <a href='#' id='ac_" + strconv.Itoa(zid) + "' class='btn btn-success btn-sm addc'><span class='glyphicon glyphicon-check'></span> Add</a><div class='vic hidden' style='text-align: center;' id='vidc_" + strconv.Itoa(zid) + "'></td><td style='text-align: center;'>  <a href='#' id='rc_" + strconv.Itoa(zid) + "' class='btn btn-warning btn-sm removec'><i class='fa fa-trash-o'></i> Remove</a></td></tr>"
		}
		cn++
	}
	script+="</tbody></table>"
	if cn>10{
		script+="y"
	}
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zonesguview(w http.ResponseWriter, r *http.Request){
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
	id := r.URL.Query().Get("id")
	rows,err :=db.Query("select sgu.sgu_id,sgu.location_name from sgu inner join zone_sgu on zone_sgu.sguid=sgu.sgu_id where zone_sgu.zid='"+id+"'")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>Sgu ID</th><th style='text-align: center;'>Sgu Location</th></tr></thead><tbody>"
	for rows.Next(){
		var sguid int
		var loc string
		err=rows.Scan(&sguid,&loc)
		chkErr(err,&w)
			script+="<tr><td>"+strconv.Itoa(sguid)+"</td><td>"+loc+"</td></tr>"
	}
	script+="</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zoneadd(w http.ResponseWriter, r *http.Request){
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
	id := r.URL.Query().Get("id")
	rows,err :=db.Query("select sgu.sgu_id,sgu.location_name from sgu where sgu_id not in(select sguid from zone_sgu)")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Sgu ID</th><th style='text-align: center;'>Sgu Location</th></tr></thead><tbody>"
	for rows.Next(){
		var sguid int
		var loc string
		err=rows.Scan(&sguid,&loc)
		chkErr(err,&w)
			script+="<tr><td style='text-align:center''><input type='checkbox' name='addc_"+id+"' value='"+strconv.Itoa(sguid)+"'></td><td>"+strconv.Itoa(sguid)+"</td><td>"+loc+"</td></tr>"
	}
	script+="<tr><td colspan='3' style='text-align:center'><a href='#' id='sc_"+id+"'class='btn btn-info btn-sm savec'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Zonesguremove(w http.ResponseWriter, r *http.Request){
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
	id := r.URL.Query().Get("id")
	rows,err :=db.Query("select sgu.sgu_id,sgu.location_name from sgu where sgu_id in (select sguid from zone_sgu where zid='"+id+"')")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Sgu ID</th><th style='text-align: center;'>Sgu Location</th></tr></thead><tbody>"
	for rows.Next(){
		var sguid int
		var loc string
		err=rows.Scan(&sguid,&loc)
		chkErr(err,&w)
		script+="<tr><td style='text-align:center''><input type='checkbox' name='removec_"+id+"' value='"+strconv.Itoa(sguid)+"'></td><td>"+strconv.Itoa(sguid)+"</td><td>"+loc+"</td></tr>"
	}
	script+="<tr><td colspan='3' style='text-align:center'><a href='#' id='re_"+id+"'class='btn btn-info btn-sm saverc'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zonesgusave(w http.ResponseWriter, r *http.Request)  {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	sids := r.URL.Query().Get("ids")
	zid := r.URL.Query().Get("zid")
	ids :=strings.Split(sids,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt, err := db.Prepare("INSERT zone_sgu SET zid=?,sguid=?")
	defer stmt.Close()
	chkErr(err,&w)
	for _,val:=range ids{
		_, eorr:=stmt.Exec(zid,val)
		chkErr(eorr,&w)

	}
	io.WriteString(w, "Saved Successfully!!")
}
func Zonesgusaver(w http.ResponseWriter, r *http.Request)  {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	sids := r.URL.Query().Get("ids")
	zid := r.URL.Query().Get("zid")
	ids :=strings.Split(sids,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt, err := db.Prepare("Delete from zone_sgu where zid=? and sguid=?")
	defer stmt.Close()
	chkErr(err,&w)
	for _,val:=range ids{
		_, eorr:=stmt.Exec(zid,val)
		chkErr(eorr,&w)

	}
	io.WriteString(w, "Saved Successfully!!")
}

func Removezone(w http.ResponseWriter, r *http.Request)  {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	zid := r.URL.Query().Get("id")
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt2, err2 := db.Query("select id from zone where id='"+zid+"'")
	defer stmt2.Close()
	chkErr(err2,&w)
	if ! (stmt2.Next()){
		io.WriteString(w, "Please enter correct Zone ID!!")
		return
	}
	stmt, err := db.Prepare("Delete from zone_sgu where zid=?")
	defer stmt.Close()
	if err!=nil{
		io.WriteString(w, "Error!!")
		return
	}else {
		_, eorr:=stmt.Exec(zid)
		if eorr!=nil{
			io.WriteString(w, "Error!!")
			return
		}else{
			stmt1, err := db.Prepare("Delete from zone where id=?")
			defer stmt1.Close()
			_,err=stmt1.Exec(zid)
			if err!=nil{
				io.WriteString(w, "Error!!")
			}else{
				io.WriteString(w, "done")
			}
		}
	}


}

func Updatezone(w http.ResponseWriter, r *http.Request){
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}
	id:=r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows,err:=db.Query("Select id from zone where name='"+name+"'")
	defer rows.Close()
	chkErr(err,&w)
	for rows.Next(){
		logger.Println("Already Present!!")
		io.WriteString(w,"already")
		return
	}
	stmt, err := db.Prepare("Update zone SET name=? where id='"+id+"'")
	defer stmt.Close()
	chkErr(err,&w)
	_, eorr:=stmt.Exec(name)
	chkErr(eorr,&w)
	if(eorr==nil){
		io.WriteString(w,"done")
	}
	defer stmt.Close()
}