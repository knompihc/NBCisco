/********************************************************************
 * FileName:     mapview.go
 * Project:      Havells StreetComm
 * Module:       mapview
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package mapview

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	"dbUtils"
	"sguUtils"
	"tcpUtils"
	"userMgmt"
)

type Zone struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Group struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Tags struct {
	Tag int
}

type scu struct {
	Name   string `json:"text"`
	Lat    string `json:"lat"`
	Lng    string `json:"lng"`
	Id     string `json:"scuid"`
	Sguid  string `json:"sguid"`
	Status string `json:"status"`
}

type sgu struct {
	Sguname string `json:"text"`
	Sguid   string `json:"sguid"`
	Scus    []scu  `json:"nodes"`
	Color   string `json:"backColor"`
	Tag     Tags   `json:"tags"`
}

type res struct {
	Zname string `json:"text"`
	Zid   string `json:"zid"`
	Sgus  []sgu  `json:"nodes"`
	Tag   Tags   `json:"tags"`
}

type zo struct {
	val res
}

type resgrp struct {
	Gname string `json:"text"`
	Gid   string `json:"gid"`
	Sgus  []sgu  `json:"nodes"`
	Tag   Tags   `json:"tags"`
}

type gro struct {
	val resgrp
}

var dbController dbUtils.DbUtilsStruct
var logger *log.Logger

func InitMapview(dbCon dbUtils.DbUtilsStruct, logg *log.Logger) {
	dbController = dbCon
	logger = logg
}

func Showmap(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	zid := r.URL.Query().Get("id")
	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select * from scu where sgu_id in (select sguid from zone_sgu where zid='" + zid + "') " +
		"and location_lat is NOT NULL and location_lng is NOT NULL")
	chkErr(err, &w)
	defer rows.Close()

	data := "["
	fl := false
	for rows.Next() {
		if fl {
			data += ","
		} else {
			fl = true
		}
		var scu, sgu uint64
		var lat, lng float64
		var locname string
		err = rows.Scan(&scu, &sgu, &locname, &lat, &lng)
		trows, err := db.Query("Select status from scu_status where scu_id='" +
			strconv.FormatUint(scu, 10) + "' order by timestamp desc limit 1")
		chkErr(err, &w)
		defer trows.Close()

		var st uint64
		st = 10
		for trows.Next() {
			trows.Scan(&st)
		}
		sta := st & (0x0FF)
		logger.Println("STATUS before=", sta)
		sta = sta & 3
		tempscu := strconv.FormatUint(scu, 10)
		state := tcpUtils.GetTempStatus(string(tempscu))
		logger.Println("STATUS=", sta)
		status := "GREY"
		if state == "0" {
			status = "RED"
		} else if state == "1" {
			status = "GREEN"
		} else {
			status = "BLACK"
		}
		lats := strconv.FormatFloat(lat, 'f', -1, 64)
		lngs := strconv.FormatFloat(lng, 'f', -1, 64)
		scus := strconv.FormatUint(scu, 10)
		sgus := strconv.FormatUint(sgu, 10)
		data += "{\"lat\":\"" + lats + "\",\"lng\":\"" + lngs + "\",\"sgu\":\"" + sgus + "\",\"scu\":\"" +
			scus + "\",\"status\":\"" + status + "\"}"
	}
	data += "]"
	logger.Println(data)
	io.WriteString(w, data)
}

func Getzone(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select id,name from zone")
	chkErr(err, &w)
	defer rows.Close()

	data := []Zone{}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		tm := Zone{}
		tm.Name = name
		tm.Id = id
		data = append(data, tm)
	}
	logger.Println(data)
	if a, err := json.Marshal(data); err != nil {
		logger.Println("Error in json.Marshal: ", err)
	} else {
		w.Write(a)
	}
}

func Getgroup(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select id,name from groupscu")
	defer rows.Close()
	chkErr(err, &w)
	data := []Group{}
	for rows.Next() {
		var id, name string
		rows.Scan(&id, &name)
		tm := Group{}
		tm.Name = name
		tm.Id = id
		data = append(data, tm)
	}
	logger.Println(data)
	if a, err := json.Marshal(data); err != nil {
		logger.Println("Error in json.Marshal: ", err)
	} else {
		w.Write(a)
	}
}

func Getall(w http.ResponseWriter, r *http.Request) {
	if !userMgmt.IsSessionValid(w, r) {
		return
	}

	db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows, err := db.Query("Select zone_sgu.zid,zone.name,sgu.sgu_id,sgu.location_name,scu.scu_id," +
		"scu.location_name,scu.location_lat,scu.location_lng,scu_status.status,scu_status.timestamp " +
		"from zone_sgu inner join zone on zone.id=zone_sgu.zid inner join sgu on sgu.sgu_id=zone_sgu.sguid" +
		" inner join scu on scu.sgu_id=zone_sgu.sguid inner join scu_status on scu_status.scu_id=scu.scu_id")
	chkErr(err, &w)
	defer rows.Close()

	data := []res{}
	mz := make(map[string]res)
	mg := make(map[string]sgu)
	mc := make(map[string]scu)
	cmz := make(map[string]int)
	cmg := make(map[string]int)
	for rows.Next() {
		var zid, zname, sguid, sguname, scuid, scuname, lat, lng, ts string
		var st uint64
		st = 10
		rows.Scan(&zid, &zname, &sguid, &sguname, &scuid, &scuname, &lat, &lng, &st, &ts)
		sta := st & (0x0FF)
		logger.Println("STATUS before=", sta)
		sta = sta & 3
		logger.Println("STATUS=", sta)
		status := "GREY"
		state := tcpUtils.GetTempStatus(scuid)

		if state == "0" {
			status = "RED"
		} else if state == "1" {
			status = "GREEN"
		} else {
			status = "BLACK"
		}

		if cmz[zid] != 1 {
			tz := res{}
			tz.Zid = zid
			tz.Zname = zname
			tg := sgu{}

			tg.Sguid = sguid
			ui, _ := strconv.ParseUint(sguid, 10, 64)
			SguIndex := sguUtils.GetSGURamListIndex(ui)
			if SguIndex != -1 {
				tg.Color = "#A2CD5A"
			} else {
				logger.Println("Sguid=", sguid, " Not Connected!!")
			}
			if len(sguname) == 0 || sguname == " " {
				tg.Sguname = sguid
			} else {
				tg.Sguname = sguname
			}
			tc := scu{}
			tc.Id = scuid
			if len(scuname) == 0 {
				tc.Name = scuid
			} else {
				tc.Name = scuname
			}
			tc.Lat = lat
			tc.Lng = lng
			tc.Sguid = sguid
			tc.Status = status
			tg.Scus = append(tg.Scus, tc)
			tg.Tag.Tag = len(tg.Scus)
			tz.Sgus = append(tz.Sgus, tg)
			tz.Tag.Tag = len(tz.Sgus)
			mg[sguid] = tg
			mz[zid] = tz
			mc[scuid] = tc
			cmg[sguid] = 1
			cmz[zid] = 1
		} else if cmg[sguid] != 1 {
			tz := mz[zid]
			tg := sgu{}

			tg.Sguid = sguid
			ui, _ := strconv.ParseUint(sguid, 10, 64)
			SguIndex := sguUtils.GetSGURamListIndex(ui)
			if SguIndex != -1 {
				tg.Color = "#A2CD5A"
			} else {
				logger.Println("Sguid=", sguid, " Not Connected!!")
			}
			if len(sguname) == 0 {
				tg.Sguname = sguid
			} else {
				tg.Sguname = sguname
			}

			tc := scu{}
			tc.Id = scuid
			if len(scuname) == 0 {
				tc.Name = scuid
			} else {
				tc.Name = scuname
			}
			tc.Lat = lat
			tc.Lng = lng
			tc.Sguid = sguid
			tc.Status = status
			tg.Scus = append(tg.Scus, tc)
			tg.Tag.Tag = len(tg.Scus)
			tz.Sgus = append(tz.Sgus, tg)
			tz.Tag.Tag = len(tz.Sgus)
			mg[sguid] = tg
			mz[zid] = tz
			mc[scuid] = tc
			cmg[sguid] = 1
			cmz[zid] = 1

		} else {
			tz := mz[zid]
			tg := mg[sguid]
			tc := scu{}
			tc.Id = scuid
			if len(scuname) == 0 {
				tc.Name = scuid
			} else {
				tc.Name = scuname
			}
			tc.Lat = lat
			tc.Lng = lng
			tc.Sguid = sguid
			tc.Status = status
			tot := 0
			for k1, re := range mz {
				for k2, te := range re.Sgus {
					if te.Sguid == sguid {
						mz[k1].Sgus[k2].Scus = append(te.Scus, tc)
						mz[k1].Sgus[k2].Tag.Tag = len(mz[k1].Sgus[k2].Scus)
					}
					if re.Zid == zid {
						tot += len(mz[k1].Sgus[k2].Scus)
					}
				}
			}
			tz.Tag.Tag = (tot)
			mg[sguid] = tg
			mz[zid] = tz
			mc[scuid] = tc
			cmg[sguid] = 1
			cmz[zid] = 1
		}
	}
	for _, re := range mz {
		data = append(data, re)
	}
	if a, err := json.Marshal(data); err != nil {
		logger.Println("Error in json.Marshal: ", err)
	} else {
		w.Write(a)
	}
}

func chkErr(err error, r *http.ResponseWriter) {
	if err != nil {
		io.WriteString(*r, (err.Error()))
		logger.Println(err)
		//panic(err);
	}
}
