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
	"configure"
	"dbUtils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mapview"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"report"
	"sguUtils"
	"strconv"
	"strings"
	"tcpServer"
	"tcpUtils"
	"time"

	"github.com/context"
	"github.com/jwt-go-master"
	"github.com/scorredoira/email"
	"github.com/sessions"
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
	SendSMSChan chan string
	logger      *log.Logger
	tokenMap    (map[string]int)
)

const (
	MaxNumSGUs             = 1024
	sguChanSize            = 16
	SendSMSChanSize        = 8
	maxNumScusPerSgu       = 100
	lampControllerChansize = maxNumScusPerSgu * 4
	remoteLog              = false
	aesPassword            = "234FHF?#@$#%%jio4323486"
)

type names struct {
	Name []string
}

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
	//	if strings.Compare(r.Form["admin_op"][0], "as_admin") == 0 {
	//		admin_op = 1
	//	} else {
	//		admin_op = 0
	//	}

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

	exp += "&&(T>=" + timestr + "&&T<=" + timeen + ")"
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

type NBFdn struct {
	System      string `json:"system"`
	Gateway     string `json:"gateway"`
	Street_lamp string `json:"street_lamp"`
}
type NBData struct {
	Brightness string `json:"brightness"`
	Message    string `json:"msg"`
	Token      string `json:token`
	Email      string `json:email`
}
type NBAllLampControlStruct struct {
	Token  string `json:"token"`
	Object string `json:"object"`
	Fdn    NBFdn  `json:"fdn"`
	Opr    string `json:"opr"`
	Data   NBData `json:"data"`
}
type NBResponseStruct struct {
	Response_status string `json:"response_status"`
	Data            NBData `json:"data"`
}

//NB Street Lamp Controll
func StreetLampControll(w http.ResponseWriter, r *http.Request) {
	var ans NBResponseStruct

	parse_err := r.ParseForm()
	if parse_err != nil {
		logger.Println(parse_err)
	}
	var NBLampStr NBAllLampControlStruct
	if len(r.FormValue("token")) == 0 {
		decoder := json.NewDecoder(r.Body)
		logger.Println(decoder)
		err := decoder.Decode(&NBLampStr)
		if err != nil {
			logger.Println(err)
		}
	} else {
		NBLampStr.Token = r.FormValue("token")
		NBLampStr.Data.Brightness = r.FormValue("data.brightness")
		NBLampStr.Fdn.System = r.FormValue("fdn.system")
		NBLampStr.Object = r.FormValue("object")
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	l_object := NBLampStr.Object
	l_opr := NBLampStr.Opr
	l_system := NBLampStr.Fdn.System
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	//l_data := NBLampStr.Data
	l_sgu := NBLampStr.Fdn.Gateway
	l_scu := NBLampStr.Fdn.Street_lamp
	l_event := NBLampStr.Data.Brightness
	if !validateSGU(l_sgu) {
		fmt.Println("SGU len", len(l_sgu))
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid of SGU"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if !validateSCU(l_scu) {
		fmt.Println("SCU len", len(l_scu))

		ans.Response_status = "fail"
		ans.Data.Message = "Invalid SCU"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if l_system == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "System Not Specified"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if l_object == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Object Not Specified"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if l_opr == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Operation Not Specified"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}

	if l_brightness == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Brightness Not Specified"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_brightness_i, _ := strconv.Atoi(l_brightness)

	if l_brightness_i < 0 || l_brightness_i > 10 {
		ans.Response_status = "fail"
		ans.Data.Message = "brightness is not in range"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		var LampController sguUtils.SguUtilsLampControllerStruct

		logger.Println(r.URL)
		u, err := url.Parse(r.URL.String())
		logger.Println(u)
		logger.Println(u.RawQuery)
		LampController.SGUID, err = strconv.ParseUint(l_sgu, 10, 64)
		if err != nil {
			logger.Println("Invalid SGUID" + l_sgu + " specified")
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SGUID"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		LampController.SCUID, err = strconv.ParseUint(l_scu, 10, 64)
		if err != nil {
			logger.Println("Invalid SCUID" + l_scu + " specified")
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SCUID"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		LampController.LampEvent, err = strconv.Atoi(l_event)
		if err != nil {
			logger.Println("Invalid lamp contral val  " + l_event + " specified")
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid lamp contral val"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}

			return
		}
		/*m, _ := url.ParseQuery(u.RawQuery)
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
			}*/
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
		ans.Response_status = "success"
		ans.Data.Message = ""
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid token "
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	//}
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
func validateSGU(p_sgu string) bool {
	if len(p_sgu) == 14 {
		return true
	} else {
		return false
	}
}
func validateSCU(p_scu string) bool {
	if len(p_scu) == 16 {
		return true
	} else {
		return false
	}
}
func TokenParse_errorChecking(myToken string) (string, bool) {
	var uid string
	/*if tokenMap[myToken]!=1{
		logger.Println("InValid Token")
		return uid,false
	}*/
	token, err := jwt.Parse(myToken, func(token *jwt.Token) (interface{}, error) {
		return ([]byte)("234F4323486HF?#@$MAZE"), nil
	})
	if err != nil {
		logger.Println("This token has invalid number of segments:", err)
		return uid, false
	}
	if token.Valid {
		uid = token.Claims["uid"].(string)
		logger.Println(uid)
		logger.Println("Valid Token")
		return uid, true
	} else if ve, ok := err.(*jwt.ValidationError); ok {
		if ve.Errors&jwt.ValidationErrorMalformed != 0 {
			logger.Println("That's not even a token")
			return uid, false
		} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
			// Token is either expired or not active yet
			logger.Println("Token Expired")
			return uid, false
		} else {
			logger.Println("Couldn't handle this token:", err)
			return uid, false
		}
	} else {
		logger.Println("Couldn't handle this token:", err)
		return uid, false
	}

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
func NBLampControl(w http.ResponseWriter, r *http.Request) {
	parse_err := r.ParseForm()
	if parse_err != nil {
		logger.Println(parse_err)
	}
	/*	session, _ := store.Get(r, "auth")
		logger.Println(session.Values["set"])
		if session.Values["set"] == 1 {
			http.Redirect(w, r, "adminlogin.html", http.StatusFound)
			return
		} else if session.Values["set"] == nil || session.Values["set"] == 0 {
			http.Redirect(w, r, "login.html", http.StatusFound)
			return
		}*/

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

//struct for json input.
type NBlogiStr struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

//struct for NB Login json output.
type sendlogin struct {
	Response_status string `json:"response_status"`
	Token           string `json:"token"`
	Netpassphrase   string `json:"networkpassphrase"`
	Isadmin         bool   `json:"isadmin"`
	Ismaster        bool   `json:"ismaster"`
	Name            string `json:"name"`
	Email           string `json:"email"`
}

// Login for North Bound Api.
func NBlogin(w http.ResponseWriter, r *http.Request) {
	tokenMap = make(map[string]int)
	r.ParseForm()
	var t NBlogiStr
	if len(r.FormValue("username")) == 0 {
		decoder := json.NewDecoder(r.Body)
		logger.Println(decoder)
		err := decoder.Decode(&t)
		if err != nil {
			logger.Println(err)
		}
	} else {
		t.Username = r.FormValue("username")
		t.Password = r.FormValue("password")
	}

	username := t.Username
	password := t.Password
	ans := NBResponseStruct{}
	if !validateEmail(username) {
		fmt.Println("Email address is invalid")
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Email Id"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		logger.Println("response status", ans.Response_status)
		return
	}
	//session, _ := store.Get(r, "authmaze")

	var b_u_name, b_pwd []byte
	if dbController.DbConnected {

		//row, err := db.Query("select user_email, password from login where user_email=?",1)
		logger.Println("User name", username)

		statement := "SELECT user_email,password FROM login where user_email='" + username + "'"

		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			logger.Println("Error quering database  for login information")
			logger.Println(err)
			ans.Response_status = "fail"
			ans.Data.Message = "Something went wrong!!"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			logger.Println("response status", ans.Response_status)
			return
		} else {

			for rows.Next() {
				rows.Scan(&b_u_name, &b_pwd)
			}

			rows.Close()
		}
	}
	s_username := string(b_u_name)
	s_password := string(b_pwd)

	if username != s_username {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Email Id"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		logger.Println("response status", ans.Response_status)
		return
	} else if s_password != password {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Password"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		logger.Println("response status", ans.Response_status)
		return
	}
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["uid"] = username
	rand.Seed(time.Now().Unix())
	token.Claims["rand"] = rand.Float64()
	tokenString, err := token.SignedString(([]byte)("234F4323486HF?#@$MAZE"))
	if err != nil {
		ans.Response_status = "fail"
		ans.Data.Message = "Something Went Wrong"
		logger.Println(err.Error())
		logger.Println("response status", ans.Response_status)
		return
	}
	fmt.Println("tokenString", tokenString)
	tokenMap[tokenString] = 1
	logger.Println(tokenString)
	logger.Println("Matching entry found. Redirecting")
	ans.Response_status = "success"
	ans.Data.Token = tokenString
	ans.Data.Email = username
	a, err := json.Marshal(ans)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	} else {
		w.Write(a)
	}
	logger.Println("response status", ans.Response_status)
	return
}
func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
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

func main() {
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

	sguConnectionChan = make(chan net.Conn, sguChanSize)
	LampControllerChannel = make(chan sguUtils.SguUtilsLampControllerStruct, lampControllerChansize)
	//energysguutilChannel = make(chan sguUtils.SguUtilsEnergyCntrStruct, lampControllerChansize)
	SendSMSChan = make(chan string, SendSMSChanSize)

	configure.InitConfigure(LampControllerChannel, dbController, logger)
	NBApis.InitNBApis(LampControllerChannel, dbController, logger)
	tcpUtils.Init(dbController, logger)
	sguUtils.Init(logger)
	mapview.InitMapview(dbController, logger)
	//go report.InitSendreport(dbController)
	report.InitReport(dbController, logger)
	report.InitSendreport(dbController)
	go report.ReportGenThread()
	//go LampController.InitTcpUtilsStruct()

	logger.Println("Strting TCP server")

	// Removed `go`
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

	http.HandleFunc("/LampControl", LampControl)
	http.HandleFunc("/AllLampControl", AllLampControl)
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
