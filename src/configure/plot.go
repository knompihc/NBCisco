/********************************************************************
 * FileName:     plot.go
 * Project:      Havells StreetComm
 * Module:       plot
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure

import (
	"net/http"
	/*"encoding/csv"
	"os"*/
	"io"
	//"fmt"
	//"time"
	"userMgmt"

	_ "github.com/go-sql-driver/mysql"
)

func Plot(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	sel := r.URL.Query().Get("sel")
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select (timestamp)," + sel + " from parameters where sgu_id='" + id + "' and timestamp>='" + sd + "' and timestamp<='" + ed + "' order by timestamp desc limit 100")
	defer rows.Close()
	chkErr(err, &w)
	data := "["
	fl := false
	for rows.Next() {
		if fl {
			data += ","
		} else {
			fl = true
		}
		var ti, val string
		err = rows.Scan(&ti, &val)
		data += "{\"ti\":\"" + ti + "\",\"val\":\"" + val + "\"}"
	}
	data += "]"
	logger.Println(data)
	io.WriteString(w, data)
}
func Csv(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	sel := r.URL.Query().Get("sel")
	sd := r.URL.Query().Get("sd")
	ed := r.URL.Query().Get("ed")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select timestamp," + sel + " from parameters where sgu_id='" + id + "' and timestamp>='" + sd + "' and timestamp<='" + ed + "' order by timestamp desc limit 100")
	defer rows.Close()
	chkErr(err, &w)
	/*ti:=time.Now()

	_, eorr := os.Stat("static/reports")
	if eorr == nil {  }
	if os.IsNotExist(eorr) {
		eorr=os.Mkdir("static/reports",os.ModePerm)
		if eorr!=nil{
			logger.Println(eorr)
		}
	}
	fname:="reports/"+ti.Format("2006-01-02 15:04:05")+".csv"
	file, err := os.Create("static/"+fname)
	if err != nil {
		logger.Println("Cannot create file ", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	*/
	da := ""
	fl := true
	for rows.Next() {
		if !fl {
			da += ","
		} else {
			fl = false
		}
		var ti, val string
		err = rows.Scan(&ti, &val)
		da += ti + "," + val
		/*starr:=[]string{ti,val}
		err := writer.Write(starr)
		if err != nil {
			logger.Println("Cannot write to file ", err)
		}*/
	}
	/*
		writer.Flush()
		logger.Println(fname)
		w.Header().Set("Content-Disposition", "attachment; filename="+fname)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))*/
	io.WriteString(w, da)
}
