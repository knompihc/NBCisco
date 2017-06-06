/********************************************************************
 * FileName:     tcpServer.go
 * Project:      Havells StreetComm
 * Module:       tcpServer
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package tcpServer



import (
    //"fmt"
	"net"
	//"bufio"
	"time"
	//"strings"
	"log"

)

var logger *log.Logger

func StartTcpServer(clientFIFO  chan net.Conn,logg *log.Logger) () {


	logger=logg

	//loop for ever
	for {


		// listen on all interfaces   
		ln, err := net.Listen("tcp", ":62000")  
	
		if err != nil {
	
	
			logger.Println("Error starting TCP server")
			logger.Println(err)
			//put thread to sleep
			time.Sleep(time.Millisecond*5000)
		} else {

	
			logger.Println("Listening for connections")

	
	
		
	  
			// accept connection on port   
			for  {

				conn, err := ln.Accept() 

				if err != nil {

					logger.Println("Error creating tcp client")
					logger.Println(err)


				} else {

					logger.Println("Accepted a new connection.")
					//add connection to channel
					clientFIFO<- conn

					time.Sleep(time.Millisecond*5000)


				}


			}

		}
	}


}
																
