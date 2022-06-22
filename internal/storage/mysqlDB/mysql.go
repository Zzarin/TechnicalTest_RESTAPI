package mysqlDB

import (
	"database/sql"
	"fmt"
	"time"
)

type ModelRequest struct {
	Currency    string
	DateUpdated time.Time
}

type ModelData struct {
	Id            int
	Currency      string
	Rate          float64
	DateUpdated   time.Time
	DateRequested time.Time
}

var databaseModel ModelData

func Select(req *ModelRequest) (ModelData, error) {
	db, err := sql.Open("mysqlDB", "root:pw@tcp(mysql:3306)/exchange_rate?parseTime=true")
	if err != nil {
		fmt.Println(err)
	}

	result := db.QueryRow("SELECT * FROM currency WHERE currency = ? AND date_updated = ?;", req.Currency, req.DateUpdated.Format("2006-01-02 15:04:05"))
	defer db.Close()
	err = result.Scan(&databaseModel.Id, &databaseModel.Currency, &databaseModel.Rate, &databaseModel.DateUpdated, &databaseModel.DateRequested)

	return databaseModel, err
}
