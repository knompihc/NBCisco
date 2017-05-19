/********************************************************************
 * FileName:     tcpUtils_test.go
 * Project:      Havells StreetComm
 * Module:       tcpUtils_test
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package tcpUtils

import "testing"
import "net"
import "time"






func testTcpUtils(t *testing.T) {

	t.Log("Running bit parser test")

	var testStruct TcpUtilsStruct


    //init struct
	testStruct.InitTcpUtilsStruct()

	testStruct.commandLineBuff[0] = 0x0e
	testStruct.commandLineBuff[1] = 0x00

	testStruct.inputBufferDipstick = 2


	x0 := testStruct.ReadOneByteFromInput()

    if 	x0 != 0x0e {
		t.Error("Failed ",0,x0)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0xe0
	testStruct.commandLineBuff[1] = 0x00

	testStruct.inputBufferDipstick = 2


	x1 := testStruct.ReadOneByteFromInput()

    if 	x1 != 0xe0 {
		t.Error("Failed ",1,x1)
	}

    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0xe0

	testStruct.inputBufferDipstick = 2


	x2 := testStruct.ReadTwoBytesFromInput()

    if 	x2 != 0xe0 {
		t.Error("Failed ",2,x2)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0xe0
	testStruct.commandLineBuff[1] = 0x00

	testStruct.inputBufferDipstick = 2


	x3 := testStruct.ReadTwoBytesFromInput()

    if 	x3 != 0xe000 {
		t.Error("Failed ",3,x3)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0xe0
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00

	testStruct.inputBufferDipstick = 4


	x4 := testStruct.ReadFourBytesFromInput()

    if 	x4 != 0xe0000000 {
		t.Error("Failed ",4,x4)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0xe0
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00

	testStruct.inputBufferDipstick = 4


	x5 := testStruct.ReadFourBytesFromInput()

    if 	x5 != 0x00e00000 {
		t.Error("Failed ",5,x5)
	}

    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0xe0
	testStruct.commandLineBuff[3] = 0x00

	testStruct.inputBufferDipstick = 4


	x6 := testStruct.ReadFourBytesFromInput()

    if 	x6 != 0x0000e000 {
		t.Error("Failed ",6,x6)
	}

    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0xe0

	testStruct.inputBufferDipstick = 4


	x7 := testStruct.ReadFourBytesFromInput()

    if 	x7 != 0x000000e0 {
		t.Error("Failed ",7,x7)
	}



    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0xf0
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00
	testStruct.commandLineBuff[4] = 0x00
	testStruct.commandLineBuff[5] = 0x00

	testStruct.inputBufferDipstick = 6


	x8 := testStruct.ReadSixBytesFromInput()

    if 	x8 != 0xf00000000000 {
		t.Error("Failed ",8,x8)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0xf0
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00
	testStruct.commandLineBuff[4] = 0x00
	testStruct.commandLineBuff[5] = 0x00

	testStruct.inputBufferDipstick = 6


	x9 := testStruct.ReadSixBytesFromInput()

    if 	x9 != 0x00f000000000 {
		t.Error("Failed ",9,x9)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0xf0
	testStruct.commandLineBuff[3] = 0x00
	testStruct.commandLineBuff[4] = 0x00
	testStruct.commandLineBuff[5] = 0x00

	testStruct.inputBufferDipstick = 6


	x10 := testStruct.ReadSixBytesFromInput()

    if 	x10 != 0x0000f0000000 {
		t.Error("Failed ",10,x10)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0xf0
	testStruct.commandLineBuff[4] = 0x00
	testStruct.commandLineBuff[5] = 0x00

	testStruct.inputBufferDipstick = 6


	x11 := testStruct.ReadSixBytesFromInput()

    if 	x11 != 0x000000f00000 {
		t.Error("Failed ",11,x11)
	}


    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00
	testStruct.commandLineBuff[4] = 0xf0
	testStruct.commandLineBuff[5] = 0x00

	testStruct.inputBufferDipstick = 6


	x12 := testStruct.ReadSixBytesFromInput()

    if 	x12 != 0x00000000f000 {
		t.Error("Failed ",12,x12)
	}

    //init struct
	testStruct.inputBufferReadPtr = 0

	testStruct.commandLineBuff[0] = 0x00
	testStruct.commandLineBuff[1] = 0x00
	testStruct.commandLineBuff[2] = 0x00
	testStruct.commandLineBuff[3] = 0x00
	testStruct.commandLineBuff[4] = 0x00
	testStruct.commandLineBuff[5] = 0xf0

	testStruct.inputBufferDipstick = 6


	x13 := testStruct.ReadSixBytesFromInput()

    if 	x13 != 0x0000000000f0 {
		t.Error("Failed ",13,x13)
	}


	t.Log("Finished bit parser test")


}



func SendPackets(ticker	*time.Ticker, tcpClient	net.Conn,t *testing.T) {

	
	var Message []byte 

	Message = make([]byte, 29)

	Message[0] = 0x7E

	//packet length
	Message[1] = 00
	Message[2] = 26


	//SGU ID
	Message[3] = 0x11
	Message[4] = 0x22
	Message[5] = 0x33
	Message[6] = 0x44
	Message[7] = 0x55
	Message[8] = 0x66


	//Timestamp
	Message[9] = 0
	Message[10] = 0
	Message[11] = 0
	Message[12] = 0
	Message[13] = 0
	Message[14] = 0
	Message[15] = 0
	Message[16] = 0
	Message[17] = 0
	Message[18] = 0
	Message[19] = 0
	Message[20] = 0
	Message[21] = 0
	Message[22] = 0


	//seq num
	Message[23] = 0
	Message[24] = 1
	Message[25] = 2
	Message[26] = 3


	//packet type
	Message[27] = 0
	Message[28] = 2







	for range ticker.C {

		//generate packets
		_, err := tcpClient.Write(Message)

		if err != nil {
			t.Error("Error writing data to socket")
			return
		}



	}

}


func TestSocketConnections(t *testing.T) {


	t.Log("Running socket connection test")


	//open a connection
	transmitter,err := net.Dial("tcp","54.185.172.55:62001")

	if err != nil {
		t.Error("Unable to open TCP connection")
		return
	}



	ticker := time.NewTicker(time.Millisecond * 5000)
	go SendPackets(ticker, transmitter,t)


	for  {

	}




}




















