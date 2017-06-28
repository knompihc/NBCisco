/********************************************************************
 * FileName:     dbUtils.go
 * Project:      Havells StreetComm
 * Module:       dbUtils
 * Company:      Havells India Limited
 * Developed by: Chipmonk Technologies Private Limited
 * Copyright and Disclaimer Notice Software:
 **************************************************************************/
package dbUtils

import (
	"database/sql"
	"log"
	"runtime"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

const (
	enableLogs  = false
	useRemoteDB = true
	//useRemoteDB = false
)

type DbUtilsStruct struct {
	Db          *sql.DB
	Tx          *sql.Tx
	Stmt        *sql.Stmt
	Err         error
	DbConnected bool
	DbSemaphore sync.Mutex
}

var logger *log.Logger

func (DbUtilsStructPtr *DbUtilsStruct) CreateDBTable(qStatement string) {

	//create non existane tables
	if enableLogs {
		logger.Println("Creating transaction")
	}

	DbUtilsStructPtr.Tx, DbUtilsStructPtr.Err = DbUtilsStructPtr.Db.Begin()
	if DbUtilsStructPtr.Err != nil {
		log.Println("Error Executing " + qStatement)
		log.Println("Error creating transaction")
		log.Println(DbUtilsStructPtr.Err)
		return
	}

	DbUtilsStructPtr.Stmt, DbUtilsStructPtr.Err = DbUtilsStructPtr.Tx.Prepare(qStatement)

	if DbUtilsStructPtr.Err != nil {
		log.Println("Error Executing " + qStatement)
		log.Println("Error creating statement")
		log.Println(DbUtilsStructPtr.Err)
		return
	}

	if enableLogs {
		logger.Println("Executing statement")
	}

	_, DbUtilsStructPtr.Err = DbUtilsStructPtr.Stmt.Exec()

	if DbUtilsStructPtr.Err != nil {
		log.Println("Error Executing " + qStatement)
		log.Println("Error Executing statement")
		log.Println(DbUtilsStructPtr.Err)
		DbUtilsStructPtr.Tx.Rollback()

	} else {
		DbUtilsStructPtr.Tx.Commit()
	}

	DbUtilsStructPtr.Stmt.Close()
}

func CreateDeployment_ParameterTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Deployment_Parameter Table")
	}

	//create deployment_parameter table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS deployment_parameter( " +
		"deployment_id	VARCHAR(128)  PRIMARY KEY NOT NULL  " +
		",scu_onoff_pkt_delay	VARCHAR(128)	NOT NULL" +
		",scu_poll_delay	VARCHAR(128)	NOT NULL" +
		",scu_schedule_pkt_delay	VARCHAR(128)	NOT NULL" +
		",scu_onoff_retry_delay	VARCHAR(128)	NOT NULL" +
		",scu_max_retry	VARCHAR(128)	NOT NULL" +
		",server_pkt_ack_delay	VARCHAR(128)	NOT NULL) \n"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateOta_ServerTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for OTA_SERVER Table")
	}

	//create ota_server table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS ota_server( " +
		"id BIGINT(20) PRIMARY KEY NOT NULL UNIQUE  AUTO_INCREMENT  " +
		",deployment_id VARCHAR(200) NULL" +
		",device VARCHAR(200) NULL" +
		",access_token VARCHAR(200) NULL,detail VARCHAR(1000) NULL,major VARCHAR(50) NULL,minor VARCHAR(50) NULL,firmware_name VARCHAR(200) NULL) \n"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateUserAuthTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for UserAuth Table")
	}

	//create login table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS UserAuth ( idUserAuth	INT(11)	PRIMARY KEY NOT NULL UNIQUE  AUTO_INCREMENT " +
		",UserID	VARCHAR(128)	PRIMARY KEY  NOT NULL UNIQUE  " +
		",UserPasswd	VARCHAR(24)	NOT NULL) \n"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateScustatusTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Scu_status Table")
	}

	//create scu_status table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `scu_status` (`scu_id` bigint(20) NOT NULL,`timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,`status` int(11) NOT NULL,PRIMARY KEY (`scu_id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateZoneTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Zone Table")
	}

	//create zone table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `zone` (`id` int(11) NOT NULL AUTO_INCREMENT,`name` varchar(120) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateSurveyTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Zone Table")
	}

	//create zone table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `survey` (\n  `id` int(11) NOT NULL AUTO_INCREMENT,\n  `usr` varchar(200) DEFAULT NULL,\n  `pno` varchar(200) DEFAULT NULL,\n  `mun` varchar(200) DEFAULT NULL,\n  `ward` varchar(200) DEFAULT NULL,\n  `loc` varchar(200) DEFAULT NULL,\n  `rw` varchar(200) DEFAULT NULL,\n  `pso` varchar(200) DEFAULT NULL,\n  `pla` varchar(200) DEFAULT NULL,\n  `height` varchar(200) DEFAULT NULL,\n  `pty` varchar(200) DEFAULT NULL,\n  `opw` varchar(200) DEFAULT NULL,\n  `lf` varchar(200) DEFAULT NULL,\n  `earth` varchar(200) DEFAULT NULL,\n  `phase` varchar(200) DEFAULT NULL,\n  `fun` varchar(200) DEFAULT NULL,\n  `lul` varchar(200) DEFAULT NULL,\n  `lat` varchar(200) DEFAULT NULL,\n  `lng` varchar(200) DEFAULT NULL,\n  PRIMARY KEY (`id`)\n)"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateGroupScuTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for groupscu Table")
	}

	//create zone table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `groupscu` (`id` int(11) NOT NULL AUTO_INCREMENT,`name` varchar(120) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateZoneSguTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for ZoneSgu Table")
	}

	//create zonesgu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `zone_sgu` (`id` int(11) NOT NULL AUTO_INCREMENT,`zid` int(11) NOT NULL,`sguid` bigint(20) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateGroupScuRelTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for GroupScuRel Table")
	}
	//create zonesgu table if not already created

	qStatement = "CREATE TABLE IF NOT EXISTS `group_scu_rel` (`id` int(11) NOT NULL AUTO_INCREMENT,`gid` int(11) NOT NULL,`scuid` bigint(20) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateParametersTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Parameters Table")
	}

	//create parameters table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `parameters` (`id` int(11) NOT NULL AUTO_INCREMENT,`sgu_id` bigint(20) NOT NULL,`timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,`Vr` double DEFAULT NULL,`Vy` double DEFAULT NULL,`Vb` double DEFAULT NULL,`Ir` double DEFAULT NULL,`Iy` double DEFAULT NULL,`Ib` double DEFAULT NULL,`Pf` double DEFAULT NULL,`KW` double DEFAULT NULL,`KVA` double DEFAULT NULL,`KWH` double DEFAULT NULL,`KVAH` double DEFAULT NULL,`rKVAH` double DEFAULT NULL,`Run_Hours` double DEFAULT NULL,`freq` double DEFAULT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateInvTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Inv Table")
	}

	//create Inv table if not already created

	qStatement = "CREATE TABLE IF NOT EXISTS `inventory` ( id INT(11)	PRIMARY KEY NOT NULL UNIQUE  AUTO_INCREMENT ,`AssetType` varchar(10) default NULL,`Description` varchar(100) default NULL,`Quantity` int(11) default NULL)"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateDeploymentTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for deployment Table")
	}

	//create Inv table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `deployment` ( `id` int(11) NOT NULL AUTO_INCREMENT,  `deployment_id` varchar(10) NOT NULL,  PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateAdminTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for admin Table")
	}

	//create Admin table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `admin` (`id` int(11) NOT NULL AUTO_INCREMENT,`name` varchar(45) NOT NULL,`email_id` varchar(45) NOT NULL,`mobile_num` varchar(20) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateIdqueryTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for UserAuth Table")
	}

	//create login table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `idquerydefinition` (id INT(11)	PRIMARY KEY NOT NULL UNIQUE  AUTO_INCREMENT ,`deviceid` int(3) default NULL,`length` int(8) default NULL,`query` varchar(500) default NULL) "

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateReportconfTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Reportconf Table")
	}

	//create reportConf if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `reportcofig` (`id` int(11) NOT NULL AUTO_INCREMENT,`reportfrequency` varchar(45) NOT NULL,`reportdef_userid` varchar(45) NOT NULL,`next` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,`type` varchar(45) NOT NULL,PRIMARY KEY (`id`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateTestLoginTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for login Table")
	}

	//create login table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS login ( user_email varbinary(200) PRIMARY KEY, password BLOB, `admin_op` int(11) DEFAULT NULL)\n"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateUserTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for user Table")
	}

	//create user table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `user` (`userid` varchar(45) NOT NULL,	PRIMARY KEY (`userid`)	)"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateSGUTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for SGU Table")
	}

	//create sgu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS sgu ( sgu_id bigint PRIMARY KEY," +
		"location_name varchar(45) NOT NULL DEFAULT ' ',location_lat double precision, location_lng double precision, major varchar(50), minor varchar(50), status varchar(100))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateScheduleTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for schedule Table")
	}

	//create schedule table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS schedule (" +
		"`idschedule` bigint NOT NULL AUTO_INCREMENT," +
		"`ScheduleStartTime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00'," +
		"`ScheduleEndTime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00'," +
		"`ScheduleExpression` varchar(200) NOT NULL," +
		"`pwm` int(11) DEFAULT NULL," +
		"`timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
		"PRIMARY KEY (`idschedule`)," +
		"UNIQUE KEY `idschedule_UNIQUE` (`idschedule`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateScuconfigureTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for scuconfigure Table")
	}

	//create scuconfigure table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS `scuconfigure` (`idSCUSchedule` bigint NOT NULL AUTO_INCREMENT,`ScheduleID` bigint DEFAULT NULL,`Timestamp` timestamp DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,`ScuID` bigint DEFAULT NULL,`SchedulingID` int(11) DEFAULT NULL,`PWM` int(11) DEFAULT NULL,`ScheduleStartTime` timestamp NULL DEFAULT NULL,`ScheduleEndTime` timestamp NULL DEFAULT NULL,	`ScheduleExpression` varchar(200) DEFAULT NULL,	PRIMARY KEY (`idSCUSchedule`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateSCUTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for SCU Table")
	}

	//create scu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS scu ( scu_id bigint PRIMARY KEY," +
		"sgu_id bigint references sgu(sgu_id)," +
		"location_name varchar(45),location_lat double precision, location_lng double precision, major varchar(50), minor varchar(50), status varchar(100))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreateLocationTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for Location Table")
	}

	//create scu table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS location ( location_name varchar(45) PRIMARY KEY," +
		"lattitude 	double precision," +
		"longitude 	double precision," +
		"creator_id   varchar(45),  creation_time time)"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func CreatesupportdefinitionTable(DbUtilsStructPtr *DbUtilsStruct) {
	var qStatement string

	if enableLogs {
		log.Println("Creating statement for schedule Table")
	}

	//create schedule table if not already created
	qStatement = "CREATE TABLE IF NOT EXISTS supportdefinition (" +
		"`idsupportdefinition` int(11) NOT NULL AUTO_INCREMENT," +
		"`Subject` varchar(45) NOT NULL," +
		"`Category` varchar(45) NOT NULL," +
		"`EmailID` varchar(45) NOT NULL," +
		"`ContactNO` varchar(12) NOT NULL," +
		"`Description` varchar(225) NOT NULL," +
		"`Status` tinyint(1) DEFAULT '0'," +
		"`timestamp` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
		"PRIMARY KEY (`idsupportdefinition`))"

	DbUtilsStructPtr.CreateDBTable(qStatement)
}

func (DbUtilsStructPtr *DbUtilsStruct) DbUtilsInit(logg *log.Logger) {
	logger = logg
	var dbinfo string

	if runtime.GOOS == "windows" {
		if !useRemoteDB {

			//dbinfo = "--host=localhost --user=root --database=test --password=admin123"
			dbinfo = "root:admin123@tcp(localhost:3306)/test"
		} else {
			//db in oragon
			//dbinfo = "--host=mysql.cqwf1pvghoch.us-west-2.rds.amazonaws.com --user=HavellsDBAdmin --database=HavellsCCMSDatabase  --password=HavellsCCMS420HavellsCCMS420"
			dbinfo = "HavellsDBAdmin:HavellsCCMS420@tcp(mysql.cqwf1pvghoch.us-west-2.rds.amazonaws.com:3306)/HavellsCCMSTest"
		}
	} else {
		if !useRemoteDB {

			//dbinfo = "--host=localhost --user=root --database=test --password=admin123"
			dbinfo = "root:pass@tcp(localhost:3306)/HavellsCCMSTest"
		} else {
			//db in oragon
			dbinfo = "HavellsDBAdmin:HavellsCCMS420@tcp(mysql.cqwf1pvghoch.us-west-2.rds.amazonaws.com:3306)/HavellsCCMSDatabase"
		}
		//use command below to access remote db from local machine
		//mysql --user=HavellsDBAdmin --password=HavellsCCMS420 --host=mysql.cqwf1pvghoch.us-west-2.rds.amazonaws.com --database=HavellsCCMSDatabase

		//use line below for Heroku
		//DbUtilsStructPtr.Db , DbUtilsStructPtr.Err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	}

	DbUtilsStructPtr.Db, DbUtilsStructPtr.Err = sql.Open("mysql", dbinfo)

	//db, Err := sql.Open("mysql", "tcp:localhost:3306*mydb/root/admin123")
	if DbUtilsStructPtr.Err != nil {
		logger.Println(DbUtilsStructPtr.Err)
		DbUtilsStructPtr.DbConnected = false
	} else {
		//	logger.Println("Connected to database")
		DbUtilsStructPtr.DbConnected = true
	}

	if !DbUtilsStructPtr.DbConnected {
		return
	}

	DbUtilsStructPtr.Err = DbUtilsStructPtr.Db.Ping()

	if DbUtilsStructPtr.Err != nil {
		logger.Println("Db connection error")
		logger.Println(DbUtilsStructPtr.Err)
		DbUtilsStructPtr.DbConnected = false
	}

	CreateTestLoginTable(DbUtilsStructPtr)

	//create non existane tables
	CreateLocationTable(DbUtilsStructPtr)
	CreateSGUTable(DbUtilsStructPtr)
	CreateSCUTable(DbUtilsStructPtr)
	CreateScheduleTable(DbUtilsStructPtr)
	CreateScuconfigureTable(DbUtilsStructPtr)
	CreateReportconfTable(DbUtilsStructPtr)
	CreateInvTable(DbUtilsStructPtr)
	CreateIdqueryTable(DbUtilsStructPtr)
	CreateDeploymentTable(DbUtilsStructPtr)
	CreateParametersTable(DbUtilsStructPtr)
	CreateAdminTable(DbUtilsStructPtr)
	CreateScustatusTable(DbUtilsStructPtr)
	CreateZoneTable(DbUtilsStructPtr)
	CreateZoneSguTable(DbUtilsStructPtr)
	CreateGroupScuTable(DbUtilsStructPtr)
	CreateGroupScuRelTable(DbUtilsStructPtr)
	//CreateUserTable(DbUtilsStructPtr)
	CreatesupportdefinitionTable(DbUtilsStructPtr)
	CreateDeployment_ParameterTable(DbUtilsStructPtr)
	CreateOta_ServerTable(DbUtilsStructPtr)
	CreateSurveyTable(DbUtilsStructPtr)
}
