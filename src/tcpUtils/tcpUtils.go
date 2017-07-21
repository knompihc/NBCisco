/********************************************************************
 * FileName:     tcpUtils.go
 * Project:      Havells StreetComm
 * Module:       tcpUtils
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package tcpUtils



import (
    //"fmt"
	"log"
	"net"
	"bufio"
	"time"
	"strconv"
	"encoding/binary"
    "dbUtils"
    "net/http"
    "io/ioutil"
	"math"
	"strings"
	"bytes"
	"sync"
	"encoding/hex"
)
var status = struct {
        sync.RWMutex
        lbuffer map[string]string
}{lbuffer: make(map[string]string)}
	   

const	(

	StartingDelimeter 		= 0x7E
	MaxInOutBufferLength 	= 1024*8
	//MaxInOutBufferLength 	= 65536
	//MaxInOutBufferLength 	= 1000000
	FixedPacketLength 		= 29

)

type TcpUtilsStruct struct {

	err							error
	tcpClient					net.Conn 
	reader						*bufio.Reader
	writer						*bufio.Writer

	responseLineBuff[]			byte
	commandLineBuff[]   		byte
	NumOfSCUs					int
	NumOfSCUsInDB				int

	InputPacketLength			int
	OutputPacketLength			int
	SGUID						uint64
	ControlSGUID				uint64
	TimeStampHi					uint64
	TimeStampLo					uint64
	InputSeqNumber				int
	OutputSeqNumber				int
	InputPacketType				int
	OutputPacketType			int
	SGULatitude					int
	SGULongitude				int
	SCUIDArray[]				uint64
	SCUIDinDBArray[]			uint64
	LampStatusArray[]			uint64
	SGUZigbeeID					uint64
	SCUAnalogP1StateArray[]		int
	ResponseReceivedArray[]		int

	inputBufferDipstick			int 
	inputBufferReadPtr			int 
	inputBufferWritePtr			int 

	outputBufferDipstick		int 
	outputBufferReadPtr			int 
	outputBufferWritePtr		int 

	InputSyncSearchStatus		int

	ConnectedToSGU 				bool
	SCUListreceived				bool
	InputPacketcounter			int

	MAXNumOFSCUS				int
	
	deviceId                    int64
    Length                      int64
    Query                       string
    set                         int
	ResponseReceivedCount		int
	AlertState					int
    AlertStateOld               int
	LampStatusCount				int

	Enable                      int
	PollingRate                 int
	ResponseRate				int
	DeviceTimeout               int
	SlaveId                     int
	SlaveIds                    int

	//for retry
	RetryHash					map[string]int
	RetryHashSCU				map[string]int64

	//firmware
	Current_pos				int64
	Firmware_seq				int
	Prev_temp_arr				[]byte
	Prev_status				int
	Is_updating				bool
	Curr_major				byte
	Curr_minor				byte
	RetryFirmware				map[int64]int64

	Is_TCP_Connected			bool
}

var(
	//firmware
	Sgu_firmware			map[int64][]byte
	Sgu_firmware_size		string
	Sgu_firmware_major		byte
	Sgu_firmware_minor		byte
	Sgu_firmware_name		string
	Sgu_firmware_bucket		int64

	Scu_firmware			map[int64][]byte
	Scu_firmware_size		string
	Scu_firmware_major		byte
	Scu_firmware_minor		byte
	Scu_firmware_name		string
	Scu_firmware_bucket		int64

	Scu_Current_pos				map[uint64]int64
	Scu_Firmware_seq			map[uint64]int
	Scu_Prev_temp_arr			map[uint64][]byte
	Scu_Prev_status				map[uint64]int
	Scu_Is_updating				map[uint64]bool
	Scu_Curr_major				map[uint64]byte
	Scu_Curr_minor				map[uint64]byte
	Scu_RetryFirmware			map[uint64]map[int64]int64
)
var dbController dbUtils.DbUtilsStruct
var logger *log.Logger
func Init(dbcon dbUtils.DbUtilsStruct,logg *log.Logger){
    dbController =dbcon
	logger=logg
	Scu_Current_pos=make(map[uint64]int64)
	Scu_Firmware_seq=make(map[uint64]int)
	Scu_Prev_temp_arr=make(map[uint64][]byte)
	Scu_Prev_status=make(map[uint64]int)
	Scu_Is_updating=make(map[uint64]bool)
	Scu_Curr_major=make(map[uint64]byte)
	Scu_Curr_minor=make(map[uint64]byte)
	Scu_RetryFirmware=make(map[uint64]map[int64]int64)
}
//For LOCAL TESTING
func (TcpUtilsStructPtr	*TcpUtilsStruct) ConnectToSGU() bool {
	//open connection
	//TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp","192.168.1.1:62000")
	TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp","54.185.172.55:62002")

	if (TcpUtilsStructPtr.err != nil) {
		logger.Println("Error opening TCP connection")
		logger.Println(TcpUtilsStructPtr.err)
		TcpUtilsStructPtr.ConnectedToSGU = false
		return false
	} else {

		logger.Println("connected")
		TcpUtilsStructPtr.reader = bufio.NewReader(TcpUtilsStructPtr.tcpClient)
		TcpUtilsStructPtr.writer = bufio.NewWriter(TcpUtilsStructPtr.tcpClient)
		TcpUtilsStructPtr.ConnectedToSGU = true
		return true

	}

}

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    CloseTcpClient()  {
	err:= TcpUtilsStructPtr.tcpClient.Close()
	if err != nil {
		logger.Println("Error closing TCP client")
		logger.Println(err)
	}

	TcpUtilsStructPtr.reader = nil
	TcpUtilsStructPtr.writer = nil
	/*TcpUtilsStructPtr.responseLineBuff = nil
	TcpUtilsStructPtr.commandLineBuff = nil
	TcpUtilsStructPtr.SCUIDArray = nil
	TcpUtilsStructPtr.SCUIDinDBArray = nil
	TcpUtilsStructPtr.ResponseReceivedArray = nil
	TcpUtilsStructPtr.LampStatusArray = nil
	TcpUtilsStructPtr.ConnectedToSGU = false
*/

}



func (TcpUtilsStructPtr	*TcpUtilsStruct) AddTcpClientToSGU(newTcpClient net.Conn) {



	//TcpUtilsStructPtr.MAXNumOFSCUS = MAXNumOFSCUS
	TcpUtilsStructPtr.tcpClient = newTcpClient
	TcpUtilsStructPtr.reader = bufio.NewReader(TcpUtilsStructPtr.tcpClient)
	TcpUtilsStructPtr.writer = bufio.NewWriter(TcpUtilsStructPtr.tcpClient)
	/*
	TcpUtilsStructPtr.responseLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.commandLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.SCUIDArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.SCUIDinDBArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.ResponseReceivedArray = make([]int, MAXNumOFSCUS)
	TcpUtilsStructPtr.LampStatusArray	= make([]uint64,MAXNumOFSCUS)


	TcpUtilsStructPtr.SCUAnalogP1StateArray = make([]int, MAXNumOFSCUS)

	TcpUtilsStructPtr.RetryHash=make(map[string]int)
	TcpUtilsStructPtr.RetryHashSCU=make(map[string]int64)
	TcpUtilsStructPtr.ConnectedToSGU = true
	TcpUtilsStructPtr.InputPacketcounter = 0
	TcpUtilsStructPtr.LampStatusCount = 0
	TcpUtilsStructPtr.SGUID = 0
	TcpUtilsStructPtr.SCUListreceived=false;

	//firmware
	TcpUtilsStructPtr.Is_updating=false;
	TcpUtilsStructPtr.Prev_temp_arr=make([]byte,1028)
*/
	TcpUtilsStructPtr.ConnectedToSGU = true
	TcpUtilsStructPtr.Is_TCP_Connected=true

}

func (TcpUtilsStructPtr *TcpUtilsStruct) InitializeBufferParams(MAXNumOFSCUS int) {
	TcpUtilsStructPtr.MAXNumOFSCUS = MAXNumOFSCUS
	//TcpUtilsStructPtr.tcpClient = newTcpClient
	//	TcpUtilsStructPtr.reader = bufio.NewReader(TcpUtilsStructPtr.tcpClient)
	//	TcpUtilsStructPtr.writer = bufio.NewWriter(TcpUtilsStructPtr.tcpClient)

	TcpUtilsStructPtr.responseLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.commandLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.SCUIDArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.SCUIDinDBArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.ResponseReceivedArray = make([]int, MAXNumOFSCUS)
	TcpUtilsStructPtr.LampStatusArray = make([]uint64, MAXNumOFSCUS)

	TcpUtilsStructPtr.SCUAnalogP1StateArray = make([]int, MAXNumOFSCUS)

	TcpUtilsStructPtr.RetryHash = make(map[string]int)
	TcpUtilsStructPtr.RetryHashSCU = make(map[string]int64)
	//TcpUtilsStructPtr.ConnectedToSGU = true
	TcpUtilsStructPtr.InputPacketcounter = 0
	TcpUtilsStructPtr.LampStatusCount = 0
	TcpUtilsStructPtr.SGUID = 0
	TcpUtilsStructPtr.SCUListreceived = false

	//firmware
	TcpUtilsStructPtr.Is_updating = false
	TcpUtilsStructPtr.Prev_temp_arr = make([]byte, 1028)

	//TcpUtilsStructPtr.Is_TCP_Connected = true

}

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  RewindInputBuffer(nBytes int) {

	//if (TcpUtilsStructPtr.inputBufferDipstick < nBytes){
    //	logger.Printf("Rewinding %d bytes when only %d bytes in FIFO\n",nBytes,TcpUtilsStructPtr.inputBufferDipstick);
    //    return;
    //}
        
        
    TcpUtilsStructPtr.inputBufferDipstick += nBytes;
    TcpUtilsStructPtr.inputBufferReadPtr -= nBytes;
    TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength-1);    
	


}	 

func (TcpUtilsStructPtr *TcpUtilsStruct) AddByteToInputBuff(newByte byte) {
	if TcpUtilsStructPtr.inputBufferDipstick < MaxInOutBufferLength {
		if TcpUtilsStructPtr.inputBufferWritePtr < MaxInOutBufferLength && TcpUtilsStructPtr.inputBufferWritePtr >= 0 {
			TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferWritePtr] = newByte
			TcpUtilsStructPtr.inputBufferWritePtr++
			TcpUtilsStructPtr.inputBufferWritePtr &= (MaxInOutBufferLength - 1)
			TcpUtilsStructPtr.inputBufferDipstick++
		} else {
			logger.Println("Error Writing to input buffer: Index out of range")
		}

	} else {
		//should be spinning here till thread empties buffer.
		//TBD
		TcpUtilsStructPtr.ReadNBytesFromInput(len(TcpUtilsStructPtr.commandLineBuff))
		logger.Println("Warning! Input Buff is full")
	}
}

/*func  (TcpUtilsStructPtr	*TcpUtilsStruct)  AddByteToInputBuff(newByte byte ){
        if (TcpUtilsStructPtr.inputBufferDipstick < MaxInOutBufferLength){
            TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferWritePtr] = newByte
			TcpUtilsStructPtr.inputBufferWritePtr++
            TcpUtilsStructPtr.inputBufferWritePtr &= (MaxInOutBufferLength-1);
            TcpUtilsStructPtr.inputBufferDipstick++;
            
        } else {
            //should be spinning here till thread empties buffer.
            //TBD
            TcpUtilsStructPtr.ReadNBytesFromInput(len(TcpUtilsStructPtr.commandLineBuff))
            logger.Println("Warning! Input Buff is full");
        }
    }
*/

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  ReadOneByteFromInput() byte {

	if TcpUtilsStructPtr.inputBufferDipstick > 0 {
		if TcpUtilsStructPtr.inputBufferReadPtr < len(TcpUtilsStructPtr.commandLineBuff) && TcpUtilsStructPtr.inputBufferReadPtr >= 0 {
			var newByte = TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferReadPtr]
			//logger.Printf("%x\n",newByte)
			TcpUtilsStructPtr.inputBufferReadPtr++
			TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength - 1)
			TcpUtilsStructPtr.inputBufferDipstick--
			return newByte
		} else {
			logger.Println("Error reading from input buffer: Incomplete packet")
			return 0
		}

	} else {
		//should be spinning here till thread fills  buffer.
		//TBD

		logger.Println("Warning! Input Buff is empty")
		return 0
	} 

/*       if (TcpUtilsStructPtr.inputBufferDipstick >0){
            var newByte = TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferReadPtr]
			//logger.Printf("%x\n",newByte)
			TcpUtilsStructPtr.inputBufferReadPtr++
            TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.inputBufferDipstick--
            return newByte
            
        } else {
            //should be spinning here till thread fills  buffer.
            //TBD
            
            logger.Println("Warning! Input Buff is empty");
            return 0;
        }*/
        
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  GetByteFromOutputBuff() byte{
     
	if TcpUtilsStructPtr.outputBufferDipstick > 0 {
		if TcpUtilsStructPtr.outputBufferReadPtr < len(TcpUtilsStructPtr.responseLineBuff) && TcpUtilsStructPtr.outputBufferReadPtr >= 0 {
			var newByte = TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferReadPtr]
			TcpUtilsStructPtr.outputBufferReadPtr++
			TcpUtilsStructPtr.outputBufferReadPtr &= (MaxInOutBufferLength - 1)
			TcpUtilsStructPtr.outputBufferDipstick--
			return newByte
		} else {

			logger.Println("Error reading fromm output buffer")
			return 0
		}

	} else {
		//should be spinning here till thread fills  buffer.
		//TBD
		logger.Println("Warning! Output Buff is empty")
		return 0
	}

/*   if (TcpUtilsStructPtr.outputBufferDipstick >0){
            var	newByte = TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferReadPtr]
			TcpUtilsStructPtr.outputBufferReadPtr++
            TcpUtilsStructPtr.outputBufferReadPtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.outputBufferDipstick--
            return newByte
            
        } else {
            //should be spinning here till thread fills  buffer.
            //TBD
            logger.Println("Warning! Output Buff is empty");
            return 0;
        }*/
        
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   AddByteToOutputBuff( newByte byte){
        //logger.Println("####=",TcpUtilsStructPtr.responseLineBuff)
    
	if TcpUtilsStructPtr.outputBufferDipstick < MaxInOutBufferLength {
		if TcpUtilsStructPtr.outputBufferWritePtr >=0 && TcpUtilsStructPtr.outputBufferWritePtr < MaxInOutBufferLength {
			TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferWritePtr] = newByte
		 	TcpUtilsStructPtr.outputBufferWritePtr++
			TcpUtilsStructPtr.outputBufferWritePtr &= (MaxInOutBufferLength - 1)
			TcpUtilsStructPtr.outputBufferDipstick++
		}else{
			logger.Println("Error write to output : Index out of range")
		}
		

	} else {
		//should be spinning here till thread empties buffer.
		//TBD
		logger.Println("Warning! Output Buff is full")
	}

 /*   if (TcpUtilsStructPtr.outputBufferDipstick < MaxInOutBufferLength){
            TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferWritePtr] = newByte
			TcpUtilsStructPtr.outputBufferWritePtr++
            TcpUtilsStructPtr.outputBufferWritePtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.outputBufferDipstick++
            
        }else {
            //should be spinning here till thread empties buffer.
            //TBD
            logger.Println("Warning! Output Buff is full");
        }*/
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   ReadTwoBytesFromInput() int {
        var tTemp int
        tTemp = (int) ((((int) (TcpUtilsStructPtr.ReadOneByteFromInput()) << 8))  | ((int) (TcpUtilsStructPtr.ReadOneByteFromInput() & 0x00FF))); 
        return tTemp & 0x0000FFFF;
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    ReadFourBytesFromInput() int {
        return ((TcpUtilsStructPtr.ReadTwoBytesFromInput() << 16) | (TcpUtilsStructPtr.ReadTwoBytesFromInput() & 0x00FFFF));     
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    ReadFiveBytesFromInput() uint64 {
        return ((uint64) (TcpUtilsStructPtr.ReadOneByteFromInput()) << 32) |
        	((uint64) (TcpUtilsStructPtr.ReadFourBytesFromInput()) & 0x00000000FFFFFFFF)

    }
func  (TcpUtilsStructPtr	*TcpUtilsStruct)    ReadSixBytesFromInput() uint64 {
        return ((uint64) (TcpUtilsStructPtr.ReadTwoBytesFromInput()) << 32) |
        	((uint64) (TcpUtilsStructPtr.ReadFourBytesFromInput()) & 0x00000000FFFFFFFF)

    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)     ReadEightBytesFromInput() uint64 {
        return (((uint64)(TcpUtilsStructPtr.ReadFourBytesFromInput()) << 32) | ((uint64) (TcpUtilsStructPtr.ReadFourBytesFromInput())))     
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    ReadNBytesFromInput(BytesToRead int ) {
        //not really reading bytes, just dumping data
        if (TcpUtilsStructPtr.inputBufferDipstick >=BytesToRead){
            TcpUtilsStructPtr.inputBufferReadPtr += BytesToRead;
            TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.inputBufferDipstick-=BytesToRead;
                  
        } else {
            //should be speeeing here till thread fills  buffer.
            //TBD
            logger.Println("Warning! Input Buff is empty while jumping ahead")
 
        }
    }


func  (TcpUtilsStructPtr	*TcpUtilsStruct)    WriteTwoBytesToOutput(i int) {
        TcpUtilsStructPtr.AddByteToOutputBuff((byte)(i >> 8))
        TcpUtilsStructPtr.AddByteToOutputBuff((byte)(i))    
    }
    
func  (TcpUtilsStructPtr	*TcpUtilsStruct)   WriteFourBytesToOutput(i int ) {
       TcpUtilsStructPtr.WriteTwoBytesToOutput(i >> 16)
       TcpUtilsStructPtr. WriteTwoBytesToOutput(i)    
    }
    
func  (TcpUtilsStructPtr	*TcpUtilsStruct)    WriteSixBytesToOutput(i uint64) {
        TcpUtilsStructPtr.WriteTwoBytesToOutput((int)(i >> 32))
        TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i))    
    }
    
 
func  (TcpUtilsStructPtr	*TcpUtilsStruct)     WriteEightBytesToOutput(i uint64) {
       TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i >> 32))
       TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i))    
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)     ReceiveSocketData() {





	
        if (!TcpUtilsStructPtr.ConnectedToSGU) {
            //logger.Println("Not Connected ! Attempting to re-connect ")
            //if (!TcpUtilsStructPtr.ConnectToSGU()){
                logger.Println("Not Connected !")
                return;
            //}           
        }
            

        var  bytesAvailable int


		TcpUtilsStructPtr.tcpClient.SetReadDeadline(time.Now().Add(time.Millisecond*500))


		_,err := TcpUtilsStructPtr.reader.Peek(1)

		if err != nil {
			return
		}


		bytesAvailable=TcpUtilsStructPtr.reader.Buffered()
            
        if (bytesAvailable == 0) {
			//logger.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )
            return;                                   
        }                     
        if (bytesAvailable!=0) {
       	    //logger.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )
        	 
        }
        for ;bytesAvailable>0;bytesAvailable-- {
			tByte, err := TcpUtilsStructPtr.reader.ReadByte()

			if err != nil {
				logger.Println("Error reading from socket")
				//TcpUtilsStructPtr.ConnectedToSGU = false
				TcpUtilsStructPtr.CloseTcpClient()
				return


			} else {
            	TcpUtilsStructPtr.AddByteToInputBuff(tByte)
            }        
    	}
	}

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    SendSocketData() {
        if (!TcpUtilsStructPtr.ConnectedToSGU) {
            //try conneting
            //if (!TcpUtilsStructPtr.ConnectToSGU()){
            logger.Println("Not Connected, can not send data!")
            return
            //}           
        }
        
       // logger.Printf("Sending %d Bytes\n", TcpUtilsStructPtr.outputBufferDipstick)
	 pktdata := make([]byte,1)
        for ;TcpUtilsStructPtr.outputBufferDipstick>0; {

			tByte := TcpUtilsStructPtr.GetByteFromOutputBuff()
			//logger.Printf("%x", tByte)
           		pktdata = append(pktdata, tByte) 
   			err :=  TcpUtilsStructPtr.writer.WriteByte(tByte)
			if err != nil {
                logger.Println("Could not  write to socket");
				TcpUtilsStructPtr.Is_TCP_Connected=false
				TcpUtilsStructPtr.CloseTcpClient()
				return
            }
       }
	tpktdata := hex.EncodeToString(pktdata)
	logger.Println("Debugging packet data being sent: ",tpktdata)
logger.Println("func SendSocketData")	
err_flush:=TcpUtilsStructPtr.writer.Flush()
if err_flush != nil {
	logger.Println("Flush Error: ",err_flush)
	TcpUtilsStructPtr.CloseTcpClient()

}
defer func() {
		rec := recover()
		if rec != nil {
			logger.Println("Deffered Recovery")
			logger.Println("Cause:", rec)
			logger.Println("recovered in SendSocketData", rec)

		}

	}()
    }


func  (TcpUtilsStructPtr	*TcpUtilsStruct)    GetSCUIndexFromSCUID(SCUID uint64) int {
        
        var i int
        for i=0;i<TcpUtilsStructPtr.NumOfSCUsInDB;i++ {
		logger.Println("searching current SCU=",TcpUtilsStructPtr.SCUIDinDBArray[i])
            if (TcpUtilsStructPtr.SCUIDinDBArray[i]==SCUID) {
		    logger.Println("SCUID found!!!")
                return i;
			}
        }
	logger.Println("SCUID not found!!!")
        return -1;
    
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    PacketTypeToPacketLength(PacketType int) int {
        
        switch (PacketType) {
            case 0x0001: {
                return FixedPacketLength + 12;       
            }
            
            case 0x0002: {
                return FixedPacketLength;
            }

            case 0x0003: {
                return FixedPacketLength + 28;
            }
            
            case 0x0004: {
                return FixedPacketLength + 9;
            }
            
            case 0x0005: {
                return FixedPacketLength + 10;
            }
            
            case 0xe000: {
                return FixedPacketLength + 2 + TcpUtilsStructPtr.NumOfSCUs*24;
                
            }

            case 0x0011: {
                return FixedPacketLength + 1;
                
            }
            case 0x0022: {
                return FixedPacketLength + 1;
                
            }
            case 0x0023: {
                return FixedPacketLength + 1;
                
            }

            case 0x0024: {
                return FixedPacketLength + 1;
                
            }
            case 0x0025: {
                return FixedPacketLength + 1;
                
            }


            case 0xe001: {
                return FixedPacketLength + 1;
                
            }




            case 0x1000:{
                return FixedPacketLength;
                
            }

            case 0x1001:{
                return FixedPacketLength + 11;
                
            }
            case 0x2000: {
                return FixedPacketLength + 8;
                
            }
            case 0x2001: {
                return FixedPacketLength + 34;
            }
            
            case 0x3000: {
                return FixedPacketLength + 14;
            }
            
            case 0x3001: {
                return FixedPacketLength + 15;
            }
            case 0x4000: {
                return FixedPacketLength+8;
            }

            case 0x4001: {
                return FixedPacketLength+23;
            }
            case 0x5000: {
                return FixedPacketLength+22;
            }
            case 0x5001: {
                return FixedPacketLength+9;
            }

            case 0x6000: {
                return FixedPacketLength+8;
            }
            case 0x6001: {
                return FixedPacketLength+24;
            }
            case 0x7000: {
                return FixedPacketLength+65;
            }
            case 0x7001: {
                return FixedPacketLength+66;
            }
            case 0x8000: {
                return FixedPacketLength+11;
            }
            case 0x8001: {
                return FixedPacketLength+11;
            }
            case 0x9000: {
                return FixedPacketLength+5;
            }
            case 0x9001: {
                return FixedPacketLength+6;
            }
			case 0xA000: {
                return FixedPacketLength+13;
            }	
			case 0xA001: {
                return FixedPacketLength+14;
            }				
            case 0xB000: {
                return FixedPacketLength;
            }
			case 0xD000: {
				return FixedPacketLength-1;
			}
            case 0xB001: {
                return FixedPacketLength+4;
            }
		case 0x1024: {
			return FixedPacketLength+1;
		}
	case 0xC000: {
		return FixedPacketLength+9;
	}
	case 0x1022: {
		return FixedPacketLength+1;
	}
	case 0x1025: {
		return FixedPacketLength+9;
	}
            default: {
                logger.Printf("Invalid Packet Type  %x Specifid", PacketType);
            }            
        }
  	return 0;

    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   SendResponsePacket( OutputPacketType int, SCUID uint64,  StatusByte int, expression   []byte, expressionLength int) {

	if OutputPacketType==0x1024 && StatusByte==0x01{
		if TcpUtilsStructPtr.Curr_major==Sgu_firmware_major&&TcpUtilsStructPtr.Curr_minor==Sgu_firmware_minor{
		logger.Println("SGU=",TcpUtilsStructPtr.SGUID," already updated!!")
			return
		}
	}
		TcpUtilsStructPtr.OutputPacketType = OutputPacketType
        //first add the delimeter
        TcpUtilsStructPtr.AddByteToOutputBuff(StartingDelimeter);
        TcpUtilsStructPtr.OutputPacketLength = TcpUtilsStructPtr.PacketTypeToPacketLength(TcpUtilsStructPtr.OutputPacketType);


		switch OutputPacketType {


			case 0x8000: {
				TcpUtilsStructPtr.OutputPacketLength += expressionLength
				break
			}


			case 0x3000: {
				if (StatusByte & 0x00FF00)==0 {
					//for get mode, packet is smaller
					TcpUtilsStructPtr.OutputPacketLength -= 5
				}

			}

			case 0xB000: {
				TcpUtilsStructPtr.OutputPacketLength += expressionLength
				break
			}
		case 0xD000: {
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}
		case 0x1022: {
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}
		case 0x1025: {
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}
		}


        TcpUtilsStructPtr.OutputPacketLength -= 3; //FixedPacketLength;
		//logger.Println("Packet Type = %d, Packet Length = %x\n", OutputPacketType, TcpUtilsStructPtr.OutputPacketLength)

        TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketLength);
        TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.SGUID);
        //TcpUtilsStructPtr.WriteEightBytesToOutput(TcpUtilsStructPtr.TimeStampHi);
        //TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.TimeStampLo);
	/*	currentTime := time.Now().Local()
        newFormat := currentTime.Format("20060102150405")*/
//IST Format for newFormat
	loc, err := time.LoadLocation("Asia/Calcutta")
	var newFormat string
	if err != nil {
		logger.Println("could not load IST time Zone.... sending in UTC only")
		currentTime := time.Now().Local()
		newFormat = currentTime.Format("20060102150405")
	} else {

		currentTime := time.Now().Local()
		currentTime = currentTime.In(loc)
		newFormat = currentTime.Format("20060102150405")
	}
	logger.Println("newFormat", newFormat)

		for k:=0;k<14;k++ {
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(newFormat[k]))
		}

		
        TcpUtilsStructPtr.WriteFourBytesToOutput(TcpUtilsStructPtr.OutputSeqNumber);
        TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketType);
        logger.Println("Sending For packet type=",TcpUtilsStructPtr.OutputPacketType," seqNo=",TcpUtilsStructPtr.OutputSeqNumber)
        //done with common part.
        switch (TcpUtilsStructPtr.OutputPacketType) {
            

            case 0x0011: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0x0022: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0x0023: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0x0024: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0x0025: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0xe001: {
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
                break;                
            }
            case 0x1000:{
            	break;
            }

            case 0x2000: {
                break;
            }
            case 0x3000: {
                //separate LampId and LampVal;
                lampVal := StatusByte & 0x0FF;
				getSetByte := (StatusByte >> 8) & 0x0FF
log.Println("getSetByte--:", getSetByte)
                TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID);
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(getSetByte));
				//TcpUtilsStructPtr.AddByteToOutputBuff((byte)(0x09));
				if (getSetByte==1) {
                	//for set need to set additional fields
					/*if lampVal==0{
						TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
					}else{
						x:=0x09
						TcpUtilsStructPtr.AddByteToOutputBuff((byte)(x));
					}*/
					logger.Println("Dim=",lampVal);
					tt:=0x00
					if lampVal==1{
						tt=0x01	
					}else if lampVal==2{
						tt=0x02
					}else if lampVal==3{
						tt=0x03
					}else if lampVal==4{
						tt=0x04
					}else if lampVal==5{
						tt=0x05
					}else if lampVal==6{
						tt=0x06
					}else if lampVal==7{
						tt=0x07
					}else if lampVal==8{
						tt=0x08
					}else if lampVal==9{
						tt=0x09
					}else if lampVal==10{
						tt=0x0a
					}
					TcpUtilsStructPtr.AddByteToOutputBuff((byte)(tt));
					tval:=0x01
					if lampVal==0{
						tval=0
					}
                	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(tval ));
                	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(tval ));
                	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(tval ));
                	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(tval ));
				}
                
                break;
                
            }

            case 0x4000: {
                break;
            }


            case 0x5000: {
                TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID);
				for k:=0;k<14;k++ {
					TcpUtilsStructPtr.AddByteToOutputBuff((byte)(newFormat[k]))
				}

                
                break;
                
            }

            case 0x8000: {
                TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID);
                //write Get/Set which is byte0 of StatusByte
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x0FF));



                //write scheduling id  which is byte1 of StatusByte
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)((StatusByte >> 8)  & 0x0FF));

                //write pwm  state  which is byte2 of StatusByte
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)((StatusByte >> 16)  & 0x0FF));
		logger.Println("Sending 8000 with PWM=",(int)((StatusByte >> 16)  & 0x0FF))
				for k := 0; k < expressionLength; k++ {
					 TcpUtilsStructPtr.AddByteToOutputBuff(expression[k])
				
				}
				                

                
                break;
                
            }

			case 0x9000: {
			logger.Println("Entered into 0X9000 Packet..");
                //write Get/Set which is nibble-0  of StatusByte
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x00FF));
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[0] & 0x00F);
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[1] & 0x00F);
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[2] & 0x00F);
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[3] & 0x00F); 
				logger.Println("Output :",expression[0]);
				logger.Println("Output :",expression[1]);
				logger.Println("Output :",expression[2]);
				logger.Println("Output :",expression[3]);
			logger.Println("Exited from 0X9000 Packet..");				
                break;
                
            } 
			
            case 0xA000: {
			    logger.Println("Entered into 0XA000 Packet..");
			/*	TcpUtilsStructPtr.Enable=1;
				TcpUtilsStructPtr.PollingRate=600;
				TcpUtilsStructPtr.ResponseRate=600;
				TcpUtilsStructPtr.DeviceTimeout=100;
				TcpUtilsStructPtr.SlaveId=1; */
				
				//TcpUtilsStructPtr.SlaveIds=0xffffffff;
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x00FF));
                /*TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[0]));
                TcpUtilsStructPtr.WriteTwoBytesToOutput((int)(expression[1]));
                TcpUtilsStructPtr.WriteTwoBytesToOutput((int)(expression[2]));
                TcpUtilsStructPtr.WriteTwoBytesToOutput((int)(expression[3]));
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[4]));
				TcpUtilsStructPtr.WriteFourBytesToOutput(TcpUtilsStructPtr.SlaveIds);*/
				for k := 0; k < expressionLength; k++ {
					TcpUtilsStructPtr.AddByteToOutputBuff(expression[k])
				}
				/*logger.Println("Output :",(byte)(expression[0]));
				logger.Println("Output :",(int)(expression[1]));
				logger.Println("Output :",(int)(expression[2]));
				logger.Println("Output :",(int)(expression[3]));
				logger.Println("Output :",(byte)(expression[4]));*/
			logger.Println("Exited from 0XA000 Packet..");	
                break;
                
            }

            case 0xB000:{
			logger.Println("Entered into 0XB000 Packet..");
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x00FF));
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[0]);
                TcpUtilsStructPtr.AddByteToOutputBuff(expression[1]);
                for i:=0;i<expressionLength-2;i++{
                    //tmp,_:=strconv.ParseInt(TcpUtilsStruc0tPtr.Query[i:i+2],10,32)
					logger.Println("Output :",(expression[i+2]));
                    TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[i+2]))
                }
			logger.Println("Exited from 0XB000 Packet..");	
                break;

            }
		case 0xD000:{
			logger.Println("Entered into 0XD000 Packet..");
			//TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x00FF));
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[0]);
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[1]);
			for i:=0;i<expressionLength-2;i++{
				//tmp,_:=strconv.ParseInt(TcpUtilsStruc0tPtr.Query[i:i+2],10,32)
				logger.Println("Output :",(expression[i+2]));
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[i+2]))
			}
			logger.Println("Exited from 0XD000 Packet..");
			break;

		}
		case 0x1024:{
			logger.Println("Entered into 0x1024 Packet..");
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
			break;

		}
		case 0xC000:{
			logger.Println("Entered into 0xC000 Packet..");
			Scu_Current_pos[SCUID]=0
			TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID);
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte));
			break;

		}
		case 0x1022:{
			logger.Println("Entered into 0x1022 Packet..");
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			for i:=0;i<expressionLength;i++{
				TcpUtilsStructPtr.AddByteToOutputBuff(expression[i])
			}
		}
		case 0x1025:{
			logger.Println("Entered into 0x1025 Packet..");
			TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID);
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			for i:=0;i<expressionLength;i++{
				TcpUtilsStructPtr.AddByteToOutputBuff(expression[i])
			}
		}
            default: {
                logger.Printf("Invalid Output Packet Type %x Specifid\n",TcpUtilsStructPtr.OutputPacketType);
            }            
        } 
        if (TcpUtilsStructPtr.outputBufferDipstick < (TcpUtilsStructPtr.OutputPacketLength + 3)) {
            logger.Println("Output Packet Formating Error");
        }
        TcpUtilsStructPtr.SendSocketData();
    }


func  (TcpUtilsStructPtr	*TcpUtilsStruct)   ParseInputPacket() {

        //add data from socket buffer to local buffer
        TcpUtilsStructPtr.ReceiveSocketData();

        
        if (TcpUtilsStructPtr.InputSyncSearchStatus==0) {
            //need to search for start delimiter
            for ;TcpUtilsStructPtr.inputBufferDipstick>0; {
                tByte := TcpUtilsStructPtr.ReadOneByteFromInput();
                if (tByte==StartingDelimeter) {
                    //found sync. Confirm by parsing and looking at next sync.
                    //TBD
                    //rewind dipstick and read pointer
                    TcpUtilsStructPtr.RewindInputBuffer(1);
                    TcpUtilsStructPtr.InputSyncSearchStatus = 1; 
                    break;
                }
            }
            //could not sync, so just return
            if (TcpUtilsStructPtr.InputSyncSearchStatus==0) {
                return;
			}
        }

        
        TcpUtilsStructPtr.ReceiveSocketData();
        //here dipstick has to be minimum, else we can
        //not parse fixed header


        if (TcpUtilsStructPtr.inputBufferDipstick < (FixedPacketLength)) {
            return;
        
		}
        
        
    	//confirm start delimiter
    	if (TcpUtilsStructPtr.ReadOneByteFromInput() != StartingDelimeter) {
            logger.Println("Failed to  match start delimiter");
            TcpUtilsStructPtr.InputSyncSearchStatus = 0;
            return;
    	}


    	TcpUtilsStructPtr.InputPacketLength = TcpUtilsStructPtr.ReadTwoBytesFromInput();
    	if (TcpUtilsStructPtr.InputPacketLength > 0x8000) {
    		logger.Printf("Invalid Packet Length = %x   \n",TcpUtilsStructPtr.InputPacketLength); 
    		
    	} else {
    		//System.out.printf("Packet Length = %x   \n",InputPacketLength); 
    	}

    
        
        TcpUtilsStructPtr.ReceiveSocketData();


        //make sure entire packet is in buffer
        if (TcpUtilsStructPtr.inputBufferDipstick < (TcpUtilsStructPtr.InputPacketLength) ) {
            //insufficient data in buuer
            //rewind pointers and return
            //need to rewind by 3 bytes
            TcpUtilsStructPtr.RewindInputBuffer(3);
            return;
        }
        if TcpUtilsStructPtr.inputBufferReadPtr>0&&TcpUtilsStructPtr.inputBufferReadPtr+TcpUtilsStructPtr.inputBufferDipstick>0&&TcpUtilsStructPtr.inputBufferReadPtr+TcpUtilsStructPtr.inputBufferDipstick<len(TcpUtilsStructPtr.commandLineBuff){
        	logger.Printf("Packet=%x",TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferReadPtr:TcpUtilsStructPtr.inputBufferReadPtr+TcpUtilsStructPtr.inputBufferDipstick]);           
        }
	
        
        //get 8 bytes of SGU id
       // TcpUtilsStructPtr.SGUID = TcpUtilsStructPtr.ReadSixBytesFromInput();
        SGUID := TcpUtilsStructPtr.ReadSixBytesFromInput()
	if SGUID != 0 {
		TcpUtilsStructPtr.SGUID = SGUID
	}
        logger.Printf("SGU ID  %d \n",TcpUtilsStructPtr.SGUID);        
    	
		//TimeStampString  := make([]byte,14)


    	//get first 8 bytes of timestamp
        TcpUtilsStructPtr.TimeStampHi = TcpUtilsStructPtr.ReadEightBytesFromInput();

    	//get remaining 6 bytes of timestamp
        TcpUtilsStructPtr.TimeStampLo = TcpUtilsStructPtr.ReadSixBytesFromInput();
        
        //get 4 bytes of input sequence number
        TcpUtilsStructPtr.InputSeqNumber = TcpUtilsStructPtr.ReadFourBytesFromInput();
        
    	TcpUtilsStructPtr.InputPacketType = TcpUtilsStructPtr.ReadTwoBytesFromInput();
        
    	logger.Printf("Received packet type %d \n",TcpUtilsStructPtr.InputPacketType);

		//TimeStampString[0:7] = TcpUtilsStructPtr.TimeStampHi[0:7]

		tArray := make([]byte, 8)

		binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampHi)
		TimeStampString := string(tArray[:8])

		binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampLo)
		TimeStampString += string(tArray[:6])

		logger.Println(TimeStampString)

		//logger.Println("Packet");
		//for i:=TcpUtilsStructPtr.inputBufferReadPtr;i<TcpUtilsStructPtr.inputBufferDipstick+TcpUtilsStructPtr.inputBufferReadPtr;i++{
			
		//}

        switch (TcpUtilsStructPtr.InputPacketType) {
            case 0x0001: {
		    TcpUtilsStructPtr.Is_updating=false
		    //Reset Indication
                //send packet of type 0x11
                //read 12 bytes from buffer and junk them
                TcpUtilsStructPtr.ReadNBytesFromInput(12);     
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                logger.Println("Received packet type 0x0001 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x11,0,0,nil,0);
                TcpUtilsStructPtr.SCUListreceived = true
		break;
            }
            case 0x0002: {
		    TcpUtilsStructPtr.Is_updating=false
		    //Keep Alive
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;               
                logger.Println("Received packet type 0x0002 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x22,0,0,nil,0);
                break;               
            }
            case 0x0003: {
		    TcpUtilsStructPtr.Is_updating=false
		    //SCU Lit
                //parse packet
                //TBD
                TcpUtilsStructPtr.NumOfSCUs = TcpUtilsStructPtr.ReadTwoBytesFromInput()

				logger.Printf("Found  %d SCUs in list\n", TcpUtilsStructPtr.NumOfSCUs -1 )


				if (TcpUtilsStructPtr.NumOfSCUs > TcpUtilsStructPtr.MAXNumOFSCUS) {

					logger.Printf("Max num of SCUs exceeded. Received %d\n", TcpUtilsStructPtr.NumOfSCUs )
					TcpUtilsStructPtr.NumOfSCUs = TcpUtilsStructPtr.MAXNumOFSCUS

				}
				//first ID is zigbeed Id of SGU itself
				TcpUtilsStructPtr.SGUZigbeeID = TcpUtilsStructPtr.ReadEightBytesFromInput()
				//read and dump reserved byte
				TcpUtilsStructPtr.ReadNBytesFromInput(1)

				TcpUtilsStructPtr.NumOfSCUs--

                
                for i:=0;i<TcpUtilsStructPtr.NumOfSCUs;i++ {
                    //TcpUtilsStructPtr.SCUIDArray[i] = TcpUtilsStructPtr.ReadEightBytesFromInput();
                    scuid := TcpUtilsStructPtr.ReadEightBytesFromInput()
				if scuid != 0{
					TcpUtilsStructPtr.SCUIDArray[i] = scuid
				}
		    //read and dump reserved byte
                    TcpUtilsStructPtr.ReadOneByteFromInput();
                }
                
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;              
                logger.Println("Received packet type 0x0003 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x23,0,0,nil,0);
                TcpUtilsStructPtr.SCUListreceived=true;
                break;               
            }
            case 0x0004: {      //SCU Deleted
                //parse packet
                //just dump 9 bytes
                TcpUtilsStructPtr.ReadNBytesFromInput(9);
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                logger.Println("Received packet type 0x0004 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x24,0,0,nil,0);
                break;               
            }  
            case 0x0005: {
		    TcpUtilsStructPtr.Is_updating=false
		    //SCU Added
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(8);
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                logger.Println("Received packet type 0x0005 successfully with seqNo=",TcpUtilsStructPtr.InputSeqNumber);
                TcpUtilsStructPtr.SendResponsePacket(0x25,0,0,nil,0);
                break;               
            }
            case 0xe000: {
		    TcpUtilsStructPtr.Is_updating=false
		    //Input Status
                //parse packet
                //logger.Println(TcpUtilsStructPtr.responseLineBuff)
                NumSCUPlusSGU := TcpUtilsStructPtr.ReadTwoBytesFromInput();
                TcpUtilsStructPtr.ControlSGUID = TcpUtilsStructPtr.ReadEightBytesFromInput();

				sguSTATUS := TcpUtilsStructPtr.ReadOneByteFromInput()
				DigitalInput1 := TcpUtilsStructPtr.ReadOneByteFromInput()
				DigitalInput2 := TcpUtilsStructPtr.ReadOneByteFromInput()
				DigitalInput3 := TcpUtilsStructPtr.ReadOneByteFromInput()
				TcpUtilsStructPtr.AlertState = 0

				if (sguSTATUS==0) {
					if (DigitalInput1 != 1) {
                        logger.Println("DigitalInput1 tripped")
						TcpUtilsStructPtr.AlertState = 1
                       // go TcpUtilsStructPtr.SendAlertSMS()
					}
					if (DigitalInput2 != 1) {
                        logger.Println("DigitalInput2 tripped")
						TcpUtilsStructPtr.AlertState |= 2
                       // go TcpUtilsStructPtr.SendAlertSMS()
					}
					if (DigitalInput3 != 1) {
                        logger.Println("DigitalInput3 tripped")
						TcpUtilsStructPtr.AlertState |= 4
					}
					//go TcpUtilsStructPtr.SendAlertSMS()
				}

                TcpUtilsStructPtr.ReadNBytesFromInput(12);



				logger.Printf("Found status info for %d SCUs\n",NumSCUPlusSGU-1)
                
                
                //read the SGU 

				tempCounter := TcpUtilsStructPtr.LampStatusCount;
                        
                for i:=0;i<NumSCUPlusSGU-1;i++ {
                    SCUID := TcpUtilsStructPtr.ReadEightBytesFromInput();
                    scuIndex := TcpUtilsStructPtr.GetSCUIndexFromSCUID(SCUID);
                    //dump next 5 bytes as they are not used for now


                    scuStatus := TcpUtilsStructPtr.ReadOneByteFromInput();

				
					//dump next 4 bytes
                    TcpUtilsStructPtr.ReadNBytesFromInput(4);
                    
                    
                    tempAnalog := TcpUtilsStructPtr.ReadFourBytesFromInput();
                    
                    

                    tempDigital := TcpUtilsStructPtr.ReadFiveBytesFromInput();

                    logger.Println("Before Received packet type 0xe000 successfully for scuindex=",scuIndex," with status=",(tempDigital& (0x0FF))," SCUID=",SCUID);
                    //dume next 2 bytes
                    TcpUtilsStructPtr.ReadNBytesFromInput(2);

					if (scuStatus==0) {

                    	if (scuIndex>=0) {
                        	TcpUtilsStructPtr.SCUAnalogP1StateArray[tempCounter] = tempAnalog;
							tempDigital = tempDigital | (((uint64)(scuIndex)) << 40)
                        	TcpUtilsStructPtr.LampStatusArray[tempCounter] = tempDigital;
							tempCounter++
                    	} else {
                        	logger.Println("Unindentified SCU specified");
                        
                    	}


					}
                    logger.Println("Received packet type 0xe000 successfully for scuid=",scuIndex," with status=",(tempDigital& (0x0FF)));
                             
                }
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
				TcpUtilsStructPtr.LampStatusCount = tempCounter
                logger.Println("Received packet type 0xe000 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0xe001,0,0,nil,0);
                break;               
            }
            
            //response from SGU for queries
            
            case 0x1001: {      //Get SGU details
                //parse packet
            	
                TcpUtilsStructPtr.ReadNBytesFromInput(3);  
                TcpUtilsStructPtr.SGULatitude = TcpUtilsStructPtr.ReadTwoBytesFromInput() 
                TcpUtilsStructPtr.SGULongitude = TcpUtilsStructPtr.ReadTwoBytesFromInput()
		    TcpUtilsStructPtr.Curr_major=byte(TcpUtilsStructPtr.ReadOneByteFromInput())
		    TcpUtilsStructPtr.Curr_minor=byte(TcpUtilsStructPtr.ReadOneByteFromInput())
                TcpUtilsStructPtr.ReadNBytesFromInput(2);
	break;
		    logger.Println("Running version, major=",TcpUtilsStructPtr.Curr_major," minor=",TcpUtilsStructPtr.Curr_minor)
	logger.Println("SGU ID at 10001 resp",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)) 
	stmt,_:=dbController.Db.Prepare("update sgu set major=?,minor=? where sgu_id='"+strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)+"'")
		    _,eorr:=stmt.Exec(TcpUtilsStructPtr.Curr_major,TcpUtilsStructPtr.Curr_minor)
	logger.Println("update sgu query executed DB")
		    defer stmt.Close()
		    if eorr!=nil{
			    logger.Println(eorr)
		    }
                logger.Printf("Received sgu coordinates: %f\n" , TcpUtilsStructPtr.SGULatitude);
                logger.Printf("Received sgu coordinates: %f\n" , TcpUtilsStructPtr.SGULongitude);
                logger.Println("Received packet type 0x1001 successfully");
                break;               
            }
            
            case 0x2001: {      //Get SCU details
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(34); 
                logger.Println("Received packet type 0x2001 successfully");
                break;               
            }
         
            case 0x3001: {      //Get/Set Digital Output State
                //parse packet

				
				//get status
				status := TcpUtilsStructPtr.ReadOneByteFromInput()

				SCUID := TcpUtilsStructPtr.ReadEightBytesFromInput()
				scuIndex := TcpUtilsStructPtr.GetSCUIndexFromSCUID(SCUID)
                gs:=TcpUtilsStructPtr.ReadOneByteFromInput()

				if (scuIndex >=0) {
                    tempDigital :=TcpUtilsStructPtr.ReadFiveBytesFromInput()
					//for retry
					strHash:=strconv.Itoa(0x3000)+"#"+strconv.FormatUint(SCUID,10)+"#"+strconv.Itoa(int(gs))+"#"+strconv.FormatUint((tempDigital>> 32) & 0x0FF,10)
					logger.Println("Hash received=",strHash)
					if TcpUtilsStructPtr.RetryHash[strHash]==1{
						TcpUtilsStructPtr.RetryHash[strHash]=2
					}
		            if (TcpUtilsStructPtr.ResponseReceivedCount <  TcpUtilsStructPtr.MAXNumOFSCUS) {
						if (status==0) {
							TcpUtilsStructPtr.ResponseReceivedArray[TcpUtilsStructPtr.ResponseReceivedCount] = ((scuIndex << 16) & (0xFFFF0000)) | (((int)(tempDigital & 0x01)))
                            tempDigital = tempDigital | (((uint64)(scuIndex)) << 40)
                            TcpUtilsStructPtr.LampStatusArray[TcpUtilsStructPtr.LampStatusCount] = tempDigital;
                            
						} else {

							TcpUtilsStructPtr.ResponseReceivedArray[TcpUtilsStructPtr.ResponseReceivedCount] = ((scuIndex << 16) & (0xFFFF0000)) | 2						
                            tempDigital = 2 | (((uint64)(scuIndex)) << 40)
                            TcpUtilsStructPtr.LampStatusArray[TcpUtilsStructPtr.LampStatusCount] = tempDigital;

						}

						TcpUtilsStructPtr.ResponseReceivedCount++
						TcpUtilsStructPtr.LampStatusCount++
					} else {
						logger.Println("Too many responses pending. Buffer full")
					}
				} else {
					logger.Println("Response received when SCU not in list. ")
                	TcpUtilsStructPtr.ReadNBytesFromInput(5)

				} 
                
                  
                logger.Printf("Received packet type 0x3001 successfully. status = %2.2x\n",status);
                break;               
            }
            case 0x4001: {      //Get Time Stamp
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(23);
                logger.Println("Received packet type 0x4001 successfully");
                break;               
            }
            case 0x5001: {      //Set Time Stamp
                //parse packet
                status := TcpUtilsStructPtr.ReadOneByteFromInput();
                scuID :=  TcpUtilsStructPtr.ReadEightBytesFromInput() 
                if status==0 {
                	logger.Printf("Time set successfully for SCU %d\n", scuID)
                } else {
                
                	logger.Printf("Error setting time for SCU %d\n", scuID)
                }
                  
                logger.Println("Received packet type 0x5001 successfully");
                break;               
            }
            case 0x6001: {      //Get Input Status
                //parse packet
                //ReadNBytesFromInput(24); 
                TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
                logger.Println("Received packet type 0x6001 successfully");
                break;               
            }
            case 0x7001: {      //Set Input Status
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(66); 
                logger.Println("Received packet type 0x7001 successfully");
                
                break;               
            }
        case 0x0006: {
		TcpUtilsStructPtr.Is_updating=false

            //skip no. of devices (1 byte)
			/*logger.Println(TcpUtilsStructPtr.commandLineBuff)
			logger.Println("len=",TcpUtilsStructPtr.InputPacketLength)*/
			//return
			//TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
            TcpUtilsStructPtr.ReadOneByteFromInput();
            //skip device id (1 byte)
            TcpUtilsStructPtr.ReadOneByteFromInput();
            //get status (1 byte)
            status:=TcpUtilsStructPtr.ReadOneByteFromInput();
            //status==0 for successful response
            if status==0{
                //skip this length, device id and modbus type(3 bytes).
                TcpUtilsStructPtr.ReadOneByteFromInput();
                TcpUtilsStructPtr.ReadTwoBytesFromInput();
				//skip 1 byte response length
				TcpUtilsStructPtr.ReadOneByteFromInput();
				//read Watts Total 4bytes
				wa2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
                wa1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				wa:=(wa1 << 16) | (wa2 & 0x00FFFF)
				nn := uint32(wa)
				waf := math.Float32frombits(nn)
				kwa:=waf/1000.0
				logger.Println("KWA=",kwa)

                //skip 28 bytes to get to the Pf
				TcpUtilsStructPtr.ReadNBytesFromInput(28);

				//read Pf Total 4bytes
				pf2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				pf1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				pf:=(pf1 << 16) | (pf2 & 0x00FFFF)
				pfnn := uint32(pf)
				pff := math.Float32frombits(pfnn)
				logger.Println("PF=",pff)

				//skip 12 bytes to get to the Va
				TcpUtilsStructPtr.ReadNBytesFromInput(12);

				//read Va Total 4bytes
				va2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				va1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				va:=(va1 << 16) | (va2 & 0x00FFFF)
				vann := uint32(va)
				vaf := math.Float32frombits(vann)
				kva:=vaf/1000.0
				logger.Println("kva=",kva)

				//skip 32 bytes to get to the Va
				TcpUtilsStructPtr.ReadNBytesFromInput(32);

				//read Vr Total 4bytes
				vr2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vr1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vr:=(vr1 << 16) | (vr2 & 0x00FFFF)
				vrnn := uint32(vr)
				vrf := math.Float32frombits(vrnn)
				logger.Println("vr=",vrf)

				//read Vy Total 4bytes
				vy2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vy1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vy:=(vy1 << 16) | (vy2 & 0x00FFFF)
				vynn := uint32(vy)
				vyf := math.Float32frombits(vynn)
				logger.Println("vy=",vyf)

				//read Vb Total 4bytes
				vb2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vb1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				vb:=(vb1 << 16) | (vb2 & 0x00FFFF)
				vbnn := uint32(vb)
				vbf := math.Float32frombits(vbnn)
				logger.Println("vb=",vbf)

				//skip 4 bytes to get to the Ir
				TcpUtilsStructPtr.ReadNBytesFromInput(4);

				//read Ir Total 4bytes
				ir2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				ir1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				ir:=(ir1 << 16) | (ir2 & 0x00FFFF)
				irnn := uint32(ir)
				irf := math.Float32frombits(irnn)
				logger.Println("ir=",irf)

				//read Iy Total 4bytes
				iy2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				iy1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				iy:=(iy1 << 16) | (iy2 & 0x00FFFF)
				iynn := uint32(iy)
				iyf := math.Float32frombits(iynn)
				logger.Println("iy=",iyf)
				
				//read Ib Total 4bytes
				ib2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				ib1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				ib:=(ib1 << 16) | (ib2 & 0x00FFFF)
				ibnn := uint32(ib)
				ibf := math.Float32frombits(ibnn)
				logger.Println("ib=",ibf)

				//read freq Total 4bytes
				fre2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				fre1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				fre:=(fre1 << 16) | (fre2 & 0x00FFFF)
				frenn := uint32(fre)
				fref := math.Float32frombits(frenn)
				logger.Println("fre=",fref)

				//read whq Total 4bytes
				wh2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				wh1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				wh:=(wh1 << 16) | (wh2 & 0x00FFFF)
				whnn := uint32(wh)
				whf := math.Float32frombits(whnn)
				kwh:=whf/1000.0
				logger.Println("kwh=",kwh)

				logger.Println("SGUID=",TcpUtilsStructPtr.SGUID)
				//dbController.DbSemaphore.Lock()
				//defer dbController.DbSemaphore.Unlock()
				db := dbController.Db
				stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,KW=?,Pf=?,KVA=?,Vr=?,Vy=?,Vb=?,Ir=?,Iy=?,Ib=?,KWH=?,freq=?")
				defer stmt.Close()
				_, eorr:=stmt.Exec(TcpUtilsStructPtr.SGUID,kwa,pff,kva,vrf,vyf,vbf,irf,iyf,ibf,kwh,fref)
				if eorr!=nil{
					logger.Println(eorr)
				}else {
					logger.Println("Inserting 0x0006 Packet Successfully with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
				}
				/*//skip 116 bytes to get to the Ir
				TcpUtilsStructPtr.ReadNBytesFromInput(116);

				//read rhq Total 4bytes
				rh2:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				rh1:=TcpUtilsStructPtr.ReadTwoBytesFromInput();
				rh:=(rh1 << 16) | (rh2 & 0x00FFFF)
				logger.Println("rh=",rh)*/

				//skip 9 bytes
				//TcpUtilsStructPtr.ReadNBytesFromInput(9);
                //if length==28 read W/Pf/VA else read others
                /*if length==28{
                    logger.Println("Recieved 0x0006 Packet Successfully with length 28")
                    //read watts (4 bytes)
                    watts:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    //skip 12 bytes to get pf
                    TcpUtilsStructPtr.ReadSixBytesFromInput();
                    TcpUtilsStructPtr.ReadSixBytesFromInput();
                    //read pf (4 bytes)
                    pf:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    //skip 4 bytes to get va
                    TcpUtilsStructPtr.ReadFourBytesFromInput();
                    // read va (4 bytes)
                    va:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    dbController.DbSemaphore.Lock()
                    defer dbController.DbSemaphore.Unlock()
                    db := dbController.Db
                    rows,err :=db.Query("Select KVA from parameters where sgu_id='"+strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)+"'order by timestamp desc limit 1")
                    defer rows.Close()
                    if err!=nil{
                        logger.Println(err)
                    }else {
                        if rows.Next(){
                            var KVA string
                            err=rows.Scan(&KVA)
                            if err!=nil{
                                logger.Println(err)
                            }else {
                                if len(KVA)!=0{
                                    stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,KW=?,Pf=?,KVA=?")
                                    defer stmt.Close()
                                    _, eorr:=stmt.Exec(TcpUtilsStructPtr.SGUID,watts,pf,va)
                                    if eorr!=nil{
                                        logger.Println(err)
                                    }else {
                                        logger.Println("Inserting new 0x0006 Packet Successfully with length 28 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                                    }
                                }else {
                                    stmt1, err := dbController.Db.Prepare("update  parameters set sgu_id=?,KW=?,Pf=?,KVA=?")
                                    defer stmt1.Close()
                                    if err != nil {
                                        logger.Println(err)
                                    }
                                    _, eorr:=stmt1.Exec(TcpUtilsStructPtr.SGUID,watts,pf,va)
                                    if eorr!=nil{
                                        logger.Println(err)
                                    }else {
                                        logger.Println("Updating 0x0006 Packet Successfully with length 28 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                                    }
                                }
                            }
                        } else{
                            stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,KW=?,Pf=?,KVA=?")
                            defer stmt.Close()
                            _, eorr:=stmt.Exec(TcpUtilsStructPtr.SGUID,watts,pf,va)
                            if eorr!=nil{
                                logger.Println(err)
                            }else {
                                logger.Println("Inserting new 0x0006 Packet Successfully with length 28 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                            }
                        }
                    }
                } else if length==18{
                    logger.Println("Recieved 0x0006 Packet Successfully with length 18")
                    vr:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    vy:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    vb:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    //skip 4 bytes
                    TcpUtilsStructPtr.ReadFourBytesFromInput();

                    ir:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    iy:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    ib:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    freq:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    wh:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    vah:=TcpUtilsStructPtr.ReadFourBytesFromInput();
                    dbController.DbSemaphore.Lock()
                    defer dbController.DbSemaphore.Unlock()
                    db := dbController.Db
                    rows,err :=db.Query("Select Vr from parameters where sgu_id='"+strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)+"'order by timestamp desc limit 1")
                    defer rows.Close()
                    if err!=nil{
                        logger.Println(err)
                    }else {
                        if rows.Next(){
                            var Vr string
                            err=rows.Scan(&Vr)
                            if err!=nil{
                                logger.Println(err)
                            }else {
                                if len(Vr)!=0{
                                    stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,Vr=?,Vy=?,Vb=?,Ir=?,Iy=?,Ib=?,KVAH=?,KWH=?,freq=?")
                                    defer stmt.Close()
                                    _, eorr:=stmt.Exec(TcpUtilsStructPtr.SGUID,vr,vy,vb,ir,iy,ib,vah,wh,freq)
                                    if eorr!=nil{
                                        logger.Println(err)
                                    }else {
                                        logger.Println("Inserting new 0x0006 Packet Successfully with length 18 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                                    }
                                }else {
                                    stmt1, err := dbController.Db.Prepare("update  parameters set sgu_id=?,Vr=?,Vy=?,Vb=?,Ir=?,Iy=?,Ib=?,KVAH=?,KWH=?,freq=?")
                                    defer stmt1.Close()
                                    if err != nil {
                                        logger.Println(err)
                                    }
                                    _, eorr:=stmt1.Exec(TcpUtilsStructPtr.SGUID,vr,vy,vb,ir,iy,ib,vah,wh,freq)
                                    if eorr!=nil{
                                        logger.Println(err)
                                    }else {
                                        logger.Println("Updating 0x0006 Packet Successfully with length 18 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                                    }
                                }
                            }
                        }else {
                            stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,Vr=?,Vy=?,Vb=?,Ir=?,Iy=?,Ib=?,KVAH=?,KWH=?,freq=?")
                            defer stmt.Close()
                            _, eorr:=stmt.Exec(TcpUtilsStructPtr.SGUID,vr,vy,vb,ir,iy,ib,vah,wh,freq)
                            if eorr!=nil{
                                logger.Println(err)
                            }else {
                                logger.Println("Inserting new 0x0006 Packet Successfully with length 18 with sguid=",strconv.FormatUint(TcpUtilsStructPtr.SGUID,10))
                            }
                        }
                    }
                }*/
            } else {
                logger.Println("Recieved 0x0006 Packet with invalid status")
            }
        }
           case 0x9001: {      //Get Modbus details
                //parse packet
                modbusStatus9001 :=TcpUtilsStructPtr.ReadOneByteFromInput(); 
				modbusSet9001 :=TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusBaudRate9001 :=TcpUtilsStructPtr.ReadOneByteFromInput(); 
                modbusStopbits9001 := TcpUtilsStructPtr.ReadOneByteFromInput(); 
                modbusParity9001 := TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusNumberOfBits9001 := TcpUtilsStructPtr.ReadOneByteFromInput();
				logger.Printf("Received Modbus Status: %d\n" , modbusStatus9001);
				logger.Printf("Received Modbus Set: %d\n" , modbusSet9001);
				logger.Printf("Received Modbus BaudRate: %d\n" , modbusBaudRate9001);
				logger.Printf("Received Modbus Stopbits: %d\n" , modbusStopbits9001);
				logger.Printf("Received Modbus Parity: %d\n" , modbusParity9001);
				logger.Printf("Received Modbus NumberOfBits: %d\n" , modbusNumberOfBits9001);
                logger.Println("Received packet type 0x9001 successfully");
                break;               
            } 
            case 0xA001: {      //Get Modbus details
                //parse packet
                modbusStatus :=TcpUtilsStructPtr.ReadOneByteFromInput(); 
				modbusSet :=TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusEnabled :=TcpUtilsStructPtr.ReadOneByteFromInput(); 
                modbusPollingRate := TcpUtilsStructPtr.ReadTwoBytesFromInput(); 
                modbusResponseRate := TcpUtilsStructPtr.ReadTwoBytesFromInput();
				modbusDeviceTimeout := TcpUtilsStructPtr.ReadTwoBytesFromInput();
				modbusSlaveId := TcpUtilsStructPtr.ReadOneByteFromInput();
				TcpUtilsStructPtr.ReadNBytesFromInput(4);
				logger.Printf("Received Modbus Status: %d\n" , modbusStatus);
				logger.Printf("Received Modbus Set: %d\n" , modbusSet);
				logger.Printf("Received Modbus Enable: %d\n" , modbusEnabled);
				logger.Printf("Received Modbus PollingRate: %d\n" , modbusPollingRate);
				logger.Printf("Received Modbus ResponseRate: %d\n" , modbusResponseRate);
				logger.Printf("Received Modbus ResponseRate: %d\n" , modbusDeviceTimeout);
				logger.Printf("Received Modbus SlaveIds: %d\n" , modbusSlaveId);
                logger.Println("Received packet type 0xA001 successfully");
                break;               
            }
            case 0xB001: {      //get modbus Data
                //parse packet
                modbusStatus1 :=TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusSet1 :=TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusDeviceId1 :=TcpUtilsStructPtr.ReadOneByteFromInput();
				modbusDataLength :=TcpUtilsStructPtr.ReadOneByteFromInput();
				//x:=8
				TcpUtilsStructPtr.ReadNBytesFromInput((int)(modbusDataLength));
				logger.Printf("Received Modbus Status1: %d\n" , modbusStatus1);
				logger.Printf("Received Modbus Set1: %d\n" , modbusSet1);
				logger.Printf("Received Modbus DeviceId1: %d\n" , modbusDeviceId1);
				logger.Printf("Received Modbus DataLength: %d\n" , modbusDataLength);
                logger.Println("Received packet type 0xB001 successfully");
                
                break;               
            }
			case 0x8001: {
				//parse packet
				status :=TcpUtilsStructPtr.ReadOneByteFromInput();
				scuid :=TcpUtilsStructPtr.ReadEightBytesFromInput();
				gs :=TcpUtilsStructPtr.ReadOneByteFromInput();
				scid :=(int)(TcpUtilsStructPtr.ReadOneByteFromInput());
				pwm :=(int)(TcpUtilsStructPtr.ReadOneByteFromInput());
				if status==0{
					logger.Println("Recieved 0x8001 Packet with status=",status)
					//read expression
					//len to read=packet data length - fixed length till 8001 - 12
					lentoread:=TcpUtilsStructPtr.InputPacketLength-26-12;
					expArr := make ([]byte,lentoread)
					for i:=0;i<lentoread;i++{
						expArr[i]=TcpUtilsStructPtr.ReadOneByteFromInput();
					}
					sz := bytes.IndexByte(expArr, 0)
					if sz==-1{
						sz=lentoread
					}
					logger.Println("expARR=",expArr," sz=",sz)
					exp := string(expArr[:sz])
					logger.Println("SCUID=",scuid)
					logger.Println("get/set=",gs)
					logger.Println("Scheduling ID=",scid)
					logger.Println("PWM=",pwm)
					exp=strings.Trim(exp," ")
					logger.Println("Expression=",exp)
					db:=dbController.Db
					stmtt, _ := db.Prepare("Select idschedule,ScheduleStartTime,ScheduleEndTime from schedule where ScheduleExpression=?")
					trows,err :=stmtt.Query(exp)
					defer trows.Close()
					logger.Println(stmtt)
					if err!=nil{
						logger.Println(err)
					}else{
						var sst,set,val string
						logger.Println(trows)
						if trows.Next(){
							logger.Println("inside!!!")
							trows.Scan(&val,&sst,&set)
						}
						if len(sst)==0{
							sst="NA"
						}
						if len(set)==0{
							set="NA"
						}
						logger.Println("val=",val,"start-time=",sst," end-time=",set)
						ttrows,err :=db.Query("Select Timestamp from scuconfigure where ScheduleID='"+val+"' and ScuID='"+strconv.FormatUint(scuid,10)+"' and SchedulingID='"+strconv.Itoa(scid)+"' and PWM='"+strconv.Itoa(pwm)+"' and ScheduleExpression='"+exp+"'")
						defer ttrows.Close()
						if err!=nil{
							logger.Println(err)
						}
						if ttrows.Next(){
							logger.Println("Already in DB!!")
						}else{
							stmt, _ := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
							_, eorr:=stmt.Exec(val,scuid,scid,pwm,sst,set,exp)
							defer stmt.Close()
							if eorr!=nil{
								logger.Println(eorr)
							}
						}
					}
				}else{
					logger.Println("Recieved 0x8001 Packet with invalid status=",status)
				}
				break;
			}
		//READY FOR OTA SCU
		case 0x0032:{
			logger.Println("Received packet type 0x0032 successfully");
			scuid :=TcpUtilsStructPtr.ReadEightBytesFromInput();
			status :=TcpUtilsStructPtr.ReadOneByteFromInput();
			if status==0x43{
				Scu_Is_updating[scuid]=true
				if Scu_Current_pos[scuid]!=0{
					logger.Println("SCU=",scuid," ready for OTA!! Continuing from Bucket=",Scu_Current_pos[scuid])
					Scu_RetryFirmware[scuid][Scu_Current_pos[scuid]-1]=0
					Scu_RetryFirmware[scuid][Scu_Current_pos[scuid]]=1
					logger.Println("val=",Scu_RetryFirmware[scuid][Scu_Current_pos[scuid]]," prev=",Scu_Prev_temp_arr[scuid])
					go TcpUtilsStructPtr.scu_sendFirmwarePacketWithRetry(Scu_Prev_temp_arr[scuid],Scu_Current_pos[scuid],Scu_Prev_status[scuid],scuid)
				}else{
					logger.Println("SCU=",scuid," ready for OTA!! Resetting Current pointer.")
					tempArray := make ([]byte,1028)

					Scu_RetryFirmware[scuid]=make(map[int64]int64)
					Scu_Prev_temp_arr[scuid]=make([]byte,1028)
					Scu_Current_pos[scuid]=0
					Scu_Firmware_seq[scuid]=0
					comp:=^Scu_Firmware_seq[scuid]
					tempArray[0]=byte(Scu_Firmware_seq[scuid])
					tempArray[1]=byte(comp)
					cn:=2
					na:=[]byte(Scu_firmware_name)
					for _,v:=range na{
						tempArray[cn]=v
						cn++
					}
					sz:=[]byte(Scu_firmware_size)
					for _,v:=range sz{
						tempArray[cn]=v
						cn++
					}
					tempArray[cn]=Scu_firmware_major
					cn++
					tempArray[cn]=Scu_firmware_minor
					cn++
					crc:=(Crc16(tempArray[2:1026]))
					logger.Println("CRC=",crc)
					//crc
					tempArray[1027]=byte(crc&0xff)
					tempArray[1026]=byte(crc>>8)
					Scu_Prev_temp_arr[scuid]=tempArray
					Scu_Prev_status[scuid]=0x02
					TcpUtilsStructPtr.SendResponsePacket(0x1025,scuid,0x02,tempArray,1028)
				}
			}else{
				logger.Println("Invalid OTA ready status=",status)
			}
			break;
		}
		case 0x0035:{
			logger.Println("Received packet type 0x0035 successfully");
			if TcpUtilsStructPtr.Is_TCP_Connected==false{
				logger.Println("Previous Connection Lost!!");
			}else{
				scuid :=TcpUtilsStructPtr.ReadEightBytesFromInput();
				status :=TcpUtilsStructPtr.ReadOneByteFromInput();
				if status==0x07{
					logger.Println("Received IACK for SCU=",scuid);
					if Scu_Current_pos[scuid] != 0 {
						Scu_RetryFirmware[scuid][Scu_Current_pos[scuid] - 1] = 0
					}
				}else if status==0x06{
					if Scu_Current_pos[scuid]!=0{
						Scu_RetryFirmware[scuid][Scu_Current_pos[scuid]-1]=0
					}

					Scu_RetryFirmware[scuid][Scu_Current_pos[scuid]]=1

					logger.Println("SCU=",scuid," ACK Received")
					tempArray := make ([]byte,1028)
					Scu_Firmware_seq[scuid]=(Scu_Firmware_seq[scuid]+1)%256
					comp:=^Scu_Firmware_seq[scuid]
					tempArray[0]=byte(Scu_Firmware_seq[scuid])
					tempArray[1]=byte(comp)
					cn:=2
					//logger.Println("Scu_firmware_bucket",Scu_firmware_bucket,"Scu_Current_pos[scuid]",Scu_Current_pos[scuid])
					if Scu_firmware_bucket!=Scu_Current_pos[scuid] {
						na := Scu_firmware[int64(Scu_Current_pos[scuid])]
						for _, v := range na {
							tempArray[cn] = v
							cn++
						}
					}
					crc:=(Crc16(tempArray[2:1026]))
					logger.Println("CRC=",crc)
					//crc
					tempArray[1027]=byte(crc&0xff)
					tempArray[1026]=byte(crc>>8)
					logger.Println("SCU=",scuid," Now Sending Bucket=",Scu_Current_pos[scuid]," of size=",len(Scu_firmware[int64(Scu_Current_pos[scuid])]))
					if Scu_firmware_bucket==Scu_Current_pos[scuid]{
						logger.Println("ACK Received for Last Packet, Update Completed")
						Scu_Is_updating[scuid]=false
						stmt,_:=dbController.Db.Prepare("update scu set status=?,major=?,minor=? where scu_id='"+strconv.FormatUint(scuid,10)+"'")
						_,eorr:=stmt.Exec("Completed",Scu_firmware_major,Scu_firmware_minor)
						defer stmt.Close()
						if eorr!=nil{
							logger.Println(eorr)
						}
						Scu_Prev_temp_arr[scuid]=tempArray
						Scu_Prev_status[scuid]=0x04
						logger.Println("Sending EOT Packet!!")
						go TcpUtilsStructPtr.scu_sendFirmwarePacketWithRetry(tempArray,Scu_Current_pos[scuid],0x04,scuid)
						//TcpUtilsStructPtr.SendResponsePacket(0x1022,0,0x04,tempArray,1028)
					} else{
						Scu_Prev_temp_arr[scuid]=tempArray
						Scu_Prev_status[scuid]=0x02
						go TcpUtilsStructPtr.scu_sendFirmwarePacketWithRetry(tempArray,Scu_Current_pos[scuid],0x02,scuid)
						//TcpUtilsStructPtr.SendResponsePacket(0x1022,0,0x02,tempArray,1028)
						Scu_Current_pos[scuid]++
					}

				}else if status==0x15{
					logger.Println("SCU=",scuid," NACK Received")
					//logger.Println("Resending Packet!!")
					/*if TcpUtilsStructPtr.Current_pos!=0{
						TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos-1]=0
					}
					TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos]=1
					go TcpUtilsStructPtr.sendFirmwarePacketWithRetry(TcpUtilsStructPtr.Prev_temp_arr,TcpUtilsStructPtr.Current_pos,TcpUtilsStructPtr.Prev_status)
		*/


				}else if status==0x18{
					logger.Println("SCU=",scuid," Large Size Received")
					Scu_Is_updating[scuid]=false
					stmt,_:=dbController.Db.Prepare("update scu set status=? where scu_id='"+strconv.FormatUint(scuid,10)+"'")
					_,eorr:=stmt.Exec("Error")
					defer stmt.Close()
					if eorr!=nil{
						logger.Println(eorr)
					}
				}
			}

			break;
		}
		//READY FOR OTA
		case 0x0038:{
			logger.Println("Received packet type 0x0038 successfully");
			status :=TcpUtilsStructPtr.ReadOneByteFromInput();
			if status==0x43{
				TcpUtilsStructPtr.Is_updating=true
				logger.Println("SGU=",TcpUtilsStructPtr.SGUID," ready for OTA!! Resetting Current pointer.")
				tempArray := make ([]byte,1028)

				TcpUtilsStructPtr.RetryFirmware=make(map[int64]int64)

				TcpUtilsStructPtr.Current_pos=0
				TcpUtilsStructPtr.Firmware_seq=0
				comp:=^TcpUtilsStructPtr.Firmware_seq
				tempArray[0]=byte(TcpUtilsStructPtr.Firmware_seq)
				tempArray[1]=byte(comp)
				cn:=2
				na:=[]byte(Sgu_firmware_name)
				for _,v:=range na{
					tempArray[cn]=v
					cn++
				}
				sz:=[]byte(Sgu_firmware_size)
				for _,v:=range sz{
					tempArray[cn]=v
					cn++
				}
				tempArray[cn]=Sgu_firmware_major
				cn++
				tempArray[cn]=Sgu_firmware_minor
				cn++
				crc:=(Crc16(tempArray[2:1026]))
				logger.Println("CRC=",crc)
				//crc
				tempArray[1027]=byte(crc&0xff)
				tempArray[1026]=byte(crc>>8)
				TcpUtilsStructPtr.Prev_temp_arr=tempArray
				TcpUtilsStructPtr.Prev_status=0x02
				TcpUtilsStructPtr.SendResponsePacket(0x1022,0,0x02,tempArray,1028)
			}else{
				logger.Println("Invalid OTA ready status=",status)
			}
			break;
		}
		case 0x0034:{
			logger.Println("Received packet type 0x0034 successfully");
			status :=TcpUtilsStructPtr.ReadOneByteFromInput();
			if status==0x06{
				if TcpUtilsStructPtr.Current_pos!=0{
					TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos-1]=0
				}

				TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos]=1

				logger.Println("SGU=",TcpUtilsStructPtr.SGUID," ACK Received")
				tempArray := make ([]byte,1028)
				TcpUtilsStructPtr.Firmware_seq=(TcpUtilsStructPtr.Firmware_seq+1)%256
				comp:=^TcpUtilsStructPtr.Firmware_seq
				tempArray[0]=byte(TcpUtilsStructPtr.Firmware_seq)
				tempArray[1]=byte(comp)
				cn:=2
				if Sgu_firmware_bucket!=TcpUtilsStructPtr.Current_pos {
					na := Sgu_firmware[int64(TcpUtilsStructPtr.Current_pos)]
					for _, v := range na {
						tempArray[cn] = v
						cn++
					}
				}
				crc:=(Crc16(tempArray[2:1026]))
				logger.Println("CRC=",crc)
				//crc
				tempArray[1027]=byte(crc&0xff)
				tempArray[1026]=byte(crc>>8)
				logger.Println("SGU=",TcpUtilsStructPtr.SGUID," Now Sending Bucket=",TcpUtilsStructPtr.Current_pos," of size=",len(Sgu_firmware[int64(TcpUtilsStructPtr.Current_pos)]))
				if Sgu_firmware_bucket==TcpUtilsStructPtr.Current_pos{
					logger.Println("ACK Received for Last Packet, Update Completed")
					TcpUtilsStructPtr.Is_updating=false
					stmt,_:=dbController.Db.Prepare("update sgu set status=?,major=?,minor=? where sgu_id='"+strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)+"'")
					_,eorr:=stmt.Exec("Completed",Sgu_firmware_major,Sgu_firmware_minor)
					defer stmt.Close()
					if eorr!=nil{
						logger.Println(eorr)
					}
					TcpUtilsStructPtr.Prev_temp_arr=tempArray
					TcpUtilsStructPtr.Prev_status=0x04
					logger.Println("Sending EOT Packet!!")
					go TcpUtilsStructPtr.sendFirmwarePacketWithRetry(tempArray,TcpUtilsStructPtr.Current_pos,0x04)
					//TcpUtilsStructPtr.SendResponsePacket(0x1022,0,0x04,tempArray,1028)
				} else{
					TcpUtilsStructPtr.Prev_temp_arr=tempArray
					TcpUtilsStructPtr.Prev_status=0x02
					go TcpUtilsStructPtr.sendFirmwarePacketWithRetry(tempArray,TcpUtilsStructPtr.Current_pos,0x02)
					//TcpUtilsStructPtr.SendResponsePacket(0x1022,0,0x02,tempArray,1028)
					TcpUtilsStructPtr.Current_pos++
				}

			}else if status==0x15{
				logger.Println("SGU=",TcpUtilsStructPtr.SGUID," NACK Received")
				//logger.Println("Resending Packet!!")
					/*if TcpUtilsStructPtr.Current_pos!=0{
						TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos-1]=0
					}
					TcpUtilsStructPtr.RetryFirmware[TcpUtilsStructPtr.Current_pos]=1
					go TcpUtilsStructPtr.sendFirmwarePacketWithRetry(TcpUtilsStructPtr.Prev_temp_arr,TcpUtilsStructPtr.Current_pos,TcpUtilsStructPtr.Prev_status)
*/


			}else if status==0x18{
				logger.Println("SGU=",TcpUtilsStructPtr.SGUID," Large Size Received")
				TcpUtilsStructPtr.Is_updating=false
				stmt,_:=dbController.Db.Prepare("update sgu set status=? where sgu_id='"+strconv.FormatUint(TcpUtilsStructPtr.SGUID,10)+"'")
				_,eorr:=stmt.Exec("Error")
				defer stmt.Close()
				if eorr!=nil{
					logger.Println(eorr)
				}
			}
			break;
		}

		default: {
			TcpUtilsStructPtr.Is_updating=false
                logger.Printf("Invalid Packet Type %d Specified\n",TcpUtilsStructPtr.InputPacketType); 
                TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
                
            
            }
        }
    }

func (TcpUtilsStructPtr	*TcpUtilsStruct) MonitorPackets(ticker	*time.Ticker) {

	for range ticker.C {

		logger.Println("Entering Socket ticker")
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		logger.Println("Leaving Socket ticker")

	}


}

func (TcpUtilsStructPtr	*TcpUtilsStruct) sendFirmwarePacketWithRetry(data []byte,pos int64, status int){
	attempt:=0
	du, _ := time.ParseDuration("30s")
	for ;attempt<13;attempt++{
		if TcpUtilsStructPtr.RetryFirmware[pos]==0||TcpUtilsStructPtr.Is_TCP_Connected==false{
			break
		}
		if attempt!=0{
			logger.Println("Retrying Attempt=",attempt," for Bucket=",pos)
		}
		TcpUtilsStructPtr.SendResponsePacket(0x1022,0,status,data,1028)
		time.Sleep(du)
	}
}

func (TcpUtilsStructPtr	*TcpUtilsStruct) scu_sendFirmwarePacketWithRetry(data []byte,pos int64, status int,scuid uint64){
	attempt:=0
	du, _ := time.ParseDuration("120s")
	for ;attempt<13;attempt++{
		if Scu_RetryFirmware[scuid][pos]==0{
			break
		}
		if TcpUtilsStructPtr.Is_TCP_Connected==false{
			logger.Println("Connection Lost")
			break
		}
		if attempt!=0{
			logger.Println("Retrying Attempt=",attempt," for Bucket=",pos)
		}
		TcpUtilsStructPtr.SendResponsePacket(0x1025,scuid,status,data,1028)
		time.Sleep(du)
	}
}


func (TcpUtilsStructPtr	*TcpUtilsStruct) SendAlertSMS() {
	logger.Println("old state=",TcpUtilsStructPtr.AlertStateOld)
    if (TcpUtilsStructPtr.AlertState == TcpUtilsStructPtr.AlertStateOld) {
		return
	}

    temp 	:= 	TcpUtilsStructPtr.AlertState
    temp1 	:= 	TcpUtilsStructPtr.AlertStateOld

	TcpUtilsStructPtr.AlertStateOld = TcpUtilsStructPtr.AlertState
    //format string here



    //get SGU name

    SMSstring := "Deployment:%20HAVELLSWB%20"

    SMSstring += "SGU%20ID=" + strconv.FormatUint(TcpUtilsStructPtr.SGUID,16) + "%20"


    tArray := make([]byte, 8)

    binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampHi)



    TimeStampString := string(tArray[:4])  + "-" + string(tArray[4:6]) +"-" + string(tArray[4:6]) + "%20"

    binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampLo)

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
            SMSstring += "EARTH%20FAULE%20AT:%20" + TimeStampString
        } else {
            SMSstring += "EARTH%20FAULE%20RESOLVED%20AT:%20" + TimeStampString
        }
    }


        logger.Printf("Detected Alert event %d\n",TcpUtilsStructPtr.AlertState)


    //SendSMSChan<-SMSstring
    db := dbController.Db
    rows,err :=db.Query("Select mobile_num from admin")
    defer rows.Close()
    if err!=nil{
        logger.Println(err.Error())
    }else{
        for rows.Next(){
            var to string
            err=rows.Scan(&to)
            if err!=nil{
                logger.Println(err.Error())
            }else{
                response, err := http.Get("http://login.smsgatewayhub.com/api/mt/SendSMS?APIKey=6c3f0e72-71f8-4ffa-94e2-84a0cc7f50b9&senderid=WEBSMS&channel=2&DCS=0&flashsms=0&number="+to+"&text="+SMSstring+"&route=1")
                if err != nil {
                    logger.Printf("%s\n", err)

                } else {
                    defer response.Body.Close()
                    contents, err := ioutil.ReadAll(response.Body)
                    if err != nil {
                        logger.Printf("%s\n", err)
                    }else{
                        logger.Printf("%s\n", string(contents))
                    }

                }
            }

        }

    }

    //SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld = SguUtilsStructPtr.SguTcpUtilsStruct.AlertState



}

func Sgu_firmware_init(firmware map[int64][]byte,firmware_size string,firmware_major byte,firmware_minor byte,firmware_name string,firmware_bucket int64) {
	Sgu_firmware = make(map[int64][]byte)
	Sgu_firmware = firmware
	Sgu_firmware_size = firmware_size
	Sgu_firmware_major = firmware_major
	Sgu_firmware_minor = firmware_minor
	Sgu_firmware_name = firmware_name
	Sgu_firmware_bucket = firmware_bucket
	logger.Println("SGU FIRMWARE INITIALISED!!")
}

func Scu_firmware_init(firmware map[int64][]byte,firmware_size string,firmware_major byte,firmware_minor byte,firmware_name string,firmware_bucket int64) {
	Scu_firmware = make(map[int64][]byte)
	Scu_firmware = firmware
	Scu_firmware_size = firmware_size
	Scu_firmware_major = firmware_major
	Scu_firmware_minor = firmware_minor
	Scu_firmware_name = firmware_name
	Scu_firmware_bucket = firmware_bucket
	logger.Println("SCU FIRMWARE INITIALISED!!")
}



var crc16tab = [256]uint16{
	0x0000, 0x1021, 0x2042, 0x3063, 0x4084, 0x50a5, 0x60c6, 0x70e7,
	0x8108, 0x9129, 0xa14a, 0xb16b, 0xc18c, 0xd1ad, 0xe1ce, 0xf1ef,
	0x1231, 0x0210, 0x3273, 0x2252, 0x52b5, 0x4294, 0x72f7, 0x62d6,
	0x9339, 0x8318, 0xb37b, 0xa35a, 0xd3bd, 0xc39c, 0xf3ff, 0xe3de,
	0x2462, 0x3443, 0x0420, 0x1401, 0x64e6, 0x74c7, 0x44a4, 0x5485,
	0xa56a, 0xb54b, 0x8528, 0x9509, 0xe5ee, 0xf5cf, 0xc5ac, 0xd58d,
	0x3653, 0x2672, 0x1611, 0x0630, 0x76d7, 0x66f6, 0x5695, 0x46b4,
	0xb75b, 0xa77a, 0x9719, 0x8738, 0xf7df, 0xe7fe, 0xd79d, 0xc7bc,
	0x48c4, 0x58e5, 0x6886, 0x78a7, 0x0840, 0x1861, 0x2802, 0x3823,
	0xc9cc, 0xd9ed, 0xe98e, 0xf9af, 0x8948, 0x9969, 0xa90a, 0xb92b,
	0x5af5, 0x4ad4, 0x7ab7, 0x6a96, 0x1a71, 0x0a50, 0x3a33, 0x2a12,
	0xdbfd, 0xcbdc, 0xfbbf, 0xeb9e, 0x9b79, 0x8b58, 0xbb3b, 0xab1a,
	0x6ca6, 0x7c87, 0x4ce4, 0x5cc5, 0x2c22, 0x3c03, 0x0c60, 0x1c41,
	0xedae, 0xfd8f, 0xcdec, 0xddcd, 0xad2a, 0xbd0b, 0x8d68, 0x9d49,
	0x7e97, 0x6eb6, 0x5ed5, 0x4ef4, 0x3e13, 0x2e32, 0x1e51, 0x0e70,
	0xff9f, 0xefbe, 0xdfdd, 0xcffc, 0xbf1b, 0xaf3a, 0x9f59, 0x8f78,
	0x9188, 0x81a9, 0xb1ca, 0xa1eb, 0xd10c, 0xc12d, 0xf14e, 0xe16f,
	0x1080, 0x00a1, 0x30c2, 0x20e3, 0x5004, 0x4025, 0x7046, 0x6067,
	0x83b9, 0x9398, 0xa3fb, 0xb3da, 0xc33d, 0xd31c, 0xe37f, 0xf35e,
	0x02b1, 0x1290, 0x22f3, 0x32d2, 0x4235, 0x5214, 0x6277, 0x7256,
	0xb5ea, 0xa5cb, 0x95a8, 0x8589, 0xf56e, 0xe54f, 0xd52c, 0xc50d,
	0x34e2, 0x24c3, 0x14a0, 0x0481, 0x7466, 0x6447, 0x5424, 0x4405,
	0xa7db, 0xb7fa, 0x8799, 0x97b8, 0xe75f, 0xf77e, 0xc71d, 0xd73c,
	0x26d3, 0x36f2, 0x0691, 0x16b0, 0x6657, 0x7676, 0x4615, 0x5634,
	0xd94c, 0xc96d, 0xf90e, 0xe92f, 0x99c8, 0x89e9, 0xb98a, 0xa9ab,
	0x5844, 0x4865, 0x7806, 0x6827, 0x18c0, 0x08e1, 0x3882, 0x28a3,
	0xcb7d, 0xdb5c, 0xeb3f, 0xfb1e, 0x8bf9, 0x9bd8, 0xabbb, 0xbb9a,
	0x4a75, 0x5a54, 0x6a37, 0x7a16, 0x0af1, 0x1ad0, 0x2ab3, 0x3a92,
	0xfd2e, 0xed0f, 0xdd6c, 0xcd4d, 0xbdaa, 0xad8b, 0x9de8, 0x8dc9,
	0x7c26, 0x6c07, 0x5c64, 0x4c45, 0x3ca2, 0x2c83, 0x1ce0, 0x0cc1,
	0xef1f, 0xff3e, 0xcf5d, 0xdf7c, 0xaf9b, 0xbfba, 0x8fd9, 0x9ff8,
	0x6e17, 0x7e36, 0x4e55, 0x5e74, 0x2e93, 0x3eb2, 0x0ed1, 0x1ef0}

func Crc16(bs []byte) (crc uint16) {
	l := len(bs)
	for i := 0; i < l; i++ {
		crc = ((crc << 8) & 0xff00) ^ crc16tab[((crc>>8)&0xff)^uint16(bs[i])]
	}

	return
}

func SetTempStatus(scu string, NewStatus string){

status.Lock()
status.lbuffer[scu] = NewStatus
status.Unlock()
}

func GetTempStatus(scu string) string{

status.RLock()
ans := status.lbuffer[scu]
status.RUnlock()
return ans
}

func SyncFromDB () bool {

db := dbController.Db
	dbController.DbSemaphore.Lock()
	defer dbController.DbSemaphore.Unlock()
	rows ,err := db.Query("SELECT scu_id, status from scu_status")
	defer rows.Close()
	if err != nil{
		logger.Println("error while sync status from DB: ",err)
		return false
	}else{
		var scu, state string
		status.Lock()
		for rows.Next(){
			rows.Scan(&scu,&state)
			status.lbuffer[scu] = state
		}
		status.Unlock()
	}
	return true
}
