/********************************************************************
 * FileName:     admintool.go
 * Project:      Havells StreetComm
 * Module:       admintool
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure

import (
//"fmt"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"io"
	"sguUtils"
)
type scu struct {
	Name string `json:"name"`
	Lat string `json:"lat"`
	Lng string `json:"lng"`
	Id   string `json:"id"`
}

func Getsculoc(w http.ResponseWriter, r *http.Request){
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
	rows,err:=db.Query("Select scu_id,location_name,location_lat,location_lng from scu")
	defer rows.Close()
	chkErr(err,&w)
	data := []scu{}
	for rows.Next(){
		var id,name,lat,lng string
		rows.Scan(&id,&name,&lat,&lng)
		tm:=scu{}
		tm.Name=name
		tm.Lat=lat
		tm.Lng=lng
		tm.Id=id
		data = append(data,tm)
	}
	a, err := json.Marshal(data)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	}  else {
		//logger.Println(a)
		w.Write(a)

	}
}
func Updatesculoc(w http.ResponseWriter, r *http.Request){
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
	lat := r.URL.Query().Get("lat")
	lng := r.URL.Query().Get("lng")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	stmt, err := db.Prepare("Update scu SET location_name=?,location_lat=?,location_lng=? where scu_id='"+id+"'")
	defer stmt.Close()
	chkErr(err,&w)
	_, eorr:=stmt.Exec(name,lat,lng)
	chkErr(eorr,&w)
	if(eorr==nil){
		io.WriteString(w,"done")
	}
	defer stmt.Close()
}
//--------------------------------------------------------------------------------------------------------------------------------------------------------------

type parameter struct {
	//Id string `json:"id"`
	Deployment_id string `json:"deployment_id"`
	Scu_onoff_pkt_delay string `json:"scu_onoff_pkt_delay"`
	Scu_poll_delayd   string `json:"scu_poll_delay"`
	Scu_schedule_pkt_delay   string `json:"scu_schedule_pkt_delay"`
	Scu_onoff_retry_delay   string `json:"scu_onoff_retry_delay"`
	Scu_max_retry   string `json:"scu_max_retry"`
	Server_pkt_ack_delay   string `json:"server_pkt_ack_delay"`
}

func Getdeploymentparameter(w http.ResponseWriter, r *http.Request){
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
	rows,err:=db.Query("Select deployment.deployment_id,deployment_parameter.scu_onoff_pkt_delay,deployment_parameter.scu_poll_delay,deployment_parameter.scu_schedule_pkt_delay,deployment_parameter.scu_onoff_retry_delay,deployment_parameter.scu_max_retry, deployment_parameter.server_pkt_ack_delay from deployment left join deployment_parameter on deployment.deployment_id=deployment_parameter.deployment_id")
	defer rows.Close()
	chkErr(err,&w)
	data := []parameter{}
	for rows.Next(){
		var deployment_id,scu_onoff_pkt_delay,scu_poll_delay,scu_schedule_pkt_delay,scu_onoff_retry_delay,scu_max_retry,server_pkt_ack_delay string
		rows.Scan(&deployment_id,&scu_onoff_pkt_delay,&scu_poll_delay,&scu_schedule_pkt_delay,&scu_onoff_retry_delay,&scu_max_retry,&server_pkt_ack_delay)
		tm:=parameter{}
		//tm.Id=id
		tm.Deployment_id=deployment_id
		tm.Scu_onoff_pkt_delay=scu_onoff_pkt_delay
		tm.Scu_poll_delayd=scu_poll_delay
		tm.Scu_schedule_pkt_delay=scu_schedule_pkt_delay
		tm.Scu_onoff_retry_delay=scu_onoff_retry_delay
		tm.Scu_max_retry=scu_max_retry
		tm.Server_pkt_ack_delay=server_pkt_ack_delay
		data = append(data,tm)
	}
	a, err := json.Marshal(data)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	}  else {
		//logger.Println(a)
		w.Write(a)

	}
}

func Updatedeploymentparameter(w http.ResponseWriter, r *http.Request){
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "../adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "../login.html", http.StatusFound)
		return
	}

	deployment_id := r.URL.Query().Get("deployment_prm_id")
	scu_onoff_pkt_delay := r.URL.Query().Get("deployment_pkt_delay")
	scu_poll_delay := r.URL.Query().Get("deployment_pol_delay")
	scu_schedule_pkt_delay := r.URL.Query().Get("deployment_sch_delay")
	scu_onoff_retry_delay := r.URL.Query().Get("deployment_rtry_delay")
	scu_max_retry := r.URL.Query().Get("deployment_max_try")
	server_pkt_ack_delay := r.URL.Query().Get("server_pkt_ack_delay")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	//stmt, err := db.Prepare("Update deployment_parameter SET scu_onoff_pkt_delay=?,scu_poll_delay=?,scu_schedule_pkt_delay=?,scu_onoff_retry_delay=?,scu_max_retry=? where deployment_id='"+deployment_id+"'")
	stmt,_:= db.Prepare("INSERT INTO deployment_parameter (deployment_id, scu_onoff_pkt_delay, scu_poll_delay,scu_schedule_pkt_delay,scu_onoff_retry_delay,scu_max_retry,server_pkt_ack_delay) VALUES ('"+deployment_id+"','"+ scu_onoff_pkt_delay+"','"+scu_poll_delay+"','"+scu_schedule_pkt_delay+"','"+scu_onoff_retry_delay+"','"+scu_max_retry+"','"+server_pkt_ack_delay+"') ON DUPLICATE KEY UPDATE scu_onoff_pkt_delay='"+scu_onoff_pkt_delay+"', scu_poll_delay='"+scu_poll_delay+"', scu_schedule_pkt_delay='"+scu_schedule_pkt_delay+"', scu_onoff_retry_delay='"+scu_onoff_retry_delay+"', scu_max_retry='"+scu_max_retry+"', server_pkt_ack_delay='"+server_pkt_ack_delay+"'")
	_,err:=stmt.Exec();
	defer stmt.Close()
	if(err==nil){
		*per_scu_delay=scu_onoff_pkt_delay
		sguUtils.Config_Params(scu_onoff_pkt_delay ,scu_poll_delay,scu_onoff_retry_delay,scu_max_retry)
		scu_scheduling=scu_schedule_pkt_delay
		io.WriteString(w,"done")
	}else{
		io.WriteString(w,"error")
	}
	defer stmt.Close()
}


//-----------------------------------------------------------------------------------------------------------------------------------------------------------------
func Adduser(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "auth")
	logger.Println(session.Values["set"])
	if session.Values["set"]==1{
		http.Redirect(w, r, "adminlogin.html", http.StatusFound)
		return
	}else if session.Values["set"]==nil || session.Values["set"]==0{
		http.Redirect(w, r, "login.html", http.StatusFound)
		return
	}
	email1:=r.URL.Query().Get("userid")
	pass1:=r.URL.Query().Get("pass")
	admin:=r.URL.Query().Get("admin")
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	//statement := "select password  from login where user_email='"+email1+"'"
	statement1 :=  "SELECT CAST(AES_DECRYPT(password,'234FHF?#@$#%%jio4323486') AS CHAR(10000) CHARACTER SET utf8 ) AS password FROM login where user_email= AES_ENCRYPT('"+email1+"','234FHF?#@$#%%jio4323486');"
	stmt, err := db.Query(statement1)
	defer stmt.Close()
	chkErr(err,&w)
	if stmt.Next(){
		io.WriteString(w, "already")
		return
	}else{
				stmt1, err := db.Prepare("insert login set user_email=AES_ENCRYPT('"+email1+"','234FHF?#@$#%%jio4323486'),password=AES_ENCRYPT('"+pass1+"','234FHF?#@$#%%jio4323486'),admin_op='"+admin+"'  ;")
				defer stmt1.Close()
				logger.Println("insert login set user_email=AES_ENCRYPT('"+email1+"','234FHF?#@$#%%jio4323486'),password=AES_ENCRYPT('"+pass1+"','234FHF?#@$#%%jio4323486'),admin_op='"+admin+"'  ;")
				_, eorr:=stmt1.Exec()
				if eorr==nil{
					logger.Println(err)
					io.WriteString(w,"done")
				}

	}
}
