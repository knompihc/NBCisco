/********************************************************************
 * FileName:     group.go
 * Project:      Havells StreetComm
 * Module:       group
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"userMgmt"
)

func Addgroup(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	name := r.URL.Query().Get("name")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select id from groupscu where name='" + name + "'")
	defer rows.Close()
	chkErr(err, &w)
	for rows.Next() {
		logger.Println("Already Present!!")
		io.WriteString(w, "already")
		return
	}
	stmt, err := db.Prepare("INSERT groupscu SET name=?")
	defer stmt.Close()
	chkErr(err, &w)
	_, eorr := stmt.Exec(name)
	chkErr(eorr, &w)
	if eorr == nil {
		io.WriteString(w, "done")
	}
	defer stmt.Close()
}

func Groupconfigure(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	pid := r.URL.Query().Get("pid")
	var script string

	rows, err := db.Query("Select id,name from groupscu limit " + (pid) + ",11")
	defer rows.Close()
	chkErr(err, &w)

	cn := 0
	script += "<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;' >Group Name</th><th style='text-align: center;'>Group ID</th><th style='text-align: center;'>View Added SCUS</th><th style='text-align: center;'>Add SCUS</th><th style='text-align: center;'>Remove SCUS</th></tr></thead><tbody>"
	for rows.Next() {
		var name string
		var gid int
		err = rows.Scan(&gid, &name)
		chkErr(err, &w)
		if cn <= 9 {
			script += "<tr><td style='text-align: center;'>" + name + "</td><td style='text-align: center;'>" + strconv.Itoa(gid) + "</td><td style='text-align: center;'><a href='#' id='vc_" + strconv.Itoa(gid) + "' class='btn btn-info btn-sm viewc'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td style='text-align: center;'>  <a href='#' id='ac_" + strconv.Itoa(gid) + "' class='btn btn-success btn-sm addc'><span class='glyphicon glyphicon-check'></span> Add</a><div class='vic hidden' style='text-align: center;' id='vidc_" + strconv.Itoa(gid) + "'></td><td style='text-align: center;'>  <a href='#' id='rc_" + strconv.Itoa(gid) + "' class='btn btn-warning btn-sm removec'><i class='fa fa-trash-o'></i> Remove</a></td></tr>"
		}
		cn++
	}
	script += "</tbody></table>"
	if cn > 10 {
		script += "y"
	}
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Groupscuview(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	id := r.URL.Query().Get("id")
	rows, err := db.Query("select scu.scu_id,scu.location_name from scu inner join group_scu_rel on group_scu_rel.scuid=scu.scu_id where group_scu_rel.gid='" + id + "'")
	defer rows.Close()
	chkErr(err, &w)
	var script string
	script = "<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>Scu ID</th><th style='text-align: center;'>Scu Location</th></tr></thead><tbody>"
	for rows.Next() {
		var sguid int
		var loc string
		err = rows.Scan(&sguid, &loc)
		chkErr(err, &w)
		script += "<tr><td>" + strconv.Itoa(sguid) + "</td><td>" + loc + "</td></tr>"
	}
	script += "</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Groupadd(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	id := r.URL.Query().Get("id")
	rows, err := db.Query("select scu.scu_id,scu.location_name from scu where scu_id not in(select scuid from group_scu_rel)")
	defer rows.Close()
	chkErr(err, &w)
	var script string
	script = "<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scu ID</th><th style='text-align: center;'>Scu Location</th></tr></thead><tbody>"
	for rows.Next() {
		var sguid int
		var loc string
		err = rows.Scan(&sguid, &loc)
		chkErr(err, &w)
		script += "<tr><td style='text-align:center''><input type='checkbox' name='addc_" + id + "' value='" + strconv.Itoa(sguid) + "'></td><td>" + strconv.Itoa(sguid) + "</td><td>" + loc + "</td></tr>"
	}
	script += "<tr><td colspan='3' style='text-align:center'><a href='#' id='sc_" + id + "'class='btn btn-info btn-sm savec'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Groupscuremove(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	id := r.URL.Query().Get("id")
	rows, err := db.Query("select scu.scu_id,scu.location_name from scu where scu_id in (select scuid from group_scu_rel where gid='" + id + "')")
	defer rows.Close()
	chkErr(err, &w)
	var script string
	script = "<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scu ID</th><th style='text-align: center;'>Scu Location</th></tr></thead><tbody>"
	for rows.Next() {
		var sguid int
		var loc string
		err = rows.Scan(&sguid, &loc)
		chkErr(err, &w)
		script += "<tr><td style='text-align:center''><input type='checkbox' name='removec_" + id + "' value='" + strconv.Itoa(sguid) + "'></td><td>" + strconv.Itoa(sguid) + "</td><td>" + loc + "</td></tr>"
	}
	script += "<tr><td colspan='3' style='text-align:center'><a href='#' id='re_" + id + "'class='btn btn-info btn-sm saverc'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Groupscusave(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	sids := r.URL.Query().Get("ids")
	gid := r.URL.Query().Get("gid")
	ids := strings.Split(sids, ",")
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt, err := db.Prepare("INSERT group_scu_rel SET gid=?,scuid=?")
	defer stmt.Close()
	chkErr(err, &w)
	for _, val := range ids {
		_, eorr := stmt.Exec(gid, val)
		chkErr(eorr, &w)

	}
	io.WriteString(w, "Saved Successfully!!")
}
func Groupscusaver(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	sids := r.URL.Query().Get("ids")
	gid := r.URL.Query().Get("gid")
	ids := strings.Split(sids, ",")
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt, err := db.Prepare("Delete from group_scu_rel where gid=? and scuid=?")
	defer stmt.Close()
	chkErr(err, &w)
	for _, val := range ids {
		_, eorr := stmt.Exec(gid, val)
		chkErr(eorr, &w)

	}
	io.WriteString(w, "Saved Successfully!!")
}

func Removegroup(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	gid := r.URL.Query().Get("id")
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	stmt2, err2 := db.Query("select id from groupscu where id='" + gid + "'")
	defer stmt2.Close()
	chkErr(err2, &w)
	if !(stmt2.Next()) {
		io.WriteString(w, "Please enter correct Group ID!!")
		return
	}
	stmt, err := db.Prepare("Delete from group_sgu where gid=?")
	defer stmt.Close()
	if err != nil {
		io.WriteString(w, "Error!!")
		return
	} else {
		_, eorr := stmt.Exec(gid)
		if eorr != nil {
			io.WriteString(w, "Error!!")
			return
		} else {
			stmt1, err := db.Prepare("Delete from groupscu where id=?")
			defer stmt1.Close()
			_, err = stmt1.Exec(gid)
			if err != nil {
				io.WriteString(w, "Error!!")
			} else {
				io.WriteString(w, "done")
			}
		}
	}

}

func Updategroup(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select id from groupscu where name='" + name + "'")
	defer rows.Close()
	chkErr(err, &w)
	for rows.Next() {
		logger.Println("Already Present!!")
		io.WriteString(w, "already")
		return
	}
	stmt, err := db.Prepare("Update groupscu SET name=? where id='" + id + "'")
	defer stmt.Close()
	chkErr(err, &w)
	_, eorr := stmt.Exec(name)
	chkErr(eorr, &w)
	if eorr == nil {
		io.WriteString(w, "done")
	}
	defer stmt.Close()
}
