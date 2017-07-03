/********************************************************************
 * FileName:     NBApis.go
 * Project:      Havells StreetComm
 * Module:       NBApis
 * Company:      North Bound Cisco
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package NBApis

import (
	//	"configure"
	"dbUtils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	//	"mapview"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	//	"net/smtp"
	"net/url"
	//	"os"
	"regexp"
	//	"report"
	"sguUtils"
	"strconv"
	"strings"
	//	"tcpServer"
	"sync"
	"tcpUtils"
	"time"
	//	"github.com/context"
	"github.com/jwt-go-master"
	//	"github.com/scorredoira/email"
	//	"github.com/sessions"
)

var (
	dbController          dbUtils.DbUtilsStruct
	sguConnectionChan     chan net.Conn
	err                   error
	LampControllerChannel chan sguUtils.SguUtilsLampControllerStruct
	//energytcputil           tcpUtils.TcpUtilsStruct
	//energysguutil           sguUtils.SguUtilsEnergyCntrStruct
	//energysguutilChannel    chan   	sguUtils.SguUtilsEnergyCntrStruct
	SendSMSChan    chan string
	logger         *log.Logger
	tokenMap       (map[string]int)
	per_scu_delay  string
	scu_scheduling string
)

var status = struct {
	sync.RWMutex
	lbuffer map[string]string
}{lbuffer: make(map[string]string)}

func InitNBApis(LampConChannel chan sguUtils.SguUtilsLampControllerStruct,
	dbcon dbUtils.DbUtilsStruct, logg *log.Logger) {
	logger = logg
	LampControllerChannel = LampConChannel
	dbController = dbcon
}

func Config_Params(scu_sch, per_scudelay string) {
	scu_scheduling = scu_sch + "s"
	per_scu_delay = per_scudelay
}

type NBFdn struct {
	System      string `json:"system"`
	Gateway     string `json:"gateway"`
	Street_lamp string `json:"street_lamp"`
	Group       string `json:"group"`
	Id          string `json:"id"`
	Zone        string `json:"zone_id"`
}

type NBData struct {
	Brightness    string                                  `json:"brightness"`
	Message       string                                  `json:"msg"`
	Token         string                                  `json:"token"`
	Email         string                                  `json:email`
	Discovery_map map[string]map[string]map[string]string `json:discovery`
	FromDate      string                                  `json:"from_date"`
	ToDate        string                                  `json:"to_date"`
	FromTime      string                                  `json:"from_time"`
	ToTime        string                                  `json:"to_time"`
	Username      string                                  `json:"username"`
	Password      string                                  `json:"password"`
	Ids           []string                                `json:"ids"`
	GroupId       string                                  `json:"group_id"`
	ZoneId        string                                  `json:"zone_id"`
	Schedule_Id   string                                  `json:"schedule_id"`
	Priority      string                                  `json:"priority"`
	Zone          string                                  `json:"zone_name"`
	//Group         string                                  `json:"group_id"`
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
	End             string `json:"end"`
}

type GatewayResponse struct {
	Response_status string `json:"response_status"`
	//data            interface{} `json:"data"`
	Data map[string]StreetLampDetails `json:"data"`
	End  string                       `json:"end"`
}

type StreetLampDetails struct {
	Location_name string `json:"location_name"`
	Location_lat  string `json:"location_lat"`
	Location_lng  string `json:"location_lng"`
	Status        string `json:"status"`
}

type ScheduleStr struct {
	//Id         string `json:"schedule_id"`
	SSDate     string `json:"from_date"`
	SEDate     string `json:"to_date"`
	SSTime     string `json:"from_time"`
	SETime     string `json:"to_time"`
	Brightness string `json:"brightness"`
	//Priority   string `json:"priority"`
}

type SCUVIewStr struct {
	//Id         string `json:"schedule_id"`
	SSDate     string `json:"from_date"`
	SEDate     string `json:"to_date"`
	SSTime     string `json:"from_time"`
	SETime     string `json:"to_time"`
	Brightness string `json:"brightness"`
	Priority   string `json:"priority"`
}

type ScheduleResp struct {
	Response_status string                            `json:"response_status"`
	Data            map[string]map[string]ScheduleStr `json:"data"`
	End             string                            `json:"end"`
}

type SCUViewResp struct {
	Response_status string                           `json:"response_status"`
	Data            map[string]map[string]SCUVIewStr `json:"data"`
	End             string                           `json:"end"`
}

/*type SguView struct {
	Response_status string `json:"resposen_status"`
	Data             `json:"data"`
	End             string `json:"end"`
}*/
type LampPowerStatus struct {
	Response_status string            `json:"resposne_status"`
	Data            map[string]string `json:"data"`
	End             string            `json:"end"`
	//Status 			string				`json:"current_status"`
}

//NB Street Lamp Controll
func StreetLampControll(w http.ResponseWriter, r *http.Request) {
	logger.Println("StreetLampControll()")
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
	//not sure 100% about object validation.
	if l_object == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Object"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_opr := NBLampStr.Opr
	if l_opr == "" || l_opr != "set_street_lamp_power_status" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_system := NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	//l_data := NBLampStr.Data
	var NewStatus string
	if l_brightness != "0" {
		NewStatus = "1"
	} else {
		NewStatus = "0"
	}
	l_sgu := NBLampStr.Fdn.Gateway
	l_scu := NBLampStr.Fdn.Street_lamp
	l_event := NBLampStr.Data.Brightness
	if !validateSGU(l_sgu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid SGU"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if !IsSGUInDb(l_sgu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Gateway Not Found"
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
	if !IsSCUInDb(l_scu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Street Lamp Not Found"
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
	Levent := l_brightness_i
	if Levent > 0 {
		if Levent <= 2 {
			l_event = "5"
		} else if Levent > 2 && Levent <= 4 {
			l_event = "6"
		} else if Levent > 4 && Levent <= 6 {
			l_event = "7"
		} else if Levent > 6 && Levent <= 8 {
			l_event = "8"
		}
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
		tcpUtils.SetTempStatus(l_scu, NewStatus)
		logger.Println("Lamp event sent to channel")
		du, _ := time.ParseDuration(per_scu_delay + "s")
		time.Sleep(du)
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
/*func NBlogin(w http.ResponseWriter, r *http.Request) {
	logger.Println("NBlogin()")
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
	fmt.Println("username:", username)
	password := t.Password
	fmt.Println("password:", password)
	ans := NBResponseStruct{}
	if !validateEmail(username) {
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
}*/

// To validate email format
func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
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

//type NBDiscoveryResponseStruct struct {
//	Response_status string `json:"response_status"`
//	Message
//	Data map[string]map[string]map[string]string `json:"data"`
//}

func Discovery(w http.ResponseWriter, r *http.Request) {
	logger.Println("Discovery()")
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
	logger.Println("l_object", l_object)
	/*l_opr := NBLampStr.Opr
	logger.Println("l_opr", l_opr)
	l_system := NBLampStr.Fdn.System
	logger.Println("l_system", l_system)*/
	l_opr := NBLampStr.Opr
	if l_opr == "" || l_opr != "discovery" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_system := NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	logger.Println("l_brightness", l_brightness)
	//l_data := NBLampStr.Data

	//Map for system
	var NBsys map[string]map[string]map[string]string
	NBsys = make(map[string]map[string]map[string]string)
	NBsys["gateway1"] = make(map[string]map[string]string)
	NBsys["gateway1"]["scu1"] = make(map[string]string)
	//token validation
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		db := dbController.Db
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		//rows, err := db.Query("Select sgu.sgu_id,sgu.location_name,scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from zone_sgu inner join zone on zone.id=zone_sgu.zid inner join sgu on sgu.sgu_id=zone_sgu.sguid inner join scu on scu.sgu_id=zone_sgu.sguid inner join scu_status on scu_status.scu_id=scu.scu_id")
		//rows, err := db.Query("Select zone_sgu.zid,zone.name,sgu.sgu_id,sgu.location_name,scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from zone_sgu inner join zone on zone.id=zone_sgu.zid inner join sgu on sgu.sgu_id=zone_sgu.sguid inner join scu on scu.sgu_id=zone_sgu.sguid inner join scu_status on scu_status.scu_id=scu.scu_id")
		rows, err := db.Query("Select sgu.sgu_id,sgu.location_name,scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from scu inner join sgu on scu.sgu_id=sgu.sgu_id inner join scu_status on scu_status.scu_id=scu.scu_id")
		defer rows.Close()
		chkErr(err, &w)
		//scanninng data from the query result
		for rows.Next() {
			//var zid, zname, sguid, sguname, scuid, scuname, lat, lng string
			var sguid, sguname, scuid, scuname, lat, lng, sched_id string
			var st uint64
			st = 10
			//rows.Scan(&zid, &zname, &sguid, &sguname, &scuid, &scuname, &lat, &lng, &st)
			rows.Scan(&sguid, &sguname, &scuid, &scuname, &lat, &lng, &st)
			sta := st & (0x0FF)
			logger.Println("STATUS before=", sta)
			sta = sta & 3
			logger.Println("STATUS=", sta)
			//To get schedule Id
			statement := "SELECT ScheduleID FROM scuconfigure where scuid='" + scuid + "'"

			logger.Println(statement)
			rows, err := dbController.Db.Query(statement)
			defer rows.Close()
			if err != nil {
				logger.Println("Error quering database  for Schedule Id")
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
					rows.Scan(&sched_id)
				}
				rows.Close()
			}

			for sguKey, sguData := range NBsys {
				if sguKey == sguid {
					for scuKey, scuData := range sguData {
						scuData["op_mode"] = ""
						scuData["location_name"] = scuname
						scuData["location_lat"] = lat
						scuData["location_lng"] = lng
						scuData["status"] = strconv.FormatUint(sta, 10)
						scuData["schedule_id"] = sched_id
						if scuKey == "scu1" {
							delete(sguData, "scu1")
							sguData[scuid] = scuData
						} else {
							sguData[scuid] = scuData
						}
					}
					if sguKey == "gateway1" {
						delete(NBsys, "gateway1")
						NBsys[sguid] = sguData
					}
				} else {
					for sguKey, sguData := range NBsys {
						for scuKey, scuData := range sguData {
							if scuKey == "scu1" {
								scuKey = scuid
							}
							scuData["op_mode"] = ""
							scuData["location_name"] = scuname
							scuData["location_lat"] = lat
							scuData["location_lng"] = lng
							scuData["status"] = strconv.FormatUint(sta, 10)
							scuData["schedule_id"] = sched_id
							if scuKey == "scu1" {
								delete(sguData, "scu1")
								sguData[scuid] = scuData
							} else {
								sguData[scuid] = scuData
							}
						}
						if sguKey == "gateway1" {
							delete(NBsys, "gateway1")
							NBsys[sguid] = sguData
						} else {
							NBsys[sguid] = sguData
						}
					}
				}
			}
		}
		ans.Response_status = "success"
		//		ans.Data.Token = tokenString
		//		ans.Data.Email = username
		ans.Data.Discovery_map = NBsys
		ans.End = "^^End of Discovery^^"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		logger.Println("response status", ans.Response_status)
		logger.Println("response status", ans.Data.Discovery_map)
		return
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
}

func chkErr(err error, r *http.ResponseWriter) {
	if err != nil {
		io.WriteString(*r, (err.Error()))
		logger.Println(err)
		//panic(err);
	}
}

//Gateway's streetlamp controll
func GatewayStreetLampControll(w http.ResponseWriter, r *http.Request) {
	logger.Println("GatewayStreetLampControll()")
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
	/*l_opr := NBLampStr.Opr
	l_system := NBLampStr.Fdn.System*/
	l_opr := NBLampStr.Opr
	if l_opr == "" || l_opr != "set_gateway_power_status" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_system := NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	//l_data := NBLampStr.Data
	var NewStatus string
	if l_brightness != "0" {
		NewStatus = "1"
	} else {
		NewStatus = "0"
	}

	l_sgu := NBLampStr.Fdn.Gateway
	l_scu := NBLampStr.Fdn.Street_lamp
	l_event := NBLampStr.Data.Brightness
	if !validateSGU(l_sgu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Gateway"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	if !IsSGUInDb(l_sgu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Gateway Not Found"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	/*	if !validateSCU(l_scu) {
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
	}*/
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
	if l_opr == "" || l_opr != "set_gateway_power_status" {
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
	} else {
		Levent := l_brightness_i
		if Levent > 0 {
			if Levent <= 2 {
				l_event = "5"
			} else if Levent > 2 && Levent <= 4 {
				l_event = "6"
			} else if Levent > 4 && Levent <= 6 {
				l_event = "7"
			} else if Levent > 6 && Levent <= 8 {
				l_event = "8"
			}
		}
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

		/*LampController.SCUID, err = strconv.ParseUint(l_scu, 10, 64)
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
		}*/
		//get the SCU's from DB for the Particular SGU Id
		if dbController.DbConnected {

			statement := "select scu_id from scu where sgu_id='" + l_sgu + "'"
			logger.Println(statement)
			rows, err := dbController.Db.Query(statement)
			fmt.Println("rows", rows)
			if err != nil {
				logger.Println("Error quering database  for login information")
				logger.Println(err)
			} else {
				fmt.Println("1")
				for rows.Next() {
					fmt.Println("2")
					var scu_id_db_s string
					fmt.Println("3")
					err := rows.Scan(&scu_id_db_s)
					fmt.Println("4")
					fmt.Println("scu_id_db_s", scu_id_db_s)
					if err != nil {
						logger.Println(err)
						logger.Println("Unable to scan SCU Id from DB")
						ans.Response_status = "fail"
						ans.Data.Message = "Unable to scan SCU Id from DB"
						a, err_m := json.Marshal(ans)
						if err_m != nil {
							logger.Println("Error in json.Marshal")
							logger.Println(err_m)
						} else {
							w.Write(a)
						}
						return
					}

					//tcpUtils.SetTempStatus(scu_id_db_s,NewStatus)
					//logger.Println("Status Set")
					LampController.SCUID, err = strconv.ParseUint(scu_id_db_s, 10, 64)
					logger.Println("LampController.SCUID", LampController.SCUID)
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
					tcpUtils.SetTempStatus(scu_id_db_s, NewStatus)
					logger.Println("Lamp event sent to channel for SCU Id :", LampController.SCUID, "Of SGU Id", LampController.SGUID)

				}
				ans.Response_status = "success"
				ans.Data.Message = ""
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				rows.Close()
			}
		}
		//endget the SCU's from DB for the Particular SGU Id
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
}

//Group's streetlamp controll
func GroupStreetLampControll(p_NBLampStr *NBAllLampControlStruct) NBResponseStruct {
	logger.Println("GroupStreetLampControll()")
	var ans NBResponseStruct
	NBLampStr := p_NBLampStr
	/*parse_err := r.ParseForm()
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
	}*/
	//l_token := NBLampStr.Token
	l_object := NBLampStr.Object
	/*l_opr := NBLampStr.Opr
	l_system := NBLampStr.Fdn.System*/
	l_opr := NBLampStr.Opr
	if l_opr == "" || l_opr != "set_group_power_status" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
		return ans
	}
	l_system := NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		return ans
	}
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	var NewStatus string
	if l_brightness == "0" {
		NewStatus = "0"
	} else {
		NewStatus = "1"
	}
	//l_data := NBLampStr.Data
	//l_sgu := NBLampStr.Fdn.Gateway
	//l_scu := NBLampStr.Fdn.Street_lamp
	l_group_s := NBLampStr.Fdn.Group
	l_event := NBLampStr.Data.Brightness
	/*if !validateSGU(l_sgu) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid SGU"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}*/
	/*if !validateSCU(l_scu) {
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
	}*/
	if l_group_s == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Group Not Specified"
		return ans
	}
	if !IsGroupIDInDB(l_group_s) {
		ans.Response_status = "fail"
		ans.Data.Message = "Group Not Found"
		return ans
	}
	if l_object == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Object Not Specified"
		return ans
	}
	if l_brightness == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Brightness Not Specified"
		return ans
	}
	l_brightness_i, _ := strconv.Atoi(l_brightness)

	if l_brightness_i < 0 || l_brightness_i > 10 {
		ans.Response_status = "fail"
		ans.Data.Message = "brightness is not in range"
		return ans
	}
	Levent := l_brightness_i
	if Levent > 0 {
		if Levent <= 2 {
			l_event = "5"
		} else if Levent > 2 && Levent <= 4 {
			l_event = "6"
		} else if Levent > 4 && Levent <= 6 {
			l_event = "7"
		} else if Levent > 6 && Levent <= 8 {
			l_event = "8"
		}
	}
	var LampController sguUtils.SguUtilsLampControllerStruct

	/*	logger.Println(r.URL)
		u, err := url.Parse(r.URL.String())
		logger.Println(u)
		logger.Println(u.RawQuery)*/

	LampController.LampEvent, err = strconv.Atoi(l_event)
	if err != nil {
		logger.Println("Invalid lamp contral val  " + l_event + " specified")
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid lamp contral val"
		return ans
	}

	//get the SCU's from DB for the Particular Group Id
	if dbController.DbConnected {

		statement := "select scuid from group_scu_rel where gid='" + l_group_s + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)

		if err != nil {
			logger.Println("Error quering database  for login information")
			logger.Println(err)
		} else {

			for rows.Next() {
				var scu_id_db_s string
				err := rows.Scan(&scu_id_db_s)
				if err != nil {
					logger.Println(err)
					logger.Println("Unable to scan SCU Id from DB")
					ans.Response_status = "fail"
					ans.Data.Message = "Unable to scan SCU Id from DB"
					return ans
				}
				// Select SGU Id for SCU Id from DB
				statement := "select sgu_id from scu where scu_id='" + scu_id_db_s + "'"
				logger.Println(statement)
				rows_sgu, err := dbController.Db.Query(statement)
				if err != nil {
					logger.Println("Error quering database  for SGU ID")
					logger.Println(err)
				} else {
					for rows_sgu.Next() {
						var sgu_id_db_s string
						err := rows_sgu.Scan(&sgu_id_db_s)
						if err != nil {
							logger.Println(err)
							logger.Println("Unable to scan SGU Id from DB")
							ans.Response_status = "fail"
							ans.Data.Message = "Unable to scan SGU Id from DB"
							return ans
						}
						LampController.SCUID, err = strconv.ParseUint(scu_id_db_s, 10, 64)
						if err != nil {
							logger.Println("Invalid SCUID ")
							ans.Response_status = "fail"
							ans.Data.Message = "Invalid SCUID"
							return ans
						}
						LampController.SGUID, err = strconv.ParseUint(sgu_id_db_s, 10, 64)
						if err != nil {
							logger.Println("Invalid SGUID")
							ans.Response_status = "fail"
							ans.Data.Message = "Invalid SGUID"
							return ans
						}
						LampController.LampEvent |= 0x100
						LampController.PacketType = 0x3000
						LampController.ConfigArray = nil
						LampController.ConfigArrayLength = 0

						//LampController.ResponseSend  = make(chan bool)
						//LampController.ResponseSend  = make(chan bool)
						LampController.W = nil
						LampController.ResponseSend = nil
						du, _ := time.ParseDuration(scu_scheduling)
						time.Sleep(du)
						logger.Println("sent")
						LampControllerChannel <- LampController
						tcpUtils.SetTempStatus(scu_id_db_s, NewStatus)
						logger.Println("Lamp event sent to channel for SCU Id :", LampController.SCUID, "Of SGU Id", LampController.SGUID)
					}
				}
				rows_sgu.Close()
				// end Select SGU Id for SCU Id from DB
				//GetSet field is set to set mode
			}
			rows.Close()
			ans.Response_status = "success"
			ans.Data.Message = ""
			return ans
		}
	}
	//endget the SCU's from DB for the Particular SGU Id
	return ans
}

func CraeteSchedule(w http.ResponseWriter, r *http.Request) {
	logger.Println("CraeteSchedule()")
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
	fmt.Println("1")
	var en string
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_object := NBLampStr.Object
	fmt.Println("l_object", l_object)
	/*l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	l_system := NBLampStr.Fdn.System
	fmt.Println("l_system", l_system)*/
	l_opr := NBLampStr.Opr
	if l_opr == "" || l_opr != "set_street_lamp_power_status" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
	l_system := NBLampStr.Fdn.System
	if l_system == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
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
	fmt.Println("2")
	//l_fdn := NBLampStr.Fdn
	ssd := NBLampStr.Data.FromDate
	//ssd := r.FormValue("data.from_date")
	//ssd := r.Form["data.from_date"][0]
	//ssd := r.Form["from_date"][0]
	fmt.Println("3")
	logger.Println("ScheduleStartDate:", ssd)
	sst := NBLampStr.Data.FromTime
	logger.Println("ScheduleStartDate:", sst)
	edd := NBLampStr.Data.ToDate
	logger.Println("ScheduleStartDate:", edd)
	et := NBLampStr.Data.ToTime
	logger.Println("ScheduleStartDate:", et)
	pwm := NBLampStr.Data.Brightness
	if pwm == "" {
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
	logger.Println("pwm value:", pwm)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		datestr := ssd
		timestr := sst
		logger.Println("timestr", timestr)
		dateinymd := strings.Split(datestr, "-")
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

		dateinymd = strings.Split(dateen, "-")
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
		fmt.Println("isti :", isti)

		ieti, _ := strconv.Atoi(strings.Split(timeen, ":")[0])
		fmt.Println("ieti:", ieti)
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
		rows, err := dbController.Db.Query("Select idschedule from schedule where ScheduleStartTime='" + ssd + " " + sst + "' and ScheduleEndTime='" + edd + " " + et + "' and pwm='" + pwm + "'")
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
		res, err := stmt.Exec(ssd+" "+sst, edd+" "+et, pwm, exp)
		if err != nil {
			io.WriteString(w, "Something Went Wrong!")
			return
		}
		if res == nil {
			//fmt.Fprint(w,"no data stored in database")
			//http.Redirect(w,r,"errormessage.html",http.StatusFound)
			io.WriteString(w, "Something Went Wrong!")
		} else {
			//io.WriteString(w, "Schedule Added Successfully!")
			ans.Response_status = "success"
			ans.Data.Message = ""
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
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
}

// New Login from Inside Data Parameter for North Bound Api.
func NBlogin(w http.ResponseWriter, r *http.Request) {
	logger.Println("NBlogin()")
	ans := NBResponseStruct{}
	tokenMap = make(map[string]int)
	r.ParseForm()
	var NBLampStr NBAllLampControlStruct
	if len(r.FormValue("username")) == 0 {
		decoder := json.NewDecoder(r.Body)
		logger.Println(decoder)
		err := decoder.Decode(&NBLampStr)
		if err != nil {
			logger.Println(err)
		}
	} else {
		NBLampStr.Opr = r.FormValue("third_party_login")
		NBLampStr.Fdn.System = r.FormValue("fdn.system")
		NBLampStr.Data.Username = r.FormValue("data.username")
		NBLampStr.Data.Password = r.FormValue("data.password")
	}
	opr_l := NBLampStr.Opr
	if opr_l == "" || opr_l != "third_party_login" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Operation"
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
	system_l := NBLampStr.Fdn.System
	//System number Yet to be Fixed
	if !ValidateSystem(system_l) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
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
	username := NBLampStr.Data.Username
	fmt.Println("username:", username)
	password := NBLampStr.Data.Password
	fmt.Println("password:", password)

	if !validateEmail(username) {
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

func IsSGUInDb(p_sgu string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_sgu", p_sgu)
		statement := "SELECT sgu_id FROM sgu where sgu_id='" + p_sgu + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_sgu string
			for rows.Next() {
				rows.Scan(&l_sgu)
				if strings.EqualFold(p_sgu, l_sgu) == true {
					l_resp = true
				} else {
					l_resp = false
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func IsSCUInDb(p_scu string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_scu", p_scu)
		statement := "SELECT scu_id FROM scu where scu_id='" + p_scu + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_sgu string
			for rows.Next() {
				rows.Scan(&l_sgu)
				if strings.EqualFold(p_scu, l_sgu) == true {
					l_resp = true
				} else {
					l_resp = false
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func ValidateSystem(p_system string) bool {
	if p_system == "" || p_system != "5" {
		return false
	} else {
		return true
	}
}

func IsGroupInDB(p_group string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_group", p_group)
		statement := "SELECT name FROM groupscu where name='" + p_group + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_group string
			for rows.Next() {
				rows.Scan(&l_group)
				if strings.EqualFold(p_group, l_group) == true {
					l_resp = true
				} else {
					l_resp = false
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func SystemGroupURL(w http.ResponseWriter, r *http.Request) {
	logger.Println("SystemGroupURL()")
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	l_opr := NBLampStr.Opr
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		switch l_opr {
		case "add_street_lamps_to_a_group":
			ans = AddStreetLampsToGroup(&NBLampStr)
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		case "set_group_power_status":
			ans = GroupStreetLampControll(&NBLampStr)
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func AddStreetLampsToGroup(p_NBLampStr *NBAllLampControlStruct) NBResponseStruct {
	logger.Println("AddStreetLampsToGroup()")
	var ans NBResponseStruct

	l_object := p_NBLampStr.Object
	if l_object == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Object"
		return ans
	}
	l_system := p_NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		return ans
	}
	l_group_id := p_NBLampStr.Data.GroupId
	if l_group_id == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Group Not Specified"
		return ans

	}
	if !IsGroupIDInDB(l_group_id) {
		ans.Response_status = "fail"
		ans.Data.Message = "Group Not Found"
		return ans
	}
	l_scu_ids := p_NBLampStr.Data.Ids
	for i := 0; i < len(l_scu_ids); i++ {
		l_scu_id := l_scu_ids[i]
		_, err := strconv.ParseUint(l_scu_id, 10, 64)
		if err != nil {
			logger.Println("Invalid SCUID ")
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SCUID"
			return ans
		}
		if !validateSCU(l_scu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SCU"
			return ans
		}
		if !IsSCUInDb(l_scu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "Street Lamp Not Found"
			return ans

		}
		if !IsSCUInGroup(l_group_id, l_scu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "SCU already in the group"
			logger.Println("SCU: %s already in the group", l_scu_id)
			return ans
		}
	}
	if dbController.DbConnected {
		fmt.Println("len(l_scu_ids)", len(l_scu_ids))
		for i := 0; i < len(l_scu_ids); i++ {
			fmt.Println("i", i)
			stmt, err := dbController.Db.Prepare("INSERT group_scu_rel SET scuid=?,gid=?")
			logger.Println(stmt)
			if err != nil {
				logger.Println(err)
				ans.Response_status = "fail"
				ans.Data.Message = "Error While Preparing query"
				return ans
			}

			res, err := stmt.Exec(l_scu_ids[i], l_group_id)
			if err != nil {
				logger.Println(err)
				ans.Response_status = "fail"
				ans.Data.Message = "Error While Executing query"
				return ans
			}
			if res == nil {
				ans.Response_status = "fail"
				ans.Data.Message = "some thing went wrong while Executing query"
				return ans
			}
		}
		ans.Response_status = "success"
		ans.Data.Message = ""
		return ans
	}
	return ans
}

func IsGroupIDInDB(p_group string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_group", p_group)
		statement := "SELECT id FROM groupscu where id='" + p_group + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_group string
			for rows.Next() {
				rows.Scan(&l_group)
				if strings.EqualFold(p_group, l_group) == true {
					l_resp = true
				} else {
					l_resp = false
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func IsSCUInGroup(p_group_id, p_scu string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_group_id", p_group_id)
		statement := "SELECT scuid FROM group_scu_rel where gid='" + p_group_id + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_scu string
			for rows.Next() {
				rows.Scan(&l_scu)
				if strings.EqualFold(p_scu, l_scu) == true {
					l_resp = false
				} else {
					l_resp = true
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func SystemZoneURL(w http.ResponseWriter, r *http.Request) {
	logger.Println("SystemGroupURL()")
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		switch l_opr {
		case "add_gateways_to_a_zone":
			ans = AddGatewayToAZone(&NBLampStr)
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func AddGatewayToAZone(p_NBLampStr *NBAllLampControlStruct) NBResponseStruct {
	logger.Println("AddGatewayToAZone()")
	var ans NBResponseStruct

	l_object := p_NBLampStr.Object
	if l_object == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Object"
		return ans
	}
	l_system := p_NBLampStr.Fdn.System
	if !ValidateSystem(l_system) {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid System"
		return ans
	}
	l_zone_id := p_NBLampStr.Data.ZoneId
	if l_zone_id == "" {
		ans.Response_status = "fail"
		ans.Data.Message = "Zone Not Specified"
		return ans

	}
	if !IsZoneIDInDB(l_zone_id) {
		ans.Response_status = "fail"
		ans.Data.Message = "Zone Not Found"
		return ans
	}
	l_sgu_ids := p_NBLampStr.Data.Ids
	for i := 0; i < len(l_sgu_ids); i++ {
		l_sgu_id := l_sgu_ids[i]
		_, err := strconv.ParseUint(l_sgu_id, 10, 64)
		if err != nil {
			logger.Println("Invalid SGU ID ")
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SGU ID"
			return ans
		}
		if !validateSGU(l_sgu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SGU"
			return ans
		}
		if !IsSGUInDb(l_sgu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "Street Lamp Not Found"
			return ans

		}
		if !IsSGUInZone(l_zone_id, l_sgu_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "SGU already in the group"
			logger.Println("SGU: %s already in the group", l_sgu_id)
			return ans
		}
	}
	if dbController.DbConnected {
		fmt.Println("len(l_scu_ids)", len(l_sgu_ids))
		for i := 0; i < len(l_sgu_ids); i++ {
			fmt.Println("i", i)
			stmt, err := dbController.Db.Prepare("INSERT zone_sgu SET sguid=?,zid=?")
			logger.Println(stmt)
			if err != nil {
				logger.Println(err)
				ans.Response_status = "fail"
				ans.Data.Message = "Error While Preparing query"
				return ans
			}

			res, err := stmt.Exec(l_sgu_ids[i], l_zone_id)
			if err != nil {
				logger.Println(err)
				ans.Response_status = "fail"
				ans.Data.Message = "Error While Executing query"
				return ans
			}
			if res == nil {
				ans.Response_status = "fail"
				ans.Data.Message = "some thing went wrong while Executing query"
				return ans
			}
		}
		ans.Response_status = "success"
		ans.Data.Message = ""
		return ans
	}
	return ans
}

func IsSGUInZone(p_zone_id, p_sgu string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_zone_id", p_zone_id)
		statement := "SELECT sguid FROM zone_sgu where zid='" + p_zone_id + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_sgu_db string
			for rows.Next() {
				rows.Scan(&l_sgu_db)
				if strings.EqualFold(p_sgu, l_sgu_db) == true {
					l_resp = false
				} else {
					l_resp = true
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func IsZoneIDInDB(p_zone string) bool {
	var l_resp bool
	if dbController.DbConnected {
		fmt.Println("p_zone", p_zone)
		statement := "SELECT id FROM zone where id='" + p_zone + "'"
		logger.Println(statement)
		rows, err := dbController.Db.Query(statement)
		defer rows.Close()
		if err != nil {
			l_resp = false
		} else {
			var l_zone_db string
			for rows.Next() {
				rows.Scan(&l_zone_db)
				if strings.EqualFold(p_zone, l_zone_db) == true {
					l_resp = true
				} else {
					l_resp = false
				}
			}
			rows.Close()
		}
	}
	return l_resp
}

func SetScheduleForScu(w http.ResponseWriter, r *http.Request) {
	logger.Println("SetScheduleForScu()")
	var LampController sguUtils.SguUtilsLampControllerStruct
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token

	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		sids := NBLampStr.Data.Schedule_Id
		if sids == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Schedule Id"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		scid := NBLampStr.Fdn.Id
		if !validateSCU(scid) {
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
		if !IsSCUInDb(scid) {
			ans.Response_status = "fail"
			ans.Data.Message = "Street Lamp Not Found"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		pris := NBLampStr.Data.Priority
		if pris == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Priority"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		//value will be 1,2,3
		ids := strings.Split(sids, ",")
		tpris := strings.Split(pris, ",")
		logger.Println(ids)
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		db := dbController.Db
		//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
		//defer stmt.Close()
		//chkErr(err,&w)
		ttrows, err := db.Query("Select sgu_id from scu where scu_id='" + (scid) + "'")
		defer ttrows.Close()
		chkErr(err, &w)
		var sguid int
		if ttrows.Next() {
			err = ttrows.Scan(&sguid)
			chkErr(err, &w)
		}
		cnt := 0
		for _, val := range ids {
			shgid, _ := strconv.ParseInt(tpris[cnt], 10, 64)
			cnt++
			trows, err := db.Query("Select * from schedule where idschedule='" + val + "'")
			defer trows.Close()
			chkErr(err, &w)
			var schid, pwm int
			var sst, set, se, tss string
			for trows.Next() {
				err = trows.Scan(&schid, &sst, &set, &se, &pwm, &tss)
				chkErr(err, &w)
			}
			//_, eorr:=stmt.Exec(val,scid,shgid,pwm,sst,set,se)
			//chkErr(eorr,&w)
			status := 0
			status = ((int)(1)) & 0x00FF
			status |= ((((int)(shgid)) << 8) & 0x00FF00)
			status |= ((((int)(pwm)) << 16) & 0x00FF0000)
			//for testing
			//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
			le := len(se)
			isc, _ := (strconv.Atoi(scid))

			LampController.SGUID = uint64(sguid)
			LampController.SCUID = uint64(isc)
			LampController.ConfigArray = []byte(se)
			LampController.ConfigArrayLength = le
			LampController.PacketType = 0x8000
			LampController.LampEvent = status
			LampController.ResponseSend = make(chan bool)
			LampControllerChannel <- LampController

		}
		//io.WriteString(w, "Saved Successfully!!")
		ans.Response_status = "Success"
		ans.Data.Message = "Saved Successfully!!"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func SetScheduleForSgu(w http.ResponseWriter, r *http.Request) {
	logger.Println("SetScheduleForSgu()")
	var LampController sguUtils.SguUtilsLampControllerStruct
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token

	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		sids := NBLampStr.Data.Schedule_Id
		if sids == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Schedule Id"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		sgid := NBLampStr.Fdn.Id
		if !validateSGU(sgid) {
			ans.Response_status = "fail"
			ans.Data.Message = "Invalid SGU"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		if !IsSGUInDb(sgid) {
			ans.Response_status = "fail"
			ans.Data.Message = "Gateway Not Found"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		pris := NBLampStr.Data.Priority
		if pris == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Priority"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		//value will be 1,2,3
		ids := strings.Split(sids, ",")
		tpris := strings.Split(pris, ",")
		logger.Println(ids)
		db := dbController.Db
		tmprows, err := db.Query("Select scu_id from scu where sgu_id='" + sgid + "'")
		/*	var tmprows_s [][]string
			tmprows_s = string(tmprows)
			if len(tmprows_s) == 0 {
				ans.Response_status = "fail"
				ans.Data.Message = "No SCU Found For Gateway"
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				return
			}*/
		defer tmprows.Close()
		chkErr(err, &w)
		for tmprows.Next() {
			var tscuid int
			err = tmprows.Scan(&tscuid)
			if err != nil {
				logger.Println(err)
				ans.Response_status = "fail"
				ans.Data.Message = "No SCU Found For Gateway"
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				return
			}
			defer tmprows.Close()
			chkErr(err, &w)
			//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
			//defer stmt.Close()
			//chkErr(err,&w)
			cnt := 0
			for _, val := range ids {
				trows, err := db.Query("Select idSCUSchedule from scuconfigure where ScuID='" + strconv.Itoa(tscuid) + "' and ScheduleID='" + val + "'")
				defer trows.Close()
				chkErr(err, &w)
				if trows.Next() {
					cnt++
					continue
				}
				shgid, _ := strconv.ParseInt(tpris[cnt], 10, 64)
				cnt++
				ttrows, err := db.Query("Select * from schedule where idschedule='" + val + "'")
				defer ttrows.Close()
				chkErr(err, &w)
				var schid, pwm int
				var sst, set, se, tss string
				for ttrows.Next() {
					err = ttrows.Scan(&schid, &sst, &set, &se, &pwm, &tss)
					chkErr(err, &w)
				}
				//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
				status := 0
				status = ((int)(1)) & 0x00FF
				status |= ((((int)(shgid)) << 8) & 0x00FF00)
				status |= ((((int)(pwm)) << 16) & 0x00FF0000)
				//for testing
				//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
				logger.Println("exp=", se)
				le := len(se)
				LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
				chkErr(err, &w)
				LampController.SCUID = uint64(tscuid)
				LampController.ConfigArray = []byte(se)
				LampController.ConfigArrayLength = le
				LampController.PacketType = 0x8000
				LampController.LampEvent = status
				LampController.ResponseSend = make(chan bool)
				du, _ := time.ParseDuration(scu_scheduling)
				time.Sleep(du)
				logger.Println("sent")
				LampControllerChannel <- LampController
				//chkErr(eorr,&w)
			}
		}
		//io.WriteString(w, "Saved Successfully!!")
		ans.Response_status = "Success"
		ans.Data.Message = "Saved Successfully!!"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func SetScheduleForZone(w http.ResponseWriter, r *http.Request) {
	logger.Println("SetScheduleForSgu()")
	var LampController sguUtils.SguUtilsLampControllerStruct
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token

	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		sids := NBLampStr.Data.Schedule_Id
		if sids == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Schedule Id"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		sgid := NBLampStr.Fdn.Id
		if sgid == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "Zone Not Specified"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		if !IsZoneIDInDB(sgid) {
			ans.Response_status = "fail"
			ans.Data.Message = "Zone Not Found"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		pris := NBLampStr.Data.Priority
		if pris == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Priority"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		//value will be 1,2,3
		ids := strings.Split(sids, ",")
		tpris := strings.Split(pris, ",")
		logger.Println(ids)
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		db := dbController.Db
		tmprows, err := db.Query("Select scu_id, sgu_id from scu where sgu_id in (select sguid from zone_sgu where zid='" + sgid + "')")
		defer tmprows.Close()
		chkErr(err, &w)
		for tmprows.Next() {
			var tscuid int
			var tsguid int
			err = tmprows.Scan(&tscuid, &tsguid)
			defer tmprows.Close()
			chkErr(err, &w)
			//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
			//defer stmt.Close()
			//chkErr(err,&w)
			cnt := 0
			for _, val := range ids {
				trows, err := db.Query("Select idSCUSchedule from scuconfigure where ScuID='" + strconv.Itoa(tscuid) + "' and ScheduleID='" + val + "'")
				defer trows.Close()
				chkErr(err, &w)
				if trows.Next() {
					cnt++
					continue
				}
				shgid, _ := strconv.ParseInt(tpris[cnt], 10, 64)
				cnt++
				ttrows, err := db.Query("Select * from schedule where idschedule='" + val + "'")
				defer ttrows.Close()
				chkErr(err, &w)
				var schid, pwm int
				var sst, set, se, tss string
				for ttrows.Next() {
					err = ttrows.Scan(&schid, &sst, &set, &se, &pwm, &tss)
					chkErr(err, &w)
				}
				//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
				status := 0
				status = ((int)(1)) & 0x00FF
				status |= ((((int)(shgid)) << 8) & 0x00FF00)
				status |= ((((int)(pwm)) << 16) & 0x00FF0000)
				//for testing
				//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
				logger.Println("exp=", se)
				le := len(se)
				//LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
				LampController.SGUID = uint64(tsguid)
				chkErr(err, &w)
				LampController.SCUID = uint64(tscuid)
				LampController.ConfigArray = []byte(se)
				LampController.ConfigArrayLength = le
				LampController.PacketType = 0x8000
				LampController.LampEvent = status
				LampController.ResponseSend = make(chan bool)
				du, _ := time.ParseDuration(scu_scheduling)
				time.Sleep(du)
				logger.Println("sent")
				LampControllerChannel <- LampController
				//chkErr(eorr,&w)
			}
		}
		//io.WriteString(w, "Saved Successfully!!")
		ans.Response_status = "Success"
		ans.Data.Message = "Saved Successfully!!"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func SetScheduleForGroup(w http.ResponseWriter, r *http.Request) {
	logger.Println("SetScheduleForGroup()")
	var LampController sguUtils.SguUtilsLampControllerStruct
	var ans NBResponseStruct
	fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		sids := NBLampStr.Data.Schedule_Id
		if sids == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Schedule Id"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		sgid := NBLampStr.Fdn.Id
		if sgid == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "Group Not Specified"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		if !IsGroupIDInDB(sgid) {
			ans.Response_status = "fail"
			ans.Data.Message = "Group Not Found"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		pris := NBLampStr.Data.Priority
		if pris == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "No Priority"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		//value will be 1,2,3
		ids := strings.Split(sids, ",")
		tpris := strings.Split(pris, ",")
		logger.Println(ids)
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		db := dbController.Db
		//tmprows, err := db.Query("select scuid from group_scu_rel where gid='" + sgid + "'")
		tmprows, err := db.Query("select scu_id, sgu_id from scu where scu_id in(select scuid from group_scu_rel where gid='" + sgid + "')")
		defer tmprows.Close()
		chkErr(err, &w)
		for tmprows.Next() {
			var tscuid, tsguid int
			err = tmprows.Scan(&tscuid, &tsguid)
			defer tmprows.Close()
			chkErr(err, &w)
			//stmt, err := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
			//defer stmt.Close()
			//chkErr(err,&w)
			cnt := 0
			for _, val := range ids {
				trows, err := db.Query("Select idSCUSchedule from scuconfigure where ScuID='" + strconv.Itoa(tscuid) + "' and ScheduleID='" + val + "'")
				defer trows.Close()
				chkErr(err, &w)
				if trows.Next() {
					cnt++
					continue
				}
				shgid, _ := strconv.ParseInt(tpris[cnt], 10, 64)
				cnt++
				ttrows, err := db.Query("Select * from schedule where idschedule='" + val + "'")
				defer ttrows.Close()
				chkErr(err, &w)
				var schid, pwm int
				var sst, set, se, tss string
				for ttrows.Next() {
					err = ttrows.Scan(&schid, &sst, &set, &se, &pwm, &tss)
					chkErr(err, &w)
				}
				//_, eorr:=stmt.Exec(val,tscuid,shgid,pwm,sst,set,se)
				status := 0
				status = ((int)(1)) & 0x00FF
				status |= ((((int)(shgid)) << 8) & 0x00FF00)
				status |= ((((int)(pwm)) << 16) & 0x00FF0000)
				//for testing
				//se="((D==24&&M==12&&Y==2015)&&(T>=19:10&&T<=19:30))"
				logger.Println("exp=", se)
				le := len(se)
				//LampController.SGUID, err = (strconv.ParseUint(sgid, 10, 64))
				LampController.SGUID = uint64(tsguid)
				chkErr(err, &w)
				LampController.SCUID = uint64(tscuid)
				LampController.ConfigArray = []byte(se)
				LampController.ConfigArrayLength = le
				LampController.PacketType = 0x8000
				LampController.LampEvent = status
				LampController.ResponseSend = make(chan bool)
				du, _ := time.ParseDuration(scu_scheduling)
				time.Sleep(du)
				logger.Println("sent")
				LampControllerChannel <- LampController
				//chkErr(eorr,&w)
			}
		}
		//io.WriteString(w, "Saved Successfully!!")
		ans.Response_status = "Success"
		ans.Data.Message = "Saved Successfully!!"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func DeleteGateWaysFromZone(w http.ResponseWriter, r *http.Request) {
	logger.Println("SetScheduleForGroup()")
	var ans NBResponseStruct
	fmt.Println("ans", ans)
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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	fmt.Println("l_opr", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_gateway_ids_array := NBLampStr.Data.Ids
		if len(l_gateway_ids_array) == 0 {
			ans.Response_status = "fail"
			ans.Data.Message = "Gateways Not Specified"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		l_zone_id := NBLampStr.Data.ZoneId
		for i := 0; i < len(l_gateway_ids_array); i++ {
			if !validateSGU(l_gateway_ids_array[i]) {
				ans.Response_status = "fail"
				ans.Data.Message = "Invalid SGU"
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				return
			}
			if !IsSGUInDb(l_gateway_ids_array[i]) {
				ans.Response_status = "fail"
				ans.Data.Message = "Gateway Not Found"
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				return
			}
			if IsSGUInZone(l_zone_id, l_gateway_ids_array[i]) {
				ans.Response_status = "fail"
				ans.Data.Message = "No SGU To Delete"
				logger.Println("SGU: %s No SGU To Delete", l_gateway_ids_array[i])
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("Error in json.Marshal")
					logger.Println(err)
				} else {
					w.Write(a)
				}
				return
			}
		}

		if l_zone_id == "" {
			ans.Response_status = "fail"
			ans.Data.Message = "Zone Id Not Specified"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}

		if !IsZoneIDInDB(l_zone_id) {
			ans.Response_status = "fail"
			ans.Data.Message = "Zone Id Not Specified"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("Error in json.Marshal")
				logger.Println(err)
			} else {
				w.Write(a)
			}
			return
		}
		dbController.DbSemaphore.Lock()
		defer dbController.DbSemaphore.Unlock()
		db := dbController.Db
		stmt, err := db.Prepare("Delete from zone_sgu where zid=? and sguid=?")
		defer stmt.Close()
		chkErr(err, &w)

		for i := 0; i < len(l_gateway_ids_array); i++ {
			_, eorr := stmt.Exec(l_zone_id, l_gateway_ids_array[i])
			chkErr(eorr, &w)
		}
		//io.WriteString(w, "Saved Successfully!!")
		ans.Response_status = "Success"
		ans.Data.Message = "Deleted Successfully!!"
		ans.Data.Ids = l_gateway_ids_array
		ans.Data.ZoneId = l_zone_id
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return

	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.Write(a)
		}
		return
	}
}

func GroupView(w http.ResponseWriter, r *http.Request) {
	logger.Println("GroupVIew()")
	var ans SCUViewResp

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
			return
		}
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)

	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_obj := NBLampStr.Object
		logger.Println("Object is:", l_obj)
		if l_obj != "group" {
			logger.Println("Invalid Object")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		l_opr := NBLampStr.Opr
		logger.Println("operation is:", l_opr)
		if l_opr == "get_schedule" {

			var temp SCUVIewStr
			temp1 := make(map[string]SCUVIewStr)
			tempData := make(map[string]map[string]SCUVIewStr)
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			db := dbController.Db
			gid := NBLampStr.Fdn.Group
			rows, err := db.Query("Select scuconfigure.ScheduleID,scuconfigure.ScheduleStartTime,scuconfigure.ScheduleEndTime,scuconfigure.pwm,scuconfigure.SchedulingID from scuconfigure inner join group_scu_rel on group_scu_rel.scuid=scuconfigure.scuID and group_scu_rel.gid=?", gid)
			defer rows.Close()
			if err != nil {
				logger.Println("Error while fetching schedule from DB: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500-internal server error"))
			} else {
				for rows.Next() {
					var sid, sst, set, pwm, prt string
					rows.Scan(&sid, &sst, &set, &pwm, &prt)
					sstSlice := strings.Split(sst, " ")
					ssd := sstSlice[0]
					sst = sstSlice[1]
					setSlice := strings.Split(set, " ")
					sed := setSlice[0]
					set = setSlice[1]
					temp.Brightness = pwm
					temp.Priority = prt
					temp.SEDate = sed
					temp.SETime = set
					temp.SSDate = ssd
					temp.SSTime = sst

					temp1[sid] = temp
				}
				tempData["schedules"] = temp1
				ans.Data = tempData
				ans.Response_status = "success"
				ans.End = "end"
			}
		}

		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("error in marshalling: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - internal server error"))
		} else {
			w.Write(a)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}

func ZoneView(w http.ResponseWriter, r *http.Request) {
	logger.Println("ZoneView()")
	var ans SCUViewResp

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
			return
		}
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)

	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_obj := NBLampStr.Object
		logger.Println("Object is:", l_obj)
		if l_obj != "zone" {
			logger.Println("Invalid Object")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		l_opr := NBLampStr.Opr
		logger.Println("operation is:", l_opr)
		if l_opr == "get_schedule" {

			var temp SCUVIewStr
			temp1 := make(map[string]SCUVIewStr)
			tempData := make(map[string]map[string]SCUVIewStr)
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			db := dbController.Db
			id := NBLampStr.Fdn.Zone
			rows, err := db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID from scuconfigure inner join scu on scu.scu_id=scuconfigure.scuID and scu.sgu_id IN(select sguid from zone_sgu where zid=?)", id)
			defer rows.Close()
			if err != nil {
				logger.Println("Error while fetching schedule from DB: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500-internal server error"))
			} else {
				for rows.Next() {
					var sid, sst, set, pwm, prt string
					rows.Scan(&sid, &sst, &set, &pwm, &prt)
					sstSlice := strings.Split(sst, " ")
					ssd := sstSlice[0]
					sst = sstSlice[1]
					setSlice := strings.Split(set, " ")
					sed := setSlice[0]
					set = setSlice[1]
					temp.Brightness = pwm
					temp.Priority = prt
					temp.SEDate = sed
					temp.SETime = set
					temp.SSDate = ssd
					temp.SSTime = sst

					temp1[sid] = temp
				}
				tempData["schedules"] = temp1
				ans.Data = tempData
				ans.Response_status = "success"
				ans.End = "end"
			}
		}

		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("error in marshalling: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - internal server error"))
		} else {
			w.Write(a)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}

func SGUView(w http.ResponseWriter, r *http.Request) {
	logger.Println("SGUView()")
	var ans SCUViewResp

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
			return
		}
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)

	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_obj := NBLampStr.Object
		logger.Println("Object is:", l_obj)
		if l_obj != "gateway" {
			logger.Println("Invalid Object")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		l_opr := NBLampStr.Opr
		logger.Println("operation is:", l_opr)
		if l_opr == "get_schedule" {

			var temp SCUVIewStr
			temp1 := make(map[string]SCUVIewStr)
			tempData := make(map[string]map[string]SCUVIewStr)
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			db := dbController.Db
			id := NBLampStr.Fdn.Gateway
			rows, err := db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID from scuconfigure inner join scu on scu.scu_id=scuconfigure.scuID and scu.sgu_id=?", id)
			defer rows.Close()
			if err != nil {
				logger.Println("Error while fetching schedule from DB: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500-internal server error"))
			} else {
				for rows.Next() {
					var sid, sst, set, pwm, prt string
					rows.Scan(&sid, &sst, &set, &pwm, &prt)
					sstSlice := strings.Split(sst, " ")
					ssd := sstSlice[0]
					sst = sstSlice[1]
					setSlice := strings.Split(set, " ")
					sed := setSlice[0]
					set = setSlice[1]
					temp.Brightness = pwm
					temp.Priority = prt
					temp.SEDate = sed
					temp.SETime = set
					temp.SSDate = ssd
					temp.SSTime = sst

					temp1[sid] = temp
				}
				tempData["schedules"] = temp1
				ans.Data = tempData
				ans.Response_status = "success"
				ans.End = "end"
			}
		}

		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("error in marshalling: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - internal server error"))
		} else {
			w.Write(a)
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}

func SCUView(w http.ResponseWriter, r *http.Request) {
	logger.Println("SCUView()")
	var ans SCUViewResp

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
			return
		}
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)

	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_obj := NBLampStr.Object
		logger.Println("Object is:", l_obj)
		if l_obj != "street_lamp" {
			logger.Println("Invalid Object")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		l_opr := NBLampStr.Opr
		logger.Println("operation is:", l_opr)
		if l_opr == "get_schedule" {

			var temp SCUVIewStr
			temp1 := make(map[string]SCUVIewStr)
			tempData := make(map[string]map[string]SCUVIewStr)
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			db := dbController.Db
			id := NBLampStr.Fdn.Street_lamp
			rows, err := db.Query("Select ScheduleID,ScheduleStartTime,ScheduleEndTime,pwm,SchedulingID from scuconfigure where ScuID='" + id + "'")
			defer rows.Close()
			if err != nil {
				logger.Println("Error while fetching schedule from DB: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500-internal server error"))
			} else {

				for rows.Next() {
					var sid, sst, set, pwm, prt string
					rows.Scan(&sid, &sst, &set, &pwm, &prt)
					sstSlice := strings.Split(sst, " ")
					ssd := sstSlice[0]
					sst = sstSlice[1]
					setSlice := strings.Split(set, " ")
					sed := setSlice[0]
					set = setSlice[1]
					temp.Brightness = pwm
					temp.Priority = prt
					temp.SEDate = sed
					temp.SETime = set
					temp.SSDate = ssd
					temp.SSTime = sst

					temp1[sid] = temp
				}
				tempData["schedules"] = temp1
				ans.Data = tempData
				ans.Response_status = "success"
				ans.End = "end"
			}
		}

		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("error in marshalling: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - internal server error"))
		} else {
			w.Write(a)
		}

	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}

func GetSchedule(w http.ResponseWriter, r *http.Request) {
	logger.Println("GetSchedule()")
	var ans ScheduleResp

	parse_err := r.ParseForm()
	if parse_err != nil {
		logger.Println(parse_err)

	}
	system := r.Form["system_id"]
	logger.Println("System id is :", system)
	var NBLampStr NBAllLampControlStruct
	if len(r.FormValue("token")) == 0 {
		decoder := json.NewDecoder(r.Body)
		logger.Println(decoder)
		err := decoder.Decode(&NBLampStr)
		if err != nil {
			logger.Println(err)
			return
		}
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	logger.Println("operation is:", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		if l_opr == "get_schedule" {

			db := dbController.Db
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			temp := make(map[string]ScheduleStr)
			var temp1 ScheduleStr
			var id, pwm, sst, set string

			//			loc, _ := time.LoadLocation("Asia/Calcutta")
			logger.Println("fetching schedules from DB")
			rows, err := db.Query("SELECT idschedule, ScheduleStartTime, ScheduleEndTime, pwm from schedule")
			if err != nil {
				logger.Println("Error while reading from DB: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				ans.Response_status = "fail"
			} else {
				ans.Response_status = "Success"

				for rows.Next() {
					rows.Scan(&id, &sst, &set, &pwm)
					//sst1, _ := time.ParseInLocation("2006-01-02 15:04:05", sst, loc)
					//set1, _ := time.ParseInLocation("2006-01-02 15:04:05", set, loc)

					sstSlice := strings.Split(sst, " ")
					ssd := sstSlice[0]
					sst := sstSlice[1]
					setSlice := strings.Split(set, " ")
					sed := setSlice[0]
					set := setSlice[1]
					temp1.SSDate = ssd
					temp1.SEDate = sed
					temp1.SSTime = sst
					temp1.SETime = set
					temp1.Brightness = pwm
					temp[id] = temp1
				}

			}
			tempData := make(map[string]map[string]ScheduleStr)
			tempData["schedules"] = temp
			ans.Data = tempData
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("error in marshalling: ", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Write(a)
			}
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}

func CreateZone(w http.ResponseWriter, r *http.Request) {
	logger.Println("CreateZone()")
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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	logger.Println("operation is:", l_opr)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		if l_opr == "create_zone" {

			zone := NBLampStr.Data.Zone
			logger.Println("zone name:", zone)
			db := dbController.Db
			dbController.DbSemaphore.Lock()
			defer dbController.DbSemaphore.Unlock()
			logger.Println("checking in DB")
			rows, err := db.Query("SELECT id FROM zone WHERE name=?", zone)
			if err != nil {

				logger.Println("error while reading from DB", err)
				return
			} else {
				defer rows.Close()
				if rows.Next() {
					ans.Data.Message = "Zone name already exist"
					ans.Response_status = "fail"

					a, err := json.Marshal(ans)
					if err != nil {

						logger.Println("Error in marshelling")
						w.WriteHeader(http.StatusInternalServerError)

						return
					} else {
						w.Write(a)
					}
					return
				}

			}

			stmt, err := db.Prepare("INSERT into zone set name=?")
			defer stmt.Close()
			if err != nil {
				logger.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, eorr := stmt.Exec(zone)
			if eorr != nil {
				logger.Println(eorr)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ans.Response_status = "success"
			ans.Data.Message = "zone created successfully"
			a, err := json.Marshal(ans)
			if err != nil {
				logger.Println("error in marshelling", err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.Write(a)
			}
			return
		}
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(a)
		}
		return
	}

}

func DeleteScufromGroup(scuarray []string, gid string) NBResponseStruct {
	logger.Println("inside DeleteScufromGroup() ")
	var ans NBResponseStruct
	scuids := scuarray
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	logger.Println("scuids:", scuids)
	logger.Println("executing query in DB")
	//	query := "delete from group_scu_rel where scuid IN (%s)," + strings.Join(strings.Split(strings.Repeat("?", (len(scuids))-1), ""), ",") + "and gid=?"
	//	rows, err := db.Query("delete from group_scu_rel where scuid IN ("+scuids+") and gid=?", gid)
	//rows, err := db.Query("delete from group_scu_rel where scuid IN (%s)", strings.Join(strings.Split(strings.Repeat("?", (len(scuids))-1), ""), ",")+"and gid=?", gid)
	for i := 0; i < len(scuids); i++ {

		logger.Println("inside for loop to delete SCUIds from group", scuids[i], gid)
		scu := scuids[i]
		_, err := db.Query("delete from group_scu_rel where scuid=? and gid=?", scu, gid)

		if err != nil {
			logger.Println(err)
			ans.Response_status = "fail"
			logger.Println(ans.Response_status)
			break

		}
	}
	//stmt, _ := db.Prepare(query)
	//rows, _ := stmt.Query(scuids, gid)
	logger.Println("deletion done")
	if ans.Response_status != "fail" {
		ans.Response_status = "success"
		ans.Data.Message = ""
		logger.Println("Deleted SCUs from Group successfully")
	}

	return ans
}

func GetStatus(w http.ResponseWriter, r *http.Request) {

	logger.Println("GetStatus()")

	parse_err := r.ParseForm()
	if parse_err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500- Internal server error"))
		return
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
		//NBLampStr.Fdn = r.FormValue("fdn")
		NBLampStr.Fdn.System = r.FormValue("fdn.system")
		logger.Println("fdn system is :", NBLampStr.Fdn.System)
		NBLampStr.Object = r.FormValue("object")
		NBLampStr.Opr = r.FormValue("opr")
	}
	//logger.Println("body data:", NBLampStr)
	l_token := NBLampStr.Token

	logger.Println("token is :", l_token)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_object := NBLampStr.Object
		l_opr := NBLampStr.Opr
		l_system := NBLampStr.Fdn.System
		if l_system == "" {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid system"))
			return
		}
		if l_object == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400- Invalid Object"))
			return
		}
		if l_opr == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400- Invalid Operation"))
			return
		} else {
			logger.Println("Operation:", l_opr)
			switch l_opr {

			case "get_street_lamp_status":
				var ans LampPowerStatus
				lamp := NBLampStr.Fdn.Street_lamp
				ans = GetLampPowerStatus(lamp)
				if ans.Response_status == "fail" {
					w.WriteHeader(http.StatusInternalServerError)
					//w.Write()
					return
				} else {
					a, err := json.Marshal(ans)
					if err != nil {
						logger.Println(err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					} else {

						w.Write(a)
					}
					return
				}
			case "get_gateway_power_status":
				var ans LampPowerStatus
				sguid := NBLampStr.Fdn.Gateway
				ans = GetSGUStatus(sguid)
				if ans.Response_status == "fail" {
					w.WriteHeader(http.StatusInternalServerError)
				} else {

					a, err := json.Marshal(ans)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
					} else {
						w.Write(a)
					}
				}

			}

		}
	} else {

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401- Invalid Token"))
	}
	return
}

func GetSGUStatus(sguid string) LampPowerStatus {

	logger.Println("Inside GetSGUStatus()")
	db := dbController.Db
	var response LampPowerStatus
	var LastUpdateTime, scuid, status string
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("SELECT scu_status.status, scu_status.timestamp, scu_status.scu_id from scu_status inner join scu on scu_status.scu_id=scu.scu_id and scu.sgu_id=?", sguid)
	if err != nil {
		logger.Println(err)
		response.Response_status = "fail"
		return response
	}
	defer rows.Close()
	temp := make(map[string]string)
	for rows.Next() {
		response.Response_status = "pass"
		rows.Scan(&status, &LastUpdateTime, &scuid)
		logger.Println("data from DB is %s, %s, %s", status, LastUpdateTime, scuid)
		loc, _ := time.LoadLocation("UTC")
		statime, _ := time.ParseInLocation("2006-01-02 15:04:05", LastUpdateTime, loc)
		currTime := time.Now().Add(-3 * time.Minute)
		if currTime.After(statime) {
			temp[scuid] = "Operation failed"
			//response.Data[scuid] = "Operation failed"
		} else {

			temp[scuid] = "Operation Successfull"
			switch status {

			case "0":
				status = "OFF"
			case "1":
				status = "ON"
			default:
				status = "To be updated"
			}
			temp[scuid] = "Operation Successfull current status" + status
		}

	}
	response.Data = temp
	return response
}

func GetLampPowerStatus(lamp string) LampPowerStatus {
	logger.Println("Inside GetLampPowerStatus()")
	db := dbController.Db
	var response LampPowerStatus
	var LastUpdateTime, status, scuid string
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	temp := make(map[string]string)
	rows, err := db.Query("select scu_id,timestamp,status from scu_status where scu_id =?", lamp)
	if err != nil {
		logger.Println(err)
		response.Response_status = "fail"
		return response
	}
	defer rows.Close()
	if rows.Next() {

		rows.Scan(&scuid, &LastUpdateTime, &status)
	}
	currTime := time.Now().Add(-3 * time.Minute)
	loc, _ := time.LoadLocation("UTC")
	statime, _ := time.ParseInLocation("2006-01-02 15:04:05", LastUpdateTime, loc)
	if currTime.After(statime) {
		response.Response_status = "pass"
		temp[scuid] = "Operation failed"
		//response.Data[scuid] = "Operation failed"
	} else {
		response.Response_status = "pass"
		//response.Data = "Operation Successfull"
		switch status {

		case "0":
			status = "OFF"
		case "1":
			status = "ON"
		default:
			status = "To be updated"
		}

		temp[scuid] = "Operation Successfull current status is " + status
	}
	response.Data = temp
	return response

}

func GetStreetLamp(w http.ResponseWriter, r *http.Request) {
	parse_err := r.ParseForm()
	if parse_err != nil {

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500- Internal server error"))
		return
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
		//NBLampStr.Fdn = r.FormValue("fdn")
		NBLampStr.Fdn.System = r.FormValue("fdn.system")
		logger.Println("fdn system is :", NBLampStr.Fdn.System)
		NBLampStr.Object = r.FormValue("object")
		NBLampStr.Opr = r.FormValue("opr")
	}
	//logger.Println("body data:", NBLampStr)
	l_token := NBLampStr.Token

	logger.Println("token is :", l_token)
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		l_object := NBLampStr.Object
		l_opr := NBLampStr.Opr
		l_system := NBLampStr.Fdn.System
		if l_system == "" {

			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid system"))
			return
		}
		if l_object == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400- Invalid Object"))
			return
		}
		if l_opr != "" {

			if l_opr == "get_street_lamp_list" {

				var ans GatewayResponse
				logger.Println("fdn:", NBLampStr.Fdn)
				sguid := NBLampStr.Fdn.Gateway
				logger.Println("sgu id is:", sguid)
				ans.Data = StreetLampInfo(sguid)
				ans.Response_status = "success"
				ans.End = "end"
				//logger.Println("response is:", ans.data)
				//logger.Println("END:", ans.data)
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("error in marshelling")
				} else {

					logger.Println(string(a))
					w.Write(a)
				}
			}
		}

	} else {

		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Invalid token"))
	}
	return
}
func StreetLampInfo(SGUId string) map[string]StreetLampDetails {
	logger.Println("Inside StreetLampInfo()")
	var ans = make(map[string]StreetLampDetails)

	logger.Println("sgu id is:", SGUId)
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from scu inner join scu_status on scu_status.scu_id = scu.scu_id and scu.sgu_id=?", SGUId)
	if err != nil {
		return ans
	}
	//logger.Println("Rows received from DB:", rows)
	var scu, l, la, ln, st string
	var temp StreetLampDetails
	//defer rows.Close()
	for rows.Next() {

		//err = rows.Scan(&scu, &temp.location_name, &temp.location_lat, &temp.location_lng, &temp.status)
		err := rows.Scan(&scu, &l, &la, &ln, &st)
		//err := rows.Scan(&mac)
		if err != nil {
			return ans
		}
		logger.Println("Db status:", st)
		temp.Location_lat = la
		temp.Location_lng = ln
		temp.Location_name = l
		tempStatus := tcpUtils.GetTempStatus(scu)
		if tempStatus == "0" {
			temp.Status = "OFF"
		} else if tempStatus == "1" {
			temp.Status = "ON"
		} else {
			temp.Status = "UNKNOWN"
		}

		/*switch st {

		case "0":
			temp.Status = "OFF"

		case "1":
			temp.Status = "ON"

		default:
			temp.Status = "UNKNOWN"
		}*/
		ans[scu] = temp
	}
	status.RUnlock()
	rows.Close()
	//ans = LampDetails
	logger.Println("Ans is:", ans)
	return ans

}

func DeleteLamp(w http.ResponseWriter, r *http.Request) {
	logger.Println("SystemGroupURL()")
	var ans NBResponseStruct
	//fmt.Println("ans", ans)

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
		NBLampStr.Opr = r.FormValue("opr")
	}
	l_token := NBLampStr.Token
	fmt.Println("l_token", l_token)
	l_opr := NBLampStr.Opr
	_, bv := TokenParse_errorChecking(l_token)
	if bv {
		logger.Println("operation is:", l_opr)
		switch l_opr {

		case "del_street_lamps_from_a_group":
			ids := NBLampStr.Data.Ids
			gid := NBLampStr.Data.GroupId
			logger.Println("values for SCUids to be deleted:", ids)
			logger.Println("group id:", gid)
			ans = DeleteScufromGroup(ids, gid)

			if ans.Response_status == "fail" {

				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 - " + ans.Data.Message))
			} else {
				a, err := json.Marshal(ans)
				if err != nil {
					logger.Println("error in Marshalling: ", err)
					w.WriteHeader(http.StatusInternalServerError)
				} else {
					w.Write(a)
				}
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400- Invalid Operation"))
		}
	} else {
		ans.Response_status = "fail"
		ans.Data.Message = "Invalid Token"
		a, err := json.Marshal(ans)
		if err != nil {
			logger.Println("Error in json.Marshal")
			logger.Println(err)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(a)
		}
	}
	return
}
