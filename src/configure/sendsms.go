/********************************************************************
 * FileName:     sendsms.go
 * Project:      Havells StreetComm
 * Module:       sendsms
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package configure

import (
	"io/ioutil"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	//"fmt"
)

var (
	enableLogs = true
)

func Sendsms(msg string) {
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	db := dbController.Db
	rows, err := db.Query("Select mobile_num from admin")
	defer rows.Close()
	if err != nil {
		logger.Println(err.Error())
	} else {
		for rows.Next() {
			var to string
			err = rows.Scan(&to)
			if err != nil {
				logger.Println(err.Error())
			} else {
				response, err := http.Get("http://login.smsgatewayhub.com/api/mt/SendSMS?APIKey=6c3f0e72-71f8-4ffa-94e2-84a0cc7f50b9&senderid=WEBSMS&channel=2&DCS=0&flashsms=0&number=" + to + "&text=" + msg + "&route=1")
				if err != nil {
					logger.Printf("%s\n", err)

				} else {
					defer response.Body.Close()
					contents, err := ioutil.ReadAll(response.Body)
					if err != nil {
						logger.Printf("%s\n", err)
					} else {
						logger.Printf("%s\n", string(contents))
					}

				}
			}

		}

	}

}

func StartSendSMSThread(SendSMSChan chan string) chan bool {

	if enableLogs {

		logger.Println("Starting SMS thread")
	}

	closeThread := make(chan bool)

	go func() {

		for {

			select {

			case tempString := <-SendSMSChan:
				{

					if enableLogs {

						logger.Println("Received Message to send on SMS")
						logger.Println(tempString)
					}

					Sendsms(tempString)

				}

			case <-closeThread:
				{

					if enableLogs {

						logger.Println("Closing SMS thread")
					}

					return

				}

			}
		}

	}()

	return closeThread
}
