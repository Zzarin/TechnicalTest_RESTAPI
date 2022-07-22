package mysqlDB

import (
	"TechnicalTest_RESTAPI/config"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Mysql struct {
	Db *sqlx.DB
}

func NewDb(cfgDb *config.DBConfig) (*Mysql, error) {
	cfgToMysql := mysql.Config{
		User:      cfgDb.User,
		Passwd:    cfgDb.Password,
		Net:       cfgDb.Net,
		Addr:      cfgDb.Host + ":" + cfgDb.Port,
		DBName:    cfgDb.Dbname,
		ParseTime: true,
	}

	dbInstance, err := sqlx.Open("mysql", cfgToMysql.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("не удалось инициализировать базу данных с указанными DN или DSN: %w", err)
	}

	dbInstance.SetConnMaxLifetime(cfgDb.ConnMaxLifetime)
	dbInstance.SetMaxOpenConns(cfgDb.MaxOpenConns)
	dbInstance.SetMaxIdleConns(cfgDb.MaxIdleConns)

	//test if we can connect to database
	if err := dbInstance.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось пингануть БД: %w", err)
	}

	return &Mysql{Db: dbInstance}, nil
}

func (sql *Mysql) Shutdown() {
	sql.Db.Close()
}

//seeding the database on initialization - not good, think later
/*dateRequested := time.Now().Format("2006-01-02")
dateUpdated := time.Now().Add(time.Duration(-24) * time.Hour).Format("2006-01-02")

DBtransaction, err := dbInstance.MustBegin()
err := DBtransaction.MustExec("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('RUB', '61.38', ?,?);", dateUpdated, dateRequested)
err := DBtransaction.MustExec("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('USD', '1.00', ?,?);", dateUpdated, dateRequested)
err := DBtransaction.MustExec("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('EUR', '0.94', ?,?);", dateUpdated, dateRequested)
err := DBtransaction.MustExec("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('JPY', '132.12', ?,?);", dateUpdated, dateRequested)

err = DBtransaction.Commit()
*/
