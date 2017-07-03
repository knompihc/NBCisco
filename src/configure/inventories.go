/********************************************************************
 * FileName:     inventories.go
 * Project:      Havells StreetComm
 * Module:       inventories
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/

package configure

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"userMgmt"
)

func AddInventory(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	inventorytype := r.URL.Query().Get("name")
	desc := r.URL.Query().Get("description")

	logger.Println("adding inventories", inventorytype, desc)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	var invenType, invenDesc string
	//for database connectivity.
	stmt, err := db.Query("select AssetType,Description from inventory")
	chkErr(err, &w)
	defer stmt.Close()

	logger.Println("Feching inventories")
	for stmt.Next() {
		err := stmt.Scan(&invenType, &invenDesc)
		if err != nil {
			logger.Println(err)
			return
		}
		if (invenType == inventorytype) && (invenDesc == desc) {
			io.WriteString(w, "Inventory Already Exists!!")
			return
		}
	}
	stmt2, err := db.Prepare("INSERT inventory SET AssetType=?,Description=?")
	chkErr(err, &w)
	defer stmt2.Close()

	_, eorr := stmt2.Exec(inventorytype, desc)
	chkErr(eorr, &w)
	if eorr == nil {
		io.WriteString(w, "DataSaved Successfully")
	}
}

func Viewinventories(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	pid := r.URL.Query().Get("pid")
	var script string
	rows, err := db.Query("SELECT id,AssetType,Description,Quantity FROM inventory limit " + pid + ",6")
	chkErr(err, &w)
	defer rows.Close()

	cn := 0
	script += "<table id='inven' class='table table-bordered table-hover'><thead><tr><th style='text-align:center'>" +
		"Inventory Type</th><th style='text-align:center'>Description</th><th style='text-align:center;background-color:" +
		" lightyellow;'>Quantity</th><th style='text-align:center'>Edit</th></tr></thead><tbody>"
	for rows.Next() {
		var inventype, invendesc, id1 string
		var qtys int
		err = rows.Scan(&id1, &inventype, &invendesc, &qtys)
		chkErr(err, &w)
		if cn <= 4 {
			script += "<tr id='" + id1 + "'><td style='text-align:center'>" + inventype +
				"</td><td style='text-align:center'>" + invendesc + "</td><td contenteditable='true' id='t" + id1 +
				"' style='text-align:center;background-color: lightyellow;'>" + strconv.Itoa(qtys) +
				"</td><td style='text-align:center'> <a href='#' id='inven_" + id1 +
				"'class='btn btn-info btn-sm saveinven'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a>" +
				"  </td></tr>"
		}
		cn++
	}
	script += "</tbody></table>"
	if cn > 5 {
		script += "y"
	}
	io.WriteString(w, script)
}

func Updateinventories(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	id := r.URL.Query().Get("sid")
	scid := r.URL.Query().Get("ids")
	number, eorr := strconv.ParseInt(scid, 10, 0)
	if eorr != nil {
		logger.Println(eorr)
		io.WriteString(w, "Please Enter a Number")
		return
	}
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db

	logger.Println("Feching inventories", id, scid, number)
	stmt1, err := db.Prepare("update inventory set Quantity=? where id=?")

	if err != nil {
		logger.Println(err)
		return
	}
	defer stmt1.Close()
	if number >= 0 {
		rows, _ := stmt1.Exec(scid, id)
		if rows == nil {
			//fmt.Fprint(w,"no data stored in database")
		} else {
			fmt.Fprint(w, "DataSaved Successfuly")
			logger.Println("Updated successfully")
		}
	} else {
		io.WriteString(w, "No Negative Values!!")
	}
}
