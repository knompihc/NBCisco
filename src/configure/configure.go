/********************************************************************
 * FileName:     configure.go
 * Project:      Havells StreetComm
 * Module:       configure
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure
import (
	//"fmt"
	"log"
	"net/http"
	"io"
	"strconv"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"sguUtils"
	"dbUtils"
	"time"
	"github.com/sessions"
)
var store = sessions.NewCookieStore([]byte("something-very-secret"))
var LampController	sguUtils.SguUtilsLampControllerStruct
var LampControllerChannel	chan   sguUtils.SguUtilsLampControllerStruct
var dbController dbUtils.DbUtilsStruct
var logger *log.Logger
var scu_scheduling string
var per_scu_delay *string
func InitConfigure(LampConChannel	chan   sguUtils.SguUtilsLampControllerStruct,dbcon dbUtils.DbUtilsStruct,logg *log.Logger){
	logger=logg
	LampControllerChannel =LampConChannel
	dbController =dbcon
}
func Config_Params(scu_sch string,perscu_delay *string){
	scu_scheduling=scu_sch+"s";
	per_scu_delay=perscu_delay;
}
/*func main(){
	logger.Println("starting server on http://localhost:8888/\nvalue is")
	//http.HandleFunc("/", IndexHandler)
	http.Handle("/", http.FileServer(http.Dir("ht")))
	http.HandleFunc("/scuconfigure", scuconfigure)
	http.HandleFunc("/scuview", scuview)
	http.HandleFunc("/scuadd", scuadd)
	http.HandleFunc("/scusave", scusave)
	http.HandleFunc("/sguconfigure", sguconfigure)
	http.HandleFunc("/sguadd", sguadd)
	http.HandleFunc("/sgusave", sgusave)
	http.ListenAndServe(":8000", nil)
}
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./ht/"))))
}*/
func Scuconfigure(w http.ResponseWriter, r *http.Request){
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

	rows,err :=db.Query("Select scu_id,location_name from scu limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)

	cn:=0
	script+="<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;' >SCU Location</th><th style='text-align: center;'>SCU ID</th><th style='text-align: center;'>View Attached Schedules</th><th style='text-align: center;'>Attach Schedules</th></tr></thead><tbody>"
	for rows.Next(){
		var loc string
		var scuid int
		err=rows.Scan(&scuid,&loc)
		chkErr(err,&w)
		if cn<=9 {
			script += "<tr><td>" + loc + "</td><td>" + strconv.Itoa(scuid) + "</td><td><a href='#' id='vc_" + strconv.Itoa(scuid) + "' class='btn btn-info btn-sm viewc'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td>  <a href='#' id='ac_" + strconv.Itoa(scuid) + "' class='btn btn-success btn-sm addc'><span class='glyphicon glyphicon-check'></span> Add</a><div class='vic hidden' id='vidc_" + strconv.Itoa(scuid) + "'></td></tr>"
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
func Scuview(w http.ResponseWriter, r *http.Request){
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
	rows,err :=db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID from scuconfigure where ScuID='"+id+"'")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var schid,pwm int
		var from,to,pri string
		err=rows.Scan(&schid,&from,&to,&pwm,&pri)
		chkErr(err,&w)
		sd:=strings.Split(from," ")
		ed:=strings.Split(to," ")
		if len(from)!=0&&len(to)!=0{
			script+="<tr><td>"+pri+"</td><td>"+strconv.Itoa(schid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else if len(from)!=0{
			script+="<tr><td>"+pri+"</td><td>"+strconv.Itoa(schid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else if len(to)!=0{
			script+="<tr><td>"+pri+"</td><td>"+strconv.Itoa(schid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else{
			script+="<tr><td>"+pri+"</td><td>"+strconv.Itoa(schid)+"</td><td>" + "" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}
	}
	script+="</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Scuadd(w http.ResponseWriter, r *http.Request){
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
	rows,err :=db.Query("Select idschedule,ScheduleStartTime,ScheduleEndTime,pwm from schedule where idschedule not in (select distinct ScheduleID from scuconfigure where ScuID='"+id+"')")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scheduling Priority</th><th>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var schid,pwm int
		var from,to string
		err=rows.Scan(&schid,&from,&to,&pwm)
		chkErr(err,&w)
		sd:=strings.Split(from," ")
		ed:=strings.Split(to," ")
		if len(from)!=0&&len(to)!=0{
			script+="<tr><td style='text-align:center''><input type='checkbox' name='addc_"+id+"' value='"+strconv.Itoa(schid)+"'></td>"
			script+="<td><select class='form-control' name='pric_"+strconv.Itoa(schid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
			script+="<td>"+strconv.Itoa(schid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else if len(from)!=0{
			script+="<tr><td style='text-align:center''><input type='checkbox' name='addc_"+id+"' value='"+strconv.Itoa(schid)+"'></td>"
			script+="<td><select class='form-control' name='pric_"+strconv.Itoa(schid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
			script+="<td>"+strconv.Itoa(schid)+"</td><td>" + sd[0] + "</td><td>" +"" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else if len(to)!=0{
			script+="<tr><td style='text-align:center''><input type='checkbox' name='addc_"+id+"' value='"+strconv.Itoa(schid)+"'></td>"
			script+="<td><select class='form-control' name='pric_"+strconv.Itoa(schid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
			script+="<td>"+strconv.Itoa(schid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}else{
			script+="<tr><td style='text-align:center''><input type='checkbox' name='addc_"+id+"' value='"+strconv.Itoa(schid)+"'></td>"
			script+="<td><select class='form-control' name='pric_"+strconv.Itoa(schid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
			script+="<td>"+strconv.Itoa(schid)+"</td><td>" + "" + "</td><td>" + "" +"</td><td>"+"" +"</td><td>"+"" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
		}
	}
	script+="<tr><td colspan='8' style='text-align:center'><a href='#' id='sc_"+id+"'class='btn btn-info btn-sm savec'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Scusave(w http.ResponseWriter, r *http.Request)  {
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
	scid := r.URL.Query().Get("sid")
	pris := r.URL.Query().Get("pri")
	ids :=strings.Split(sids,",");
	tpris:=strings.Split(pris,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
	//defer stmt.Close()
	//chkErr(err,&w)
	ttrows,err :=db.Query("Select sgu_id from scu where scu_id='"+(scid)+"'")
	defer ttrows.Close()
	chkErr(err,&w)
	var sguid int
	if ttrows.Next(){
		err=ttrows.Scan(&sguid)
		chkErr(err,&w)
	}
	cnt:=0
	for _,val:=range ids{
		shgid,_:=strconv.ParseInt(tpris[cnt],10,64)
		cnt++
		trows,err :=db.Query("Select * from schedule where idschedule='"+val+"'")
		defer trows.Close()
		chkErr(err,&w)
		var schid,pwm int
		var sst,set,se,tss string
		for trows.Next(){
			err=trows.Scan(&schid,&sst,&set,&se,&pwm,&tss)
			chkErr(err,&w)
		}
		//_, eorr:=stmt.Exec(val,scid,shgid,pwm,sst,set,se)
		//chkErr(eorr,&w)
		status:=0
		status = ((int)( 1)) & 0x00FF
		status |= ((((int)(shgid)) << 8) & 0x00FF00)
		status |= ((((int)(pwm)) << 16) & 0x00FF0000)
		//for testing
		//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
		le:=len(se)
		isc,_:=(strconv.Atoi(scid))

		LampController.SGUID=uint64(sguid)
		LampController.SCUID=uint64(isc)
		LampController.ConfigArray=[]byte(se)
		LampController.ConfigArrayLength=le
		LampController.PacketType=0x8000
		LampController.LampEvent=status
		LampController.ResponseSend  = make(chan bool)
		LampControllerChannel<-LampController

	}
	io.WriteString(w, "Saved Successfully!!")
}
func Sguconfigure(w http.ResponseWriter, r *http.Request){
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
	pid := r.URL.Query().Get("pid")
	var script string
	logger.Println("Select sgu_id,location_name from sgu limit "+pid+",11")
	rows,err :=db.Query("Select sgu_id,location_name from sgu limit "+pid+",11")
	defer rows.Close()
	chkErr(err,&w)
	cn:=0
	script+="<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;'>SGU Location</th><th style='text-align: center;'>SGU ID</th><th style='text-align: center;'>View Attached Schedules</th><th style='text-align: center;'>Attach Schedules</th></tr></thead><tbody>"
	for rows.Next(){
		var loc string
		var sguid int
		err=rows.Scan(&sguid,&loc)
		chkErr(err,&w)
		if(cn<=9) {
			script += "<tr><td>" + loc + "</td><td>" + strconv.Itoa(sguid) + "</td><td><a href='#' id='vg_" + strconv.Itoa(sguid) + "' class='btn btn-info btn-sm viewg'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td> <a href='#' id='ag_" + strconv.Itoa(sguid) + "' class='btn btn-success btn-sm addg'><span class='glyphicon glyphicon-check'></span> Add</a><div class='vig hidden' id='vidg_" + strconv.Itoa(sguid) + "'></td></tr>"
		}
		cn++
	}
	script+="</tbody></table>"
	if(cn>10) {
		script += "y"
	}
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Sguadd(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id='"+id+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
	}
	rows,err :=db.Query("Select distinct idschedule,ScheduleStartTime,ScheduleEndTime,pwm from schedule")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to string
		err=rows.Scan(&scid,&from,&to,&pwm)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"'")
		defer trows.Close()
		chkErr(err,&w)
		fl:=0
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
		}
		if fl==1{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addg_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prig_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addg_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prig_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addg_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prig_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addg_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prig_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" +"" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="<tr><td colspan='8' style='text-align:center'><a href='#' id='sg_"+id+"'class='btn btn-info btn-sm saveg'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Sguview(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id='"+id+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
	}
	rows,err :=db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID,ScuID from scuconfigure where ScuID in (select scu_id from scu where sgu_id='"+id+"')")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>SCU ID</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to,pri,scuid string
		err=rows.Scan(&scid,&from,&to,&pwm,&pri,&scuid)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"' and ScuID in (select scu_id from scu where sgu_id='"+id+"')")
		defer trows.Close()
		chkErr(err,&w)
		fl:=1
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
			fl=0
		}
		if fl==0{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Sgusave(w http.ResponseWriter, r *http.Request)  {
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
	sgid := r.URL.Query().Get("sid")
	pris := r.URL.Query().Get("pri")
	ids :=strings.Split(sids,",");
	tpris:=strings.Split(pris,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id='"+sgid+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		defer tmprows.Close()
		chkErr(err,&w)
		//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
		//defer stmt.Close()
		//chkErr(err,&w)
		cnt:=0
		for _,val:=range ids{
			trows,err :=db.Query("Select idSCUSchedule from scuconfigure where ScuID='"+strconv.Itoa(tscuid)+"' and ScheduleID='"+val+"'")
			defer trows.Close()
			chkErr(err,&w)
			if trows.Next(){
				cnt++
				continue
			}
			shgid,_:=strconv.ParseInt(tpris[cnt],10,64)
			cnt++
			ttrows,err :=db.Query("Select * from schedule where idschedule='"+val+"'")
			defer ttrows.Close()
			chkErr(err,&w)
			var schid,pwm int
			var sst,set,se,tss string
			for ttrows.Next(){
				err=ttrows.Scan(&schid,&sst,&set,&se,&pwm,&tss)
				chkErr(err,&w)
			}
			//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
			status:=0
			status = ((int)( 1)) & 0x00FF
			status |= ((((int)(shgid)) << 8) & 0x00FF00)
			status |= ((((int)(pwm)) << 16) & 0x00FF0000)
			//for testing
			//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
			logger.Println("exp=",se)
			le := len(se)
			LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
			chkErr(err, &w)
			LampController.SCUID = uint64(tscuid)
			LampController.ConfigArray = []byte(se)
			LampController.ConfigArrayLength = le
			LampController.PacketType = 0x8000
			LampController.LampEvent = status
			LampController.ResponseSend  = make(chan bool)
			du, _ := time.ParseDuration(scu_scheduling)
			time.Sleep(du)
			logger.Println("sent")
			LampControllerChannel <- LampController
			//chkErr(eorr,&w)
		}
	}
	io.WriteString(w, "Saved Successfully!!")
}

func Zoneconfiguresc(w http.ResponseWriter, r *http.Request){
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
	pid := r.URL.Query().Get("pid")
	var script string
	rows,err :=db.Query("Select id,name from zone limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)
	cn:=0
	script+="<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;'>ZONE Name</th><th style='text-align: center;'>ZONE ID</th><th style='text-align: center;'>View Attached Schedules</th><th style='text-align: center;'>Attach Schedules</th></tr></thead><tbody>"
	for rows.Next(){
		var loc string
		var zid int
		err=rows.Scan(&zid,&loc)
		chkErr(err,&w)
		if(cn<=9) {
			script += "<tr><td>" + loc + "</td><td>" + strconv.Itoa(zid) + "</td><td><a href='#' id='vz_" + strconv.Itoa(zid) + "' class='btn btn-info btn-sm viewz'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td> <a href='#' id='az_" + strconv.Itoa(zid) + "' class='btn btn-success btn-sm addz'><span class='glyphicon glyphicon-check'></span> Add</a><div class='viz hidden' id='vidz_" + strconv.Itoa(zid) + "'></td></tr>"
		}
		cn++
	}
	script+="</tbody></table>"
	if(cn>10) {
		script += "y"
	}
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zoneaddsc(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id in (select sguid from zone_sgu where zid='"+id+"')")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
	}
	rows,err :=db.Query("Select distinct idschedule,ScheduleStartTime,ScheduleEndTime,pwm from schedule")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to string
		err=rows.Scan(&scid,&from,&to,&pwm)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"'")
		defer trows.Close()
		chkErr(err,&w)
		fl:=1
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
			fl=0
		}
		if fl==1{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addz_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='priz_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addz_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='priz_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addz_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='priz_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addz_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='priz_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" +"" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="<tr><td colspan='8' style='text-align:center'><a href='#' id='sz_"+id+"'class='btn btn-info btn-sm savez'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zoneview(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id in (select sguid from zone_sgu where zid='"+id+"')")
	defer tmprows.Close()
	chkErr(err,&w)
	scuids:="("
	tfl:=false
	for tmprows.Next(){
		if(tfl) {
			scuids += ",";
		}else {
			tfl = true;
		}
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
		scuids+=strconv.Itoa(tscuid)
	}
	if !tfl{
		scuids+="NULL"
	}
	scuids+=")"
	//logger.Println(scuids)
	rows,err :=db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID,ScuID from scuconfigure where ScuID in "+scuids)
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>SCU ID</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to,pri,scuid string
		err=rows.Scan(&scid,&from,&to,&pwm,&pri,&scuid)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"' and ScuID in "+scuids)
		defer trows.Close()
		chkErr(err,&w)
		fl:=1
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
			fl=0
		}
		if fl==0{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Zonesavesc(w http.ResponseWriter, r *http.Request)  {
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
	sgid := r.URL.Query().Get("sid")
	pris := r.URL.Query().Get("pri")
	ids :=strings.Split(sids,",");
	tpris:=strings.Split(pris,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	tmprows,err :=db.Query("Select scu_id from scu where sgu_id in (select sguid from zone_sgu where zid='"+sgid+"')")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		defer tmprows.Close()
		chkErr(err,&w)
		//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
		//defer stmt.Close()
		//chkErr(err,&w)
		cnt:=0
		for _,val:=range ids{
			trows,err :=db.Query("Select idSCUSchedule from scuconfigure where ScuID='"+strconv.Itoa(tscuid)+"' and ScheduleID='"+val+"'")
			defer trows.Close()
			chkErr(err,&w)
			if trows.Next(){
				cnt++
				continue
			}
			shgid,_:=strconv.ParseInt(tpris[cnt],10,64)
			cnt++
			ttrows,err :=db.Query("Select * from schedule where idschedule='"+val+"'")
			defer ttrows.Close()
			chkErr(err,&w)
			var schid,pwm int
			var sst,set,se,tss string
			for ttrows.Next(){
				err=ttrows.Scan(&schid,&sst,&set,&se,&pwm,&tss)
				chkErr(err,&w)
			}
			//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
			status:=0
			status = ((int)( 1)) & 0x00FF
			status |= ((((int)(shgid)) << 8) & 0x00FF00)
			status |= ((((int)(pwm)) << 16) & 0x00FF0000)
			//for testing
			//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
			logger.Println("exp=",se)
			le := len(se)
			LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
			chkErr(err, &w)
			LampController.SCUID = uint64(tscuid)
			LampController.ConfigArray = []byte(se)
			LampController.ConfigArrayLength = le
			LampController.PacketType = 0x8000
			LampController.LampEvent = status
			LampController.ResponseSend  = make(chan bool)
			du, _ := time.ParseDuration(scu_scheduling)
			time.Sleep(du)
			logger.Println("sent")
			LampControllerChannel <- LampController
			//chkErr(eorr,&w)
		}
	}
	io.WriteString(w, "Saved Successfully!!")
}

func Groupconfiguresc(w http.ResponseWriter, r *http.Request){
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
	pid := r.URL.Query().Get("pid")
	var script string
	rows,err :=db.Query("Select id,name from groupscu limit "+(pid)+",11")
	defer rows.Close()
	chkErr(err,&w)
	cn:=0
	script+="<table class='table table-striped table-hover'><thead><tr><th style='text-align: center;'>Group Name</th><th style='text-align: center;'>Group ID</th><th style='text-align: center;'>View Attached Schedules</th><th style='text-align: center;'>Attach Schedules</th></tr></thead><tbody>"
	for rows.Next(){
		var loc string
		var zid int
		err=rows.Scan(&zid,&loc)
		chkErr(err,&w)
		if(cn<=9) {
			script += "<tr><td>" + loc + "</td><td>" + strconv.Itoa(zid) + "</td><td><a href='#' id='vgr_" + strconv.Itoa(zid) + "' class='btn btn-info btn-sm viewgr'><span class='glyphicon glyphicon-eye-open'></span> View</a></td><td> <a href='#' id='agr_" + strconv.Itoa(zid) + "' class='btn btn-success btn-sm addgr'><span class='glyphicon glyphicon-check'></span> Add</a><div class='vigr hidden' id='vidgr_" + strconv.Itoa(zid) + "'></td></tr>"
		}
		cn++
	}
	script+="</tbody></table>"
	if(cn>10) {
		script += "y"
	}
	io.WriteString(w, script)
	//w.Write([]byte(script))
}

func Groupaddsc(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("select scuid from group_scu_rel where gid='"+id+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
	}
	rows,err :=db.Query("Select distinct idschedule,ScheduleStartTime,ScheduleEndTime,pwm from schedule")
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th width='30px' style='text-align: center;'>Select</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to string
		err=rows.Scan(&scid,&from,&to,&pwm)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"'")
		defer trows.Close()
		chkErr(err,&w)
		fl:=1
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
			fl=0
		}
		if fl==1{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addgr_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prigr_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addgr_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prigr_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addgr_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prigr_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td style='text-align:center''><input type='checkbox' name='addgr_"+id+"' value='"+strconv.Itoa(scid)+"'></td>"
				script+="<td><select class='form-control' name='prigr_"+strconv.Itoa(scid)+"'><option>1</option><option>2</option><option>3</option><option>4</option><option>5</option><option>6</option><option>7</option><option>8</option><option>9</option><option>10</option></select></td>"
				script+="<td>"+strconv.Itoa(scid)+"</td><td>" +"" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="<tr><td colspan='8' style='text-align:center'><a href='#' id='sgr_"+id+"'class='btn btn-info btn-sm savegr'><span class='glyphicon glyphicon-floppy-saved'></span> Save</a></td></tr></tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func Groupsavesc(w http.ResponseWriter, r *http.Request)  {
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
	sgid := r.URL.Query().Get("sid")
	pris := r.URL.Query().Get("pri")
	ids :=strings.Split(sids,",");
	tpris:=strings.Split(pris,",");
	logger.Println(ids)
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	tmprows,err :=db.Query("select scuid from group_scu_rel where gid='"+sgid+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	for tmprows.Next(){
		var tscuid int
		err=tmprows.Scan(&tscuid)
		defer tmprows.Close()
		chkErr(err,&w)
		//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
		//defer stmt.Close()
		//chkErr(err,&w)
		cnt:=0
		for _,val:=range ids{
			trows,err :=db.Query("Select idSCUSchedule from scuconfigure where ScuID='"+strconv.Itoa(tscuid)+"' and ScheduleID='"+val+"'")
			defer trows.Close()
			chkErr(err,&w)
			if trows.Next(){
				cnt++
				continue
			}
			shgid,_:=strconv.ParseInt(tpris[cnt],10,64)
			cnt++
			ttrows,err :=db.Query("Select * from schedule where idschedule='"+val+"'")
			defer ttrows.Close()
			chkErr(err,&w)
			var schid,pwm int
			var sst,set,se,tss string
			for ttrows.Next(){
				err=ttrows.Scan(&schid,&sst,&set,&se,&pwm,&tss)
				chkErr(err,&w)
			}
			//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
			status:=0
			status = ((int)( 1)) & 0x00FF
			status |= ((((int)(shgid)) << 8) & 0x00FF00)
			status |= ((((int)(pwm)) << 16) & 0x00FF0000)
			//for testing
			//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
			logger.Println("exp=",se)
			le := len(se)
			LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
			chkErr(err, &w)
			LampController.SCUID = uint64(tscuid)
			LampController.ConfigArray = []byte(se)
			LampController.ConfigArrayLength = le
			LampController.PacketType = 0x8000
			LampController.LampEvent = status
			LampController.ResponseSend  = make(chan bool)
			du, _ := time.ParseDuration(scu_scheduling)
			time.Sleep(du)
			logger.Println("sent")
			LampControllerChannel <- LampController
			//chkErr(eorr,&w)
		}
	}
	io.WriteString(w, "Saved Successfully!!")
}

func Groupview(w http.ResponseWriter, r *http.Request){
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
	set := make(map[int]int)
	tmprows,err :=db.Query("select scuid from group_scu_rel where gid='"+id+"'")
	defer tmprows.Close()
	chkErr(err,&w)
	scuids:="("
	tfl:=false
	for tmprows.Next(){
		if(tfl) {
			scuids += ",";
		}else {
			tfl = true;
		}
		var tscuid int
		err=tmprows.Scan(&tscuid)
		chkErr(err,&w)
		set[tscuid]=1
		scuids+=strconv.Itoa(tscuid)
	}
	if !tfl{
		scuids+="NULL"
	}
	scuids+=")"
	//logger.Println(scuids)
	rows,err :=db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID,ScuID from scuconfigure where ScuID in "+scuids)
	defer rows.Close()
	chkErr(err,&w)
	var script string
	script="<table class='table table-bordered table-hover'><thead><tr class='info'><th style='text-align: center;'>SCU ID</th><th style='text-align: center;'>Scheduling Priority</th><th style='text-align: center;'>Schedule ID</th><th style='text-align: center;'>Start Date</th><th style='text-align: center;'>End Date</th><th style='text-align: center;'>Start Time</th><th style='text-align: center;'>End Time</th><th style='text-align: center;'>Brightness level</th></tr></thead><tbody>"
	for rows.Next(){
		var scid,pwm int
		var from,to,pri,scuid string
		err=rows.Scan(&scid,&from,&to,&pwm,&pri,&scuid)
		chkErr(err,&w)
		trows,err :=db.Query("Select distinct ScuID from scuconfigure where ScheduleID='"+strconv.Itoa(scid)+"' and ScuID in "+scuids)
		defer trows.Close()
		chkErr(err,&w)
		fl:=1
		tset := make(map[int]int)
		for trows.Next(){
			var scuid int
			err=trows.Scan(&scuid)
			chkErr(err,&w)
			tset[scuid]=1;
		}
		for k,_:=range set{
			if tset[k]!=1{
				fl=1;
				break;
			}
			fl=0
		}
		if fl==0{
			sd:=strings.Split(from," ")
			ed:=strings.Split(to," ")
			if len(from)!=0&&len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + ed[0] +"</td><td>"+ sd[1] +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(from)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + sd[0] + "</td><td>" + "" +"</td><td>"+ sd[1] +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else if len(to)!=0{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + ed[0] +"</td><td>"+ "" +"</td><td>"+ ed[1] +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}else{
				script+="<tr><td>"+scuid+"</td><td>"+pri+"</td><td>"+strconv.Itoa(scid)+"</td><td>" + "" + "</td><td>" + "" +"</td><td>"+ "" +"</td><td>"+ "" +"</td><td>"+strconv.Itoa(pwm)+"</td></tr>"
			}
		}
	}
	script+="</tbody></table>"
	io.WriteString(w, script)
	//w.Write([]byte(script))
}
func chkErr(err error,r *http.ResponseWriter){
	if err!=nil{
		//io.WriteString(*r, ("Some Problem Occured!!"))
		logger.Println(err.Error());
		//panic(err);
	}
}
