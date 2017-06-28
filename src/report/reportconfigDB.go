/********************************************************************
 * FileName:     reportconfigDB.go
 * Project:      Havells StreetComm
 * Module:       reportconfigDB
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package report

import (
	"log"
	//"fmt"
	"html/template"
	"net/http"
	// "strings"
	"dbUtils"
	"userMgmt"

	_ "github.com/go-sql-driver/mysql"
)

var dbController dbUtils.DbUtilsStruct
var logger *log.Logger

func InitReport(dbcon dbUtils.DbUtilsStruct, logg *log.Logger) {
	dbController = dbcon
	logger = logg
}
func Report(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	if r.Method == "GET" {
		t, _ := template.ParseFiles("report.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in

		rpttype := r.Form["report_type"]
		logger.Println("report_type:", rpttype[0])
		rpttype1 := rpttype[0]
		rptfq := r.Form["report_frequency"]
		logger.Println("report_frequency:", rptfq[0])
		rptfq1 := rptfq[0]
		/*rptgentime :=r.Form["nextreportgen_time"]
		logger.Println("nextreportgen_date:", rptgentime[0])
		rptgentime1:=rptgentime[0]
		rptgendate :=r.Form["nextreportgen_date"]
		logger.Println("nextreportgen_date:", rptgendate[0])
		rptgendate1:=rptgendate[0]*/
		rptgenid := r.Form["report_gen_userid"]
		logger.Println("report_gen_userid:", rptgenid[0])
		rptgenid1 := rptgenid[0]

		//for database connectivity.
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		db := dbController.Db
		tstmt, err := db.Query("SELECT id FROM reportcofig where reportfrequency='" + rptfq1 + "' and reportdef_userid='" + rptgenid1 + "' and type='" + rpttype1 + "'")
		defer tstmt.Close()
		if err != nil {
			logger.Println(err)
		}
		for tstmt.Next() {
			logger.Println("Already Present!!")
			http.Redirect(w, r, "success.html", http.StatusFound)
			return
		}
		stmt, err := db.Prepare("INSERT reportcofig SET reportfrequency=?,reportdef_userid=?,type=?")
		defer stmt.Close()
		if err != nil {
			logger.Println(err)
		}
		res, err := stmt.Exec(rptfq1, rptgenid1, rpttype1)
		if err != nil {
			logger.Println(err)
		}
		if res == nil {
			logger.Println("Faild to store data")
			http.Redirect(w, r, "errormessage.html", http.StatusFound)
		} else {
			http.Redirect(w, r, "success.html", http.StatusFound)
		}
		logger.Println("hiiiii")
	}
	logger.Println("hellloooo")
}

/*
func main() {

   http.HandleFunc("/report", report)
    err := http.ListenAndServe(":9090", nil) // setting listening port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
	}
}*/
