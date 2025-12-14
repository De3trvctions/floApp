package db

import (
	"database/sql"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB(syncDB bool) error {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "rootpassword"
	dbHost := "localhost"
	dbPort := "3306"
	dbName := "flo"

	// Construct data source name (DSN)
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8"

	// Register MySQL database driver
	orm.RegisterDriver(dbDriver, orm.DRMySQL)

	// Register default database
	orm.RegisterDataBase("default", dbDriver, dsn)

	// Open a database connection
	var err error
	db, err = sql.Open(dbDriver, dsn)
	if err != nil {
		logs.Error("[InitDB] Open DB fail:", err)
		return err
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		logs.Error("[InitDB] Ping DB fail:", err)
		return err
	}

	aliasName := "default"
	err = orm.SetDataBaseTZ(aliasName, time.Local)
	if err != nil {
		logs.Error("[InitDB] Set TimeZone Error:", err)
		return err
	}

	if syncDB {
		err = orm.RunSyncdb(aliasName, false, true)
		if err != nil {
			logs.Error("[InitDB] Sync DB Error:", err)
			return err
		}
	}

	logs.Info("[InitDB] Init DB Success")
	// Enable SQL query debugging in console
	orm.Debug = true

	// Create a custom writer that forwards SQL logs to the Elasticsearch logger
	sqlLogWriter := &SQLLogWriter{}
	orm.DebugLog = orm.NewLog(sqlLogWriter)

	return nil
}

// SQLLogWriter implements io.Writer to capture ORM SQL logs
type SQLLogWriter struct{}

// Write implements io.Writer interface to capture SQL logs
func (w *SQLLogWriter) Write(p []byte) (n int, err error) {
	// Log the SQL query with [ORM] prefix for Elasticsearch extraction
	logs.Info(string(p))
	return len(p), nil
}

func GetDB() *sql.DB {
	return db
}

func CloseDB() error {
	if db != nil {
		logs.Info("[CloseDB] Closing database connection")
		err := db.Close()
		if err != nil {
			logs.Error("[CloseDB] Error closing database connection:", err)
			return err
		}
		logs.Info("[CloseDB] Database connection closed successfully")
	}
	return nil
}
