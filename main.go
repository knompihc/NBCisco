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
	"os/exec"
	"github.com/gocron"
	"log"
	"net"
	"os"
	"bufio"
	"github.com/go-socket.io"
	"net/http"
	"time"
	"github.com/bolt"
	"bytes"
	"encoding/json"
	"encoding/binary"
	"io"
)
var (
	logger					*log.Logger
	server					*socketio.Server
	so 						socketio.Socket
	buff 					[]string
	isconn					bool
	db						*bolt.DB
)
const (
	remoteLog  = false
	maxbufflength = 5000
)

type data struct {
	Time string
	Val  string
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func get(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	loc, _ := time.LoadLocation("Asia/Kolkata")
	tf,_:=time.ParseInLocation("01/02/2006 3:04 PM",from,loc)
	tt,_:=time.ParseInLocation("01/02/2006 3:04 PM",to,loc)
	tf=tf.Add(-1*time.Minute)
	tt=tt.Add(1*time.Minute)
	logger.Println("from=",tf," to=",tt)
	restr:=""
	db.View(func(tx *bolt.Tx) error {
		// Assume our events bucket exists and has RFC3339 encoded time keys.
		c := tx.Bucket([]byte("Ind")).Cursor()
		min:=tf.Format(time.RFC3339)
		max:=tt.Format(time.RFC3339)
		logger.Println("From=",min," To=",max)
		//x:=0
		_, v := c.Seek([]byte(min))
		//sk:=string(k[:])
		mi:=v
		//sv:=binary.BigEndian.Uint64(v)
		//skt,_:=time.ParseInLocation(time.RFC3339,sk,loc)
		//logger.Println( skt," ", sv)

		_, v = c.Seek([]byte(max))
		//sk=string(k[:])
		ma:=v
		if len(v)==0{
			_,ma=c.Last()
		}
		//sv=binary.BigEndian.Uint64(v)
		//skt,_=time.ParseInLocation(time.RFC3339,sk,loc)
		//logger.Println( skt," ", sv)

		c = tx.Bucket([]byte("Log")).Cursor()
		x:=0
		for k, v := c.Seek((mi)); k != nil && bytes.Compare(k, (ma)) <= 0; k, v = c.Next() {
			//logger.Printf("%s: %s\n", k, v)
			if x!=0{
				restr+="##"
			}
			da:=data{}
			err:=json.Unmarshal(v,&da)
			if err!=nil{
				logger.Println(err)
			}
			//logger.Println(kk,vv)
			//st,_:=time.Parse(time.RFC3339,da.Time)
			t,_:=time.ParseInLocation(time.RFC3339,da.Time,loc)
			ss:=t.Format(time.RFC850)
			ss+="=>"+da.Val
			restr+=ss
			x++
		}
		return nil
	})
//	logger.Println(restr)
	io.WriteString(w,restr)
}
func updateDB(t time.Time,ss string){
	id:=uint64(0)
	eorr := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Log"))
		id, _ = b.NextSequence()
		//logger.Println("id=",id)
		da:=data{}
		da.Time=t.Format(time.RFC3339)
		da.Val=ss
		buf, err := json.Marshal(da)
		if err != nil {
			return err
		}
		err = b.Put(itob((id)), buf)
		return err
	})
	if eorr!=nil{
		logger.Println(eorr)
	}
	eorr = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Ind"))
		err := b.Put([]byte(t.Format(time.RFC3339)), itob(id))
		return err
	})
	if eorr!=nil{
		logger.Println(eorr)
	}
}
func runServer(port string) {
	out, err := exec.Command("fuser", port+"/tcp").Output()
	if err!=nil{
		logger.Println(err)
	}
	s := string(out[:])
	if len(s)==0{
		logger.Println("Build Started!!!")
		_,eorr:=exec.Command("go","build","Havels").Output()
		if eorr!=nil{
			logger.Println(eorr)
		}
		logger.Println("Starting Server!!!")
		cmd:=exec.Command("./Havels")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Println(err)
		}
		// start the command after having set up the pipe
		if err := cmd.Start(); err != nil {
			logger.Println(err)
		}
		// read command's stdout line by line
		loc, _ := time.LoadLocation("Asia/Kolkata")
		in := bufio.NewScanner(stdout)
				for in.Scan() {
					//t:=time.Now().Add(330*time.Minute)
					t,_:=time.ParseInLocation("01/02/2006 3:04:05 PM",time.Now().Add(330*time.Minute).Format("01/02/2006 3:04:05 PM"),loc)
					ss:=in.Text()
					go updateDB(t,ss)
					str:=t.Format(time.RFC850)+"=>"
					str+=ss
					logger.Println("out=",str) // write each line to your log, or anything you need
					if len(buff)>maxbufflength{
						buff=buff[1:]
					}
					buff=append(buff,str)
					if isconn{
						so.Emit("chat message", str)
						so.BroadcastTo("chat","chat message", str)
					}

				}
		if err := in.Err(); err != nil {
			logger.Println("error: %s", err)
		}
	}else{
		logger.Println("Server Running with Pid=",s)
	}

}
func startCron(port string){
	gocron.Every(5).Seconds().Do(runServer,port)
	_, time := gocron.NextRun()
	logger.Println("CRON JOB SET AT=",time)
	<-gocron.Start()
}
func main() {
	var err error
	db, err = bolt.Open("log.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	maidd:=uint64(0)
	eorr := db.Update(func(tx *bolt.Tx) error {
		//tx.DeleteBucket([]byte("Log"))
		//tx.DeleteBucket([]byte("Ind"))
		_, err := tx.CreateBucketIfNotExists([]byte("Log"))
		if err != nil {
			logger.Println("create bucket: ", err)
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Ind"))
		if err != nil {
			logger.Println("create bucket: ", err)
			return err
		}
		return nil
	})
	if eorr!=nil{
		logger.Println(eorr)
	}
	wl, err := net.Dial("udp", "logs3.papertrailapp.com:32240")
	defer wl.Close()
	if remoteLog{
		logger =log.New(wl, "runServer: ", log.Lshortfile)
		if err != nil {
			log.Fatal("error")
		}

	}else {
		logger =log.New(os.Stdout, "runServer: ", log.Lshortfile)
	}
	port := os.Getenv("PORT")
	go startCron(port)
	server, err = socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	loc, _ := time.LoadLocation("Asia/Kolkata")
	server.On("connection", func(so1 socketio.Socket) {
		/*eorr := db.Update(func(tx *bolt.Tx) error {
			maidd,_ = tx.Bucket([]byte("Log")).NextSequence()
			return nil
		})
		if eorr!=nil{
			logger.Println(eorr)
		}*/
		so=so1
		so.Join("chat")
		isconn=true
		log.Println("on connection")
		restr:=""
		db.View(func(tx *bolt.Tx) error {
			// Assume our events bucket exists and has RFC3339 encoded time keys.
			c := tx.Bucket([]byte("Log")).Cursor()
			maid,_:=c.Last()
			if len(maid)!=0{
				maidd=binary.BigEndian.Uint64(maid)
			}else{
				maidd=uint64(0)
			}
			id:=(maidd)
			logger.Println("max=",id)
			min := uint64(0)
			if id>2000{
				min=id-2000
			}
			max := id
			x:=0
			for k, v := c.Seek(itob(min)); k != nil && bytes.Compare(k, itob(max)) <= 0; k, v = c.Next() {
				//logger.Printf("%s: %s\n", k, v)
				if x!=0{
					restr+="##"
				}
				da:=data{}
				err:=json.Unmarshal(v,&da)
				if err!=nil{
					logger.Println(err)
				}
				//logger.Println(kk,vv)
				//st,_:=time.Parse(time.RFC3339,da.Time)
				t,_:=time.ParseInLocation(time.RFC3339,da.Time,loc)
				ss:=t.Format(time.RFC850)
				ss+="=>"+da.Val
				restr+=ss
				x++
			}

			return nil
		})
		if restr!=""{
			so.Emit("chat message", restr)
		}

	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})
	http.Handle("/socket.io/", server)
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	http.HandleFunc("/get", get)
	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", nil))
}
