package mock

import (
	"GuGoTik/src/utils/logging"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DBMock sqlmock.Sqlmock
var Conn *sql.DB

func init() {
	logger := logging.LogService("MockDB")
	var err error
	Conn, DBMock, err = sqlmock.New()
	_, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 Conn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		logger.Error("an error '%s' was not expected when opening a stub database connection", err)
	}
}
