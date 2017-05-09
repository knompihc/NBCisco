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
	//	"fmt"
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
	//	"strings"
	//	"tcpServer"
	//	"tcpUtils"
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
	SendSMSChan chan string
	logger      *log.Logger
	tokenMap    (map[string]int)
)

func InitNBApis(LampConChannel chan sguUtils.SguUtilsLampControllerStruct, dbcon dbUtils.DbUtilsStruct, logg *log.Logger) {
	logger = logg
	LampControllerChannel = LampConChannel
	dbController = dbcon
}

type NBFdn struct {
	System      string `json:"system"`
	Gateway     string `json:"gateway"`
	Street_lamp string `json:"street_lamp"`
}
type NBData struct {
	Brightness    string                                  `json:"brightness"`
	Message       string                                  `json:"msg"`
	Token         string                                  `json:token`
	Email         string                                  `json:email`
	Discovery_map map[string]map[string]map[string]string `json:discovery`
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
	l_opr := NBLampStr.Opr
	l_system := NBLampStr.Fdn.System
	//l_fdn := NBLampStr.Fdn
	l_brightness := NBLampStr.Data.Brightness
	//l_data := NBLampStr.Data
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
	password := t.Password
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
}

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
	l_opr := NBLampStr.Opr
	logger.Println("l_opr", l_opr)
	l_system := NBLampStr.Fdn.System
	logger.Println("l_system", l_system)
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
		rows, err := db.Query("Select sgu.sgu_id,sgu.location_name,scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from zone_sgu inner join zone on zone.id=zone_sgu.zid inner join sgu on sgu.sgu_id=zone_sgu.sguid inner join scu on scu.sgu_id=zone_sgu.sguid inner join scu_status on scu_status.scu_id=scu.scu_id")
		//rows, err := db.Query("Select zone_sgu.zid,zone.name,sgu.sgu_id,sgu.location_name,scu.scu_id,scu.location_name,scu.location_lat,scu.location_lng,scu_status.status from zone_sgu inner join zone on zone.id=zone_sgu.zid inner join sgu on sgu.sgu_id=zone_sgu.sguid inner join scu on scu.sgu_id=zone_sgu.sguid inner join scu_status on scu_status.scu_id=scu.scu_id")
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
