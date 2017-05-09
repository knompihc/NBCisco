/********************************************************************
 * FileName:     main.go
 * Project:      Havells StreetComm
 * Module:       main
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-socket.io"
	"github.com/gocron"
)

var (
	logger *log.Logger
	server *socketio.Server
	so     socketio.Socket
	buff   []string
	isconn bool
)

const (
	remoteLog     = false
	maxbufflength = 5000
)

func runServer(port string) {
	out, err := exec.Command("fuser", port+"/tcp").Output()
	if err != nil {

		logger.Println(err)
	}
	s := string(out[:])
	if len(s) == 0 {
		logger.Println("Build Started!!!")
		_, eorr := exec.Command("go", "build", "Havels").Output()
		if eorr != nil {
			logger.Println(eorr)
		}
		logger.Println("Starting Server!!!")
		cmd := exec.Command("./Havels")
		logger.Println("cmd", cmd)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Println(err)
		}
		//create log file
		f, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			logger.Println(err)
		}
		defer f.Close()
		// start the command after having set up the pipe
		if err := cmd.Start(); err != nil {
			logger.Println(err)
		}
		// read command's stdout line by line
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			str := time.Now().Format(time.RFC850) + "=>"
			str += in.Text()
			if _, err = f.WriteString(str + "\n"); err != nil {
				logger.Println(err)
			}
			logger.Println("out=", str) // write each line to your log, or anything you need
			if len(buff) > maxbufflength {
				buff = buff[1:]
			}
			buff = append(buff, str)
			if isconn {
				so.Emit("chat message", str)
				so.BroadcastTo("chat", "chat message", str)
			}

		}
		if err := in.Err(); err != nil {
			logger.Println("error: %s", err)
		}

	} else {
		logger.Println("Server Running with Pid=", s)
	}

}
func startCron(port string) {
	gocron.Every(5).Seconds().Do(runServer, port)
	_, time := gocron.NextRun()
	logger.Println("CRON JOB SET AT=", time)
	<-gocron.Start()
}
func main() {
	wl, err := net.Dial("udp", "logs3.papertrailapp.com:32240")
	defer wl.Close()
	if remoteLog {
		logger = log.New(wl, "runServer: ", log.Lshortfile)
		if err != nil {
			log.Fatal("error")
		}

	} else {
		logger = log.New(os.Stdout, "runServer: ", log.Lshortfile)
	}
	port := os.Getenv("PORT")
	go startCron(port)
	server, err = socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	server.On("connection", func(so1 socketio.Socket) {
		so = so1
		so.Join("chat")
		isconn = true
		log.Println("on connection")
		restr := ""
		for k, v := range buff {
			if k != 0 {
				restr += "##"
			}
			restr += v
			//so.BroadcastTo("chat","chat message", v)
		}
		if restr != "" {
			so.Emit("chat message", restr)
		}

	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})
	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
