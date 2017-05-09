package dbUtils



import (
    "fmt"
	"log"
	"database/sql"
	//_ "github.com/go-sql-driver/mysql"
	"os"
	_ "github.com/lib/pq"
	"runtime"
)


const	(
	enableLogs=false
	)
	   


type DbUtilsStruct struct {

	Db				*sql.DB
	tx				*sql.Tx
	stmt			*sql.Stmt
	err				error
	DbConnected     bool

}



func (DbUtilsStructPtr	*DbUtilsStruct) CreateDBTable  (qStatement string) () {




	//create non existane tables
	if (enableLogs) {
		fmt.Println("Creating transaction")
	}

	DbUtilsStructPtr.tx, DbUtilsStructPtr.err = DbUtilsStructPtr.Db.Begin()
	if DbUtilsStructPtr.err != nil {
	    log.Fatal("Error Executing " + qStatement)
	    log.Fatal("Error creating transaction")
		log.Fatal(DbUtilsStructPtr.err)
		return

	}


	DbUtilsStructPtr.stmt, DbUtilsStructPtr.err = DbUtilsStructPtr.tx.Prepare(qStatement)


	if (DbUtilsStructPtr.err != nil ) {
	    log.Fatal("Error Executing " + qStatement)
		log.Fatal("Error creating statement")
		log.Fatal(DbUtilsStructPtr.err)
		return
		
	}

    if (enableLogs) {
		fmt.Println("Executing statement")
	}

	_,DbUtilsStructPtr.err = DbUtilsStructPtr.stmt.Exec()

	if DbUtilsStructPtr.err != nil {
	    log.Fatal("Error Executing " + qStatement)
		log.Fatal("Error Executing statement")
		log.Fatal(DbUtilsStructPtr.err)
		DbUtilsStructPtr.tx.Rollback()

	} else {

		DbUtilsStructPtr.tx.Commit()

    }


	DbUtilsStructPtr.stmt.Close()



}								




func CreateUserAuthTable (DbUtilsStructPtr	*DbUtilsStruct) () {


	var qStatement string  


    if (enableLogs) {
		log.Println("Creating statement for UserAuth Table")
	}

	//create login table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS UserAuth ( idUserAuth	INT(11)	PRIMARY KEY NOT NULL UNIQUE  AUTO_INCREMENT " +
	             ",UserID	VARCHAR(128)	PRIMARY KEY  NOT NULL UNIQUE  "	+
				 ",UserPasswd	VARCHAR(24)	NOT NULL) \n"



	DbUtilsStructPtr.CreateDBTable(qStatement)


}

func CreateTestLoginTable (DbUtilsStructPtr	*DbUtilsStruct) () {


	var qStatement string  

    if (enableLogs) {
		log.Println("Creating statement for login Table")
	}


	//create login table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS login ( user_email varchar(45) PRIMARY KEY, password varchar(45))\n" 



	DbUtilsStructPtr.CreateDBTable(qStatement)


}



func CreateSGUTable (DbUtilsStructPtr	*DbUtilsStruct) () {


	var qStatement string  

    if (enableLogs) {
		log.Println("Creating statement for SGU Table")
	}


	//create sgu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS sgu ( sgu_id bigint(20) PRIMARY KEY," + 
		"location_name varchar(45),location_lat double, location_lng double)" 



	DbUtilsStructPtr.CreateDBTable(qStatement)


}


func CreateSCUTable (DbUtilsStructPtr	*DbUtilsStruct) () {


	var qStatement string  

    if (enableLogs) {
		log.Println("Creating statement for SCU Table")
	}


	//create scu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS scu1 ( sgu_id bigint(20) PRIMARY KEY," + 
		"location_name varchar(45),location_lat double, location_lng double)" 



	DbUtilsStructPtr.CreateDBTable(qStatement)


}								
								
								
								

func (DbUtilsStructPtr	*DbUtilsStruct) DbUtilsInit() {



	//DbUtilsStructPtr.Db , DbUtilsStructPtr.err = sql.Open("mysql", "root:admin123@/test")
	//DbUtilsStructPtr.Db , DbUtilsStructPtr.err = sql.Open("mysql", "root:admin123@/test")


	if runtime.GOOS == "windows" {

		dbinfo := "user=postgres password=admin123 dbname=postgres sslmode=disable"

		DbUtilsStructPtr.Db , DbUtilsStructPtr.err = sql.Open("postgres", dbinfo)

	} else {

		DbUtilsStructPtr.Db , DbUtilsStructPtr.err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	}


	//db, err := sql.Open("mysql", "tcp:localhost:3306*mydb/root/admin123")
	if DbUtilsStructPtr.err != nil {
		fmt.Println(DbUtilsStructPtr.err)
		DbUtilsStructPtr.DbConnected = false
	} else {
		fmt.Println("Connected to database")
		DbUtilsStructPtr.DbConnected = true
	}

    if !DbUtilsStructPtr.DbConnected {
		return
	}



	DbUtilsStructPtr.err =  DbUtilsStructPtr.Db.Ping()

    if DbUtilsStructPtr.err != nil {
		fmt.Println("Db connection error")
		fmt.Println(DbUtilsStructPtr.err)
		DbUtilsStructPtr.DbConnected = false
	}


	CreateTestLoginTable(DbUtilsStructPtr)

	//create non existane tables
	//CreateSGUTable(DbUtilsStructPtr)

	//CreateSCUTable(DbUtilsStructPtr)

	

}
