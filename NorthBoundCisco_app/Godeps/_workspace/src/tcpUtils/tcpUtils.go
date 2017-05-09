package tcpUtils



import (
    "fmt"
	"net"
	"bufio"
	"time"
	"strings"


)
	   

const	(

	StartingDelimeter 		= 0x7E
	MaxInOutBufferLength 	= 1024*8
	MAXNumOFSCUS			= 100
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
	SCUAnalogP1StateArray[]		int

	inputBufferDipstick			int 
	inputBufferReadPtr			int 
	inputBufferWritePtr			int 

	outputBufferDipstick		int 
	outputBufferReadPtr			int 
	outputBufferWritePtr		int 

	InputSyncSearchStatus		int

	ConnectedToSGU 				bool
	SCUListreceived				bool



}									

func (TcpUtilsStructPtr	*TcpUtilsStruct) ConnectToSGU() bool {
	//open connection
	//TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp","192.168.1.1:62000")
	TcpUtilsStructPtr.tcpClient, TcpUtilsStructPtr.err = net.Dial("tcp","54.185.172.55:62002")

	if (TcpUtilsStructPtr.err != nil) {
		fmt.Println("Error opening TCP connection")
		fmt.Println(TcpUtilsStructPtr.err)
		TcpUtilsStructPtr.ConnectedToSGU = false
		return false
	} else {

		fmt.Println("connected")
		TcpUtilsStructPtr.reader = bufio.NewReader(TcpUtilsStructPtr.tcpClient)
		TcpUtilsStructPtr.writer = bufio.NewWriter(TcpUtilsStructPtr.tcpClient)
		TcpUtilsStructPtr.ConnectedToSGU = true
		return true

	}

}

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  RewindInputBuffer(nBytes int) {

	//if (TcpUtilsStructPtr.inputBufferDipstick < nBytes){
    //	fmt.Printf("Rewinding %d bytes when only %d bytes in FIFO\n",nBytes,TcpUtilsStructPtr.inputBufferDipstick);
    //    return;
    //}
        
        
    TcpUtilsStructPtr.inputBufferDipstick += nBytes;
    TcpUtilsStructPtr.inputBufferReadPtr -= nBytes;
    TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength-1);    
	


}	 

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  AddByteToInputBuff(newByte byte ){
        if (TcpUtilsStructPtr.inputBufferDipstick < MaxInOutBufferLength){
            TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferWritePtr] = newByte
			TcpUtilsStructPtr.inputBufferWritePtr++
            TcpUtilsStructPtr.inputBufferWritePtr &= (MaxInOutBufferLength-1);
            TcpUtilsStructPtr.inputBufferDipstick++;
            
        } else {
            //should be spinning here till thread empties buffer.
            //TBD
            fmt.Println("Warning! Input Buff is full");
        }
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  ReadOneByteFromInput() byte {
        if (TcpUtilsStructPtr.inputBufferDipstick >0){
            var newByte = TcpUtilsStructPtr.commandLineBuff[TcpUtilsStructPtr.inputBufferReadPtr]
			//fmt.Printf("%x\n",newByte)
			TcpUtilsStructPtr.inputBufferReadPtr++
            TcpUtilsStructPtr.inputBufferReadPtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.inputBufferDipstick--
            return newByte
            
        } else {
            //should be spinning here till thread fills  buffer.
            //TBD
            fmt.Println("Warning! Input Buff is empty");
            return 0;
        }
        
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)  GetByteFromOutputBuff() byte{
        if (TcpUtilsStructPtr.outputBufferDipstick >0){
            var	newByte = TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferReadPtr]
			TcpUtilsStructPtr.outputBufferReadPtr++
            TcpUtilsStructPtr.outputBufferReadPtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.outputBufferDipstick--
            return newByte
            
        } else {
            //should be spinning here till thread fills  buffer.
            //TBD
            fmt.Println("Warning! Output Buff is empty");
            return 0;
        }
        
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   AddByteToOutputBuff( newByte byte){
        if (TcpUtilsStructPtr.outputBufferDipstick < MaxInOutBufferLength){
            TcpUtilsStructPtr.responseLineBuff[TcpUtilsStructPtr.outputBufferWritePtr] = newByte
			TcpUtilsStructPtr.outputBufferWritePtr++
            TcpUtilsStructPtr.outputBufferWritePtr &= (MaxInOutBufferLength-1)
            TcpUtilsStructPtr.outputBufferDipstick++
            
        } else {
            //should be spinning here till thread empties buffer.
            //TBD
            fmt.Println("Warning! Output Buff is full");
        }
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   ReadTwoBytesFromInput() int {
        var tTemp int
        tTemp = (int) ((((int) (TcpUtilsStructPtr.ReadOneByteFromInput()) << 8))  | ((int) (TcpUtilsStructPtr.ReadOneByteFromInput() & 0x00FF))); 
        return tTemp & 0x0000FFFF;
    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    ReadFourBytesFromInput() int {
        return ((TcpUtilsStructPtr.ReadTwoBytesFromInput() << 16) | (TcpUtilsStructPtr.ReadTwoBytesFromInput() & 0x00FFFF));     
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
            fmt.Println("Warning! Input Buff is empty while jumping ahead")
 
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
            fmt.Println("Not Connected ! Attempting to re-connect ")
            if (!TcpUtilsStructPtr.ConnectToSGU()){
                fmt.Println("Not Connected !")
                return;
            }           
        }
            

        var  bytesAvailable int


		TcpUtilsStructPtr.tcpClient.SetReadDeadline(time.Now().Add(time.Millisecond*500))


		_,err := TcpUtilsStructPtr.reader.Peek(1)

		if err != nil {
			return
		}


		bytesAvailable=TcpUtilsStructPtr.reader.Buffered()
            
        if (bytesAvailable == 0) {
			//fmt.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )
            return;                                   
        }                     
        if (bytesAvailable!=0) {
       	    fmt.Printf("Adding %d Bytes to Buffer\n",bytesAvailable )
        	 
        }
        for ;bytesAvailable>0;bytesAvailable-- {
			tByte, err := TcpUtilsStructPtr.reader.ReadByte()

			if err != nil {
				fmt.Println("Error reading from socket")
				TcpUtilsStructPtr.ConnectedToSGU = false
				return


			} else {
            	TcpUtilsStructPtr.AddByteToInputBuff(tByte)
            }        
    	}
	}

func  (TcpUtilsStructPtr	*TcpUtilsStruct)    SendSocketData() {
        if (!TcpUtilsStructPtr.ConnectedToSGU) {
            //try conneting
            if (!TcpUtilsStructPtr.ConnectToSGU()){
                fmt.Println("Not Connected, can not send data!")
                return
            }           
        }
        
        //fmt.Printf("Sending %d Bytes\n", TcpUtilsStructPtr.outputBufferDipstick)

        for ;TcpUtilsStructPtr.outputBufferDipstick>0; {

			tByte := TcpUtilsStructPtr.GetByteFromOutputBuff()
			//fmt.Printf("%x\n", tByte)
            
   			err :=  TcpUtilsStructPtr.writer.WriteByte(tByte)
			if err != nil {
                fmt.Println("Could not  write to socket");
				return
            }
        }
        if TcpUtilsStructPtr.writer.Flush() != nil {
            fmt.Println("Error while flushing output stream");
        }
    }


func  (TcpUtilsStructPtr	*TcpUtilsStruct)    GetSCUIndexFromSCUID(SCUID uint64) int {
        
        var i int
        for i=0;i<TcpUtilsStructPtr.NumOfSCUs;i++ {
        
            if (TcpUtilsStructPtr.SCUIDArray[i]==SCUID) {
                return i;
			}
        }
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
                return FixedPacketLength + 11;
                
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
            

            default: {
                fmt.Printf("Invalid Packet Type  %x Specifid", PacketType);
            }            
        }
  	return 0;

    }

func  (TcpUtilsStructPtr	*TcpUtilsStruct)   SendResponsePacket( OutputPacketType int,  StatusByte int) {
		
        
		TcpUtilsStructPtr.OutputPacketType = OutputPacketType
        //first add the delimeter
        TcpUtilsStructPtr.AddByteToOutputBuff(StartingDelimeter);
        TcpUtilsStructPtr.OutputPacketLength = TcpUtilsStructPtr.PacketTypeToPacketLength(TcpUtilsStructPtr.OutputPacketType);
        TcpUtilsStructPtr.OutputPacketLength -= 3; //FixedPacketLength;
	//fmt.Println("Packet Type = %d, Packet Length = %x\n", OutputPacketType, TcpUtilsStructPtr.OutputPacketLength)

        TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketLength);
        TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.SGUID);
        TcpUtilsStructPtr.WriteEightBytesToOutput(TcpUtilsStructPtr.TimeStampHi);
        TcpUtilsStructPtr.WriteSixBytesToOutput(TcpUtilsStructPtr.TimeStampLo);
        TcpUtilsStructPtr.WriteFourBytesToOutput(TcpUtilsStructPtr.OutputSeqNumber);
        TcpUtilsStructPtr.WriteTwoBytesToOutput(TcpUtilsStructPtr.OutputPacketType);
        
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
            case 0x4000: {
                break;
            }
            case 0x3000: {
                //separate LampId and LampVal;
                lampID := (StatusByte >> 8) & 0x0FF;
                lampVal := StatusByte & 0x01;
                TcpUtilsStructPtr.WriteEightBytesToOutput(TcpUtilsStructPtr.SCUIDArray[lampID]);
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(1));
                
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
                TcpUtilsStructPtr.AddByteToOutputBuff((byte)(lampVal ));
                
                break;
                
            }
            default: {
                fmt.Printf("Invalid Output Packet Type %x Specifid\n",TcpUtilsStructPtr.OutputPacketType);
            }            
        } 
        if (TcpUtilsStructPtr.outputBufferDipstick < (TcpUtilsStructPtr.OutputPacketLength + 3)) {
            fmt.Println("Output Packet Formating Error");
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
            fmt.Println("Failed to  match start delimiter");
            TcpUtilsStructPtr.InputSyncSearchStatus = 0;
            return;
    	}

    	
    	TcpUtilsStructPtr.InputPacketLength = TcpUtilsStructPtr.ReadTwoBytesFromInput(); 
    	if (TcpUtilsStructPtr.InputPacketLength > 0x8000) {
    		fmt.Printf("Invalid Packet Length = %x   \n",TcpUtilsStructPtr.InputPacketLength); 
    		
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
           

        
        //get 8 bytes of SGU id
        TcpUtilsStructPtr.SGUID = TcpUtilsStructPtr.ReadSixBytesFromInput();
        
        fmt.Printf("SGU ID  %x \n",TcpUtilsStructPtr.SGUID);        
    	
    	//get first 8 bytes of timestamp
        TcpUtilsStructPtr.TimeStampHi = TcpUtilsStructPtr.ReadEightBytesFromInput();

    	//get remaining 6 bytes of timestamp
        TcpUtilsStructPtr.TimeStampLo = TcpUtilsStructPtr.ReadSixBytesFromInput();
        
        //get 4 bytes of input sequence number
        TcpUtilsStructPtr.InputSeqNumber = TcpUtilsStructPtr.ReadFourBytesFromInput();
        
    	TcpUtilsStructPtr.InputPacketType = TcpUtilsStructPtr.ReadTwoBytesFromInput();
        
    	fmt.Printf("Received packet type %d \n",TcpUtilsStructPtr.InputPacketType);
        
        

        switch (TcpUtilsStructPtr.InputPacketType) {
            case 0x0001: {      //Reset Indication
                //send packet of type 0x11
                //read 12 bytes from buffer and junk them
                TcpUtilsStructPtr.ReadNBytesFromInput(12);     
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                fmt.Println("Received packet type 0x0001 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x11,0);                
                break;
            }
            case 0x0002: {      //Keep Alive
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;               
                fmt.Println("Received packet type 0x0002 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x22,0);
                break;               
            }
            case 0x0003: {      //SCU Lit
                //parse packet
                //TBD
                TcpUtilsStructPtr.NumOfSCUs = TcpUtilsStructPtr.ReadTwoBytesFromInput();
                
                for i:=0;i<TcpUtilsStructPtr.NumOfSCUs;i++ {
                    TcpUtilsStructPtr.ReadEightBytesFromInput();
                    //read and dump reserved byte
                    TcpUtilsStructPtr.ReadOneByteFromInput();
                }
                
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;              
                fmt.Println("Received packet type 0x0003 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x23,0);
                TcpUtilsStructPtr.SCUListreceived=true;
                break;               
            }
            case 0x0004: {      //SCU Deleted
                //parse packet
                //just dump 9 bytes
                TcpUtilsStructPtr.ReadNBytesFromInput(9);
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                fmt.Println("Received packet type 0x0004 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x24,0);
                break;               
            }  
            case 0x0005: {      //SCU Added
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(8);
                fmt.Println("Received packet type 0x0005 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0x24,0);                
                TcpUtilsStructPtr.SendResponsePacket(0x25,0);
                break;               
            }
            case 0xe000: {      //Input Status
                //parse packet

                NumSCUPlusSGU := TcpUtilsStructPtr.ReadTwoBytesFromInput();
                TcpUtilsStructPtr.ControlSGUID = TcpUtilsStructPtr.ReadEightBytesFromInput();
                TcpUtilsStructPtr.ReadNBytesFromInput(16);
                
                
                //read the SGU 
                        
                for i:=0;i<NumSCUPlusSGU-1;i++ {
                    SCUID := TcpUtilsStructPtr.ReadEightBytesFromInput();
                    j := TcpUtilsStructPtr.GetSCUIndexFromSCUID(SCUID);
                    //dump next 5 bytes as they are not used for now
                    TcpUtilsStructPtr.ReadNBytesFromInput(5);
                    
                    
                    temp := TcpUtilsStructPtr.ReadTwoBytesFromInput();
                    
                    
                    if (j>=0) {
                        TcpUtilsStructPtr.SCUAnalogP1StateArray[j] = temp;
                    } else {
                        fmt.Println("Unindentified SCU specified");
                        
                    }
                    //dume next 8 bytes
                    TcpUtilsStructPtr.ReadNBytesFromInput(9);
                             
                }
                TcpUtilsStructPtr.OutputSeqNumber = TcpUtilsStructPtr.InputSeqNumber;
                fmt.Println("Received packet type 0xe000 successfully");
                TcpUtilsStructPtr.SendResponsePacket(0xe001,0);
                break;               
            }
            
            //response from SGU for queries
            
            case 0x1001: {      //Get SGU details
                //parse packet
            	
                TcpUtilsStructPtr.ReadNBytesFromInput(3);  
                TcpUtilsStructPtr.SGULatitude = TcpUtilsStructPtr.ReadTwoBytesFromInput() 
                TcpUtilsStructPtr.SGULongitude = TcpUtilsStructPtr.ReadTwoBytesFromInput()
                TcpUtilsStructPtr.ReadNBytesFromInput(3); 
                fmt.Printf("Received sgu coordinates: %f\n" , TcpUtilsStructPtr.SGULatitude);
                fmt.Printf("Received sgu coordinates: %f\n" , TcpUtilsStructPtr.SGULongitude);
                fmt.Println("Received packet type 0x1001 successfully");
                break;               
            }
            
            case 0x2001: {      //Get SCU details
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(34); 
                fmt.Println("Received packet type 0x2001 successfully");
                break;               
            }
         
            case 0x3001: {      //Get/Set Digital Output State
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(15);  
                fmt.Println("Received packet type 0x3001 successfully");
                break;               
            }
            case 0x4001: {      //Get Time Stamp
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(23);
                fmt.Println("Received packet type 0x4001 successfully");
                break;               
            }
            case 0x5001: {      //Set Time Stamp
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(9);    
                fmt.Println("Received packet type 0x5001 successfully");
                break;               
            }
            case 0x6001: {      //Get Input Status
                //parse packet
                //ReadNBytesFromInput(24); 
                TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
                fmt.Println("Received packet type 0x6001 successfully");
                break;               
            }
            case 0x7001: {      //Set Input Status
                //parse packet
                TcpUtilsStructPtr.ReadNBytesFromInput(66); 
                fmt.Println("Received packet type 0x7001 successfully");
                
                break;               
            }
            
            default: {
                fmt.Printf("Invalid Packet Type %d Specified\n",TcpUtilsStructPtr.InputPacketType); 
                TcpUtilsStructPtr.ReadNBytesFromInput(TcpUtilsStructPtr.InputPacketLength + 3 - FixedPacketLength);
                
            
            }
        }
    }


func (TcpUtilsStructPtr	*TcpUtilsStruct) SendLightControl( lightID int,  lightVal int) {
        
        //add sanity checks later
        lightID &= 0x00FF;
        lightVal &= 0x00FF;
        
        fmt.Printf("Sending packet 0x3000, LampId = %d, LampControl = %d\n",lightID,lightVal);
        TcpUtilsStructPtr.SendResponsePacket(0x3000, ((lightID << 8) | lightVal));
    }


func (TcpUtilsStructPtr	*TcpUtilsStruct) MonitorPackets(ticker	*time.Ticker) {

	for range ticker.C {

		fmt.Println("Entering Socket ticker")
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		TcpUtilsStructPtr.ParseInputPacket()
		fmt.Println("Leaving Socket ticker")

	}


}

func (TcpUtilsStructPtr	*TcpUtilsStruct) InitTcpUtilsStruct()  {

	

	//allocate buffers
	TcpUtilsStructPtr.responseLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.commandLineBuff = make([]byte, MaxInOutBufferLength)
	TcpUtilsStructPtr.SCUIDArray = make([]uint64, MAXNumOFSCUS)
	TcpUtilsStructPtr.SCUAnalogP1StateArray = make([]int, MAXNumOFSCUS)

	TcpUtilsStructPtr.SCUIDArray[0] = 0x0013A20040BD3EBF;
    TcpUtilsStructPtr.SCUIDArray[1] = 0x0013A20040BD3E99;
    TcpUtilsStructPtr.SCUIDArray[2] = 0x0013A20040C76B8C;
    TcpUtilsStructPtr.SCUIDArray[3] = 0x0013A20040C76C05;
    TcpUtilsStructPtr.SCUIDArray[4] = 0x0013A20040C76C0D;
    TcpUtilsStructPtr.SCUIDArray[5] = 0x0013A20040995D0F;
    TcpUtilsStructPtr.SCUIDArray[6] = 0x0013A20040C76B98;
    TcpUtilsStructPtr.SCUIDArray[7] = 0x0013A200408E058C;
    TcpUtilsStructPtr.SCUIDArray[8] = 0x0013A20040BD3EC1;
    TcpUtilsStructPtr.SCUIDArray[9] = 0x0013A20040C76B8C;


	ticker := time.NewTicker(time.Millisecond * 5000)
	go TcpUtilsStructPtr.MonitorPackets(ticker)


	TcpUtilsStructPtr.ConnectedToSGU = TcpUtilsStructPtr.ConnectToSGU()



}

func  TcpServer() {


  	// listen on all interfaces   
  	ln, _ := net.Listen("tcp", "52.5.38.201:62000")   
  	// accept connection on port   
  	conn, _ := ln.Accept()   
  	// run loop forever (or until ctrl-c)   
  	for {     
		// will listen for message to process ending in newline (\n)     
		message, _ := bufio.NewReader(conn).ReadString('\n')     
		// output message received     
		fmt.Print("Message Received:", string(message))     
		// sample process for string received     
		newmessage := strings.ToUpper(message)     
		// send new string back to client     
		conn.Write([]byte(newmessage + "\n"))
		
	} 

}


