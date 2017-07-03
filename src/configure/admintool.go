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
	"encoding/json"
	"io"
	"net/http"

	_ "github.com/go-sql-driver/mysql"

	"sguUtils"
	"userMgmt"
)

type scu struct {
	Name string `json:"name"`
	Lat  string `json:"lat"`
	Lng  string `json:"lng"`
	Id   string `json:"id"`
}

func Getsculoc(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select scu_id,location_name,location_lat,location_lng from scu")
	chkErr(err, &w)
	defer rows.Close()

	data := []scu{}
	for rows.Next() {
		var id, name, lat, lng string
		rows.Scan(&id, &name, &lat, &lng)
		data = append(data, scu{name, lat, lng, id})
	}

	if a, err := json.Marshal(data); err != nil {
		logger.Println("Error in json.Marshal", err)
	} else {
		w.Write(a)
	}
}

func Updatesculoc(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")
	lat := r.URL.Query().Get("lat")
	lng := r.URL.Query().Get("lng")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	stmt, err := db.Prepare("Update scu SET location_name=?,location_lat=?,location_lng=? where scu_id='" + id + "'")
	chkErr(err, &w)
	defer stmt.Close()

	_, eorr := stmt.Exec(name, lat, lng)
	chkErr(eorr, &w)
	if eorr == nil {
		io.WriteString(w, "done")
	}
}

type parameter struct {
	//Id string `json:"id"`
	Deployment_id          string `json:"deployment_id"`
	Scu_onoff_pkt_delay    string `json:"scu_onoff_pkt_delay"`
	Scu_poll_delayd        string `json:"scu_poll_delay"`
	Scu_schedule_pkt_delay string `json:"scu_schedule_pkt_delay"`
	Scu_onoff_retry_delay  string `json:"scu_onoff_retry_delay"`
	Scu_max_retry          string `json:"scu_max_retry"`
	Server_pkt_ack_delay   string `json:"server_pkt_ack_delay"`
}

func Getdeploymentparameter(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select deployment.deployment_id,deployment_parameter.scu_onoff_pkt_delay," +
		"deployment_parameter.scu_poll_delay,deployment_parameter.scu_schedule_pkt_delay," +
		"deployment_parameter.scu_onoff_retry_delay,deployment_parameter.scu_max_retry, " +
		"deployment_parameter.server_pkt_ack_delay from deployment left join deployment_parameter on d" +
		"eployment.deployment_id=deployment_parameter.deployment_id")
	chkErr(err, &w)
	defer rows.Close()

	data := []parameter{}
	for rows.Next() {
		var deployment_id, scu_onoff_pkt_delay, scu_poll_delay, scu_schedule_pkt_delay,
			scu_onoff_retry_delay, scu_max_retry, server_pkt_ack_delay string
		rows.Scan(&deployment_id, &scu_onoff_pkt_delay, &scu_poll_delay, &scu_schedule_pkt_delay,
			&scu_onoff_retry_delay, &scu_max_retry, &server_pkt_ack_delay)

		data = append(data, parameter{deployment_id, scu_onoff_pkt_delay, scu_poll_delay,
			scu_schedule_pkt_delay, scu_onoff_retry_delay, scu_max_retry, server_pkt_ack_delay})
	}

	if a, err := json.Marshal(data); err != nil {
		logger.Println("Error in json.Marshal: ", err)
	} else {
		w.Write(a)
	}
}

func Updatedeploymentparameter(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
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
	stmt, _ := db.Prepare("INSERT INTO deployment_parameter (deployment_id, scu_onoff_pkt_delay, " +
		"scu_poll_delay,scu_schedule_pkt_delay,scu_onoff_retry_delay,scu_max_retry,server_pkt_ack_delay) VALUES ('" +
		deployment_id + "','" + scu_onoff_pkt_delay + "','" + scu_poll_delay + "','" + scu_schedule_pkt_delay + "','" +
		scu_onoff_retry_delay + "','" + scu_max_retry + "','" + server_pkt_ack_delay + "') ON DUPLICATE KEY UPDATE " +
		"scu_onoff_pkt_delay='" + scu_onoff_pkt_delay + "', scu_poll_delay='" + scu_poll_delay + "', " +
		"scu_schedule_pkt_delay='" + scu_schedule_pkt_delay + "', scu_onoff_retry_delay='" + scu_onoff_retry_delay +
		"', scu_max_retry='" + scu_max_retry + "', server_pkt_ack_delay='" + server_pkt_ack_delay + "'")
	defer stmt.Close()
	_, err := stmt.Exec()

	if err == nil {
		*per_scu_delay = scu_onoff_pkt_delay
		sguUtils.Config_Params(scu_onoff_pkt_delay, scu_poll_delay, scu_onoff_retry_delay, scu_max_retry)
		scu_scheduling = scu_schedule_pkt_delay
		io.WriteString(w, "done")
	} else {
		io.WriteString(w, "error")
	}
}

func Adduser(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	email1 := r.URL.Query().Get("userid")
	pass1 := r.URL.Query().Get("pass")
	admin := r.URL.Query().Get("admin")
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	statement1 := "SELECT CAST(AES_DECRYPT(password,'234FHF?#@$#%%jio4323486') AS CHAR(10000) " +
		"CHARACTER SET utf8 ) AS password FROM login where user_email= AES_ENCRYPT('" + email1 +
		"','234FHF?#@$#%%jio4323486');"
	stmt, err := db.Query(statement1)
	chkErr(err, &w)
	defer stmt.Close()

	if stmt.Next() {
		io.WriteString(w, "already")
		return
	} else {
		stmt1, err := db.Prepare("insert login set user_email=AES_ENCRYPT('" + email1 +
			"','234FHF?#@$#%%jio4323486'),password=AES_ENCRYPT('" + pass1 + "','234FHF?#@$#%%jio4323486'),admin_op='" +
			admin + "'  ;")
		defer stmt1.Close()
		logger.Println("insert login set user_email=AES_ENCRYPT('" + email1 +
			"','234FHF?#@$#%%jio4323486'),password=AES_ENCRYPT('" + pass1 + "','234FHF?#@$#%%jio4323486'),admin_op='" +
			admin + "'  ;")
		_, eorr := stmt1.Exec()
		if eorr == nil {
			logger.Println(err)
			io.WriteString(w, "done")
		}
	}
}
