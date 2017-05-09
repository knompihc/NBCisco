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
	"bufio"
	"dbUtils"
	"encoding/binary"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"
)

const (
	StartingDelimeter    = 0x7E
	MaxInOutBufferLength = 1024 * 8
	//MaxInOutBufferLength 	= 1000000
	FixedPacketLength = 29
)

type TcpUtilsStruct struct {
	err       error
	tcpClient net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer

	responseLineBuff []byte
	commandLineBuff  []byte
	NumOfSCUs        int
	NumOfSCUsInDB    int

	InputPacketLength     int
	OutputPacketLength    int
	SGUID                 uint64
	ControlSGUID          uint64
	TimeStampHi           uint64
	TimeStampLo           uint64
	InputSeqNumber        int
	OutputSeqNumber       int
	InputPacketType       int
	OutputPacketType      int
	SGULatitude           int
	SGULongitude          int
	SCUIDArray            []uint64
	SCUIDinDBArray        []uint64
	LampStatusArray       []uint64
	SGUZigbeeID           uint64
	SCUAnalogP1StateArray []int
	ResponseReceivedArray []int

	inputBufferDipstick int
	inputBufferReadPtr  int
	inputBufferWritePtr int

	outputBufferDipstick int
	outputBufferReadPtr  int
	outputBufferWritePtr int

	InputSyncSearchStatus int

	ConnectedToSGU     bool
	SCUListreceived    bool
	InputPacketcounter int

	MAXNumOFSCUS int

	deviceId              int64
	Length                int64
	Query                 string
	set                   int
	ResponseReceivedCount int
	AlertState            int
	AlertStateOld         int
	LampStatusCount       int

	Enable        int
	PollingRate   int
	ResponseRate  int
	DeviceTimeout int
	SlaveId       int
	SlaveIds      int

	//for retry
	RetryHash    map[string]int
	RetryHashSCU map[string]int64
}

var dbController dbUtils.DbUtilsStruct
var logger *log.Logger

func Init(dbcon dbUtils.DbUtilsStruct, logg *log.Logger) {
	dbController = dbcon
	logger = logg
}

//For LOCAL TESTING
func (TcpUtilsStructPtr *TcpUtilsStruct) ConnectToSGU() bool {
	//open connection
	//TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp","192.168.1.1:62000")
	TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp", "54.185.172.55:62002")

	if TcpUtilsStructPtr.err != nil {
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

func (TcpUtilsStructPtr *TcpUtilsStruct) CloseTcpClient() {
	err := TcpUtilsStructPtr.tcpClient.Close()
	if err != nil {
		logger.Println("Error closing TCP client")
		logger.Println(err)
	}

	TcpUtilsStructPtr.reader = nil
	TcpUtilsStructPtr.writer = nil
	TcpUtilsStructPtr.responseLineBuff = nil
	TcpUtilsStructPtr.commandLineBuff = nil
	TcpUtilsStructPtr.SCUIDArray = nil
	TcpUtilsStructPtr.SCUIDinDBArray = nil
	TcpUtilsStructPtr.ResponseReceivedArray = nil
	TcpUtilsStructPtr.LampStatusArray = nil
	TcpUtilsStructPtr.ConnectedToSGU = false

}

func (TcpUtilsStructPtr *TcpUtilsStruct) AddTcpClientToSGU(newTcpClient net.Conn, MAXNumOFSCUS int) {

	TcpUtilsStructPtr.MAXNumOFSCUS = MAXNumOFSCUS
	TcpUtilsStructPtr.tcpClient = newTcpClient
	TcpUtilsStructPtr.reader = bufio.NewReader(TcpUtilsStructPtr.tcpClient)
	TcpUtilsStructPtr.writer = bufio.NewWriter(TcpUtilsStructPtr.tcpClient)

	TcpUtilsStructPtr.responseLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.commandLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.SCUIDArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.SCUIDinDBArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.ResponseReceivedArray = make([]int, MAXNumOFSCUS)
	TcpUtilsStructPtr.LampStatusArray = make([]uint64, MAXNumOFSCUS)

	TcpUtilsStructPtr.SCUAnalogP1StateArray = make([]int, MAXNumOFSCUS)

	TcpUtilsStructPtr.RetryHash = make(map[string]int)
	TcpUtilsStructPtr.RetryHashSCU = make(map[string]int64)
	TcpUtilsStructPtr.ConnectedToSGU = true
	TcpUtilsStructPtr.InputPacketcounter = 0
	TcpUtilsStructPtr.LampStatusCount = 0
	TcpUtilsStructPtr.SGUID = 0
	TcpUtilsStructPtr.SCUListreceived = false

}

func (TcpUtilsStructPtr *TcpUtilsStruct) RewindInputBuffer(nBytes int) {

	//if (TcpUtilsStructPtr.inputBufferDipstick < nBytes){
	//	logger.Printf("Rewinding %d bytes when only %d bytes in FIFO\n",nBytes,TcpUtilsStructPtr.inputBufferDipstick);
	//    return;
	//}

	TcpUtilsStructPtr.inputBufferDipstick += nBytes
	TcpUtilsStructPtr.inputBufferReadPtr -= nBytes
	TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength - 1)

}

func (TcpUtilsStructPtr *TcpUtilsStruct) AddByteToInputBuff(newByte byte) {
	if TcpUtilsStructPtr.inputBufferDipstick < MaxInOutBufferLength {
		TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferWritePtr] = newByte
		TcpUtilsStructPtr.inputBufferWritePtr++
		TcpUtilsStructPtr.inputBufferWritePtr &= (MaxInOutBufferLength - 1)
		TcpUtilsStructPtr.inputBufferDipstick++

	} else {
		//should be spinning here till thread empties buffer.
		//TBD
		logger.Println("Warning! Input Buff is full")
	}
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadOneByteFromInput() byte {
	if TcpUtilsStructPtr.inputBufferDipstick > 0 {
		var newByte = TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferReadPtr]
		//logger.Printf("%x\n",newByte)
		TcpUtilsStructPtr.inputBufferReadPtr++
		TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength - 1)
		TcpUtilsStructPtr.inputBufferDipstick--
		return newByte

	} else {
		//should be spinning here till thread fills  buffer.
		//TBD
		logger.Println("Warning! Input Buff is empty")
		return 0
	}

}

func (TcpUtilsStructPtr *TcpUtilsStruct) GetByteFromOutputBuff() byte {
	if TcpUtilsStructPtr.outputBufferDipstick > 0 {
		var newByte = TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferReadPtr]
		TcpUtilsStructPtr.outputBufferReadPtr++
		TcpUtilsStructPtr.outputBufferReadPtr &= (MaxInOutBufferLength - 1)
		TcpUtilsStructPtr.outputBufferDipstick--
		return newByte

	} else {
		//should be spinning here till thread fills  buffer.
		//TBD
		logger.Println("Warning! Output Buff is empty")
		return 0
	}

}

func (TcpUtilsStructPtr *TcpUtilsStruct) AddByteToOutputBuff(newByte byte) {
	//logger.Println("####=",TcpUtilsStructPtr.responseLineBuff)
	if TcpUtilsStructPtr.outputBufferDipstick < MaxInOutBufferLength {
		TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferWritePtr] = newByte
		TcpUtilsStructPtr.outputBufferWritePtr++
		TcpUtilsStructPtr.outputBufferWritePtr &= (MaxInOutBufferLength - 1)
		TcpUtilsStructPtr.outputBufferDipstick++

	} else {
		//should be spinning here till thread empties buffer.
		//TBD
		logger.Println("Warning! Output Buff is full")
	}
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadTwoBytesFromInput() int {
	var tTemp int
	tTemp = (int)(((int)(TcpUtilsStructPtr.ReadOneByteFromInput()) << 8) | ((int)(TcpUtilsStructPtr.ReadOneByteFromInput() & 0x00FF)))
	return tTemp & 0x0000FFFF
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadFourBytesFromInput() int {
	return ((TcpUtilsStructPtr.ReadTwoBytesFromInput() << 16) | (TcpUtilsStructPtr.ReadTwoBytesFromInput() & 0x00FFFF))
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadFiveBytesFromInput() uint64 {
	return ((uint64)(TcpUtilsStructPtr.ReadOneByteFromInput()) << 32) |
		((uint64)(TcpUtilsStructPtr.ReadFourBytesFromInput()) & 0x00000000FFFFFFFF)

}
func (TcpUtilsStructPtr *TcpUtilsStruct) ReadSixBytesFromInput() uint64 {
	return ((uint64)(TcpUtilsStructPtr.ReadTwoBytesFromInput()) << 32) |
		((uint64)(TcpUtilsStructPtr.ReadFourBytesFromInput()) & 0x00000000FFFFFFFF)

}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadEightBytesFromInput() uint64 {
	return (((uint64)(TcpUtilsStructPtr.ReadFourBytesFromInput()) << 32) | ((uint64)(TcpUtilsStructPtr.ReadFourBytesFromInput())))
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReadNBytesFromInput(BytesToRead int) {
	//not really reading bytes, just dumping data
	if TcpUtilsStructPtr.inputBufferDipstick >= BytesToRead {
		TcpUtilsStructPtr.inputBufferReadPtr += BytesToRead
		TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength - 1)
		TcpUtilsStructPtr.inputBufferDipstick -= BytesToRead

	} else {
		//should be speeeing here till thread fills  buffer.
		//TBD
		logger.Println("Warning! Input Buff is empty while jumping ahead")

	}
}

func (TcpUtilsStructPtr *TcpUtilsStruct) WriteTwoBytesToOutput(i int) {
	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(i >> 8))
	TcpUtilsStructPtr.AddByteToOutputBuff((byte)(i))
}

func (TcpUtilsStructPtr *TcpUtilsStruct) WriteFourBytesToOutput(i int) {
	TcpUtilsStructPtr.WriteTwoBytesToOutput(i >> 16)
	TcpUtilsStructPtr.WriteTwoBytesToOutput(i)
}

func (TcpUtilsStructPtr *TcpUtilsStruct) WriteSixBytesToOutput(i uint64) {
	TcpUtilsStructPtr.WriteTwoBytesToOutput((int)(i >> 32))
	TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i))
}

func (TcpUtilsStructPtr *TcpUtilsStruct) WriteEightBytesToOutput(i uint64) {
	TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i >> 32))
	TcpUtilsStructPtr.WriteFourBytesToOutput((int)(i))
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ReceiveSocketData() {

	if !TcpUtilsStructPtr.ConnectedToSGU {
		//logger.Println("Not Connected ! Attempting to re-connect ")
		//if (!TcpUtilsStructPtr.ConnectToSGU()){
		logger.Println("Not Connected !")
		return
		//}
	}

	var bytesAvailable int

	TcpUtilsStructPtr.tcpClient.SetReadDeadline(time.Now().Add(time.Millisecond * 500))

	_, err := TcpUtilsStructPtr.reader.Peek(1)

	if err != nil {
		return
	}

	bytesAvailable = TcpUtilsStructPtr.reader.Buffered()

	if bytesAvailable == 0 {
		//logger.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )
		return
	}
	if bytesAvailable != 0 {
		//logger.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )

	}
	for ; bytesAvailable > 0; bytesAvailable-- {
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

func (TcpUtilsStructPtr *TcpUtilsStruct) SendSocketData() {
	if !TcpUtilsStructPtr.ConnectedToSGU {
		//try conneting
		//if (!TcpUtilsStructPtr.ConnectToSGU()){
		logger.Println("Not Connected, can not send data!")
		return
		//}
	}

	//logger.Printf("Sending %d Bytes\n", TcpUtilsStructPtr.outputBufferDipstick)

	for TcpUtilsStructPtr.outputBufferDipstick > 0 {

		tByte := TcpUtilsStructPtr.GetByteFromOutputBuff()
		//logger.Printf("%x\n", tByte)

		err := TcpUtilsStructPtr.writer.WriteByte(tByte)
		if err != nil {
			logger.Println("Could not  write to socket")
			TcpUtilsStructPtr.CloseTcpClient()
			return
		}
	}
	//	if TcpUtilsStructPtr.writer.Flush() != nil {
	//		logger.Println("Error while flushing output stream")
	//		TcpUtilsStructPtr.CloseTcpClient()
	//	}
	err_flush := TcpUtilsStructPtr.writer.Flush()
	if err_flush != nil {
		logger.Println("Error while flushing output stream")
		TcpUtilsStructPtr.CloseTcpClient()
		logger.Println("flush error", err_flush)
	}
	defer func() {
		if rec := recover(); rec != nil {
			logger.Println("recovered in SendSocketData", rec)
		}

	}()

}

func (TcpUtilsStructPtr *TcpUtilsStruct) GetSCUIndexFromSCUID(SCUID uint64) int {

	var i int
	for i = 0; i < TcpUtilsStructPtr.NumOfSCUsInDB; i++ {

		if TcpUtilsStructPtr.SCUIDinDBArray[i] == SCUID {
			return i
		}
	}
	return -1

}

func (TcpUtilsStructPtr *TcpUtilsStruct) PacketTypeToPacketLength(PacketType int) int {

	switch PacketType {
	case 0x0001:
		{
			return FixedPacketLength + 12
		}

	case 0x0002:
		{
			return FixedPacketLength
		}

	case 0x0003:
		{
			return FixedPacketLength + 28
		}

	case 0x0004:
		{
			return FixedPacketLength + 9
		}

	case 0x0005:
		{
			return FixedPacketLength + 10
		}

	case 0xe000:
		{
			return FixedPacketLength + 2 + TcpUtilsStructPtr.NumOfSCUs*24

		}

	case 0x0011:
		{
			return FixedPacketLength + 1

		}
	case 0x0022:
		{
			return FixedPacketLength + 1

		}
	case 0x0023:
		{
			return FixedPacketLength + 1

		}

	case 0x0024:
		{
			return FixedPacketLength + 1

		}
	case 0x0025:
		{
			return FixedPacketLength + 1

		}

	case 0xe001:
		{
			return FixedPacketLength + 1

		}

	case 0x1000:
		{
			return FixedPacketLength + 11

		}

	case 0x1001:
		{
			return FixedPacketLength + 11

		}
	case 0x2000:
		{
			return FixedPacketLength + 8

		}
	case 0x2001:
		{
			return FixedPacketLength + 34
		}

	case 0x3000:
		{
			return FixedPacketLength + 14
		}

	case 0x3001:
		{
			return FixedPacketLength + 15
		}
	case 0x4000:
		{
			return FixedPacketLength + 8
		}

	case 0x4001:
		{
			return FixedPacketLength + 23
		}
	case 0x5000:
		{
			return FixedPacketLength + 22
		}
	case 0x5001:
		{
			return FixedPacketLength + 9
		}

	case 0x6000:
		{
			return FixedPacketLength + 8
		}
	case 0x6001:
		{
			return FixedPacketLength + 24
		}
	case 0x7000:
		{
			return FixedPacketLength + 65
		}
	case 0x7001:
		{
			return FixedPacketLength + 66
		}
	case 0x8000:
		{
			return FixedPacketLength + 11
		}
	case 0x8001:
		{
			return FixedPacketLength + 11
		}
	case 0x9000:
		{
			return FixedPacketLength + 5
		}
	case 0x9001:
		{
			return FixedPacketLength + 6
		}
	case 0xA000:
		{
			return FixedPacketLength + 13
		}
	case 0xA001:
		{
			return FixedPacketLength + 14
		}
	case 0xB000:
		{
			return FixedPacketLength
		}
	case 0xD000:
		{
			return FixedPacketLength - 1
		}
	case 0xB001:
		{
			return FixedPacketLength + 4
		}
	default:
		{
			logger.Printf("Invalid Packet Type  %x Specifid", PacketType)
		}
	}
	return 0

}

func (TcpUtilsStructPtr *TcpUtilsStruct) SendResponsePacket(OutputPacketType int, SCUID uint64, StatusByte int, expression []byte, expressionLength int) {

	TcpUtilsStructPtr.OutputPacketType = OutputPacketType
	//first add the delimeter
	TcpUtilsStructPtr.AddByteToOutputBuff(StartingDelimeter)
	TcpUtilsStructPtr.OutputPacketLength = TcpUtilsStructPtr.PacketTypeToPacketLength(TcpUtilsStructPtr.OutputPacketType)

	switch OutputPacketType {

	case 0x8000:
		{
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}

	case 0x3000:
		{
			if (StatusByte & 0x00FF00) == 0 {
				//for get mode, packet is smaller
				TcpUtilsStructPtr.OutputPacketLength -= 5
			}

		}

	case 0xB000:
		{
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}
	case 0xD000:
		{
			TcpUtilsStructPtr.OutputPacketLength += expressionLength
			break
		}
	}

	TcpUtilsStructPtr.OutputPacketLength -= 3 //FixedPacketLength;
	//logger.Println("Packet Type = %d, Packet Length = %x\n", OutputPacketType, TcpUtilsStructPtr.OutputPacketLength)

	TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketLength)
	TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.SGUID)
	//TcpUtilsStructPtr.WriteEightBytesToOutput(TcpUtilsStructPtr.TimeStampHi);
	//TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.TimeStampLo);
	currentTime := time.Now().Local()
	newFormat := currentTime.Format("20060102150405")

	for k := 0; k < 14; k++ {
		TcpUtilsStructPtr.AddByteToOutputBuff((byte)(newFormat[k]))
	}

	TcpUtilsStructPtr.WriteFourBytesToOutput(TcpUtilsStructPtr.OutputSeqNumber)
	TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketType)

	//done with common part.
	switch TcpUtilsStructPtr.OutputPacketType {

	case 0x0011:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0x0022:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0x0023:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0x0024:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0x0025:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0xe001:
		{
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte))
			break
		}
	case 0x1000:
		{
			break
		}

	case 0x2000:
		{
			break
		}
	case 0x3000:
		{
			//separate LampId and LampVal;
			lampVal := StatusByte & 0x01
			getSetByte := (StatusByte >> 8) & 0x0FF
			TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID)
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(getSetByte))
			//TcpUtilsStructPtr.AddByteToOutputBuff((byte)(0x09));
			if getSetByte == 1 {
				//for set need to set additional fields
				if lampVal == 0 {
					TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal))
				} else {
					x := 0x09
					TcpUtilsStructPtr.AddByteToOutputBuff((byte)(x))
				}

				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal))
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal))
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal))
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal))
			}

			break

		}

	case 0x4000:
		{
			break
		}

	case 0x5000:
		{
			TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID)
			for k := 0; k < 14; k++ {
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(newFormat[k]))
			}

			break

		}

	case 0x8000:
		{
			TcpUtilsStructPtr.WriteEightBytesToOutput(SCUID)
			//write Get/Set which is byte0 of StatusByte
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte & 0x0FF))

			//write scheduling id  which is byte1 of StatusByte
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)((StatusByte >> 8) & 0x0FF))

			//write pwm  state  which is byte2 of StatusByte
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)((StatusByte >> 16) & 0x0FF))

			for k := 0; k < expressionLength; k++ {
				TcpUtilsStructPtr.AddByteToOutputBuff(expression[k])

			}

			break

		}

	case 0x9000:
		{
			logger.Println("Entered into 0X9000 Packet..")
			//write Get/Set which is nibble-0  of StatusByte
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte & 0x00FF))
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[0] & 0x00F)
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[1] & 0x00F)
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[2] & 0x00F)
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[3] & 0x00F)
			logger.Println("Output :", expression[0])
			logger.Println("Output :", expression[1])
			logger.Println("Output :", expression[2])
			logger.Println("Output :", expression[3])
			logger.Println("Exited from 0X9000 Packet..")
			break

		}

	case 0xA000:
		{
			logger.Println("Entered into 0XA000 Packet..")
			/*	TcpUtilsStructPtr.Enable=1;
				TcpUtilsStructPtr.PollingRate=600;
				TcpUtilsStructPtr.ResponseRate=600;
				TcpUtilsStructPtr.DeviceTimeout=100;
				TcpUtilsStructPtr.SlaveId=1; */

			//TcpUtilsStructPtr.SlaveIds=0xffffffff;
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte & 0x00FF))
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
			logger.Println("Exited from 0XA000 Packet..")
			break

		}

	case 0xB000:
		{
			logger.Println("Entered into 0XB000 Packet..")
			TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte & 0x00FF))
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[0])
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[1])
			for i := 0; i < expressionLength-2; i++ {
				//tmp,_:=strconv.ParseInt(TcpUtilsStruc0tPtr.Query[i:i+2],10,32)
				logger.Println("Output :", (expression[i+2]))
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[i+2]))
			}
			logger.Println("Exited from 0XB000 Packet..")
			break

		}
	case 0xD000:
		{
			logger.Println("Entered into 0XD000 Packet..")
			//TcpUtilsStructPtr.AddByteToOutputBuff((byte)(StatusByte  & 0x00FF));
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[0])
			TcpUtilsStructPtr.AddByteToOutputBuff(expression[1])
			for i := 0; i < expressionLength-2; i++ {
				//tmp,_:=strconv.ParseInt(TcpUtilsStruc0tPtr.Query[i:i+2],10,32)
				logger.Println("Output :", (expression[i+2]))
				TcpUtilsStructPtr.AddByteToOutputBuff((byte)(expression[i+2]))
			}
			logger.Println("Exited from 0XD000 Packet..")
			break

		}
	default:
		{
			logger.Printf("Invalid Output Packet Type %x Specifid\n", TcpUtilsStructPtr.OutputPacketType)
		}
	}
	if TcpUtilsStructPtr.outputBufferDipstick < (TcpUtilsStructPtr.OutputPacketLength + 3) {
		logger.Println("Output Packet Formating Error")
	}
	TcpUtilsStructPtr.SendSocketData()
}

func (TcpUtilsStructPtr *TcpUtilsStruct) ParseInputPacket() {

	//add data from socket buffer to local buffer
	TcpUtilsStructPtr.ReceiveSocketData()

	if TcpUtilsStructPtr.InputSyncSearchStatus == 0 {
		//need to search for start delimiter
		for TcpUtilsStructPtr.inputBufferDipstick > 0 {
			tByte := TcpUtilsStructPtr.ReadOneByteFromInput()
			if tByte == StartingDelimeter {
				//found sync. Confirm by parsing and looking at next sync.
				//TBD
				//rewind dipstick and read pointer
				TcpUtilsStructPtr.RewindInputBuffer(1)
				TcpUtilsStructPtr.InputSyncSearchStatus = 1
				break
			}
		}
		//could not sync, so just return
		if TcpUtilsStructPtr.InputSyncSearchStatus == 0 {
			return
		}
	}

	TcpUtilsStructPtr.ReceiveSocketData()
	//here dipstick has to be minimum, else we can
	//not parse fixed header

	if TcpUtilsStructPtr.inputBufferDipstick < (FixedPacketLength) {
		return

	}

	//confirm start delimiter
	if TcpUtilsStructPtr.ReadOneByteFromInput() != StartingDelimeter {
		logger.Println("Failed to  match start delimiter")
		TcpUtilsStructPtr.InputSyncSearchStatus = 0
		return
	}

	TcpUtilsStructPtr.InputPacketLength = TcpUtilsStructPtr.ReadTwoBytesFromInput()
	if TcpUtilsStructPtr.InputPacketLength > 0x8000 {
		logger.Printf("Invalid Packet Length = %x   \n", TcpUtilsStructPtr.InputPacketLength)

	} else {
		//System.out.printf("Packet Length = %x   \n",InputPacketLength);
		logger.Println("Packet Length = %x   \n", TcpUtilsStructPtr.InputPacketLength)
	}

	TcpUtilsStructPtr.ReceiveSocketData()

	//make sure entire packet is in buffer
	if TcpUtilsStructPtr.inputBufferDipstick < (TcpUtilsStructPtr.InputPacketLength) {
		//insufficient data in buuer
		//rewind pointers and return
		//need to rewind by 3 bytes
		TcpUtilsStructPtr.RewindInputBuffer(3)
		return
	}

	//get 8 bytes of SGU id
	TcpUtilsStructPtr.SGUID = TcpUtilsStructPtr.ReadSixBytesFromInput()

	logger.Printf("SGU ID  %d \n", TcpUtilsStructPtr.SGUID)

	//TimeStampString  := make([]byte,14)

	//get first 8 bytes of timestamp
	TcpUtilsStructPtr.TimeStampHi = TcpUtilsStructPtr.ReadEightBytesFromInput()

	//get remaining 6 bytes of timestamp
	TcpUtilsStructPtr.TimeStampLo = TcpUtilsStructPtr.ReadSixBytesFromInput()

	//get 4 bytes of input sequence number
	TcpUtilsStructPtr.InputSeqNumber = TcpUtilsStructPtr.ReadFourBytesFromInput()

	TcpUtilsStructPtr.InputPacketType = TcpUtilsStructPtr.ReadTwoBytesFromInput()

	logger.Printf("Received packet type %d \n", TcpUtilsStructPtr.InputPacketType)

	//TimeStampString[0:7] = TcpUtilsStructPtr.TimeStampHi[0:7]

	tArray := make([]byte, 8)

	binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampHi)
	TimeStampString := string(tArray[:8])

	binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampLo)
	TimeStampString += string(tArray[:6])

	logger.Println(TimeStampString)

	switch TcpUtilsStructPtr.InputPacketType {
	case 0x0001:
		{ //Reset Indication
			//send packet of type 0x11
			//read 12 bytes from buffer and junk them
			TcpUtilsStructPtr.ReadNBytesFromInput(12)
			TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber
			logger.Println("Received packet type 0x0001 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0x11, 0, 0, nil, 0)
			break
		}
	case 0x0002:
		{ //Keep Alive
			TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber
			logger.Println("Received packet type 0x0002 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0x22, 0, 0, nil, 0)
			break
		}
	case 0x0003:
		{ //SCU Lit
			//parse packet
			//TBD
			TcpUtilsStructPtr.NumOfSCUs = TcpUtilsStructPtr.ReadTwoBytesFromInput()

			logger.Printf("Found  %d SCUs in list\n", TcpUtilsStructPtr.NumOfSCUs-1)

			if TcpUtilsStructPtr.NumOfSCUs > TcpUtilsStructPtr.MAXNumOFSCUS {

				logger.Printf("Max num of SCUs exceeded. Received %d\n", TcpUtilsStructPtr.NumOfSCUs)
				TcpUtilsStructPtr.NumOfSCUs = TcpUtilsStructPtr.MAXNumOFSCUS

			}
			//first ID is zigbeed Id of SGU itself
			TcpUtilsStructPtr.SGUZigbeeID = TcpUtilsStructPtr.ReadEightBytesFromInput()
			//read and dump reserved byte
			TcpUtilsStructPtr.ReadNBytesFromInput(1)

			TcpUtilsStructPtr.NumOfSCUs--

			for i := 0; i < TcpUtilsStructPtr.NumOfSCUs; i++ {
				TcpUtilsStructPtr.SCUIDArray[i] = TcpUtilsStructPtr.ReadEightBytesFromInput()
				//read and dump reserved byte
				TcpUtilsStructPtr.ReadOneByteFromInput()
			}

			TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber
			logger.Println("Received packet type 0x0003 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0x23, 0, 0, nil, 0)
			TcpUtilsStructPtr.SCUListreceived = true
			break
		}
	case 0x0004:
		{ //SCU Deleted
			//parse packet
			//just dump 9 bytes
			TcpUtilsStructPtr.ReadNBytesFromInput(9)
			TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber
			logger.Println("Received packet type 0x0004 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0x24, 0, 0, nil, 0)
			break
		}
	case 0x0005:
		{ //SCU Added
			//parse packet
			TcpUtilsStructPtr.ReadNBytesFromInput(8)
			logger.Println("Received packet type 0x0005 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0x25, 0, 0, nil, 0)
			break
		}
	case 0xe000:
		{ //Input Status
			//parse packet
			//logger.Println(TcpUtilsStructPtr.responseLineBuff)
			NumSCUPlusSGU := TcpUtilsStructPtr.ReadTwoBytesFromInput()
			TcpUtilsStructPtr.ControlSGUID = TcpUtilsStructPtr.ReadEightBytesFromInput()

			sguSTATUS := TcpUtilsStructPtr.ReadOneByteFromInput()
			DigitalInput1 := TcpUtilsStructPtr.ReadOneByteFromInput()
			DigitalInput2 := TcpUtilsStructPtr.ReadOneByteFromInput()
			DigitalInput3 := TcpUtilsStructPtr.ReadOneByteFromInput()
			TcpUtilsStructPtr.AlertState = 0

			if sguSTATUS == 0 {
				if DigitalInput1 != 1 {
					logger.Println("DigitalInput1 tripped")
					TcpUtilsStructPtr.AlertState = 1
					// go TcpUtilsStructPtr.SendAlertSMS()
				}
				if DigitalInput2 != 1 {
					logger.Println("DigitalInput2 tripped")
					TcpUtilsStructPtr.AlertState |= 2
					// go TcpUtilsStructPtr.SendAlertSMS()
				}
				if DigitalInput3 != 1 {
					logger.Println("DigitalInput3 tripped")
					TcpUtilsStructPtr.AlertState |= 4
				}
				//go TcpUtilsStructPtr.SendAlertSMS()
			}

			TcpUtilsStructPtr.ReadNBytesFromInput(12)

			logger.Printf("Found status info for %d SCUs\n", NumSCUPlusSGU-1)

			//read the SGU

			tempCounter := TcpUtilsStructPtr.LampStatusCount

			for i := 0; i < NumSCUPlusSGU-1; i++ {
				SCUID := TcpUtilsStructPtr.ReadEightBytesFromInput()
				scuIndex := TcpUtilsStructPtr.GetSCUIndexFromSCUID(SCUID)
				//dump next 5 bytes as they are not used for now

				scuStatus := TcpUtilsStructPtr.ReadOneByteFromInput()

				//dump next 4 bytes
				TcpUtilsStructPtr.ReadNBytesFromInput(4)

				tempAnalog := TcpUtilsStructPtr.ReadFourBytesFromInput()

				tempDigital := TcpUtilsStructPtr.ReadFiveBytesFromInput()

				logger.Println("Before Received packet type 0xe000 successfully for scuid=", scuIndex, " with status=", (tempDigital & (0x0FF)))
				//dume next 2 bytes
				TcpUtilsStructPtr.ReadNBytesFromInput(2)

				if scuStatus == 0 {

					if scuIndex >= 0 {
						TcpUtilsStructPtr.SCUAnalogP1StateArray[tempCounter] = tempAnalog
						tempDigital = tempDigital | (((uint64)(scuIndex)) << 40)
						TcpUtilsStructPtr.LampStatusArray[tempCounter] = tempDigital
						tempCounter++
					} else {
						logger.Println("Unindentified SCU specified")

					}

				}
				logger.Println("Received packet type 0xe000 successfully for scuid=", scuIndex, " with status=", (tempDigital & (0x0FF)))

			}
			TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber
			TcpUtilsStructPtr.LampStatusCount = tempCounter
			logger.Println("Received packet type 0xe000 successfully")
			TcpUtilsStructPtr.SendResponsePacket(0xe001, 0, 0, nil, 0)
			break
		}

	//response from SGU for queries

	case 0x1001:
		{ //Get SGU details
			//parse packet

			TcpUtilsStructPtr.ReadNBytesFromInput(3)
			TcpUtilsStructPtr.SGULatitude = TcpUtilsStructPtr.ReadTwoBytesFromInput()
			TcpUtilsStructPtr.SGULongitude = TcpUtilsStructPtr.ReadTwoBytesFromInput()
			TcpUtilsStructPtr.ReadNBytesFromInput(3)
			logger.Printf("Received sgu coordinates: %f\n", TcpUtilsStructPtr.SGULatitude)
			logger.Printf("Received sgu coordinates: %f\n", TcpUtilsStructPtr.SGULongitude)
			logger.Println("Received packet type 0x1001 successfully")
			break
		}

	case 0x2001:
		{ //Get SCU details
			//parse packet
			TcpUtilsStructPtr.ReadNBytesFromInput(34)
			logger.Println("Received packet type 0x2001 successfully")
			break
		}

	case 0x3001:
		{ //Get/Set Digital Output State
			//parse packet

			//get status
			status := TcpUtilsStructPtr.ReadOneByteFromInput()

			SCUID := TcpUtilsStructPtr.ReadEightBytesFromInput()
			scuIndex := TcpUtilsStructPtr.GetSCUIndexFromSCUID(SCUID)
			gs := TcpUtilsStructPtr.ReadOneByteFromInput()

			if scuIndex >= 0 {
				tempDigital := TcpUtilsStructPtr.ReadFiveBytesFromInput()
				//for retry
				strHash := strconv.Itoa(0x3000) + "#" + strconv.FormatUint(SCUID, 10) + "#" + strconv.Itoa(int(gs)) + "#" + strconv.FormatUint((tempDigital>>32)&0x0FF, 10)
				logger.Println("Hash received=", strHash)
				if TcpUtilsStructPtr.RetryHash[strHash] == 1 {
					TcpUtilsStructPtr.RetryHash[strHash] = 2
				}
				if TcpUtilsStructPtr.ResponseReceivedCount < TcpUtilsStructPtr.MAXNumOFSCUS {
					if status == 0 {
						TcpUtilsStructPtr.ResponseReceivedArray[TcpUtilsStructPtr.ResponseReceivedCount] = ((scuIndex << 16) & (0xFFFF0000)) | ((int)(tempDigital & 0x01))
						tempDigital = tempDigital | (((uint64)(scuIndex)) << 40)
						TcpUtilsStructPtr.LampStatusArray[TcpUtilsStructPtr.LampStatusCount] = tempDigital

					} else {

						TcpUtilsStructPtr.ResponseReceivedArray[TcpUtilsStructPtr.ResponseReceivedCount] = ((scuIndex << 16) & (0xFFFF0000)) | 2
						tempDigital = 2 | (((uint64)(scuIndex)) << 40)
						TcpUtilsStructPtr.LampStatusArray[TcpUtilsStructPtr.LampStatusCount] = tempDigital

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

			logger.Printf("Received packet type 0x3001 successfully. status = %2.2x\n", status)
			break
		}
	case 0x4001:
		{ //Get Time Stamp
			//parse packet
			TcpUtilsStructPtr.ReadNBytesFromInput(23)
			logger.Println("Received packet type 0x4001 successfully")
			break
		}
	case 0x5001:
		{ //Set Time Stamp
			//parse packet
			status := TcpUtilsStructPtr.ReadOneByteFromInput()
			scuID := TcpUtilsStructPtr.ReadEightBytesFromInput()
			if status == 0 {
				logger.Printf("Time set successfully for SCU %d\n", scuID)
			} else {

				logger.Printf("Error setting time for SCU %d\n", scuID)
			}

			logger.Println("Received packet type 0x5001 successfully")
			break
		}
	case 0x6001:
		{ //Get Input Status
			//parse packet
			//ReadNBytesFromInput(24);
			TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength)
			logger.Println("Received packet type 0x6001 successfully")
			break
		}
	case 0x7001:
		{ //Set Input Status
			//parse packet
			TcpUtilsStructPtr.ReadNBytesFromInput(66)
			logger.Println("Received packet type 0x7001 successfully")

			break
		}
	case 0x0006:
		{
			//skip no. of devices (1 byte)
			/*logger.Println(TcpUtilsStructPtr.commandLineBuff)
			logger.Println("len=",TcpUtilsStructPtr.InputPacketLength)*/
			//return
			//TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
			TcpUtilsStructPtr.ReadOneByteFromInput()
			//skip device id (1 byte)
			TcpUtilsStructPtr.ReadOneByteFromInput()
			//get status (1 byte)
			status := TcpUtilsStructPtr.ReadOneByteFromInput()
			//status==0 for successful response
			if status == 0 {
				//skip this length, device id and modbus type(3 bytes).
				TcpUtilsStructPtr.ReadOneByteFromInput()
				TcpUtilsStructPtr.ReadTwoBytesFromInput()
				//skip 1 byte response length
				TcpUtilsStructPtr.ReadOneByteFromInput()
				//read Watts Total 4bytes
				wa2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				wa1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				wa := (wa1 << 16) | (wa2 & 0x00FFFF)
				nn := uint32(wa)
				waf := math.Float32frombits(nn)
				kwa := waf / 1000.0
				logger.Println("KWA=", kwa)

				//skip 28 bytes to get to the Pf
				TcpUtilsStructPtr.ReadNBytesFromInput(28)

				//read Pf Total 4bytes
				pf2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				pf1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				pf := (pf1 << 16) | (pf2 & 0x00FFFF)
				pfnn := uint32(pf)
				pff := math.Float32frombits(pfnn)
				logger.Println("PF=", pff)

				//skip 12 bytes to get to the Va
				TcpUtilsStructPtr.ReadNBytesFromInput(12)

				//read Va Total 4bytes
				va2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				va1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				va := (va1 << 16) | (va2 & 0x00FFFF)
				vann := uint32(va)
				vaf := math.Float32frombits(vann)
				kva := vaf / 1000.0
				logger.Println("kva=", kva)

				//skip 32 bytes to get to the Va
				TcpUtilsStructPtr.ReadNBytesFromInput(32)

				//read Vr Total 4bytes
				vr2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vr1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vr := (vr1 << 16) | (vr2 & 0x00FFFF)
				vrnn := uint32(vr)
				vrf := math.Float32frombits(vrnn)
				logger.Println("vr=", vrf)

				//read Vy Total 4bytes
				vy2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vy1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vy := (vy1 << 16) | (vy2 & 0x00FFFF)
				vynn := uint32(vy)
				vyf := math.Float32frombits(vynn)
				logger.Println("vy=", vyf)

				//read Vb Total 4bytes
				vb2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vb1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				vb := (vb1 << 16) | (vb2 & 0x00FFFF)
				vbnn := uint32(vb)
				vbf := math.Float32frombits(vbnn)
				logger.Println("vb=", vbf)

				//skip 4 bytes to get to the Ir
				TcpUtilsStructPtr.ReadNBytesFromInput(4)

				//read Ir Total 4bytes
				ir2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				ir1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				ir := (ir1 << 16) | (ir2 & 0x00FFFF)
				irnn := uint32(ir)
				irf := math.Float32frombits(irnn)
				logger.Println("ir=", irf)

				//read Iy Total 4bytes
				iy2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				iy1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				iy := (iy1 << 16) | (iy2 & 0x00FFFF)
				iynn := uint32(iy)
				iyf := math.Float32frombits(iynn)
				logger.Println("iy=", iyf)

				//read Ib Total 4bytes
				ib2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				ib1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				ib := (ib1 << 16) | (ib2 & 0x00FFFF)
				ibnn := uint32(ib)
				ibf := math.Float32frombits(ibnn)
				logger.Println("ib=", ibf)

				//read freq Total 4bytes
				fre2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				fre1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				fre := (fre1 << 16) | (fre2 & 0x00FFFF)
				frenn := uint32(fre)
				fref := math.Float32frombits(frenn)
				logger.Println("fre=", fref)

				//read whq Total 4bytes
				wh2 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				wh1 := TcpUtilsStructPtr.ReadTwoBytesFromInput()
				wh := (wh1 << 16) | (wh2 & 0x00FFFF)
				whnn := uint32(wh)
				whf := math.Float32frombits(whnn)
				kwh := whf / 1000.0
				logger.Println("kwh=", kwh)

				logger.Println("SGUID=", TcpUtilsStructPtr.SGUID)
				dbController.DbSemaphore.Lock()
				defer dbController.DbSemaphore.Unlock()
				db := dbController.Db
				stmt, _ := db.Prepare("INSERT parameters SET sgu_id=?,KW=?,Pf=?,KVA=?,Vr=?,Vy=?,Vb=?,Ir=?,Iy=?,Ib=?,KWH=?,freq=?")
				defer stmt.Close()
				_, eorr := stmt.Exec(TcpUtilsStructPtr.SGUID, kwa, pff, kva, vrf, vyf, vbf, irf, iyf, ibf, kwh, fref)
				if eorr != nil {
					logger.Println(eorr)
				} else {
					logger.Println("Inserting 0x0006 Packet Successfully with sguid=", strconv.FormatUint(TcpUtilsStructPtr.SGUID, 10))
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
	case 0x9001:
		{ //Get Modbus details
			//parse packet
			modbusStatus9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusSet9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusBaudRate9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusStopbits9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusParity9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusNumberOfBits9001 := TcpUtilsStructPtr.ReadOneByteFromInput()
			logger.Printf("Received Modbus Status: %d\n", modbusStatus9001)
			logger.Printf("Received Modbus Set: %d\n", modbusSet9001)
			logger.Printf("Received Modbus BaudRate: %d\n", modbusBaudRate9001)
			logger.Printf("Received Modbus Stopbits: %d\n", modbusStopbits9001)
			logger.Printf("Received Modbus Parity: %d\n", modbusParity9001)
			logger.Printf("Received Modbus NumberOfBits: %d\n", modbusNumberOfBits9001)
			logger.Println("Received packet type 0x9001 successfully")
			break
		}
	case 0xA001:
		{ //Get Modbus details
			//parse packet
			modbusStatus := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusSet := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusEnabled := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusPollingRate := TcpUtilsStructPtr.ReadTwoBytesFromInput()
			modbusResponseRate := TcpUtilsStructPtr.ReadTwoBytesFromInput()
			modbusDeviceTimeout := TcpUtilsStructPtr.ReadTwoBytesFromInput()
			modbusSlaveId := TcpUtilsStructPtr.ReadOneByteFromInput()
			TcpUtilsStructPtr.ReadNBytesFromInput(4)
			logger.Printf("Received Modbus Status: %d\n", modbusStatus)
			logger.Printf("Received Modbus Set: %d\n", modbusSet)
			logger.Printf("Received Modbus Enable: %d\n", modbusEnabled)
			logger.Printf("Received Modbus PollingRate: %d\n", modbusPollingRate)
			logger.Printf("Received Modbus ResponseRate: %d\n", modbusResponseRate)
			logger.Printf("Received Modbus ResponseRate: %d\n", modbusDeviceTimeout)
			logger.Printf("Received Modbus SlaveIds: %d\n", modbusSlaveId)
			logger.Println("Received packet type 0xA001 successfully")
			break
		}
	case 0xB001:
		{ //get modbus Data
			//parse packet
			modbusStatus1 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusSet1 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusDeviceId1 := TcpUtilsStructPtr.ReadOneByteFromInput()
			modbusDataLength := TcpUtilsStructPtr.ReadOneByteFromInput()
			//x:=8
			TcpUtilsStructPtr.ReadNBytesFromInput((int)(modbusDataLength))
			logger.Printf("Received Modbus Status1: %d\n", modbusStatus1)
			logger.Printf("Received Modbus Set1: %d\n", modbusSet1)
			logger.Printf("Received Modbus DeviceId1: %d\n", modbusDeviceId1)
			logger.Printf("Received Modbus DataLength: %d\n", modbusDataLength)
			logger.Println("Received packet type 0xB001 successfully")

			break
		}
	case 0x8001:
		{
			//parse packet
			status := TcpUtilsStructPtr.ReadOneByteFromInput()
			scuid := TcpUtilsStructPtr.ReadEightBytesFromInput()
			gs := TcpUtilsStructPtr.ReadOneByteFromInput()
			scid := TcpUtilsStructPtr.ReadOneByteFromInput()
			pwm := TcpUtilsStructPtr.ReadOneByteFromInput()
			if status == 0 {
				logger.Println("Recieved 0x8001 Packet with status=", status)
				//read expression
				//len to read=packet data length - fixed length till 8001 - 12
				lentoread := TcpUtilsStructPtr.InputPacketLength - 26 - 12
				expArr := make([]byte, lentoread)
				for i := 0; i < lentoread; i++ {
					expArr[i] = TcpUtilsStructPtr.ReadOneByteFromInput()
				}
				exp := string(expArr[:])
				logger.Println("SCUID=", scuid)
				logger.Println("get/set=", gs)
				logger.Println("Scheduling ID=", scid)
				logger.Println("PWM=", pwm)
				logger.Println("Expression=", exp)
				db := dbController.Db
				trows, err := db.Query("Select idschedule,ScheduleStartTime,ScheduleEndTime from schedule where ScheduleExpression='" + exp + "'")
				defer trows.Close()
				if err != nil {
					logger.Println(err)
				} else {
					var sst, set, val string
					for trows.Next() {
						trows.Scan(&val, &sst, &set)
					}
					stmt, _ := db.Prepare("INSERT scuconfigure SET ScheduleID=?,Timestamp=NOW(),ScuID=?,SchedulingID=?,PWM=?,ScheduleStartTime=?,ScheduleEndTime=?,ScheduleExpression=?")
					_, eorr := stmt.Exec(val, scuid, scid, pwm, sst, set, exp)
					defer stmt.Close()
					if eorr != nil {
						logger.Println(eorr)
					}
				}
			} else {
				logger.Println("Recieved 0x8001 Packet with invalid status=", status)
			}
			break
		}

	default:
		{
			logger.Printf("Invalid Packet Type %d Specified\n", TcpUtilsStructPtr.InputPacketType)
			TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength)

		}
	}
}

func (TcpUtilsStructPtr *TcpUtilsStruct) MonitorPackets(ticker *time.Ticker) {

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

func (TcpUtilsStructPtr *TcpUtilsStruct) SendAlertSMS() {
	logger.Println("old state=", TcpUtilsStructPtr.AlertStateOld)
	if TcpUtilsStructPtr.AlertState == TcpUtilsStructPtr.AlertStateOld {
		return
	}

	temp := TcpUtilsStructPtr.AlertState
	temp1 := TcpUtilsStructPtr.AlertStateOld

	TcpUtilsStructPtr.AlertStateOld = TcpUtilsStructPtr.AlertState
	//format string here

	//get SGU name

	SMSstring := "Deployment:%20HAVELLSWB%20"

	SMSstring += "SGU%20ID=" + strconv.FormatUint(TcpUtilsStructPtr.SGUID, 16) + "%20"

	tArray := make([]byte, 8)

	binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampHi)

	TimeStampString := string(tArray[:4]) + "-" + string(tArray[4:6]) + "-" + string(tArray[4:6]) + "%20"

	binary.BigEndian.PutUint64(tArray, TcpUtilsStructPtr.TimeStampLo)

	TimeStampString += string(tArray[2:4]) + ":" + string(tArray[4:6]) + ":" + string(tArray[6:8]) + "%20"

	//temp is new status
	//temp1 is old status
	//find changed bits

	temp2 := temp ^ temp1

	if (temp2 & 0x01) != 0 {
		if (temp & 0x01) != 0 {
			SMSstring += "PANEL%20OPEN%20AT:%20" + TimeStampString
		} else {
			SMSstring += "PANEL%20CLOSED%20AT:%20" + TimeStampString
		}
	}
	if (temp2 & 0x02) != 0 {
		if (temp & 0x02) != 0 {
			SMSstring += "MCB%20TRIP%20AT:%20" + TimeStampString
		} else {
			SMSstring += "MCB%20RESTORED%20AT:%20" + TimeStampString
		}

	}
	if (temp2 & 0x04) != 0 {
		if (temp & 0x04) != 0 {
			SMSstring += "EARTH%20FAULE%20AT:%20" + TimeStampString
		} else {
			SMSstring += "EARTH%20FAULE%20RESOLVED%20AT:%20" + TimeStampString
		}
	}

	logger.Printf("Detected Alert event %d\n", TcpUtilsStructPtr.AlertState)

	//SendSMSChan<-SMSstring
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
				response, err := http.Get("http://login.smsgatewayhub.com/api/mt/SendSMS?APIKey=6c3f0e72-71f8-4ffa-94e2-84a0cc7f50b9&senderid=WEBSMS&channel=2&DCS=0&flashsms=0&number=" + to + "&text=" + SMSstring + "&route=1")
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

	//SguUtilsStructPtr.SguTcpUtilsStruct.AlertStateOld = SguUtilsStructPtr.SguTcpUtilsStruct.AlertState

}
