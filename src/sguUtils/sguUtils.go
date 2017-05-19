/********************************************************************
 * FileName:     sguUtils.go
 * Project:      Havells StreetComm
 * Module:       sguUtils
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package sguUtils



import (
	//"fmt"
	"tcpUtils"
	"scuUtils"
	"net"
	"time"
	"dbUtils"
	"strconv"
	"log"
	"sync"
	"net/http"
	"encoding/json"
	"encoding/binary"
)



const (

	sguTickerTimeInMiliSeconds=5000
	HousekeepingTickerTimeInSec=120
	//sguCheckResponseTimeInMiliseconds=1000*5*60
	NumIterationsOfParsePacket = 3
	//sguCheckResponseTimeCount = sguCheckResponseTimeInMiliseconds/sguTickerTimeInMiliSeconds
	MaxNumSGUs=100
	enableLogs=true
	enableSemaphoreLogs=false
	//maxRetry=5
	//retryDelay="59s"
	maxRetryHash=1000
)


//enums for SGU state
const	(

	SGUstateUnassigned=iota
	SGUstateSGUidFound
	SGUstateSGUassigned
	SGUstateSGUReady
	SGUstateSGUnotInList
	SGUstateSGUDisconnected


)

var	(

	NumSGUSconnected			int
	SguBuffer					[MaxNumSGUs]*SguUtilsStruct
	CurrentSGUindex				int
	SGUIDArray					[MaxNumSGUs]uint64
	SCUIDArray					[MaxNumSGUs][]uint64
	NumSGUSinDB					int
	NumOfSCUsInDb				[MaxNumSGUs]int

	DbController   				dbUtils.DbUtilsStruct
	HousekeepingTicker			*time.Ticker
	HousekeepingTickerChan		chan	bool
	SguUtilsSemaphore			sync.Mutex
	SendSMSChan					chan	string
	MasterLampControllerChan 	chan	SguUtilsLampControllerStruct
	maxNumScusPerSgu			int

	sguCheckResponseTimeInMiliseconds  int
	sguCheckResponseTimeCount	int
	maxRetry					int
	retryDelay					string
	per_scu_delay				string
)



type SguUtilsLampControllerStruct struct {
	PacketType					int
	SGUID						uint64
	SCUID						uint64
	ConfigArray					[]byte
	ConfigArrayLength			int
	LampEvent					int
	W							http.ResponseWriter
	ResponseSend				chan		bool

}



type SguUtilsStruct struct {

	SGUID						uint64
	//dbTransactionHandler		*sql.Tx
	sguScuUtilsArray			[]scuUtils.ScuUtilsStruct
	Lattitude					float64
	Longitude					float64
	NumOfSCUs					int
	sguState					int
	SguTcpUtilsStruct			tcpUtils.TcpUtilsStruct
	SguClose					chan	bool
	SguTicker					*time.Ticker
	NumOfSCUsInDb				int
	SguUtilsStructSemaphore		sync.Mutex
	ResponseWriterArray			[]http.ResponseWriter
	ResponseSendChan			[]chan	bool
	sguTickerCount				int


}

var logger *log.Logger
func Init(logg *log.Logger){
	logger=logg

}
func Config_Params(per_scudelay string,scu_polling string,scu_retry_delay string,scu_max_retry string){
	sguCheckResponseTimeInMiliseconds,_=strconv.Atoi(scu_polling)
	sguCheckResponseTimeInMiliseconds*=1000
	sguCheckResponseTimeCount=sguCheckResponseTimeInMiliseconds/sguTickerTimeInMiliSeconds
	maxRetry,_=strconv.Atoi(scu_max_retry)
	retryDelay=scu_retry_delay+"s"
	per_scu_delay=per_scudelay+"s"

}
/*************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)SGUInitMem() {


	SguUtilsStructPtr.sguScuUtilsArray	= make([]scuUtils.ScuUtilsStruct, maxNumScusPerSgu)
	SguUtilsStructPtr.ResponseWriterArray	= make([]http.ResponseWriter, maxNumScusPerSgu)
	SguUtilsStructPtr.ResponseSendChan	= make([]chan bool, maxNumScusPerSgu)
	SguUtilsStructPtr.sguTickerCount = 0;
	SguUtilsStructPtr.NumOfSCUs = 0;

}




/*************************************************/
func IsSGUinDB(sguID uint64) (int,bool) {


	if (enableSemaphoreLogs) {
    	logger.Println("Locking1")
	}

  	SguUtilsSemaphore.Lock()


	for k:=0;k<NumSGUSinDB;k++ {


		if SGUIDArray[k] == sguID {

			SguUtilsSemaphore.Unlock()

			if (enableSemaphoreLogs) {
				logger.Println("Unlocked1")
			}

			return k,true
		}

	}

	SguUtilsSemaphore.Unlock()
	if (enableSemaphoreLogs) {
		logger.Println("Unlocked1")
	}

	return -1,false



}

/*****************************************************************/

func IsSGUinRamList(SguUtilsStructPtr *SguUtilsStruct) (int, bool) {

	if (enableSemaphoreLogs) {
    	logger.Println("Locking2")
	}

	SguUtilsSemaphore.Lock()

	for k:=0; k < CurrentSGUindex; k++   {

		if SguBuffer[k].SGUID == SguUtilsStructPtr.SGUID {


			if (SguBuffer[k].sguState == SGUstateSGUReady) || (SguBuffer[k].sguState == SGUstateSGUDisconnected) {

				if (enableLogs) {
					logger.Printf("RAM LIST %d  %d %d\n",k,CurrentSGUindex, SguBuffer[k].SGUID)
				}

				SguUtilsSemaphore.Unlock()
				if (enableSemaphoreLogs) {
					logger.Println("Unlocked2")
				}
				return k, true
			} else {
				SguUtilsSemaphore.Unlock()
				if (enableSemaphoreLogs) {
				    logger.Println("SGU in RAM list but state is unexpected")
					logger.Println("Unlocked2")
				}
				return -1, false
			}
		}

	}

	SguUtilsSemaphore.Unlock()
	if (enableSemaphoreLogs) {
		logger.Println("Unlocked2")
	}
	return -1, false


}

/*****************************************************************/
func GetSGURamListIndex(SGUID uint64) (int) {

	if (enableSemaphoreLogs) {
    	logger.Println("Locking3")
	}

	SguUtilsSemaphore.Lock()

	for k:=0; k < CurrentSGUindex; k++   {

		if SguBuffer[k].SGUID == SGUID {

			SguUtilsSemaphore.Unlock()
			if (enableSemaphoreLogs) {
				logger.Println("Unlocked3")
			}
			return k
		}

	}

	SguUtilsSemaphore.Unlock()
	if (enableSemaphoreLogs) {
		logger.Println("Unlocked3")
	}
	return -1


}



func (SguUtilsStructPtr *SguUtilsStruct)CloseSGU() {


	for scuIndex:=0;scuIndex<SguUtilsStructPtr.NumOfSCUsInDb;scuIndex++ {

		if (SguUtilsStructPtr.ResponseWriterArray[scuIndex] != nil )  {
			SguUtilsStructPtr.ResponseWriterArray[scuIndex] = nil
			close(SguUtilsStructPtr.ResponseSendChan[scuIndex])
		}

	}


	SguUtilsStructPtr.sguScuUtilsArray	= nil
	SguUtilsStructPtr.ResponseWriterArray	= nil
	SguUtilsStructPtr.ResponseSendChan	= nil



}



func (SguUtilsStructPtr *SguUtilsStruct)SendResponseToUI(scuIndex int,   status int) {



   

	w := SguUtilsStructPtr.ResponseWriterArray[scuIndex]

	if w==nil {

		logger.Println("Sending response with nil writer")
		return
	}



	response := []string{}

	if (status==0) {
		response = append(response,"RED")
	} else if (status==1) {
		response = append(response,"GREEN")
	} else {

		response = append(response,"BLACK")	}


	a, err := json.Marshal(response)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	}  else {
		w.Write(a)

	}

	SguUtilsStructPtr.ResponseWriterArray[scuIndex]  = nil
	close(SguUtilsStructPtr.ResponseSendChan[scuIndex])





}


   
func SendResponseToUIImmediate(w http.ResponseWriter,   status int, ResponseSendChan chan bool) {



   


	if w==nil {

		logger.Println("Sending response with nil writer")
		return
	}



	response := []string{}

	if (status==0) {
		response = append(response,"RED")
	} else if (status==1) {
		response = append(response,"GREEN")
	} else if status==2{
		response = append(response,"BLACK")
	} else {
		response = append(response,"GREY")
	}


	a, err := json.Marshal(response)
	if err != nil {
		logger.Println("Error in json.Marshal")
		logger.Println(err)
	}  else {
		w.Write(a)

	}

	close(ResponseSendChan)



}


/*****************************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)UpdateSGUscuTimestamps() {

	var	LampController			SguUtilsLampControllerStruct

	LampController.SGUID = SguUtilsStructPtr.SGUID
    LampController.PacketType = 0x5000
    LampController.ConfigArray = nil	
    LampController.ConfigArrayLength = 0	
    LampController.LampEvent = 0	
    LampController.W = nil
    LampController.ResponseSend	= nil


	LampController.SCUID  =  SguUtilsStructPtr.SguTcpUtilsStruct.SGUZigbeeID
	MasterLampControllerChan<-LampController

	for k:=0;k<SguUtilsStructPtr.SguTcpUtilsStruct.NumOfSCUs;k++ {

		LampController.SCUID  =  SguUtilsStructPtr.SguTcpUtilsStruct.SCUIDArray[k]
		MasterLampControllerChan<-LampController

	}







}
func (SguUtilsStructPtr *SguUtilsStruct)UpdateSGUFirmwareStatus() {

	var	LampController			SguUtilsLampControllerStruct

	LampController.SGUID = SguUtilsStructPtr.SGUID
	LampController.PacketType = 0x1000
	LampController.ConfigArray = nil
	LampController.ConfigArrayLength = 0
	LampController.LampEvent = 0
	LampController.W = nil
	LampController.ResponseSend	= nil


	LampController.SCUID  =  0
	MasterLampControllerChan<-LampController
}

/*****************************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)UpdateDBWithLampStatus() {
	logger.Println("Inside UPDATE SCU STATUS",SguUtilsStructPtr.SguTcpUtilsStruct.LampStatusCount)
	if (SguUtilsStructPtr.SguTcpUtilsStruct.LampStatusCount == 0) {
		return
	}

	tempcount := SguUtilsStructPtr.SguTcpUtilsStruct.LampStatusCount
	SguUtilsStructPtr.SguTcpUtilsStruct.LampStatusCount = 0

	DbController.DbSemaphore.Lock()
	//create a new transaction
	_Tx, _Err := DbController.Db.Begin()

	if _Err != nil {
		logger.Println("Error opening a new DB transaction for updating lamp status in DB")
		logger.Println(_Err)
	   		DbController.DbSemaphore.Unlock()
		return

	}
	DbController.DbSemaphore.Unlock()



	qStatement := "insert into scu_status (scu_id, status) values(?,?) " +
		 		"on duplicate key update status=?,timestamp=now()"
	
	 


	_Stmt, _Err1 := _Tx.Prepare(qStatement)

    //close statement
	defer  _Stmt.Close()

	if _Err1 != nil {
		logger.Println("Error preparing statement while updating lamp status in DB")
		logger.Println(_Err1)
		return
	}


	for k:=0; k<tempcount;k++ {

		uint64Temp := SguUtilsStructPtr.SguTcpUtilsStruct.LampStatusArray[k]
		scuIndex := (int)((uint64Temp >> 40) & 0x0FF)
		scuid := SguUtilsStructPtr.SguTcpUtilsStruct.SCUIDinDBArray[scuIndex]
		//some problem with mysql driver.   status is defined as int(11) and 
		//we are writing a 5 byte number. However, value written in db is 0x7fffffff
		//Quick fix is only write lower 8 bits. Need to resolve this issue later.
		//uint64Temp &= 0x00FFFFFFFFFF
		uint64Temp &= 0x00FF
		logger.Println("UPDATING STATUS FOR SCUID=",strconv.FormatUint(scuid,10))
		//logger.Println("FOR SCUID==",scuid)
		_,err := _Stmt.Exec(strconv.FormatUint(scuid,10), strconv.FormatUint(uint64Temp, 10), strconv.FormatUint(uint64Temp, 10))


		if err != nil {
			logger.Println("Error  executing prepared statement while updating lamp status in DB")

		}
	
	}


	err := _Tx.Commit()


	if err != nil {
		logger.Println("Error while commiting transaction to DB for updating lamp status in db")
		_Tx.Rollback()

	}
		
	logger.Println("DONE UPDATION!!")

}


/**********************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)SGUGetLampStatus() {

	sguIndex, _ := 	IsSGUinDB(SguUtilsStructPtr.SGUID)

	if(sguIndex==-1) {
		logger.Println("Invalid SGU specified while getting lamp status")
		return
	}

	SguUtilsStructPtr.sguTickerCount++
	if (SguUtilsStructPtr.sguTickerCount < sguCheckResponseTimeCount) {
		return
	}
	SguUtilsStructPtr.sguTickerCount = 0





	var tempLampControl  SguUtilsLampControllerStruct

	tempLampControl.SGUID = SguUtilsStructPtr.SGUID
	tempLampControl.PacketType = 0x3000
	tempLampControl.W = nil
	tempLampControl.ConfigArray = nil
	tempLampControl.ConfigArrayLength = 0

	//get/set field is set to get
	tempLampControl.LampEvent = 0
	logger.Println("FOR LAMP STATUS SGUID====",SguUtilsStructPtr.SGUID," index=",sguIndex)
	//logger.Println("SCIDSGUARRAY=",SCUIDArray)
	for k:=0;k<NumOfSCUsInDb[sguIndex];k++ {

		tempLampControl.SCUID = SCUIDArray[sguIndex][k]
		MasterLampControllerChan<-tempLampControl

	}
}





/*****************************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)SendAlertSMS() {
	logger.Println("old state=",SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld)
	if (SguUtilsStructPtr.SguTcpUtilsStruct.AlertState == SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld) {
		return
	}

	temp 	:= 	SguUtilsStructPtr.SguTcpUtilsStruct.AlertState
	temp1 	:= 	SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld

	SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld = SguUtilsStructPtr.SguTcpUtilsStruct.AlertState
	//format string here



	//get SGU name

	DbController.DbSemaphore.Lock()
	//create a new transaction
	_Tx, _Err := DbController.Db.Begin()

	if _Err != nil {
		logger.Println("Error opening a new DB transaction for SendAlertSMS")
		logger.Println(_Err)
	   		DbController.DbSemaphore.Unlock()
		return

	}
	DbController.DbSemaphore.Unlock()


	

	qStatement := "select location_name from sgu where sgu_id=?"


	_Stmt, _Err1 := _Tx.Prepare(qStatement)

	if _Err1 != nil {
		logger.Println("Error preparing statement while SendAlertSMS")
		logger.Println(_Err1)
		return
	}

    //close statement
	defer  _Stmt.Close()

	var  sgu_name	string
	_Stmt.QueryRow(strconv.FormatUint(SguUtilsStructPtr.SGUID,10)).Scan(&sgu_name)

	SMSstring := "Deployment:%20HAVELLSWB%20"
	if sgu_name==" "||len(sgu_name)==0{
		SMSstring += "SGUID=" + strconv.FormatUint(SguUtilsStructPtr.SGUID,10) + "%20"
	}else{
		SMSstring += "SGU%20" + sgu_name + "%20,ID=" + strconv.FormatUint(SguUtilsStructPtr.SGUID,10) + "%20"
	}



	tArray := make([]byte, 8)

	binary.BigEndian.PutUint64(tArray, SguUtilsStructPtr.SguTcpUtilsStruct.TimeStampHi)

	if (enableLogs) {
		logger.Println(tArray)
	}


	TimeStampString := string(tArray[:4])  + "-" + string(tArray[4:6]) +"-" + string(tArray[6:8]) + "%20"

	binary.BigEndian.PutUint64(tArray, SguUtilsStructPtr.SguTcpUtilsStruct.TimeStampLo)

	TimeStampString += string(tArray[2:4]) + ":" + string(tArray[4:6]) + ":" +	string(tArray[6:8]) + "%20"


	//temp is new status
	//temp1 is old status
	//find changed bits


	temp2  := temp ^ temp1

	if ((temp2 & 0x01) != 0) {
	    if ((temp & 0x01)!=0) {
			SMSstring += "PANEL%20OPEN%20AT:%20" + TimeStampString
		} else {
			SMSstring += "PANEL%20CLOSED%20AT:%20" + TimeStampString
		}
	}
	if ((temp2 & 0x02) != 0) {
	    if ((temp & 0x02)!=0) {
			SMSstring += "MCB%20TRIP%20AT:%20" + TimeStampString
		} else {
			SMSstring += "MCB%20RESTORED%20AT:%20" + TimeStampString
		}

	}
	if ((temp2 & 0x04) != 0) {
	    if ((temp & 0x04)!=0) {
		    SMSstring += "EARTH%20FAULT%20AT:%20" + TimeStampString
		} else {
		    SMSstring += "EARTH%20FAULT%20RESOLVED%20AT:%20" + TimeStampString
		}
	}


	if (enableLogs) {
		logger.Println("sgu_name is " + sgu_name)
		logger.Println(SguUtilsStructPtr.SguTcpUtilsStruct.TimeStampHi)		
	}

	if (enableLogs) {
		logger.Printf("Detected Alert event %d\n",SguUtilsStructPtr.SguTcpUtilsStruct.AlertState)
	}


	SendSMSChan<-SMSstring



	SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld = SguUtilsStructPtr.SguTcpUtilsStruct.AlertState



}


/*****************************************************************/

func (SguUtilsStructPtr *SguUtilsStruct)HandleSguPackets() {


	for {

		select {


		case  <-SguUtilsStructPtr.SguTicker.C:  {

			//sgu status is updated only after parsing input packet.
			//so as long as sgu is connected, we should parse packet

			if (SguUtilsStructPtr.SguTcpUtilsStruct.ConnectedToSGU) {



				if (enableSemaphoreLogs) {
					logger.Println("Locking Semaphore for parsing")
				}

				SguUtilsStructPtr.SguUtilsStructSemaphore.Lock()

				for k:=0; k < NumIterationsOfParsePacket;k++ {
					SguUtilsStructPtr.SguTcpUtilsStruct.ParseInputPacket()
				}
				SguUtilsStructPtr.SguUtilsStructSemaphore.Unlock()
				if (enableSemaphoreLogs) {
					logger.Println("Unlocked Semaphore after parsing")
				}


			}

			//check for current sgu status
			switch SguUtilsStructPtr.sguState  {

			case  SGUstateUnassigned: {
				//we are still waiting for a valid packet from SGU
				if (!SguUtilsStructPtr.SguTcpUtilsStruct.SCUListreceived)	{
					break
				}

				SguUtilsStructPtr.sguState = SGUstateSGUidFound
				SguUtilsStructPtr.SGUID = SguUtilsStructPtr.SguTcpUtilsStruct.SGUID
				if (enableLogs) {
					logger.Println("SGU ID Found")
				}

			}
				fallthrough

			case  SGUstateSGUidFound: {
				//a valid packet is found. //check if ID exists in DB
				_,SGUinDBflag := IsSGUinDB(SguUtilsStructPtr.SGUID)
				if  SGUinDBflag {
					if (enableLogs) {
						logger.Printf("SGU %d is already in the dB list\n",SguUtilsStructPtr.SGUID)
					}

					//if a similar ID exists in RAM, close old SGU instance and
					//add current SGU in same location
					sguIndex, sguInList := 	IsSGUinRamList(SguUtilsStructPtr)
					if (sguInList) {


						if (enableLogs) {
							logger.Println("SGU is already in the RAM list")
						}

						//close down old instance
						if (enableSemaphoreLogs) {
    						logger.Println("Locking4")
						}
						SguUtilsSemaphore.Lock()
						SguBuffer[sguIndex].CloseSGU()
						close(SguBuffer[sguIndex].SguClose)

						//TBD. Make sure that the SGU is closed properly

						SguBuffer[sguIndex] = SguUtilsStructPtr
						SguUtilsStructPtr.sguState = SGUstateSGUassigned
						if (enableLogs) {
							logger.Println("SGU state switched to assigned")
						}

						SguUtilsSemaphore.Unlock()
						if (enableSemaphoreLogs) {
							logger.Println("Unlocked4")
						}

					} else {
						//sanity check. Make sure we are not exceeding allocated space
						if (CurrentSGUindex < MaxNumSGUs) {

							if (enableLogs) {
								logger.Println("adding SGU in the list")
							}


							//SGU not in the list. Add it
							if (enableSemaphoreLogs) {
    							logger.Println("Locking5")
							}
							SguUtilsSemaphore.Lock()
							SguBuffer[CurrentSGUindex] =  SguUtilsStructPtr
							SguUtilsStructPtr.sguState = SGUstateSGUassigned
							CurrentSGUindex++
							SguUtilsSemaphore.Unlock()
							if (enableSemaphoreLogs) {
								logger.Println("Unlocked5")
							}

							if (enableLogs) {
								logger.Println("SGU state switched to assigned")
							}

						} else {
							logger.Printf("Valid SGU detected but Max Limit of %d is reached",MaxNumSGUs)
							break

						}


					}



				} else {

					AddSGUToDB(SguUtilsStructPtr.SGUID)
					logger.Printf("NEW SGU id %d detected but was not in the authorized list",SguUtilsStructPtr.SGUID)
					SguUtilsStructPtr.sguState = SGUstateSGUassigned

					if (enableLogs) {
						logger.Println("SGU state switched to assigned")
					}

				}

			}
				fallthrough
			case  SGUstateSGUassigned: {
				//just make sure that SGU is not disconnected
				//also, for new list, update db if needed.

				//populate SCU list from DB to SGU TCP structure
				sguIndex, sguInList := 	IsSGUinDB(SguUtilsStructPtr.SGUID)

				if (sguInList) {

					SguUtilsStructPtr.SguTcpUtilsStruct.NumOfSCUsInDB = NumOfSCUsInDb[sguIndex]
					logger.Println("Number of SCU in DB=",NumOfSCUsInDb[sguIndex]," for SGUID=",SguUtilsStructPtr.SGUID," index=",sguIndex)
					for k:=0;k<NumOfSCUsInDb[sguIndex];k++ {
					   SguUtilsStructPtr.SguTcpUtilsStruct.SCUIDinDBArray[k]	= SCUIDArray[sguIndex][k]
						logger.Println("SCUID=",SCUIDArray[sguIndex][k]," for SGUID=",SguUtilsStructPtr.SGUID)
					}
					//logger.Println("SCIDSGUARRAY=",SCUIDArray)
					SguUtilsStructPtr.UpdateSGUscuTimestamps()
					SguUtilsStructPtr.UpdateSGUFirmwareStatus()

				}
				SguUtilsStructPtr.sguState = SGUstateSGUReady
			}

				fallthrough

			case SGUstateSGUReady: {



				if (!SguUtilsStructPtr.SguTcpUtilsStruct.ConnectedToSGU) {

					//SGU is now disconnected. Change state
					SguUtilsStructPtr.sguState = SGUstateSGUDisconnected
					




				} else  {

					if (SguUtilsStructPtr.SguTcpUtilsStruct.SCUListreceived) {

						SguUtilsStructPtr.UpdateSCUListInDB()
						SguUtilsStructPtr.SguTcpUtilsStruct.SCUListreceived = false
					}

					if (enableSemaphoreLogs) {
    					logger.Println("Locking7")
					}

  					SguUtilsSemaphore.Lock()

					for k:=0;k< SguUtilsStructPtr.SguTcpUtilsStruct.ResponseReceivedCount;k++ {
							temp :=  SguUtilsStructPtr.SguTcpUtilsStruct.ResponseReceivedArray[k]

							//separate index
							SguUtilsStructPtr.SendResponseToUI(((temp >> 16) & 0x00FF), (temp & 0x00FF))


					}
					SguUtilsStructPtr.SguTcpUtilsStruct.ResponseReceivedCount = 0

					if (enableSemaphoreLogs) {
    					logger.Println("Unlocked7")
					}

  					SguUtilsSemaphore.Unlock()
					SguUtilsStructPtr.SendAlertSMS()
					SguUtilsStructPtr.SGUGetLampStatus()
					SguUtilsStructPtr.UpdateDBWithLampStatus()
				}


			}
			case  SGUstateSGUnotInList: {

			}
			case  SGUstateSGUDisconnected: {

			}


			}

		}

		case <-SguUtilsStructPtr.SguClose: 	{

			//time to close timer and SGU
			SguUtilsStructPtr.SguTicker.Stop()
			//SguUtilsStructPtr.SguTicker.Close()
			SguUtilsStructPtr = nil
			return


			}

		}


	}


}


/**************************************************/
func	InitSGUIDsfromDB()	{

	//This table is update in runtime. so need synchronization


	var sguID uint64


	if DbController.DbConnected {
		DbController.DbSemaphore.Lock()

		SGUsFromDB := 0

		rows, err := DbController.Db.Query("select sgu_id from sgu ")

		if err != nil {
			logger.Println("Error tryting to read SGU list from database")
			logger.Println(err)
		}


		for rows.Next() {

			err := rows.Scan(&sguID)

			if err != nil {

				logger.Println("Error tryting to read SGU id from database")
				logger.Println(err)
			} else {
				SGUIDArray[SGUsFromDB] = sguID
				SGUsFromDB++

			}
		}

		rows.Close()
		DbController.DbSemaphore.Unlock()
		if (NumSGUSinDB !=  SGUsFromDB) {
			logger.Printf("Found %d SGUs in database\n",SGUsFromDB)
			NumSGUSinDB = SGUsFromDB
		}
	}





}

/**************************************************/
func	AddSGUToDB(sguID uint64)	{

	//This table is update in runtime. so need synchronization




	if DbController.DbConnected {

		DbController.DbSemaphore.Lock()

		//create scu table if not already created
		qStatement := "insert into sgu (sgu_id) values (" +
		strconv.FormatUint(sguID, 10) + " )"

		//create non existane tables
		if (enableLogs) {
			logger.Println("Creating transaction for inserting sgu to db")
			logger.Println("qStatement")

		}

		DbController.Tx, DbController.Err = DbController.Db.Begin()

		if DbController.Err != nil {
			log.Println("Error Executing " + qStatement)
			log.Println("Error creating transaction")
			log.Println(DbController.Err)
			return

		}


		DbController.Stmt, DbController.Err = DbController.Tx.Prepare(qStatement)


		if (DbController.Err != nil ) {
			log.Println("Error Executing " + qStatement)
			log.Println("Error creating statement")
			log.Println(DbController.Err)
			return

		}

		if (enableLogs) {
			logger.Println("Executing statement for inserting SGU in DB")
		}

		_,DbController.Err = DbController.Stmt.Exec()

		if DbController.Err != nil {
			log.Println("Error Executing " + qStatement)
			log.Println("Error Executing statement")
			log.Println(DbController.Err)
			DbController.Tx.Rollback()

		} else {

			DbController.Tx.Commit()

		}


		DbController.Stmt.Close()
		DbController.DbSemaphore.Unlock()


	}


}

/**************************************************/
func	AddSCUToDB(sguID uint64, scuID uint64)	{

	//This table is update in runtime. so need synchronization




	if DbController.DbConnected {

		DbController.DbSemaphore.Lock()

		//create scu table if not already created
		qStatement := "insert into scu (sgu_id, scu_id) values (" +
		strconv.FormatUint(sguID, 10) + ","   +
		strconv.FormatUint(scuID, 10) + ")"



		//create non existane tables
		if (enableLogs) {
			logger.Println("Creating transaction for inserting scu to db")
			logger.Println(qStatement)

		}

		DbController.Tx, DbController.Err = DbController.Db.Begin()

		if DbController.Err != nil {
			log.Println("Error Executing " + qStatement)
			log.Println("Error creating transaction")
			log.Println(DbController.Err)
			return

		}


		DbController.Stmt, DbController.Err = DbController.Tx.Prepare(qStatement)


		if (DbController.Err != nil ) {
			log.Println("Error Executing " + qStatement)
			log.Println("Error creating statement")
			log.Println(DbController.Err)
			return

		}

		if (enableLogs) {
			logger.Println("Executing statement for inserting SCU in DB")
		}

		_,DbController.Err = DbController.Stmt.Exec()

		if DbController.Err != nil {
			log.Println("Error Executing " + qStatement)
			log.Println("Error Executing statement")
			log.Println(DbController.Err)
			DbController.Tx.Rollback()

		} else {

			DbController.Tx.Commit()

		}


		DbController.Stmt.Close()
		DbController.DbSemaphore.Unlock()


	}


}


/*********************************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)IsSCUinDB(scuID uint64) (int , bool) {


	index, sguIndBflag := IsSGUinDB(SguUtilsStructPtr.SGUID)

	if sguIndBflag {

		for k:=0;k<NumOfSCUsInDb[index];k++ {


			if SCUIDArray[index][k] == scuID {
				return k, true
			}
		}

		//sgu found but no scu found
		return -1,	false


	} else {
		return -2,	false

	}

}


/*****************************************************************/
func (SguUtilsStructPtr *SguUtilsStruct)UpdateSCUListInDB() {

	//

	sguIndex, sguFlag := IsSGUinDB(SguUtilsStructPtr.SGUID)


	if (sguFlag) {

		for scuIndex := 0; scuIndex < SguUtilsStructPtr.SguTcpUtilsStruct.NumOfSCUs;scuIndex++ {
			scuID := SguUtilsStructPtr.SguTcpUtilsStruct.SCUIDArray[scuIndex]

			_,scuInDBflag := SguUtilsStructPtr.IsSCUinDB(scuID)

			if !scuInDBflag {
				AddSCUToDB(SguUtilsStructPtr.SGUID, scuID)
				logger.Printf("Found and added a new scu %d for sgu %d\n",scuID,SguUtilsStructPtr.SGUID)
				//also update RAM list
				SCUIDArray[sguIndex][NumOfSCUsInDb[sguIndex]] = scuID
				NumOfSCUsInDb[sguIndex]++
			}

		}
	}

}



/*****************************************************************/
func GetSCUsForAllSGUs() {


	for k:=0;k<NumSGUSinDB;k++ {
		InitSCUIDsFromDB(k)



	}



}



/*****************************************************************/
func	InitSCUIDsFromDB(sguIndex int)	{

	//This table is update in runtime. so need synchronization
	//TBD Add semaphore


	var scuID uint64


	if DbController.DbConnected {

		DbController.DbSemaphore.Lock()

		SCUsFromDB := 0

		rows, err := DbController.Db.Query("select scu_id from scu where sgu_id=? ",SGUIDArray[sguIndex])

		if err != nil {
			logger.Printf("Error tryting to read SCU list from database for sgu id = %8x\n",SGUIDArray[sguIndex])
			logger.Println(err)
		}


		for rows.Next() {

			err := rows.Scan(&scuID)

			if err != nil {

				logger.Printf("Error tryting to read SCU id from database for sgu id = %8x\n",SGUIDArray[sguIndex])
				logger.Println(err)
			} else {
				SCUIDArray[sguIndex][SCUsFromDB] = scuID
				SCUsFromDB++

			}
		}

		rows.Close()
		DbController.DbSemaphore.Unlock()
		if (NumOfSCUsInDb[sguIndex] !=  SCUsFromDB) {
			logger.Printf("Found %d SCUs in database for sgu id = %8x\n",SCUsFromDB, SGUIDArray[sguIndex])
			NumOfSCUsInDb[sguIndex] = SCUsFromDB
		}
	}





}


func SGUHouseKeeping() {


	for {

		select {

		case  	<-HousekeepingTicker.C:  	{

			//for now, only update SGU list from DB
			InitSGUIDsfromDB()

			GetSCUsForAllSGUs()


		}

		case 	<-HousekeepingTickerChan:	{

			HousekeepingTicker.Stop()
			return



		}
		}
	}


}



/************************************************************************************************************/
func	HandleSguConnections(sguChan chan net.Conn,  MasterdbController dbUtils.DbUtilsStruct, smsChan chan string, NumScusPerSgu int)  (chan bool) {


	maxNumScusPerSgu = NumScusPerSgu
	SendSMSChan = smsChan
	//create channel
	done := make(chan bool)

	CurrentSGUindex = 0

	DbController = MasterdbController


	for k:=0;k<MaxNumSGUs;k++ {
		SCUIDArray[k] = make([]uint64,NumScusPerSgu)
	}
		 



	//get list of all SGUs from DB
	InitSGUIDsfromDB()

	GetSCUsForAllSGUs()


	//debug hack.
	//t := new(SguUtilsStruct)
	//t.SGUID=57381672663091
	//t.SguTcpUtilsStruct.AlertState = 7
	//t.SguTcpUtilsStruct.TimeStampHi=0x3230313531323138
	//t.SguTcpUtilsStruct.TimeStampLo=0x313131333030
	//t.SendAlertSMS()



	//create housekeeping ticker

	HousekeepingTicker = time.NewTicker(time.Millisecond * 1000*HousekeepingTickerTimeInSec)

	HousekeepingTickerChan = make(chan bool)

	go SGUHouseKeeping()







	//loop and wait for new connection

	go func()  {
		for {

			select {

			case temp := <-sguChan:{

				//a new SGU connected.
				//for now create a new instance and assign it to buffer
				if (enableSemaphoreLogs) {
    				logger.Println("Locking6")
				}
				SguUtilsSemaphore.Lock()
				
				tempSGU := new(SguUtilsStruct)
				tempSGU.SguClose = make(chan bool)
				tempSGU.SGUInitMem()

				tempSGU.SguTcpUtilsStruct.AddTcpClientToSGU(temp,maxNumScusPerSgu)

				ticker := time.NewTicker(time.Millisecond * sguTickerTimeInMiliSeconds)
				tempSGU.SguTicker = ticker
				go tempSGU.HandleSguPackets()

				SguUtilsSemaphore.Unlock()
				if (enableSemaphoreLogs) {
					logger.Println("Unlocked6")
				}



			}
			case <-done:	{


				//close housekeeping function thread
				close(HousekeepingTickerChan)

				//free memory
				for k:=0;k<MaxNumSGUs;k++ {
					SCUIDArray[k] =  nil
				}


				logger.Println("closing  sgu handler  loop")
				//SguBuffer = nil

				return


			}
			}

		}

	}()


	return done


}

//Send Packet With RETRY
func (SguUtilsStructPtr *SguUtilsStruct)SendWithRetry(OutputPacketType int, SCUID uint64,  StatusByte int, expression   []byte, expressionLength int,gs int){
	attempt:=0
	strSCU:=strconv.FormatUint(SCUID,10)
	init:=SguUtilsStructPtr.SguTcpUtilsStruct.RetryHashSCU[strSCU]
	//strHash:=strconv.Itoa(temp.PacketType)+"#"+strconv.FormatUint(temp.SCUID,10)+"#"+strconv.Itoa(getSet)+"#"+strconv.FormatUint(((temp.LampEvent) & 0x0FF),10)
	strHash:=strconv.Itoa(OutputPacketType)+"#"+strconv.FormatUint(SCUID,10)+"#"+strconv.Itoa(gs)+"#"+strconv.Itoa(((StatusByte) & 0x0FF))
	/*if ((StatusByte) & 0x0FF)==1{
		strHash=strconv.Itoa(OutputPacketType)+"#"+strconv.FormatUint(SCUID,10)+"#"+strconv.Itoa(gs)+"#"+strconv.Itoa(9)
	}*/
	logger.Println("Rec. for sending 3000 Packet ")
	for SguUtilsStructPtr.SguTcpUtilsStruct.RetryHash[strHash]==1&&attempt<maxRetry{
		logger.Println("Try: ",attempt+1,", for SGUID=",SguUtilsStructPtr.SGUID,", Packet Type=",OutputPacketType,", SCUID=",SCUID,", Status=",StatusByte,", Hash=",strHash)
		SguIndex := GetSGURamListIndex(SguUtilsStructPtr.SGUID)
		if SguIndex == -1 {
			logger.Printf("Event specified for non existent SGU  %d\n",SguUtilsStructPtr.SGUID)
		}else{
			if init!=SguUtilsStructPtr.SguTcpUtilsStruct.RetryHashSCU[strSCU]{
				logger.Println("New lamp event specified for scu flushing retrys.")
				break
			}
			SguUtilsStructPtr.SguTcpUtilsStruct.SendResponsePacket(OutputPacketType, SCUID,StatusByte ,expression,expressionLength)
		}
		du, _ := time.ParseDuration(retryDelay)
		time.Sleep(du)
		attempt++;
	}
	logger.Println("Exiting with Hash =",strHash," Value =",SguUtilsStructPtr.SguTcpUtilsStruct.RetryHash[strHash])
}
/************************************************************************************************************/
func	HandleLampEvents(lampControllerChan chan	SguUtilsLampControllerStruct)  (chan bool) {



	MasterLampControllerChan = lampControllerChan
	//create channel
	done := make(chan bool)

	logger.Println("Called HandleLampEvents")

	//loop and wait for new connection

	go func()  {
		for {

			select {

			case temp := <-lampControllerChan:{

				//a new SGU lamp event.
				logger.Printf("New Lamp Event, packetType = %4.4x, SGUID=%d, SCUID=%d %d\n", temp.PacketType,temp.SGUID, temp.SCUID,temp.LampEvent)

				SguIndex := GetSGURamListIndex(temp.SGUID)
				if SguIndex == -1 {
					logger.Printf("Event specified for non existent SGU  %d\n",temp.SGUID)
				} else {


					SguUtilsStructPtr := SguBuffer[SguIndex]


					SguUtilsStructPtr.SguUtilsStructSemaphore.Lock()
					//defer SguUtilsStructPtr.SguUtilsStructSemaphore.Unlock()
					if (enableSemaphoreLogs) {
						logger.Println("Locking Semaphore for sending")
					}



					if (SguUtilsStructPtr.sguState == SGUstateSGUReady)	 {

						if (!SguUtilsStructPtr.SguTcpUtilsStruct.ConnectedToSGU) {

							//SGU is now disconnected. Change state
							SguUtilsStructPtr.sguState = SGUstateSGUDisconnected
						}

					}



					if SguUtilsStructPtr.sguState == SGUstateSGUReady {

						if SguUtilsStructPtr.SguTcpUtilsStruct.Is_updating==true{
							logger.Println("Sorry, SGU Firmware is currently updating!!!!")

						}else if (temp.PacketType==0x3000) {
							logger.Println("CHECKING")
							scuIndex := SguUtilsStructPtr.SguTcpUtilsStruct.GetSCUIndexFromSCUID(temp.SCUID)
							if scuIndex>=0{
								getSet := ((temp.LampEvent >> 8) & 0x0FF)

                                //if current command is set, need to check if old response is still pending.
								//if current command is get, no need for check

								if (SguUtilsStructPtr.ResponseWriterArray[scuIndex] != nil) && (getSet != 0) {
									logger.Println("New Event specified when still waiting for response from old event")
									logger.Println("New event will be lost")
									//close down old event
									SendResponseToUIImmediate(SguUtilsStructPtr.ResponseWriterArray[scuIndex],	3,	SguUtilsStructPtr.ResponseSendChan[scuIndex])
								}

								if (getSet != 0) {
									SguUtilsStructPtr.ResponseWriterArray[scuIndex]= temp.W
									SguUtilsStructPtr.ResponseSendChan[scuIndex]= temp.ResponseSend
								}
								if (getSet!=0){
									strHash:=strconv.Itoa(temp.PacketType)+"#"+strconv.FormatUint(temp.SCUID,10)+"#"+strconv.Itoa(getSet)+"#"+strconv.Itoa(((temp.LampEvent) & 0x0FF))
									/*if ((temp.LampEvent) & 0x0FF)==1{
										strHash=strconv.Itoa(temp.PacketType)+"#"+strconv.FormatUint(temp.SCUID,10)+"#"+strconv.Itoa(getSet)+"#"+strconv.Itoa(9)
									}*/
									if len(SguUtilsStructPtr.SguTcpUtilsStruct.RetryHash)>maxRetryHash{
										logger.Println("Too many Packets, flushing retry Hash for SGUID=",temp.SGUID)
										SguUtilsStructPtr.SguTcpUtilsStruct.RetryHash=make(map[string]int)
									}
									logger.Println("Hash=",strHash)
									SguUtilsStructPtr.SguTcpUtilsStruct.RetryHash[strHash]=1
									SguUtilsStructPtr.SguTcpUtilsStruct.RetryHashSCU[strconv.FormatUint(temp.SCUID,10)]+=1
									logger.Println("Sending 3000 Packet for",temp.SCUID,"==",SguUtilsStructPtr.SguTcpUtilsStruct.RetryHashSCU[strconv.FormatUint(temp.SCUID,10)])
									go SguUtilsStructPtr.SendWithRetry(temp.PacketType, temp.SCUID, temp.LampEvent,temp.ConfigArray,temp.ConfigArrayLength,getSet)
								}else{
									logger.Println("Sending 3000 Packet for",temp.SCUID," POLLING")
									SguUtilsStructPtr.SguTcpUtilsStruct.SendResponsePacket(temp.PacketType, temp.SCUID, temp.LampEvent,temp.ConfigArray,temp.ConfigArrayLength)
								}
								du, _ := time.ParseDuration(per_scu_delay)
								time.Sleep(du)
								//SguUtilsStructPtr.SguTcpUtilsStruct.SendResponsePacket(temp.PacketType, temp.SCUID, temp.LampEvent,temp.ConfigArray,temp.ConfigArrayLength)

							}



						} else {

							SguUtilsStructPtr.SguTcpUtilsStruct.SendResponsePacket(temp.PacketType, temp.SCUID, temp.LampEvent,temp.ConfigArray,temp.ConfigArrayLength)
						}
						//SguUtilsStructPtr.SendResponseToUI(scuIndex, temp.LampEvent ^ 1)




					} else {
						logger.Printf("Event specified when SGU not ready %d  %d \n", temp.SGUID, SguBuffer[SguIndex].sguState)

					}

					SguUtilsStructPtr.SguUtilsStructSemaphore.Unlock()
					if (enableSemaphoreLogs) {
						logger.Println("Unlocked Semaphore after sending")
					}

				}




			}
			case <-done:	{



				logger.Println("closing  lamp ebvent handler  loop")

				return


			}
			}

		}

	}()


	return done


}


func Sgu_firmware_update(sguid uint64) {
	var	LampController	SguUtilsLampControllerStruct

	LampController.SGUID = sguid
	LampController.SCUID = 0
	LampController.PacketType = 0x1024
	LampController.ConfigArray = nil
	LampController.ConfigArrayLength = 0
	LampController.LampEvent = 0x01
	LampController.W = nil
	LampController.ResponseSend	= nil
	MasterLampControllerChan<-LampController
}

func Scu_firmware_update(scuid uint64,sguid uint64) {
	var	LampController	SguUtilsLampControllerStruct

	LampController.SGUID = sguid
	LampController.SCUID = scuid
	LampController.PacketType = 0xC000
	LampController.ConfigArray = nil
	LampController.ConfigArrayLength = 0
	LampController.LampEvent = 0x01
	LampController.W = nil
	LampController.ResponseSend	= nil
	MasterLampControllerChan<-LampController
}

/**************************************************/
/****************** EOF ***************************/
/**************************************************/



