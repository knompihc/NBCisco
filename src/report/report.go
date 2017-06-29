/********************************************************************
 * FileName:     report.go
 * Project:      Havells StreetComm
 * Module:       report
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package report

import (
	"dbUtils"
	"fmt"
	"net/smtp"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gocron"
	"github.com/scorredoira/email"
	/*"os"
	"strings"*/
	"os"

	"github.com/xlsx"
)

func InitSendreport(dbcon dbUtils.DbUtilsStruct) {
	dbController = dbcon
}

func forcsv(fname string, ts string) string {
	db := dbController.Db
	logger.Println("for ti>=", ts)
	stmt1, err := db.Query("Select sgu_id,timestamp,Vr,Vy,Vb,Ir,Iy,Ib,Pf,KW,KVA,KWH,KVAH,rKVAH,Run_Hours,freq from parameters where timestamp>='" + ts + "'")
	if err != nil {
		logger.Println(err)
	}
	defer stmt1.Close()

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
	cell.Value = "SGU ID"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Timestamp"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Vr"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Vy"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Vb"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Ir"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Iy"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Ib"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Pf"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "KW"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "KVA"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "KVAH"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "rKVAH"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Run_Hours"
	cell = row.AddCell()
	cell.SetStyle(sty)
	cell.Value = "Freq"
	sty.Font = *(xlsx.NewFont(12, "Arial"))
	for stmt1.Next() {
		row = sheet.AddRow()
		cell = row.AddCell()
		var sgid string
		var ti string
		var Vr, Vy, Vb, Ir, Iy, Ib, Pf, KW, KVA, KWH, KVAH, rKVAH, Run_Hours, freq string
		err = stmt1.Scan(&sgid, &ti, &Vr, &Vy, &Vb, &Ir, &Iy, &Ib, &Pf, &KW, &KVA, &KWH, &KVAH, &rKVAH, &Run_Hours, &freq)
		cell.SetStyle(sty)
		cell.Value = sgid
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = ti
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Vr
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Vy
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Vb
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Ir
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Iy
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Ib
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Pf
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = KW
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = KVA
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = KVAH
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = rKVAH
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = Run_Hours
		cell = row.AddCell()
		cell.SetStyle(sty)
		cell.Value = freq
	}
	//defer stmt1.Close()
	err = file.Save(fname + ".xlsx")
	if err != nil {
		fmt.Printf(err.Error())
	}
	if err != nil {
		logger.Println(err)
	}
	logger.Println("^^&&&^&^")
	if stmt1 != nil {
		logger.Println("XLSX File Created")
	}
	return fname
}

type EmailConfig struct {
	Username string
	Password string
	Host     string
	Port     int
}

func foremail(currentTimeForEmailFile, reportuserid string) {
	logger.Println("inside foremail")
	// authentication configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := 25
	smtpPass := "Havells123"
	smtpUser := "havellstreetcomm@chipmonk.in"

	emailConf := &EmailConfig{smtpUser, smtpPass, smtpHost, smtpPort}
	emailauth := smtp.PlainAuth("", emailConf.Username, emailConf.Password, emailConf.Host)

	sender := "havellstreetcomm@chipmonk.in"

	receivers := []string{
		reportuserid,
	}

	message := "Please see the email attachment for the XLSX File"
	subject := "Attached Reports XLSX File!"

	emailContent := email.NewMessage(subject, message)
	emailContent.From = sender
	emailContent.To = receivers
	logger.Println("inside foremail222")
	/*	 path,_:=os.Getwd()
	var st []string
			st=strings.Split(path,"\\")
			tpath:=""
			for i,_:=range st{
				tpath=st[i]+"/"
				break
			}*/
	filename := currentTimeForEmailFile + ".xlsx"

	logger.Println("inside foremail3333")

	err := emailContent.Attach(filename)
	if err != nil {
		logger.Println(err)
	}

	eorr := email.Send(smtpHost+":"+strconv.Itoa(emailConf.Port), //convert port number from int to string
		emailauth,
		emailContent)

	if eorr != nil {
		logger.Println(eorr)
	}
	logger.Println("inside foremail5555")
}

func reportconfig() {
	var reportfrequency, reportuserid, nxt, typ, id string

	db := dbController.Db

	reportstmt, err := db.Query("SELECT id,reportfrequency,reportdef_userid,next,type FROM reportcofig")

	if err != nil {
		logger.Println("Failed data retrieve")
	}
	defer reportstmt.Close()

	t := time.Now()
	fnames := make(map[string]string)
	hr := make(map[string]string)
	hr["DAILY"] = "24h"
	hr["WEEKLY"] = "168h"
	hr["MONTHLY"] = "720h"
	ti := t.Format("20060102_1504")
	tmp := t
	v, _ := time.ParseDuration("-24h")
	tmp = tmp.Add(v)
	st := tmp.Format("2006-01-02 15:04:05")
	fnames["DAILY"] = forcsv(ti+"D", st)
	tmp = t
	v, _ = time.ParseDuration("-168h")
	tmp = tmp.Add(v)
	st = tmp.Format("2006-01-02 15:04:05")
	fnames["WEEKLY"] = forcsv(ti+"W", st)
	tmp = t
	v, _ = time.ParseDuration("-720h")
	tmp = tmp.Add(v)
	st = tmp.Format("2006-01-02 15:04:05")
	fnames["MONTHLY"] = forcsv(ti+"M", st)
	logger.Println(fnames["DAILY"])
	for reportstmt.Next() {
		err := reportstmt.Scan(&id, &reportfrequency, &reportuserid, &nxt, &typ)
		if err != nil {
			logger.Println(err)
		}
		logger.Println(reportfrequency)
		logger.Println(reportuserid)
		logger.Println(nxt)
		logger.Println(typ)
		if typ == "ENERGY REPORT" {
			ne, _ := time.Parse("2006-01-02 15:04:05", nxt)
			v, _ = time.ParseDuration("-5h30m")
			ne = ne.Add(v)
			t := time.Now().UTC()
			logger.Println("next=", ne)
			logger.Println("CURR=", t)
			if ne.Before(t) {
				logger.Println("Sending email to=", reportuserid)
				foremail(fnames[reportfrequency], reportuserid)
				v, _ = time.ParseDuration(hr[reportfrequency])
				tmp = ne
				tmp = tmp.Add(v)
				st = tmp.Format("2006-01-02 15:04:05")
				tstmt, _ := db.Prepare("update reportcofig set next=? where id=?")
				defer tstmt.Close()
				_, eorr := tstmt.Exec(st, id)
				if eorr != nil {
					logger.Println("Failed data update1")
				}
			}
		}

	}
	err = os.Remove(fnames["DAILY"] + ".xlsx")
	if err != nil {
		logger.Println(err)
	}
	err = os.Remove(fnames["WEEKLY"] + ".xlsx")
	if err != nil {
		logger.Println(err)
	}
	err = os.Remove(fnames["MONTHLY"] + ".xlsx")
	if err != nil {
		logger.Println(err)
	}
}

func ReportGenThread() {

	gocron.Every(1).Day().At("20:00").Do(reportconfig)
	_, time := gocron.NextRun()
	logger.Println("CRON JOB SET AT=", time)
	<-gocron.Start()
}
