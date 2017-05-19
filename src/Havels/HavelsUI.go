/********************************************************************
 * FileName:     HavelsUI.go
 * Project:      Havells StreetComm
 * Module:       HavelsUI
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package main

import (
	"NBApis"
	"bytes"
	"configure"
	"database/sql"
	"dbUtils"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mapview"
	//	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/smtp"
	"net/url"
	"os"
	//	"regexp"
	"report"
	"sguUtils"
	"strconv"
	"strings"
	"tcpServer"
	"tcpUtils"
	"time"

	"github.com/context"
	"github.com/go-github/github"
	//	"github.com/jwt-go-master"
	"github.com/rollbar"
	"github.com/scorredoira/email"
	"github.com/sessions"
	"github.com/xlsx"
	"golang.org/x/oauth2"
)

var store = sessions.NewCookieStore([]byte("something-very-secret"))
var (
	dbController          dbUtils.DbUtilsStruct
	sguConnectionChan     chan net.Conn
	err                   error
	LampControllerChannel chan sguUtils.SguUtilsLampControllerStruct
	//energytcputil           tcpUtils.TcpUtilsStruct
	//energysguutil           sguUtils.SguUtilsEnergyCntrStruct
	//energysguutilChannel    chan   	sguUtils.SguUtilsEnergyCntrStruct
	SendSMSChan     chan string
	logger          *log.Logger
	per_scu_delay   string
	scu_polling     string
	scu_scheduling  string
	scu_retry_delay string
	scu_max_retry   string
	ack_delay       string

	Sgu_firmware        map[int64][]byte
	Sgu_firmware_size   string
	Sgu_firmware_major  byte
	Sgu_firmware_minor  byte
	Sgu_firmware_name   string
	Sgu_firmware_bucket int64

	Scu_firmware        map[int64][]byte
	Scu_firmware_size   string
	Scu_firmware_major  byte
	Scu_firmware_minor  byte
	Scu_firmware_name   string
	Scu_firmware_bucket int64
)

const (
	Deployment_id          = "Test"
	MaxNumSGUs             = 1024
	sguChanSize            = 16
	SendSMSChanSize        = 8
	maxNumScusPerSgu       = 100
	lampControllerChansize = maxNumScusPerSgu * 4
	remoteLog              = false
	aesPassword            = "234FHF?#@$#%%jio4323486"
	rollbarToken           = "fbd0d81022b044f28f63018a7388b2bb"
	github_url             = "https://github.com/login/oauth/access_token"
	github_client_id       = "f57cb3fd2c96cb4e5148"
	github_client_secret   = "bed603e1add83829dc2dae8768fa60f983e35836"
)

type names struct {
	Name []string
}

//--------------------------------------------------------------------------------------------
type viewrep struct {
	Chk              string `json:"chk"`
	Id               string `json:"id"`
	Report_frequency string `json:"reportfrequency"`
	Reportdef_userid string `json:"reportdef_userid"`
	Type             string `json:"type"`
}

type reportview struct {
	Viewrep []viewrep `json:"data"`
}

type viewalert struct {
	Chk        string `json:"chk"`
	Id         string `json:"id"`
	Name       string `json:"name"`
	Email_id   string `json:"email_id"`
	Mobile_num string `json:"mobile_num"`
}

type alertview struct {
	Viewalert []viewalert `json:"data"`
}

type viewuser struct {
	Chk        string `json:"chk"`
	User_email string `json:"user_email"`
	Admin_op   string `json:"admin_op"`
}

type userview struct {
	Viewuser []viewuser `json:"data"`
}

type ota struct {
	Sgu []sguota `json:"data"`
}

type sguota struct {
	Chk    string `json:"chk"`
	Sgu    string `json:"sgu"`
	Curr   string `json:"curr"`
	Status string `json:"status"`
}

type cota struct {
	Scu []scuota `json:"data"`
}

type scuota struct {
	Chk    string `json:"chk"`
	Scu    string `json:"scu"`
	Curr   string `json:"curr"`
	Status string `json:"status"`
}

//----------------------------------------------------------------------------------------------

func login(w http.ResponseWriter, r *http.Request) {

	passed := true

	r.ParseForm()
	// logic part of log in

	id := r.Form["deployment_id"]

	//logger.Println(username[0])
	//logger.Println(password[0])
	session, _ := store.Get(r, "auth")

	if dbController.DbConnected {

		statement := "select * from deployment where deployment_id='" + id[0] + "'"

		rows, err := dbController.Db.Query(statement)
		//logger.Println(statement)

		if err != nil {
			logger.Println("Error quering database  for login information")
			logger.Println(err)
		} else {

			for rows.Next() {

				passed = true
				session.Values["set"] = 1
			}

			rows.Close()
		}
	}
	//session.Values["set"] = 1
	if passed {
		//logger.Println("Matching entry found. Redirecting")
		session.Save(r, w)
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)

	} else {

		fmt.Fprintf(w, "<h1>Invalid User Name Or Password</h1>")
		//http.Error(w, "Invalid User Name Or Password\n",http.StatusInternalServerError)
		//http.Redirect(w, r,  "index.html", http.StatusFound)

		//logger.Println("Timer started")
		//timer1 := time.NewTimer(time.Second * 5)
		//<-timer1.C
		//logger.Println("Timer stoped")
		//http.Redirect(w, r,  "index.html", http.StatusFound)

	}

}
func auth(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == nil || session.Values["set"] == 0 {
		io.WriteString(w, "0")
		return
	} else if session.Values["set"] == 1 {
		io.WriteString(w, "1")
	} else if session.Values["set"] == 2 {
		io.WriteString(w, "2")
	}
}
func isAdmin(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println("get admin=", session.Values["isadmin"])
	if session.Values["isadmin"] == nil || session.Values["isadmin"] == 0 {
		io.WriteString(w, "0")
		return
	} else if session.Values["isadmin"] == 1 {
		io.WriteString(w, "1")
	}
}
func signout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	session.Values["set"] = 0
	session.Save(r, w)
	logger.Println("User logged out")
	io.WriteString(w, "0")
}
func getUid(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	var uid string
	data := session.Values["uid"]
	if str, ok := data.(string); ok {
		/* act on str */
		uid = str
		logger.Println(uid)
	} else {
		uid = ""
		logger.Println("error")
		/* not string */
	}

	io.WriteString(w, uid)
}
func adminlogin(w http.ResponseWriter, r *http.Request) {

	passed := false

	r.ParseForm()
	// logic part of log in

	username := r.Form["username"]
	password := r.Form["password"]
	var admin_op int
	/*if strings.Compare(r.Form["admin_op"][0],"as_admin")== 0 {
		admin_op = 1;
	}	else {
		admin_op = 0;
	}*/
	if strings.EqualFold(r.Form["admin_op"][0], "as_admin") == true {
		admin_op = 1
	} else {
		admin_op = 0
	}
	logger.Println("admin???", admin_op)
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 2 {
		http.Redirect(w, r, "index.html", http.StatusFound)
		return
	}
	if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	//logger.Println(username[0])
	//logger.Println(password[0])

	if dbController.DbConnected {

		//row, err := db.Query("select user_email, password from login where user_email=?",1)

		statement := "select * from login where user_email=AES_ENCRYPT('" +
			username[0] + "','" + aesPassword + "') " +
			"AND password=AES_ENCRYPT('" +
			password[0] + "','" + aesPassword + "') " +
			"AND admin_op='" +
			strconv.Itoa(admin_op) + "'"

		rows, err := dbController.Db.Query(statement)
		//logger.Println(statement)

		if err != nil {
			logger.Println("Error quering database  for login information")
			logger.Println(err)
		} else {

			for rows.Next() {

				passed = true
			}

			rows.Close()
		}
	}

	if passed {
		session.Values["set"] = 2
		session.Values["uid"] = username[0]
		session.Values["isadmin"] = admin_op
		session.Save(r, w)
		logger.Println("Matching entry found. Redirecting")
		http.Redirect(w, r, "index.html", http.StatusFound)

	} else {

		fmt.Fprintf(w, "<h1>Invalid Admin User Name Or Password</h1>")
		//http.Error(w, "Invalid User Name Or Password\n",http.StatusInternalServerError)
		//http.Redirect(w, r,  "index.html", http.StatusFound)

		//logger.Println("Timer started")
		//timer1 := time.NewTimer(time.Second * 5)
		//<-timer1.C
		//logger.Println("Timer stoped")
		//http.Redirect(w, r,  "index.html", http.StatusFound)

	}

}
func LocationAdd(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var lastInsertId int
	r.ParseForm()
	// logic part of log in

	location_name := r.Form["location_name"]
	location_lat := r.Form["location_lat"]
	location_lng := r.Form["location_lng"]
	logger.Println(location_name[0])
	logger.Println(location_lat[0])
	logger.Println(location_lng[0])
	if dbController.DbConnected {

		//row, err := db.Query(INSERT INTO locations(location_name, location_lat, location_lng)VALUES ($2, $3, $4),location_name[0],location_lat[0],location_lng[0])
		err := dbController.Db.QueryRow("INSERT INTO locations(location_name, location_lat, location_lng)VALUES ($1, $2, $3)returning location_id;", location_name[0], location_lat[0], location_lng[0]).Scan(&lastInsertId)

		if err != nil {
			logger.Println(err)
			fmt.Fprintf(w, "<h1>Location Name Are Already Present!</h1>")
		} else {
			logger.Println("last inserted id =", lastInsertId)
			http.Redirect(w, r, "blank.html", http.StatusFound)
			//fmt.Fprintf(w, "<h1>Location Details Are Added Successfully!</h1><br><a href="blank.html">press back</a>")
			//http.Error(w, "Invalid User Name Or Password\n",http.StatusInternalServerError)
			//http.Redirect(w, r,  "index.html", http.StatusFound)
			//logger.Println("Timer started")
			//timer1 := time.NewTimer(time.Second * 5)
			//<-timer1.C
			//logger.Println("Timer stoped")
			//http.Redirect(w, r,  "index.html", http.StatusFound)
		}
	}

}

func getLocationNames(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var name string
	r.ParseForm()
	// logic part of log in

	logger.Println("enters into getlocations!")
	if dbController.DbConnected {

		response := []string{}
		//row, err := db.Query(INSERT INTO locations(location_name, location_lat, location_lng)VALUES ($2, $3, $4),location_name[0],location_lat[0],location_lng[0])
		rows, err := dbController.Db.Query("SELECT location_name FROM location;")
		if err != nil {
			logger.Println(err)
		} else {
			for rows.Next() {
				err := rows.Scan(&name)
				if err != nil {
					logger.Println(err)
				} else {
					//logger.Println("Getting Values From DB!")
					//logger.Println("Location Valus:",name)
					response = append(response, name)
					//fmt.Printf("%q\n", strings.Split(name))
					//w.Header(name)
					//fmt.Fprintf(w, "<h1>Location Name Are Already Present:</h1>",name)
				}
			}

		}
		rows.Close()

		{
			//fmt.Fprintf(w, response, r.URL.Path[1:])
			a, err := json.Marshal(response)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {

				w.Write(a)
			}
		}
	}

}
func sguAdd(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var lastInsertId1 int
	r.ParseForm()
	sgu_id := r.Form["sgu_id"]
	location_name := r.Form["location_name"]
	sgu_lat := r.Form["sgu_lat"]
	sgu_lng := r.Form["sgu_lng"]
	logger.Println(sgu_id[0])
	logger.Println(location_name[0])
	logger.Println(sgu_lat[0])
	logger.Println(sgu_lng[0])
	// logic part of log in

	logger.Println("enters into sgudetails!")
	if dbController.DbConnected {

		//row, err := db.Query(INSERT INTO sgus(sgu_lat, sgu_lng, location_name) VALUES (?, ?, ?);)
		err := dbController.Db.QueryRow("INSERT INTO sgus(sgu_id,sgu_lat, sgu_lng, location_name) VALUES ($1,$2, $3, $4)returning sgu_id;", sgu_id[0], sgu_lat[0], sgu_lng[0], location_name[0]).Scan(&lastInsertId1)
		if err != nil {
			logger.Println(err)
			fmt.Fprintf(w, "<h1>SGU Is Already Present!</h1>")
		} else {
			logger.Println("last inserted id =", lastInsertId1)
			fmt.Fprintf(w, "<h1>Sgu Details Are Added!</h1>")
		}
	}

}
func AddSchedule(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var en string

	r.ParseForm()
	// logic part of log in
	ssd := r.Form["startdate"][0]
	logger.Println("ScheduleStartDate:", ssd)
	sst := r.Form["starttime"][0]
	logger.Println("ScheduleStartDate:", sst)
	edd := r.Form["enddate"][0]
	logger.Println("ScheduleStartDate:", edd)
	et := r.Form["endtime"][0]
	logger.Println("ScheduleStartDate:", et)
	pwm := r.Form["pwm"]
	logger.Println("pwm value:", pwm[0])
	datestr := ssd
	timestr := sst
	logger.Println("timestr", timestr)
	dateinymd := strings.Split(datestr, "/")
	sy := dateinymd[0]
	logger.Println("sy", sy)
	sm := dateinymd[1]
	logger.Println(sm)
	sd := dateinymd[2]
	logger.Println(sd)
	logger.Println(en)
	dateen := edd
	timeen := et
	logger.Println(timeen)

	dateinymd = strings.Split(dateen, "/")
	ey := dateinymd[0]
	logger.Println(ey)
	em := dateinymd[1]
	logger.Println(em)
	ed := dateinymd[2]
	logger.Println(ed)

	iy, _ := strconv.Atoi(sy)
	fy, _ := strconv.Atoi(ey)
	im, _ := strconv.Atoi(sm)
	fm, _ := strconv.Atoi(em)
	id, _ := strconv.Atoi(sd)
	fd, _ := strconv.Atoi(ed)
	isti, _ := strconv.Atoi(strings.Split(timestr, ":")[0])
	ieti, _ := strconv.Atoi(strings.Split(timeen, ":")[0])
	var exp, vim, vid, vfm, vfd string
	if len(strconv.Itoa((im))) == 1 {
		vim = "0" + strconv.Itoa((im))
	} else {
		vim = strconv.Itoa((im))
	}
	if len(strconv.Itoa((id))) == 1 {
		vid = "0" + strconv.Itoa((id))
	} else {
		vid = strconv.Itoa((id))
	}
	if len(strconv.Itoa((fm))) == 1 {
		vfm = "0" + strconv.Itoa((fm))
	} else {
		vfm = strconv.Itoa((fm))
	}
	if len(strconv.Itoa((fd))) == 1 {
		vfd = "0" + strconv.Itoa((fd))
	} else {
		vfd = strconv.Itoa((fd))
	}
	//if(iy<fy-1){
	if iy == fy {
		if im == fm {
			if fd == id {
				exp = "(M=" + vim + "&&D=" + vid + "&&Y=" + strconv.Itoa(iy) + ")"
			} else {
				exp = "(M=" + vim + "&&D>=" + vid + "&&D<=" + vfd + "&&Y=" + strconv.Itoa(iy) + ")"
			}
		} else {
			//(((D>=20&&M==10)||(M>=11&&M<=11)||(D<=23&&M==12))&&Y==2015)
			exp = "(((D>=" + vid + "&&M=" + vim + ")||(M>" + vim + "&&M<" + vfm + ")||(D<=" + vfd + "&&M=" + vfm + "))&&Y=" + strconv.Itoa(iy) + ")"
		}
	} else {
		//((((D>=20&&M==10)||(M>10))&&Y==2015)||(Y>2015&&Y<2016)||((M<11||(M==12&&D<=23))&&Y==2016))
		exp = "((((D>=" + vid + "&&M=" + vim + ")||(M>" + vim + "))&&Y=" + strconv.Itoa(iy) + ")||(Y>" + strconv.Itoa(iy) + "&&Y<" + strconv.Itoa(fy) + ")||((M<" + vfm +
			"||(M=" + vfm + "&&D<=" + vfd + "))&&Y=" + strconv.Itoa(fy) + "))"
	}
	/*if im<12{
		exp+="||((M>="+strconv.Itoa((im)+1)+"||(D>="+strconv.Itoa(id)+"&&M=="+strconv.Itoa(im)+"))&&Y=="+strconv.Itoa(iy)+")"
		if fm>1{
			exp+="||((M<="+strconv.Itoa((fm)-1)+"||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"))&&Y=="+strconv.Itoa(fy)+")"
		}else{
			exp+="||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"&&Y=="+strconv.Itoa(fy)+")"
		}
	}else{
		exp+="||(D>="+strconv.Itoa(id)+"&&M=="+strconv.Itoa(im)+"&&Y=="+strconv.Itoa(iy)+")"
		if fm>1{
			exp+="||((M<="+strconv.Itoa((fm)-1)+"||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"))&&Y=="+strconv.Itoa(fy)+")"
		}else{
			exp+="||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"&&Y=="+strconv.Itoa(fy)+")"
		}
	}*/
	/*exp+="||((M>="+strconv.Itoa((im)+1)+"&&Y=="+strconv.Itoa(iy)+")"
	exp+="||(M<="+strconv.Itoa((fm)-1)+"&&Y=="+strconv.Itoa(fy)+")"
	exp+="||(D>="+strconv.Itoa(id)+"&&M=="+strconv.Itoa(im)+"&&Y=="+strconv.Itoa(iy)+")"
	exp+="||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"&&Y=="+strconv.Itoa(fy)+"))"*/
	/*}else{
		exp+="((M>="+strconv.Itoa((im)+1)+"&&Y=="+strconv.Itoa(iy)+")"
		exp+="||(M<="+strconv.Itoa((fm)-1)+"&&Y=="+strconv.Itoa(fy)+")"
		exp+="||(D>="+strconv.Itoa(id)+"&&M=="+strconv.Itoa(im)+"&&Y=="+strconv.Itoa(iy)+")"
		exp+="||(D<="+strconv.Itoa(fd)+"&&M=="+strconv.Itoa(fm)+"&&Y=="+strconv.Itoa(fy)+"))"
	}*/
	//exp="(Y>="+sy+"&&Y<="+ey+")"
	logger.Println("startTime=", isti, " endTime=", ieti)
	if isti <= ieti {
		exp += "&&(T>=" + timestr + "&&T<=" + timeen + ")"
	} else {
		exp += "&&(T>=" + timestr + "||T<=" + timeen + ")"
	}

	logger.Println("expression " + exp)
	rows, err := dbController.Db.Query("Select idschedule from schedule where ScheduleStartTime='" + ssd + " " + sst + "' and ScheduleEndTime='" + edd + " " + et + "' and pwm='" + pwm[0] + "'")
	defer rows.Close()
	if rows.Next() {
		var tid int64
		rows.Scan(&tid)
		io.WriteString(w, "Schedule Already Added With Schedule ID: "+strconv.FormatInt(tid, 10))
		return
	}
	stmt, err := dbController.Db.Prepare("INSERT schedule SET ScheduleStartTime=?,ScheduleEndTime=?,pwm=?,ScheduleExpression=?")
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "Something Went Wrong!")
		return
	}
	res, err := stmt.Exec(ssd+" "+sst, edd+" "+et, pwm[0], exp)
	if err != nil {
		io.WriteString(w, "Something Went Wrong!")
		return
	}
	if res == nil {
		//fmt.Fprint(w,"no data stored in database")
		//http.Redirect(w,r,"errormessage.html",http.StatusFound)
		io.WriteString(w, "Something Went Wrong!")
	} else {
		io.WriteString(w, "Schedule Added Successfully!")
		//fmt.Fprint(w,"DataSaved Successfuly")
		//http.Redirect(w,r,"success.html",http.StatusFound)
	}

}

func ViewSchedule(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	logger.Println("hi dear")
	var scheid string

	//for database connectivity.
	stmt, err := dbController.Db.Prepare("select idschedule from schedule")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Query()
	if err != nil {
		logger.Println(err)
	}
	if res == nil {
		logger.Println("no Schedule available")
	} else {
		var cnt int
		for res.Next() {
			var schid string
			err := res.Scan(&schid)
			if err != nil {
				logger.Println(err)
			}
			if cnt != 0 {
				scheid += " " + schid

			} else {
				scheid += schid
			}
			cnt++
		}

	}
	fmt.Fprint(w, scheid)

}
func scuAdd(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var lastInsertId1 int
	r.ParseForm()
	scu_id := r.Form["scu_id"]
	location_name := r.Form["location_name"]
	sgu_id := r.Form["sgu_id"]
	scu_lat := r.Form["scu_lat"]
	scu_lng := r.Form["scu_lng"]
	logger.Println(scu_id[0])
	logger.Println(location_name[0])
	logger.Println(sgu_id[0])
	logger.Println(scu_lat[0])
	logger.Println(scu_lng[0])
	// logic part of log in

	logger.Println("enters into sgudetails!")
	if dbController.DbConnected {

		//row, err := db.Query(INSERT INTO scus( scu_id, scu_lat, scu_lng, sgu_id, location_name) VALUES (?, ?, ?, ?, ?);)
		err := dbController.Db.QueryRow("INSERT INTO scus( scu_id, scu_lat, scu_lng, sgu_id, location_name) VALUES ($1,$2, $3, $4,$5)returning scu_id;", scu_id[0], scu_lat[0], scu_lng[0], sgu_id[0], location_name[0]).Scan(&lastInsertId1)
		if err != nil {
			logger.Println(err)
			fmt.Fprintf(w, "<h1>SCU Is Already Present!</h1>")
		} else {
			logger.Println("last inserted id =", lastInsertId1)
			fmt.Fprintf(w, "<h1>SCU Details Are Added!</h1>")
		}
	}

}

func AllLampControlpwm(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var LampController sguUtils.SguUtilsLampControllerStruct

	logger.Println(r.URL)
	u, err := url.Parse(r.URL.String())
	logger.Println(u)
	logger.Println(u.RawQuery)

	m, _ := url.ParseQuery(u.RawQuery)
	arrsg := strings.Split(m["SGUID"][0], " ")
	arrsc := strings.Split(m["SCUID"][0], " ")
	for i := 0; i < len(arrsg)-1; i++ {
		logger.Println("SGUID ", arrsg[i])
		logger.Println("SCUID ", arrsc[i])
		LampController.SGUID, err = strconv.ParseUint(arrsg[i], 10, 64)
		if err != nil {
			logger.Println("Invalid SGUID" + arrsg[i] + " specified")
			return
		}
		LampController.SCUID, err = strconv.ParseUint(arrsc[i], 10, 64)
		if err != nil {
			logger.Println("Invalid SCUID" + arrsc[i] + " specified")
			return
		}
		LampController.LampEvent, err = strconv.Atoi(m["LampEvent"][0])
		if err != nil {
			logger.Println("Invalid lamp contral val  " + m["lampEvent"][0] + " specified")
			return
		}
		//GetSet field is set to set mode
		LampController.LampEvent |= 0x100
		LampController.PacketType = 0x3000
		LampController.ConfigArray = nil
		LampController.ConfigArrayLength = 0

		//LampController.ResponseSend  = make(chan bool)
		//LampController.ResponseSend  = make(chan bool)
		LampController.W = nil
		LampController.ResponseSend = nil
		LampControllerChannel <- LampController
		logger.Println("Lamp event sent to channel")
	}
	//logger.Println(m)
	//logger.Println(m["SGUID"][0])
	//logger.Println(m["SCUID"][0])
	//logger.Println(m["SCUID"][0])
	//logger.Println("I am inside")

	/*LampController.SGUID, err = strconv.ParseUint(m["SGUID"][0],10,64)

	//logger.Println("Parsed SGU ID")

	if err != nil {
		logger.Println("Invalid SGUID" + m["SGUID"][0] + " specified")
		return
	}

	LampController.SCUID, err = strconv.ParseUint(m["SCUID"][0],10,64)

	//logger.Println("Parsed SCU ID")

	if err != nil {
		logger.Println("Invalid SCUID" + m["SCUID"][0] + " specified")
		return
	}

	LampController.LampEvent, err = strconv.Atoi(m["LampEvent"][0])

	//logger.Println("Parsed lampEvent")

	if err != nil {
		logger.Println("Invalid lamp contral val  " + m["lampEvent"][0] + " specified")
		return
	}

	LampController.PacketType = 0x3000
	LampController.ConfigArray = nil
	LampController.ConfigArrayLength = 0


	if (LampController.W != nil) {

		logger.Println("Lamp event specified when still waiting for response from old event")
		logger.Println("Old event will be overwritten")

	}

	LampController.W = w

	LampController.ResponseSend  = make(chan bool)
	//fmt.Printf("LampId = %d, LampVal = %d\n", LampId, LampVal)
	LampControllerChannel<-LampController
	logger.Println("Lamp event sent to channel")

	//wait for response
	//TBD. Add a timeout here
	<-LampController.ResponseSend

	*/

}

func LampControl(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}

	var LampController sguUtils.SguUtilsLampControllerStruct

	logger.Println(r.URL)
	u, err := url.Parse(r.URL.String())
	logger.Println(u)
	logger.Println(u.RawQuery)

	m, _ := url.ParseQuery(u.RawQuery)
	/*arrsg:=strings.Split(m["SGUID"][0]," ")
	arrsc:=strings.Split(m["SCUID"][0]," ")
	for i:=0;i<len(arrsg)-1;i++ {
		logger.Println("SGUID ",arrsg[i])
		logger.Println("SCUID ",arrsc[i])
		LampController.SGUID, err = strconv.ParseUint(arrsg[i],10,64)
		if err != nil {
			logger.Println("Invalid SGUID" + arrsg[i] + " specified")
			return
		}
		LampController.SCUID, err = strconv.ParseUint(arrsc[i],10,64)
		if err != nil {
			logger.Println("Invalid SCUID" + arrsc[i] + " specified")
			return
		}
		LampController.LampEvent, err = strconv.Atoi(m["LampEvent"][0])
		if err != nil {
			logger.Println("Invalid lamp contral val  " + m["lampEvent"][0] + " specified")
			return
		}

		LampController.PacketType = 0x3000
		LampController.ConfigArray = nil
		LampController.ConfigArrayLength = 0

		LampController.ResponseSend  = make(chan bool)
		LampControllerChannel<-LampController
		logger.Println("Lamp event sent to channel")
	}*/
	//logger.Println(m)
	//logger.Println(m["SGUID"][0])
	//logger.Println(m["SCUID"][0])
	//logger.Println(m["SCUID"][0])
	//logger.Println("I am inside")

	LampController.SGUID, err = strconv.ParseUint(m["SGUID"][0], 10, 64)

	//logger.Println("Parsed SGU ID")

	if err != nil {
		logger.Println("Invalid SGUID" + m["SGUID"][0] + " specified")
		return
	}

	LampController.SCUID, err = strconv.ParseUint(m["SCUID"][0], 10, 64)

	//logger.Println("Parsed SCU ID")

	if err != nil {
		logger.Println("Invalid SCUID" + m["SCUID"][0] + " specified")
		return
	}

	LampController.LampEvent, err = strconv.Atoi(m["LampEvent"][0])

	//logger.Println("Parsed lampEvent")

	if err != nil {
		logger.Println("Invalid lamp contral val  " + m["lampEvent"][0] + " specified")
		return
	}
	//GetSet field is set to set mode
	LampController.LampEvent |= 0x100

	LampController.PacketType = 0x3000
	LampController.ConfigArray = nil
	LampController.ConfigArrayLength = 0

	if LampController.W != nil {

		logger.Println("Lamp event specified when still waiting for response from old event")
		logger.Println("Old event will be overwritten")

	}

	LampController.W = w

	LampController.ResponseSend = make(chan bool)
	//fmt.Printf("LampId = %d, LampVal = %d\n", LampId, LampVal)
	LampControllerChannel <- LampController
	logger.Println("Lamp event sent to channel")

	//wait for response
	//TBD. Add a timeout here
	<-LampController.ResponseSend

}
func AllLampControl(w http.ResponseWriter, r *http.Request) {

	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var LampController sguUtils.SguUtilsLampControllerStruct

	logger.Println(r.URL)
	u, err := url.Parse(r.URL.String())
	logger.Println(u)
	logger.Println(u.RawQuery)

	m, _ := url.ParseQuery(u.RawQuery)
	arrsg := strings.Split(m["SGUID"][0], " ")
	arrsc := strings.Split(m["SCUID"][0], " ")
	for i := 0; i < len(arrsg)-1; i++ {
		logger.Println("SGUID ", arrsg[i])
		logger.Println("SCUID ", arrsc[i])
		LampController.SGUID, err = strconv.ParseUint(arrsg[i], 10, 64)
		if err != nil {
			logger.Println("Invalid SGUID" + arrsg[i] + " specified")
			return
		}
		LampController.SCUID, err = strconv.ParseUint(arrsc[i], 10, 64)
		if err != nil {
			logger.Println("Invalid SCUID" + arrsc[i] + " specified")
			return
		}
		LampController.LampEvent, err = strconv.Atoi(m["LampEvent"][0])
		if err != nil {
			logger.Println("Invalid lamp contral val  " + m["lampEvent"][0] + " specified")
			return
		}
		//GetSet field is set to set mode
		LampController.LampEvent |= 0x100
		LampController.PacketType = 0x3000
		LampController.ConfigArray = nil
		LampController.ConfigArrayLength = 0

		//LampController.ResponseSend  = make(chan bool)
		//LampController.ResponseSend  = make(chan bool)
		LampController.W = nil
		LampController.ResponseSend = nil
		LampControllerChannel <- LampController
		logger.Println("Lamp event sent to channel")
		du, _ := time.ParseDuration(per_scu_delay + "s")
		time.Sleep(du)
	}
}
func AddEnergyParameter(w http.ResponseWriter, r *http.Request) {
	var energysguutil sguUtils.SguUtilsLampControllerStruct
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	var SGUID string
	deviceId, _ := strconv.ParseInt(r.URL.Query().Get("DeviceID"), 10, 64)
	Length, _ := strconv.ParseInt(r.URL.Query().Get("Length"), 10, 64)
	Query1 := r.URL.Query().Get("Query")
	logger.Println("deviceId:", deviceId)
	logger.Println("Length:", Length)
	logger.Println("Query1:", Query1)
	stmt1, err1 := dbController.Db.Prepare("INSERT idquerydefinition SET deviceid=?,length=?,query=?")
	if err1 != nil {
		logger.Println(err)
	}
	res1, err1 := stmt1.Exec(deviceId, Length, Query1)
	if err1 != nil {
		logger.Println(err)
	}
	if res1 == nil {
		//fmt.Fprint(w,"no data stored in database")
		//fmt.Fprintf(w, "energystore", "no data stored in database")
		http.Redirect(w, r, "errormessage.html", http.StatusFound)
	} else {
		fmt.Fprint(w, "DataSaved Successfully")
		//http.Redirect(w,r,"success.html",http.StatusFound)
		stmt, err := dbController.Db.Prepare("select sgu_id from sgu")
		if err != nil {
			fmt.Println(err)
		}
		res, err := stmt.Query()
		if err != nil {
			fmt.Println(err)
		} else {

			for res.Next() {
				logger.Println("SGUID")
				err := res.Scan(&SGUID)
				logger.Println("SGUID", SGUID)
				energysguutil.SGUID, err = strconv.ParseUint(SGUID, 10, 64)
				tempArray := make([]byte, (len(Query1)/2)+3)
				tempArray[0] = (byte)(deviceId)
				logger.Println("Deice Id:", tempArray[0])
				tempArray[1] = (byte)(Length)
				logger.Println("Length:", tempArray[1])
				ind := 0
				for t := 0; t < len(Query1); t += 2 {
					tms := ""
					tms += string(Query1[t])
					tms += string(Query1[t+1])
					logger.Println("FOR:", tms)
					tmp, _ := strconv.ParseUint(tms, 16, 64)
					strtmp := strconv.FormatUint(tmp, 10)
					xtmp, _ := strconv.ParseUint(strtmp, 10, 64)
					logger.Println("CONV:", xtmp)
					/*if strtmp=="0"{
						strtmp+="0"
					}*/
					tt := ((byte)(xtmp & 0x0FF))
					tempArray[ind+2] = tt
					logger.Println("Query:", tempArray[ind+2])
					ind++
				}
				//energytcputil .SendEnargyControl(deviceId,Length,Query)

				//set get/set to 1 i.e. set mode
				energysguutil.LampEvent = 1
				//energysguutil.PacketType = 0xA000
				//energysguutil.ConfigArray = nil
				//energysguutil.ConfigArrayLength = 0
				//LampControllerChannel<-energysguutil

				energysguutil.PacketType = 0xB000
				energysguutil.ConfigArray = tempArray
				energysguutil.ConfigArrayLength = len(tempArray)
				LampControllerChannel <- energysguutil

				if err != nil {
					logger.Println(err)
				}

			}
		}
	}
}
func Communparams(w http.ResponseWriter, r *http.Request) {
	var energysguutil sguUtils.SguUtilsLampControllerStruct
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	logger.Println("Inside 9000 packet parameters ")
	var SGUID string
	baudRate, _ := strconv.ParseInt(r.URL.Query().Get("baud_rate"), 10, 64)
	stopBit, _ := strconv.ParseInt(r.URL.Query().Get("stop_bits"), 10, 64)
	parityBit, _ := strconv.ParseInt(r.URL.Query().Get("parity_bits"), 10, 64)
	dataBits, _ := strconv.ParseInt(r.URL.Query().Get("data_bits"), 10, 64)
	logger.Println("baudrate:", baudRate)
	logger.Println("stopBit:", stopBit)
	logger.Println("parityBit:", parityBit)
	logger.Println("dataBits:", dataBits)
	/* stmt1, err1 := dbController.Db.Prepare("INSERT idquerydefinition SET deviceid=?,length=?,query=?")
	if err1!=nil{
		logger.Println(err)
	}
	res1, err1 := stmt1.Exec(deviceId, Length ,Query1)
	if err1!=nil{
		logger.Println(err)
	}
	if res1==nil{
		//fmt.Fprint(w,"no data stored in database")
		//fmt.Fprintf(w, "energystore", "no data stored in database")
		http.Redirect(w,r,"errormessage.html",http.StatusFound)
	}else{} */
	//fmt.Fprint(w,"DataSaved Successfuly")
	//http.Redirect(w,r,"success.html",http.StatusFound)
	logger.Println("executing 9000 packet")
	stmt, err := dbController.Db.Prepare("select sgu_id from sgu")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Query()
	logger.Println("SGUID")
	if err != nil {
		logger.Println(err)
	} else {

		for res.Next() {
			logger.Println("SGUID")
			err := res.Scan(&SGUID)
			logger.Println("SGUID", SGUID)
			energysguutil.SGUID, err = strconv.ParseUint(SGUID, 10, 64)
			tempArray := make([]byte, 4)
			tempArray[0] = (byte)(baudRate)
			logger.Println("baudrate:", (tempArray[0] & 0x00F))
			tempArray[1] = (byte)(stopBit)
			logger.Println("Stop Bit:", (tempArray[1] & 0x00F))
			tempArray[2] = (byte)(parityBit)
			logger.Println("Parity Bit:", (tempArray[2] & 0x00F))
			tempArray[3] = (byte)(dataBits)
			logger.Println("Data Bit:", (tempArray[3] & 0x00F))
			//energytcputil .SendEnargyControl(deviceId,Length,Query)
			//set get/set to 1 i.e. set mode
			energysguutil.LampEvent = 1
			//LampControllerChannel<-energysguutil
			//sending 9000 Packet
			energysguutil.PacketType = 0x9000
			energysguutil.ConfigArray = tempArray
			energysguutil.ConfigArrayLength = len(tempArray)
			LampControllerChannel <- energysguutil
			if err != nil {
				logger.Println(err)
			}

		}
		//http.Redirect(w,r,"energystore.html",http.StatusFound)
		fmt.Fprint(w, "Packet data send Successfully")

	}
}
func Polingparams(w http.ResponseWriter, r *http.Request) {
	var energysguutil sguUtils.SguUtilsLampControllerStruct
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	logger.Println("Inside A000 packet parameters ")
	var SGUID string
	packetEnable, _ := strconv.ParseInt(r.URL.Query().Get("packet_enable"), 10, 64)
	pollingRate, _ := strconv.ParseInt(r.URL.Query().Get("polling_rate"), 10, 64)
	responseRate, _ := strconv.ParseInt(r.URL.Query().Get("response_rate"), 10, 64)
	timeOut, _ := strconv.ParseInt(r.URL.Query().Get("time_out"), 10, 64)
	slaveId, _ := strconv.ParseInt(r.URL.Query().Get("device_id"), 10, 64)
	/* stmt1, err1 := dbController.Db.Prepare("INSERT idquerydefinition SET deviceid=?,length=?,query=?")
	if err1!=nil{
		logger.Println(err)
	}
	res1, err1 := stmt1.Exec(deviceId, Length ,Query1)
	if err1!=nil{
		logger.Println(err)
	}
	if res1==nil{
		//fmt.Fprint(w,"no data stored in database")
		//fmt.Fprintf(w, "energystore", "no data stored in database")
		http.Redirect(w,r,"errormessage.html",http.StatusFound)
	}else{} */
	//fmt.Fprint(w,"DataSaved Successfuly")
	//http.Redirect(w,r,"success.html",http.StatusFound)
	logger.Println("executing A000 packet")
	stmt, err := dbController.Db.Prepare("select sgu_id from sgu")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Query()
	logger.Println("SGUID")
	if err != nil {
		logger.Println(err)
	} else {

		for res.Next() {
			fmt.Println("SGUID")
			err := res.Scan(&SGUID)
			logger.Println("SGUID", SGUID)
			energysguutil.SGUID, err = strconv.ParseUint(SGUID, 10, 64)
			tempArray := make([]byte, 12)
			tempArray[0] = (byte)(packetEnable)
			logger.Println("packet enable:", (int)(tempArray[0]))

			//tempArrayPoll := make ([]byte,2)
			//tempArrayPoll =([]byte)(pollingRate)
			tempArray[2] = ((byte)(pollingRate & 0x0FF))
			tempArray[1] = ((byte)((pollingRate >> 8) & 0x0FF))
			//value:=(int)(tempArray[1]|tempArray[2])
			value := ((int)(tempArray[2])) & 0x00FF
			value |= ((((int)(tempArray[1])) << 8) & 0x00FF00)
			logger.Println("pollingRate:", value)

			//tempArrayResp := make ([]byte,2)
			//tempArrayResp =([]byte)(responseRate)
			tempArray[4] = ((byte)(responseRate & 0x0FF))
			tempArray[3] = ((byte)((responseRate >> 8) & 0x0FF))
			value = ((int)(tempArray[4])) & 0x00FF
			value |= ((((int)(tempArray[3])) << 8) & 0x00FF00)
			logger.Println("ResponseRate:", value)

			//tempArrayTime := make ([]byte,2)
			//tempArrayTime =([]byte)(timeOut)
			tempArray[6] = ((byte)(timeOut & 0x0FF))
			tempArray[5] = ((byte)((timeOut >> 8) & 0x0FF))
			value = ((int)(tempArray[6])) & 0x00FF
			value |= ((((int)(tempArray[5])) << 8) & 0x00FF00)
			logger.Println("timeout:", value)

			//tempArraySlave := make ([]byte,5)
			//tempArrayTime =([]byte)(slaveId)

			tempArray[11] = ((byte)(slaveId & 0x0FF))
			tempArray[10] = ((byte)(slaveId & 0x0FF))
			tempArray[9] = ((byte)(slaveId & 0x0FF))
			tempArray[8] = ((byte)(slaveId & 0x0FF))
			tempArray[7] = ((byte)(slaveId & 0x0FF))
			value = ((int)(tempArray[11])) & 0x00FF
			value |= ((((int)(tempArray[10])) << 8) & 0x00FF00)
			value |= ((((int)(tempArray[9])) << 16) & 0xFF0000)
			value |= ((((int)(tempArray[8])) << 24) & 0xFF000000)
			value |= ((((int)(tempArray[7])) << 32) & 0xFF00000000)
			logger.Println("slave Id:", value)
			//energytcputil .SendEnargyControl(deviceId,Length,Query)
			//set get/set to 1 i.e. set mode
			energysguutil.LampEvent = 1
			//LampControllerChannel<-energysguutil
			//sending A000 Packet
			energysguutil.PacketType = 0xA000
			energysguutil.ConfigArray = tempArray
			energysguutil.ConfigArrayLength = len(tempArray)
			LampControllerChannel <- energysguutil
			if err != nil {
				logger.Println(err)
			}

		}
		//http.Redirect(w,r,"success.html",http.StatusFound)
		fmt.Fprint(w, "Packet data send Successfully")
	}
}

func ViewInventory(w http.ResponseWriter, r *http.Request) {
	var lampQty, sguQty, scuQty string
	var quantity string
	//for database connectivity.
	stmt, err := dbController.Db.Prepare("select Quantity from inventory")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Query()
	if err != nil {
		logger.Println("Error quering database  for login information")
		logger.Println(err)
	} else {
		var cnt int
		for res.Next() {
			if cnt == 0 {
				err := res.Scan(&lampQty)
				if err != nil {
					logger.Println(err)
				}
				logger.Println(lampQty)
				quantity += lampQty
			}
			if cnt == 1 {
				err := res.Scan(&sguQty)
				if err != nil {
					logger.Println(err)
				}
				quantity += " " + sguQty
			}
			if cnt == 2 {
				err := res.Scan(&scuQty)
				if err != nil {
					logger.Println(err)
				}
				quantity += " " + scuQty
			}
			cnt++
		}

		res.Close()
	}
	logger.Println(w, quantity)
	fmt.Fprint(w, quantity)
}

func UpdateInventory(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	logger.Println("Inside")
	var lampQty, sguQty, scuQty int64
	Lamps, _ := strconv.ParseInt(r.FormValue("Lamp")[0:], 10, 64)
	SGU, _ := strconv.ParseInt(r.FormValue("SGU")[0:], 10, 64)
	SCU, _ := strconv.ParseInt(r.FormValue("SCU")[0:], 10, 64)
	//for database connectivity.
	logger.Println("ttt")
	stmt, err := dbController.Db.Prepare("select Quantity from inventory")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Query()
	if err != nil {
		logger.Println(err)
	} else {
		var cnt int
		for res.Next() {
			if cnt == 0 {
				err := res.Scan(&lampQty)
				if err != nil {
					logger.Println(err)
				}
			}

			if cnt == 1 {
				err := res.Scan(&sguQty)

				if err != nil {
					logger.Println(err)
				}
			}

			if cnt == 2 {
				err := res.Scan(&scuQty)

				if err != nil {
					logger.Println(err)
				}
			}
			cnt++

		}

	}

	lampQty = Lamps
	sguQty = SGU
	scuQty = SCU
	fl := 0
	stmt1, err := dbController.Db.Prepare("update  inventory set Quantity=? where AssetType=?")
	if err != nil {
		logger.Println(err)
	}
	res1, err := stmt1.Exec(lampQty, "Lamps")
	if err != nil {
		logger.Println("Error quering database  for login information")
		logger.Println(err)
	} else {

		if res1 == nil {
			fl = 1
			//fmt.Fprint(w,"data not store")
		} else {

			//fmt.Fprint(w,"data store successfully")
		}

	}

	stmt2, err := dbController.Db.Prepare("update  inventory set Quantity=? where AssetType=?")
	if err != nil {
		logger.Println(err)
	}
	res2, err := stmt2.Exec(scuQty, "SCU")
	if err != nil {
		logger.Println("Error quering database  for login information")
		logger.Println(err)
	} else {

		if res2 == nil {
			fl = 1
			//logger.Fprint(w,"data not store")
		} else {

			//logger.Fprint(w,"data store successfully")
		}

	}

	stmt3, err := dbController.Db.Prepare("update  inventory set Quantity=? where AssetType=?")
	if err != nil {
		logger.Println(err)
	}
	res3, err := stmt3.Exec(sguQty, "SGU")
	if err != nil {
		logger.Println("Error quering database  for login information")
		logger.Println(err)
	} else {

		if res3 == nil {
			fl = 1
			//fmt.Fprint(w,"data not store")

		} else {
			//fmt.Fprint(w,"data store successfully")

		}

	}

	if fl == 1 {
		http.Redirect(w, r, "errormessage.html", http.StatusFound)
	} else {
		http.Redirect(w, r, "success.html", http.StatusFound)
	}

}

func supportDefinition(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	subject1 := r.URL.Query().Get("support_sub")
	logger.Println(subject1)
	category1 := r.URL.Query().Get("support_category")
	logger.Println(category1)
	email1 := r.URL.Query().Get("support_email")
	logger.Println(email1)
	contact1 := r.URL.Query().Get("support_contact")
	logger.Println(contact1)
	description1 := r.URL.Query().Get("support_desc")
	logger.Println(description1)
	status1 := 0
	logger.Println(status1)

	stmt, err := dbController.Db.Prepare("INSERT supportdefinition SET Subject=?,Category=?,EmailID=?,ContactNO=?,Description=?,Status=?")
	if err != nil {
		logger.Println(err)
	}
	res, err := stmt.Exec(subject1, category1, email1, contact1, description1, status1)
	if err != nil {
		logger.Println(err)
	}
	wapticketno, err := res.LastInsertId()
	logger.Println("last value:", wapticketno)
	wapticketno1 := strconv.FormatInt(wapticketno, 10)
	//wapticketno1:=strconv.Itoa(10000)
	logger.Println("wapticketno1 after convertin to string", wapticketno1)
	if err != nil {
		logger.Println(err)
	}
	if res == nil {
		logger.Println("Faild to store data")
	} else {
		forSupportemail(subject1, category1, email1, contact1, description1, wapticketno1)
		forticketemail(email1, (string)(wapticketno1))
		fmt.Fprint(w, "Your Ticket Successfully Placed")
	}
}

func supportCP(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	email1 := r.URL.Query().Get("support_email")
	pass := r.URL.Query().Get("pass")
	pass1 := r.URL.Query().Get("pass1")
	pass2 := r.URL.Query().Get("pass2")
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	var savepasswd string
	if pass1 != pass2 {
		io.WriteString(w, "Passwords are not matched!!")
		return
	}
	//statement := "select password  from login where user_email='"+email1+"'"
	statement1 := "SELECT CAST(AES_DECRYPT(password,'234FHF?#@$#%%jio4323486') AS CHAR(10000) CHARACTER SET utf8 ) AS password FROM login where user_email= AES_ENCRYPT('" + email1 + "','234FHF?#@$#%%jio4323486') and password=AES_ENCRYPT('" + pass + "','234FHF?#@$#%%jio4323486');"
	stmt, err := db.Query(statement1)
	defer stmt.Close()
	if err != nil {
		logger.Println(err)
	}
	if stmt == nil {
		io.WriteString(w, "Email does not exist!!")
	} else {
		for stmt.Next() {
			err := stmt.Scan(&savepasswd)
			logger.Println("Feching login password:", savepasswd)
			if err != nil {
				logger.Println(err)
			}
			if savepasswd == pass2 {
				io.WriteString(w, "This password already used!!")
				return
			} else {
				_, err := db.Query("UPDATE login SET password=AES_ENCRYPT('" + pass2 + "','234FHF?#@$#%%jio4323486') WHERE user_email=AES_ENCRYPT('" + email1 + "','234FHF?#@$#%%jio4323486');")
				if err != nil {
					logger.Println(err)
					//io.WriteString(w,"Password Successfully Updated !!!!")
				}
				if err == nil {
					forSuccessemail(email1)
					io.WriteString(w, "Password Successfully Updated !!!!")
				}
			}
		}
	}
}

type EmailConfig struct {
	Username string
	Password string
	Host     string
	Port     int
}

func forSuccessemail(loginemail1 string) {
	// authentication configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 25
	smtpPass := "Havells123"
	smtpUser := "havellstreetcomm@chipmonk.in"

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}

	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	sender := "havellstreetcomm@chipmonk.in"

	receivers := []string{
		loginemail1,
	}

	message := "YOUR PASSWORD SUCCESSFULLY CHANGED"
	subject := "HAVELLS PASSWORD"

	emailContent := email.NewMessage(subject, message)
	emailContent.From = sender
	emailContent.To = receivers
	/*files := []string{
			               "/",
			               }


	         for _, filename := range files {
	                err := emailContent.Attach(filename)
						if err != nil {
	                         logger.Println(err)
							 }
					}*/
	logger.Println("Invoking email method")

	err := email.Send(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
		emailauth,
		emailContent)

	if err != nil {
		logger.Println(err)
	}
	logger.Println("Sending email on update of change password")
}

func forSupportemail(subject1, category1, email1, contact1, description1, wapticketno1 string) {
	// authentication configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 25
	smtpPass := "Havells123"
	smtpUser := "havellstreetcomm@chipmonk.in"

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}

	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	sender := "havellstreetcomm@chipmonk.in"

	receivers := []string{
		"raman.sake@chipmonk.in",
	}

	message := "Hi, \nCategory: " + category1 + "\nand My TICKET number is: " + wapticketno1 + "\n\n" + description1 + "\n\nRegards \n" + email1 + "\n" + contact1
	subject := subject1

	emailContent := email.NewMessage(subject, message)
	emailContent.From = sender
	emailContent.To = receivers
	files := []string{
		"/",
	}

	logger.Println("Invoking addressof your file")
	// address of your own files put here

	for _, filename := range files {
		err := emailContent.Attach(filename)
		if err != nil {
			logger.Println(err)
		}
	}
	logger.Println("Invoking email method")

	err := email.Send(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
		emailauth,
		emailContent)

	if err != nil {
		logger.Println(err)
	}
	logger.Println("Sending email to admin is successful")
}

func forticketemail(email1, wapticketno1 string) {
	// authentication configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 25
	smtpPass := "Havells123"
	smtpUser := "havellstreetcomm@chipmonk.in"

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}

	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	sender := "havellstreetcomm@chipmonk.in"

	receivers := []string{
		email1,
	}

	message := "YOUR QUERY SUCCESSFULLY SUBMITTED \n AND YOUR QUERY TICKET NUMBER IS : " + wapticketno1
	subject := "HAVELLS QUERY TICKET"

	emailContent := email.NewMessage(subject, message)
	emailContent.From = sender
	emailContent.To = receivers
	files := []string{
		"/",
	}

	logger.Println("Inside the gmail method")
	// address of your own files put here

	for _, filename := range files {
		err := emailContent.Attach(filename)
		if err != nil {
			fmt.Println(err)
		}
	}
	logger.Println("Executing send gmail methd")

	err := email.Send(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
		emailauth,
		emailContent)

	if err != nil {
		logger.Println(err)
	}
	logger.Println("Email send successful")
}

/*type handler func(w http.ResponseWriter, r *http.Request)
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("yes>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>")
	h(w, r)
}*/
//-------------------------------------------------------------------------------------------------------------------------------------------------------------

func getreportperson(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select id,reportfrequency,reportdef_userid,type from reportcofig ")
	defer rows.Close()
	//chkErr(err)
	res := reportview{}
	for rows.Next() {
		tmp := viewrep{}
		rows.Scan(&tmp.Id, &tmp.Report_frequency, &tmp.Reportdef_userid, &tmp.Type)
		tmp.Chk = ""
		//tmp.Chk=""
		res.Viewrep = append(res.Viewrep, tmp)
	}
	a, err := json.Marshal(res)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		//logger.Println(a)
		w.Write(a)

	}
}

func sguotadisplay(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("SELECT sgu_id,major,minor,status FROM sgu")
	defer rows.Close()
	//chkErr(err)
	res := ota{}
	for rows.Next() {
		tmp := sguota{}
		var ma, mi, sta sql.NullString
		rows.Scan(&tmp.Sgu, &ma, &mi, &sta)
		tmp.Chk = ""
		//tmp.Avail="2.0"
		//tmp.Chk=""
		if ma.Valid && mi.Valid {
			tmp.Curr = "ver_" + ma.String + "." + mi.String
		} else {
			tmp.Curr = "Unknown"
		}
		if sta.Valid {
			tmp.Status = sta.String
		}
		//tmp.Curr="ver_"+ma+"."+mi
		//tmp.Status=sta
		logger.Println("Sta=", sta, " id=", tmp.Sgu)
		res.Sgu = append(res.Sgu, tmp)
	}
	a, err := json.Marshal(res)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		//logger.Println(a)
		w.Write(a)

	}
}

func scuotadisplay(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("SELECT scu_id,major,minor,status FROM scu")
	defer rows.Close()
	//chkErr(err)
	res := cota{}
	for rows.Next() {
		tmp := scuota{}
		var ma, mi, sta sql.NullString
		rows.Scan(&tmp.Scu, &ma, &mi, &sta)
		tmp.Chk = ""
		//tmp.Avail="2.0"
		//tmp.Chk=""
		if ma.Valid && mi.Valid {
			tmp.Curr = "ver_" + ma.String + "." + mi.String
		} else {
			tmp.Curr = "Unknown"
		}
		if sta.Valid {
			tmp.Status = sta.String
		}
		//tmp.Curr="ver_"+ma+"."+mi
		//tmp.Status=sta
		logger.Println("Sta=", sta, " id=", tmp.Scu)
		res.Scu = append(res.Scu, tmp)
	}
	a, err := json.Marshal(res)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		//logger.Println(a)
		w.Write(a)

	}
}

func getalertperson(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("SELECT * FROM admin ")
	defer rows.Close()
	//chkErr(err)
	res := alertview{}
	for rows.Next() {
		tmp := viewalert{}
		rows.Scan(&tmp.Id, &tmp.Name, &tmp.Email_id, &tmp.Mobile_num)
		tmp.Chk = ""
		//tmp.Chk=""
		res.Viewalert = append(res.Viewalert, tmp)
	}
	a, err := json.Marshal(res)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		//logger.Println(a)
		w.Write(a)

	}
}

func deletereportperson(w http.ResponseWriter, r *http.Request) {
	qids := r.URL.Query().Get("ids")

	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	//logger.Println("Select distinct van_id from booking where id in ("+qids+")")
	rows, err := dbController.Db.Query("delete from reportcofig where id in (" + qids + ")")
	defer rows.Close()
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "0")
		return
	}
	io.WriteString(w, "1")
}

func deletealertperson(w http.ResponseWriter, r *http.Request) {
	qids := r.URL.Query().Get("ids")

	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	//logger.Println("Select distinct van_id from booking where id in ("+qids+")")
	rows, err := dbController.Db.Query("delete from admin where id in (" + qids + ")")
	defer rows.Close()
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "0")
		return
	}
	io.WriteString(w, "1")
}
func alluserview(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("SELECT CAST(AES_DECRYPT(user_email,'234FHF?#@$#%%jio4323486') AS CHAR(10000) CHARACTER SET utf8 ) AS user_email,admin_op FROM login ")
	defer rows.Close()
	//chkErr(err)
	res := userview{}
	for rows.Next() {
		tmp := viewuser{}
		var chkstat string
		rows.Scan(&tmp.User_email, &chkstat)
		//fmt.Println("this is chk::"+statechk)
		if chkstat == "0" {
			tmp.Admin_op = "NO"
		} else {
			tmp.Admin_op = "YES"
		}
		tmp.Chk = ""
		//tmp.Chk=""
		res.Viewuser = append(res.Viewuser, tmp)
	}
	a, err := json.Marshal(res)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		//logger.Println(a)
		w.Write(a)

	}
}

func deleteuserperson(w http.ResponseWriter, r *http.Request) {
	qids := r.URL.Query().Get("ids")
	logger.Println(qids)
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	//logger.Println("Select distinct van_id from booking where id in ("+qids+")")
	rows, err := dbController.Db.Query("delete from login where CAST(AES_DECRYPT(user_email,'234FHF?#@$#%%jio4323486') AS CHAR(10000) CHARACTER SET utf8 ) in (" + qids + ")")
	defer rows.Close()
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "0")
		return
	}
	io.WriteString(w, "1")
}

//--------------------------------------------------------------------------------------------------------------------------------------------------------
var chttp = http.NewServeMux()

func my(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 && !strings.Contains(r.URL.Path, "adminlogin.html") && strings.Contains(r.URL.Path, ".html") {
		logger.Println("Not logged in")
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	} else if (session.Values["set"] == nil || session.Values["set"] == 0) && !strings.Contains(r.URL.Path, "login.html") && strings.Contains(r.URL.Path, ".html") {
		logger.Println("Not logged in")
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	} else if session.Values["set"] == 2 && (strings.Contains(r.URL.Path, "login.html") || strings.Contains(r.URL.Path, "adminlogin.html")) {
		http.Redirect(w, r, "../index.html", http.StatusFound)
		return
	}
	chttp.ServeHTTP(w, r)
}
func handlePanic() {
	if r := recover(); r != nil {
		log.Println(r)

		rollbar.Error(rollbar.ERR, fmt.Errorf("RAW-SERVER: %v\n", r))
		rollbar.Wait()

		panic(r)
	}
}

//For github authenticaion-----------------------------------------------------------------------------------------------------------------
type githubStruct struct {
	Client_id     string `json:"client_id"`
	Client_secret string `json:"client_secret"`
	Code          string `json:"code"`
}

type githubResp struct {
	Access_token string `json:"access_token"`
	Scope        string `json:"scope"`
	Token_type   string `json:"token_type"`
}

type gitrepos struct {
	Fullname string `json:"fullname"`
	Owner    string `json:"owner"`
}

type gitbranch struct {
	Name string `json:"name"`
}

type gitdb struct {
	Fullname string `json:"fullname"`
	Owner    string `json:"owner"`
	Branch   string `json:"branch"`
}

type uigitdb struct {
	Major string `json:"major"`
	Minor string `json:"minor"`
	Name  string `json:"name"`
	Git   gitdb  `json:"git"`
}

type survey struct {
	Usr  string `json:"usr"`
	Pno  string `json:"pno"`
	Mun  string `json:"mun"`
	Ward string `json:"ward"`
	Loc  string `json:"loc"`
	Rw   string `json:"rw"`

	Pso    string `json:"pso"`
	Pla    string `json:"pla"`
	Height string `json:"height"`
	Pty    string `json:"pty"`
	Opw    string `json:"opw"`
	Lf     string `json:"lf"`
	Earth  string `json:"earth"`
	Phase  string `json:"phase"`
	Fun    string `json:"func"`
	Lul    string `json:"lul"`
	Lat    string `json:"lat"`
	Lng    string `json:"lng"`
}

func callback(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	code := r.URL.Query().Get("code")
	logger.Println(code)

	in := githubStruct{}
	in.Client_id = github_client_id
	in.Client_secret = github_client_secret
	in.Code = code

	value, _ := json.Marshal(in)
	client := http.Client{}
	req, _ := http.NewRequest("POST", github_url, bytes.NewBuffer(value))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Println(err)
		http.Redirect(w, r, "ota.html", http.StatusFound)
		return
	}
	logger.Println(resp)
	if err != nil {
		logger.Println("Unable to reach the server.")
		http.Redirect(w, r, "ota.html", http.StatusFound)
		return
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("body=", string(body))
		var out githubResp
		err := json.Unmarshal(body, &out)
		if err != nil {
			logger.Println(err)
			http.Redirect(w, r, "ota.html", http.StatusFound)
			return
		}
		logger.Println(out.Access_token)
		if out.Scope == "repo" {
			db := dbController.Db
			rows, err := db.Query("delete from ota_server where deployment_id='" + Deployment_id + "' and device='SGU'")
			defer rows.Close()
			if err != nil {
				logger.Println(err)
				http.Redirect(w, r, "ota.html", http.StatusFound)
				return
			}
			rows1, err := db.Query("insert into ota_server (deployment_id,device,access_token) values ('" + Deployment_id + "','SGU','" + out.Access_token + "')")
			defer rows1.Close()
			if err != nil {
				logger.Println(err)
				http.Redirect(w, r, "ota.html", http.StatusFound)
				return
			}

			rows2, err := db.Query("delete from ota_server where deployment_id='" + Deployment_id + "' and device='SCU'")
			defer rows2.Close()
			if err != nil {
				logger.Println(err)
				http.Redirect(w, r, "ota.html", http.StatusFound)
				return
			}
			rows3, err := db.Query("insert into ota_server (deployment_id,device,access_token) values ('" + Deployment_id + "','SCU','" + out.Access_token + "')")
			defer rows3.Close()
			if err != nil {
				logger.Println(err)
				http.Redirect(w, r, "ota.html", http.StatusFound)
				return
			}
		}

		http.Redirect(w, r, "ota.html", http.StatusFound)
		return
	}
	http.Redirect(w, r, "ota.html", http.StatusFound)
	return
}

func setrepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	dev := r.URL.Query().Get("dev")
	//rurl:=r.URL.Query().Get("url")

	db := dbController.Db
	rows, err := db.Query("select access_token from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if rows.Next() {
		var token string
		rows.Scan(&token)
		logger.Println("For token:", token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)

		client := github.NewClient(tc)

		// list all repositories for the authenticated user
		repos, _, err := client.Repositories.List("", nil)
		if err != nil {
			logger.Println(err)
		}

		for _, v := range repos {
			logger.Println(v)
		}
		//logger.Println(repos)
	}
	if err != nil {
		logger.Println(err)
		return
	}

}

func getallrepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	dev := r.URL.Query().Get("dev")
	//rurl:=r.URL.Query().Get("url")

	db := dbController.Db
	rows, err := db.Query("select access_token from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if rows.Next() {
		var token string
		rows.Scan(&token)
		logger.Println("For token:", token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)
		opt := &github.RepositoryListOptions{
			ListOptions: github.ListOptions{PerPage: 100},
		}
		// list all repositories for the authenticated user
		repos, _, err := client.Repositories.List("", opt)
		if err != nil {
			logger.Println(err)
		}

		//logger.Println(repos)

		data := []gitrepos{}
		for _, v := range repos {
			//logger.Println(v)
			tmp := gitrepos{}
			tmp.Fullname = *(v.Name)
			tmp.Owner = *(v.Owner.Login)
			data = append(data, tmp)
		}
		a, err := json.Marshal(data)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			//logger.Println(a)
			w.Write(a)
		}
	}
	if err != nil {
		logger.Println(err)
		return
	}

}

func downloadSguRepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	logger.Println("Checking SGUs")
	sids := r.URL.Query().Get("ids")
	//rurl:=r.URL.Query().Get("url")
	dev := "SGU"
	db := dbController.Db
	rows, err := db.Query("select access_token,detail,major,minor,firmware_name from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if rows.Next() {
		var token, detail, ma, mi, na string
		rows.Scan(&token, &detail, &ma, &mi, &na)
		logger.Println("For token:", token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)

		client := github.NewClient(tc)
		str := gitdb{}
		err := json.Unmarshal([]byte(detail), &str)
		if err != nil {
			logger.Println(err)
			io.WriteString(w, "1")
			return
		}
		data, err := client.Repositories.DownloadContents(str.Owner, str.Fullname, "SGU_APP_210616.bin", nil)
		if err != nil {
			logger.Println(err)
			io.WriteString(w, "1")
			return
		}
		logger.Println(data)

		raw, err := ioutil.ReadAll(data)
		if err != nil {
			fmt.Errorf("Error reading response: %v", err)
			io.WriteString(w, "1")
			return
		}
		Sgu_firmware_size = strconv.Itoa(len(raw)) + "\000"
		ima, _ := strconv.Atoi(ma)
		imi, _ := strconv.Atoi(mi)
		Sgu_firmware_major = byte(ima)
		Sgu_firmware_minor = byte(imi)
		Sgu_firmware_name = na + "\000"

		bkt := len(raw) / 1024
		if len(raw)%1024 != 0 {
			bkt++
		}
		logger.Println(len(raw), "Bytes Size")
		logger.Println(bkt, "No. of buckets")
		rea := bytes.NewReader(raw)
		for i := 0; i < bkt; i++ {
			bc := make([]byte, 1024)
			rea.Read(bc)
			Sgu_firmware[int64(i)] = bc
			//logger.Println(bc)
		}
		Sgu_firmware_bucket = int64(bkt)
		logger.Println("Size=", Sgu_firmware_size, " Major=", Sgu_firmware_major, " Minor=", Sgu_firmware_minor, " Name=", Sgu_firmware_name, " Buckets=", Sgu_firmware_bucket)
		srows, err := dbController.Db.Query("select sgu_id,major,minor from sgu where sgu_id in (" + sids + ")")
		defer srows.Close()
		if err != nil {
			logger.Println("Error tryting to read SGU list from database")
			io.WriteString(w, "1")
			logger.Println(err)
			return
		}
		tcpUtils.Sgu_firmware_init(Sgu_firmware, Sgu_firmware_size, Sgu_firmware_major, Sgu_firmware_minor, Sgu_firmware_name, Sgu_firmware_bucket)
		for srows.Next() {
			var ma, mi byte
			var sguid uint64
			srows.Scan(&sguid, &ma, &mi)
			if ma == Sgu_firmware_major && mi == Sgu_firmware_minor {
				logger.Println("Firmware already Updated for sguid=", sguid)
				continue
			}
			go sguUtils.Sgu_firmware_update(sguid)
			stmt, _ := db.Prepare("update sgu set status=? where sgu_id='" + strconv.FormatUint(sguid, 10) + "'")
			_, eorr := stmt.Exec("Inprogress")
			defer stmt.Close()
			if eorr != nil {
				logger.Println(eorr)
			}
		}
	}
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "1")
		return
	}
	io.WriteString(w, "0")

}

func downloadScuRepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	logger.Println("Checking SCUs")
	sids := r.URL.Query().Get("ids")
	//rurl:=r.URL.Query().Get("url")
	dev := "SCU"
	db := dbController.Db
	rows, err := db.Query("select access_token,detail,major,minor,firmware_name from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if rows.Next() {
		var token, detail, ma, mi, na string
		rows.Scan(&token, &detail, &ma, &mi, &na)
		logger.Println("For token:", token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)

		client := github.NewClient(tc)
		str := gitdb{}
		err := json.Unmarshal([]byte(detail), &str)
		if err != nil {
			logger.Println(err)
			io.WriteString(w, "1")
			return
		}
		data, err := client.Repositories.DownloadContents(str.Owner, str.Fullname, "SGU_APP_210616.bin", nil)
		if err != nil {
			logger.Println(err)
			io.WriteString(w, "1")
			return
		}
		logger.Println(data)

		raw, err := ioutil.ReadAll(data)
		if err != nil {
			fmt.Errorf("Error reading response: %v", err)
			io.WriteString(w, "1")
			return
		}
		Scu_firmware_size = strconv.Itoa(len(raw)) + "\000"
		ima, _ := strconv.Atoi(ma)
		imi, _ := strconv.Atoi(mi)
		Scu_firmware_major = byte(ima)
		Scu_firmware_minor = byte(imi)
		Scu_firmware_name = na + "\000"

		bkt := len(raw) / 1024
		if len(raw)%1024 != 0 {
			bkt++
		}
		logger.Println(len(raw), "Bytes Size")
		logger.Println(bkt, "No. of buckets")
		rea := bytes.NewReader(raw)
		for i := 0; i < bkt; i++ {
			bc := make([]byte, 1024)
			rea.Read(bc)
			Scu_firmware[int64(i)] = bc
			//logger.Println(bc)
		}
		Scu_firmware_bucket = int64(bkt)
		logger.Println("Size=", Scu_firmware_size, " Major=", Scu_firmware_major, " Minor=", Scu_firmware_minor, " Name=", Scu_firmware_name, " Buckets=", Scu_firmware_bucket)
		srows, err := dbController.Db.Query("select scu_id,major,minor,sgu_id from scu where scu_id in (" + sids + ")")
		defer srows.Close()
		if err != nil {
			logger.Println("Error tryting to read SCU list from database")
			io.WriteString(w, "1")
			logger.Println(err)
			return
		}
		tcpUtils.Scu_firmware_init(Scu_firmware, Scu_firmware_size, Scu_firmware_major, Scu_firmware_minor, Scu_firmware_name, Scu_firmware_bucket)
		for srows.Next() {
			var sma, smi sql.NullString
			var lma, lmi string
			var scuid, sguid uint64
			srows.Scan(&scuid, &sma, &smi, &sguid)
			if sma.Valid {
				lma = (sma.String)
			}
			if smi.Valid {
				lmi = (smi.String)
			}
			logger.Println(sma, smi, ma, mi, lma, lmi)
			if ma == lma && mi == lmi {
				logger.Println("Firmware already Updated for scuid=", scuid)
				continue
			}
			logger.Println("SGU=", sguid, " SCU=", scuid)
			go sguUtils.Scu_firmware_update(scuid, sguid)
			stmt, _ := db.Prepare("update scu set status=? where scu_id='" + strconv.FormatUint(scuid, 10) + "'")
			_, eorr := stmt.Exec("Inprogress")
			defer stmt.Close()
			if eorr != nil {
				logger.Println(eorr)
			}
		}
	}
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "1")
		return
	}
	io.WriteString(w, "0")

}

func getallbranchesforRepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	str := gitrepos{}
	dev := r.URL.Query().Get("dev")
	str.Fullname = (r.URL.Query().Get("fname"))
	str.Owner = (r.URL.Query().Get("owner"))
	logger.Println(str.Owner, str.Fullname)
	db := dbController.Db
	rows, err := db.Query("select access_token from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if rows.Next() {
		var token string
		rows.Scan(&token)
		logger.Println("For token:", token)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)

		client := github.NewClient(tc)

		// list all repositories for the authenticated user
		branches, _, err := client.Repositories.ListBranches(str.Owner, str.Fullname, nil)
		if err != nil {
			logger.Println(err)
		}
		data := []gitbranch{}
		for _, v := range branches {
			logger.Println(*(v.Name))
			tmp := gitbranch{}
			tmp.Name = *(v.Name)
			data = append(data, tmp)
		}
		a, err := json.Marshal(data)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			//logger.Println(a)
			w.Write(a)

		}
	}
	if err != nil {
		logger.Println(err)
		return
	}

}

func connectrepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	str := gitdb{}
	dev := r.URL.Query().Get("dev")
	str.Fullname = (r.URL.Query().Get("fname"))
	str.Owner = (r.URL.Query().Get("owner"))
	str.Branch = (r.URL.Query().Get("branch"))
	ma := r.URL.Query().Get("major")
	mi := r.URL.Query().Get("minor")
	na := r.URL.Query().Get("name")
	a, err := json.Marshal(str)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
		io.WriteString(w, "0")
		return
	}
	db := dbController.Db
	stmt, _ := db.Prepare("update ota_server set detail=?,major=?,minor=?,firmware_name=? where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	_, eorr := stmt.Exec(a, ma, mi, na)
	if eorr != nil {
		logger.Println(eorr)
		io.WriteString(w, "0")
		return
	}
	io.WriteString(w, "1")

}

func getconnectedrepo(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	dev := r.URL.Query().Get("dev")
	db := dbController.Db
	rows, err := db.Query("select detail,major,minor,firmware_name from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	if err != nil {
		logger.Println(err)
		io.WriteString(w, "1")
		return
	}
	defer rows.Close()
	if rows.Next() {
		var detail, ma, mi, na string
		rows.Scan(&detail, &ma, &mi, &na)
		tmp := gitdb{}
		eorr := json.Unmarshal([]byte(detail), &tmp)
		if eorr != nil {
			logger.Println(eorr)
			io.WriteString(w, "1")
			return
		}
		val := uigitdb{}
		val.Major = ma
		val.Minor = mi
		val.Name = na
		val.Git = tmp
		x, err := json.Marshal(val)
		if err != nil {
			logger.Println(eorr)
			io.WriteString(w, "1")
			return
		}
		w.Write(x)
	}

}
func gitconnected(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	dev := r.URL.Query().Get("dev")
	db := dbController.Db
	rows, err := db.Query("select access_token from ota_server where deployment_id='" + Deployment_id + "' and device='" + dev + "'")
	defer rows.Close()
	if err != nil {
		io.WriteString(w, "0")
		return
	}
	if rows.Next() {
		var token string
		rows.Scan(&token)
		logger.Println("Connected For token:", token)
		if len(token) != 0 {
			io.WriteString(w, "1")
			return
		}
	}
	io.WriteString(w, "2")
}
func saveLocation(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	str := survey{}
	str.Usr = r.URL.Query().Get("usr")
	str.Pno = r.URL.Query().Get("pno")
	str.Mun = r.URL.Query().Get("mun")
	str.Ward = r.URL.Query().Get("ward")
	str.Loc = r.URL.Query().Get("loc")
	str.Rw = r.URL.Query().Get("rw")

	str.Pso = r.URL.Query().Get("pso")
	str.Pla = r.URL.Query().Get("pla")
	str.Height = r.URL.Query().Get("height")
	str.Pty = r.URL.Query().Get("pty")
	str.Opw = r.URL.Query().Get("opw")
	str.Lf = r.URL.Query().Get("lf")
	str.Earth = r.URL.Query().Get("earth")
	str.Phase = r.URL.Query().Get("phase")
	str.Fun = r.URL.Query().Get("func")
	str.Lul = r.URL.Query().Get("lul")
	str.Lat = r.URL.Query().Get("lat")
	str.Lng = r.URL.Query().Get("lng")

	db := dbController.Db

	rows, err := db.Query("select id from survey where lat='" + str.Lat + "' and lng='" + str.Lng + "'")
	defer rows.Close()
	if err != nil {
		io.WriteString(w, "0")
		return
	}
	if rows.Next() {
		stmt, _ := db.Prepare("update survey set usr=?,pno=?,loc=?,rw=?,pso=?,pla=?,height=?,pty=?,opw=?,lf=?,earth=?,phase=?,fun=?,lul=?,lat=?,lng=? where lat='" + str.Lat + "' and lng='" + str.Lng + "'")
		_, eorr := stmt.Exec(str.Usr, str.Pno, str.Loc, str.Rw, str.Pso, str.Pla, str.Height, str.Pty, str.Opw, str.Lf, str.Earth, str.Phase, str.Fun, str.Lul, str.Lat, str.Lng)
		if eorr != nil {
			logger.Println(eorr)
			io.WriteString(w, "0")
			return
		}
	} else {
		stmt, _ := db.Prepare("insert survey set usr=?,pno=?,loc=?,rw=?,pso=?,pla=?,height=?,pty=?,opw=?,lf=?,earth=?,phase=?,fun=?,lul=?,lat=?,lng=?")
		_, eorr := stmt.Exec(str.Usr, str.Pno, str.Loc, str.Rw, str.Pso, str.Pla, str.Height, str.Pty, str.Opw, str.Lf, str.Earth, str.Phase, str.Fun, str.Lul, str.Lat, str.Lng)
		if eorr != nil {
			logger.Println(eorr)
			io.WriteString(w, "0")
			return
		}
	}

	io.WriteString(w, "1")

}

func deleteLocation(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	str := survey{}
	str.Lat = r.URL.Query().Get("lat")
	str.Lng = r.URL.Query().Get("lng")

	db := dbController.Db

	rows, err := db.Query("delete from survey where lat='" + str.Lat + "' and lng='" + str.Lng + "'")
	defer rows.Close()
	if err != nil {
		io.WriteString(w, "0")
		return
	}
	io.WriteString(w, "1")

}

func getLocation(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db

	rows, err := db.Query("select usr,pno,loc,rw,pso,pla,height,pty,opw,lf,earth,phase,fun,lul,lat,lng from survey")
	defer rows.Close()
	if err != nil {
		io.WriteString(w, "0")
		return
	}
	ans := []survey{}
	for rows.Next() {
		str := survey{}
		var Usr, Pno, Mun, Ward, Loc, Rw, Pso, Pla, Height, Pty, Opw, Lf, Earth, Phase, Fun, Lul, Lat, Lng sql.NullString
		rows.Scan(&Usr, &Pno, &Loc, &Rw, &Pso, &Pla, &Height, &Pty, &Opw, &Lf, &Earth, &Phase, &Fun, &Lul, &Lat, &Lng)
		if Usr.Valid {
			str.Usr = Usr.String
		}
		if Pno.Valid {
			str.Pno = Pno.String
		}
		if Mun.Valid {
			str.Mun = Mun.String
		}
		if Ward.Valid {
			str.Ward = Ward.String
		}
		if Loc.Valid {
			str.Loc = Loc.String
		}
		if Rw.Valid {
			str.Rw = Rw.String
		}
		if Pso.Valid {
			str.Pso = Pso.String
		}
		if Pla.Valid {
			str.Pla = Pla.String
		}
		if Height.Valid {
			str.Height = Height.String
		}
		if Pty.Valid {
			str.Pty = Pty.String
		}
		if Opw.Valid {
			str.Opw = Opw.String
		}
		if Lf.Valid {
			str.Lf = Lf.String
		}

		if Earth.Valid {
			str.Earth = Earth.String
		}
		if Phase.Valid {
			str.Phase = Phase.String
		}
		if Fun.Valid {
			str.Fun = Fun.String
		}
		if Lul.Valid {
			str.Lul = Lul.String
		}

		if Lat.Valid {
			str.Lat = Lat.String
		}
		if Lng.Valid {
			str.Lng = Lng.String
		}
		logger.Println(str)
		ans = append(ans, str)
	}
	a, err := json.Marshal(ans)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {

		w.Write(a)
	}

}

func downloadLocation(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"] == 1 {
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	} else if session.Values["set"] == nil || session.Values["set"] == 0 {
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	db := dbController.Db

	rows, err := db.Query("select usr,pno,loc,rw,pso,pla,height,pty,opw,lf,earth,phase,fun,lul,lat,lng from survey")
	defer rows.Close()
	if err != nil {
		io.WriteString(w, "0")
		return
	}
	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	sty := xlsx.NewStyle()
	sty.Font = *(xlsx.NewFont(16, "Arial Black"))
	sty.ApplyFont = true
	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf(err.Error())
	}
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Usr"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Pno"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Mun"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Ward"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Loc"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Rw"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Pso"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Pla"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Height"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Pty"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Opw"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Lf"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Earth"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Phase"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Fun"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Lul"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Lat"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Lng"
	sty.Font = *(xlsx.NewFont(12, "Arial"))
	for rows.Next() {
		row = sheet.AddRow()
		cell = row.AddCell()
		str := survey{}
		var Usr, Pno, Mun, Ward, Loc, Rw, Pso, Pla, Height, Pty, Opw, Lf, Earth, Phase, Fun, Lul, Lat, Lng sql.NullString
		rows.Scan(&Usr, &Pno, &Loc, &Rw, &Pso, &Pla, &Height, &Pty, &Opw, &Lf, &Earth, &Phase, &Fun, &Lul, &Lat, &Lng)
		if Usr.Valid {
			str.Usr = Usr.String
		}
		if Pno.Valid {
			str.Pno = Pno.String
		}
		if Mun.Valid {
			str.Mun = Mun.String
		}
		if Ward.Valid {
			str.Ward = Ward.String
		}
		if Loc.Valid {
			str.Loc = Loc.String
		}
		if Rw.Valid {
			str.Rw = Rw.String
		}
		if Pso.Valid {
			str.Pso = Pso.String
		}
		if Pla.Valid {
			str.Pla = Pla.String
		}
		if Height.Valid {
			str.Height = Height.String
		}
		if Pty.Valid {
			str.Pty = Pty.String
		}
		if Opw.Valid {
			str.Opw = Opw.String
		}
		if Lf.Valid {
			str.Lf = Lf.String
		}

		if Earth.Valid {
			str.Earth = Earth.String
		}
		if Phase.Valid {
			str.Phase = Phase.String
		}
		if Fun.Valid {
			str.Fun = Fun.String
		}
		if Lul.Valid {
			str.Lul = Lul.String
		}

		if Lat.Valid {
			str.Lat = Lat.String
		}
		if Lng.Valid {
			str.Lng = Lng.String
		}
		cell.SetStyle(sty)
		cell.Value = str.Usr
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Pno
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Mun
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Ward
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Loc
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Rw
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Pso
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Pla
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Height
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Pty
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Opw
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Lf
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Earth
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Phase
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Fun
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Lul
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Lat
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = str.Lng
	}
	err = file.Save("survey.xlsx")
	if err != nil {
		fmt.Printf(err.Error())
	}
	Openfile, err := os.Open("survey.xlsx")
	defer Openfile.Close() //Close after function return
	if err != nil {
		//File not found, send 404
		http.Error(w, "File not found.", 404)
		return
	}

	//File is found, create and send the correct headers

	//Get the Content-Type of the file
	//Create a buffer to store the header of the file in
	FileHeader := make([]byte, 3000000)
	//Copy the headers into the FileHeader buffer
	Openfile.Read(FileHeader)
	//Get content type of file
	//FileContentType := http.DetectContentType(FileHeader)

	//Get the file size
	FileStat, _ := Openfile.Stat()                     //Get info from file
	FileSize := strconv.FormatInt(FileStat.Size(), 10) //Get file size as a string

	//Send the headers
	w.Header().Set("Content-Disposition", "attachment; filename="+"survey.xlsx")
	w.Header().Set("Content-Type", "xlsx")
	w.Header().Set("Content-Length", FileSize)

	//Send the file
	//We read 512 bytes from the file already so we reset the offset back to 0
	Openfile.Seek(0, 0)
	io.Copy(w, Openfile) //'Copy' the file to the client
	return
}

func main() {
	rollbar.Token = rollbarToken
	rollbar.Environment = "Testing"
	defer handlePanic()
	wl, err := net.Dial("udp", "logs3.papertrailapp.com:32240")
	defer wl.Close()
	if remoteLog {
		logger = log.New(wl, "HAVELLS_STREET_COMM: ", log.Lshortfile)
		if err != nil {
			log.Fatal("error")
		}

	} else {
		logger = log.New(os.Stdout, "HAVELLS_STREET_COMM: ", log.Lshortfile)
	}
	Sgu_firmware = make(map[int64][]byte)
	Scu_firmware = make(map[int64][]byte)
	port := os.Getenv("PORT")
	logger.Println("new logger")

	logger.Println("error")
	//port="8000"
	if port == "" {
		logger.Println("$PORT must be set")
	}
	logger.Println("Starting Application")

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	dbController.DbUtilsInit(logger)

	if dbController.DbConnected {
		defer dbController.Db.Close()
	}
	rows, err := dbController.Db.Query("SELECT * from deployment_parameter where deployment_id='" + Deployment_id + "'")
	defer rows.Close()
	for rows.Next() {
		var did string
		rows.Scan(&did, &per_scu_delay, &scu_polling, &scu_scheduling, &scu_retry_delay, &scu_max_retry, &ack_delay)
		//set default values if null
		logger.Println(scu_polling)
		if len(per_scu_delay) == 0 {
			per_scu_delay = "5"
		}
		if len(scu_polling) == 0 {
			scu_polling = "300"
		}
		if len(scu_scheduling) == 0 {
			scu_scheduling = "15"
		}
		if len(scu_retry_delay) == 0 {
			scu_retry_delay = "59"
		}
		if len(scu_max_retry) == 0 {
			scu_max_retry = "5"
		}
	}

	sguConnectionChan = make(chan net.Conn, sguChanSize)
	LampControllerChannel = make(chan sguUtils.SguUtilsLampControllerStruct, lampControllerChansize)
	//energysguutilChannel = make(chan  sguUtils.SguUtilsEnergyCntrStruct, lampControllerChansize)
	SendSMSChan = make(chan string, SendSMSChanSize)

	configure.InitConfigure(LampControllerChannel, dbController, logger)
	NBApis.InitNBApis(LampControllerChannel, dbController, logger)
	tcpUtils.Init(dbController, logger)
	sguUtils.Init(logger)

	sguUtils.Config_Params(per_scu_delay, scu_polling, scu_retry_delay, scu_max_retry)
	configure.Config_Params(scu_scheduling, &per_scu_delay)

	mapview.InitMapview(dbController, logger)
	//go report.InitSendreport(dbController)
	report.InitReport(dbController, logger)
	report.InitSendreport(dbController)
	go report.ReportGenThread()
	//go LampController.InitTcpUtilsStruct()

	logger.Println("Strting TCP server")

	go tcpServer.StartTcpServer(sguConnectionChan, logger)

	HandleSguConnectionsDone := sguUtils.HandleSguConnections(sguConnectionChan, dbController, SendSMSChan, maxNumScusPerSgu)

	HandleLampEventsDone := sguUtils.HandleLampEvents(LampControllerChannel)
	//HandleEnergyEventsDone := sguUtils.HandleEnergyEvents(energysguutilChannel)
	StartSendSMSThreadDone := configure.StartSendSMSThread(SendSMSChan)
	logger.Println("TCP server started successfully")

	chttp.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/", my)

	logger.Println("directory Set")

	//http.HandleFunc("/login", login)
	http.HandleFunc("/adminlogin", adminlogin)

	logger.Println("Login Set")

	http.HandleFunc("/configure/updatedeploymentparameter", configure.Updatedeploymentparameter)
	http.HandleFunc("/configure/getdeploymentparameter", configure.Getdeploymentparameter)
	http.HandleFunc("/deleteuserperson", deleteuserperson)
	http.HandleFunc("/alluserview", alluserview)
	http.HandleFunc("/setrepo", setrepo)
	http.HandleFunc("/getconnectedrepo", getconnectedrepo)
	http.HandleFunc("/getallbranchesforRepo", getallbranchesforRepo)
	http.HandleFunc("/getallrepo", getallrepo)
	http.HandleFunc("/gitconnected", gitconnected)
	http.HandleFunc("/callback", callback)
	http.HandleFunc("/connectrepo", connectrepo)
	http.HandleFunc("/update_firmware", downloadSguRepo)
	http.HandleFunc("/update_firmware_scu", downloadScuRepo)
	http.HandleFunc("/deletealertperson", deletealertperson)
	http.HandleFunc("/deletereportperson", deletereportperson)
	http.HandleFunc("/getreportperson", getreportperson)
	http.HandleFunc("/getalertperson", getalertperson)
	http.HandleFunc("/sguota", sguotadisplay)
	http.HandleFunc("/scuota", scuotadisplay)
	http.HandleFunc("/LampControl", LampControl)
	http.HandleFunc("/AllLampControl", AllLampControl)
	http.HandleFunc("/AllLampControlpwm", AllLampControlpwm)
	logger.Println("Lampcontroller Set")
	http.HandleFunc("/locationadd", LocationAdd)
	logger.Println("Location Details are adding here!")
	http.HandleFunc("/sguadd", sguAdd)
	logger.Println("sgu Details are adding here!")
	http.HandleFunc("/scuadd", scuAdd)
	logger.Println("scu Details are adding here!")
	http.HandleFunc("/LocationNames", getLocationNames)
	logger.Println("Location Names are getting here!")
	http.HandleFunc("/configure/scuconfigure", configure.Scuconfigure)
	http.HandleFunc("/configure/scuview", configure.Scuview)
	http.HandleFunc("/configure/scuadd", configure.Scuadd)
	http.HandleFunc("/configure/scusave", configure.Scusave)
	http.HandleFunc("/configure/sguconfigure", configure.Sguconfigure)
	http.HandleFunc("/configure/sguadd", configure.Sguadd)
	http.HandleFunc("/AddSchedule", AddSchedule)
	http.HandleFunc("/ViewSchedule", ViewSchedule)
	http.HandleFunc("/AddEnergyParameter", AddEnergyParameter)
	http.HandleFunc("/Communparams", Communparams)
	http.HandleFunc("/Polingparams", Polingparams)
	http.HandleFunc("/configure/updateinven", configure.Updateinventories)
	http.HandleFunc("/configure/AddInventories", configure.AddInventory)
	http.HandleFunc("/configure/viewinventories", configure.Viewinventories)
	http.HandleFunc("/configure/sgusave", configure.Sgusave)
	http.HandleFunc("/showmap", mapview.Showmap)
	http.HandleFunc("/configure/viewschedules", configure.View)
	http.HandleFunc("/report", report.Report)
	http.HandleFunc("/auth", auth)
	http.HandleFunc("/isadmin", isAdmin)
	http.HandleFunc("/signout", signout)
	http.HandleFunc("/getuid", getUid)
	http.HandleFunc("/configure/graph", configure.Graph)
	http.HandleFunc("/configure/plot", configure.Plot)
	http.HandleFunc("/configure/subscribe", configure.Subscribe)
	http.HandleFunc("/configure/addzone", configure.Addzone)
	http.HandleFunc("/configure/zoneconfigure", configure.Zoneconfigure)
	http.HandleFunc("/configure/zonesguview", configure.Zonesguview)
	http.HandleFunc("/configure/zonesguadd", configure.Zoneadd)
	http.HandleFunc("/configure/zonesgusave", configure.Zonesgusave)
	http.HandleFunc("/configure/zoneconfiguresc", configure.Zoneconfiguresc)
	http.HandleFunc("/configure/zoneaddsc", configure.Zoneaddsc)
	http.HandleFunc("/configure/zonesavesc", configure.Zonesavesc)
	http.HandleFunc("/configure/zonesguremove", configure.Zonesguremove)
	http.HandleFunc("/configure/zonesgusaver", configure.Zonesgusaver)
	http.HandleFunc("/configure/removezone", configure.Removezone)
	http.HandleFunc("/configure/updatezone", configure.Updatezone)
	http.HandleFunc("/configure/sguview", configure.Sguview)
	http.HandleFunc("/configure/zoneview", configure.Zoneview)
	http.HandleFunc("/configure/csv", configure.Csv)
	http.HandleFunc("/configure/getsculoc", configure.Getsculoc)
	http.HandleFunc("/configure/updatesculoc", configure.Updatesculoc)
	http.HandleFunc("/configure/adduser", configure.Adduser)
	http.HandleFunc("/mapview/getzone", mapview.Getzone)
	http.HandleFunc("/mapview/getall", mapview.Getall)
	http.HandleFunc("/supportCP", supportCP)
	http.HandleFunc("/supportWTS", supportDefinition)
	http.HandleFunc("/saveLocation", saveLocation)
	http.HandleFunc("/deleteLocation", deleteLocation)
	http.HandleFunc("/getLocation", getLocation)
	http.HandleFunc("/downloadLocation", downloadLocation)

	http.HandleFunc("/configure/addgroup", configure.Addgroup)
	http.HandleFunc("/configure/groupconfigure", configure.Groupconfigure)
	http.HandleFunc("/configure/groupscuview", configure.Groupscuview)
	http.HandleFunc("/configure/groupscuadd", configure.Groupadd)
	http.HandleFunc("/configure/groupscusave", configure.Groupscusave)
	http.HandleFunc("/configure/groupscuremove", configure.Groupscuremove)
	http.HandleFunc("/configure/groupscusaver", configure.Groupscusaver)
	http.HandleFunc("/configure/removegroup", configure.Removegroup)
	http.HandleFunc("/configure/updategroup", configure.Updategroup)
	http.HandleFunc("/mapview/getgroup", mapview.Getgroup)
	http.HandleFunc("/configure/groupconfiguresc", configure.Groupconfiguresc)
	http.HandleFunc("/configure/groupaddsc", configure.Groupaddsc)
	http.HandleFunc("/configure/groupsavesc", configure.Groupsavesc)
	http.HandleFunc("/configure/groupview", configure.Groupview)
	//NB Services
	http.HandleFunc("/login", NBApis.NBlogin)
	http.HandleFunc("/system/street_lamp", NBApis.StreetLampControll)
	http.HandleFunc("/system", NBApis.Discovery)
	logger.Println("Starting HTTP Server")
	//downloadSguRepo()
	err = http.ListenAndServe(":"+port, context.ClearHandler(http.DefaultServeMux))

	//err = http.ListenAndServeTLS(":"+port, "./keys/server.pem", "./keys/server.key", nil)

	if err != nil {
		logger.Println("Failed to start server")
		logger.Print(err.Error())
	}
	close(StartSendSMSThreadDone)
	close(HandleSguConnectionsDone)
	close(HandleLampEventsDone)
	//close(HandleEnergyEventsDone)

}
