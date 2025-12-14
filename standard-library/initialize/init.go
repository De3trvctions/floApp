package initilize

import (
	"floapp/standard-library/db"

	"github.com/beego/beego/v2/core/logs"
)

func InitDB() error {
	syncDB := true
	err := db.InitDB(syncDB)
	if err != nil {
		logs.Error("[InitDB] Failed to initialize database:", err)
		return err
	}
	logs.Info("[InitDB] Successfully initialized database")
	return nil
}

// CloseDB properly closes all database connections
func CloseDB() error {
	logs.Info("[CloseDB] Closing database connections")
	err := db.CloseDB()
	if err != nil {
		logs.Error("[CloseDB] Error closing database connections: %v", err)
		return err
	}
	logs.Info("[CloseDB] Database connections closed successfully")
	return nil
}
