package storage

import (
	"TechnicalTest_RESTAPI/pkg/logger"
	"github.com/jmoiron/sqlx"
	"time"
)

type RawData struct {
	currency string
	date     string
}

//нужен сервис
type ModelRequest struct {
	currency    string
	dateUpdated time.Time
}

type Model interface {
	Select()
	Insert()
	mapping()
}

var modelRequestInstance ModelRequest

//var db *sql.DB

type ModelData struct {
	Id            int
	Currency      string
	Rate          float64
	DateUpdated   time.Time
	DateRequested time.Time
}

var databaseModel ModelData

func (req *ModelRequest) Select(db *sqlx.DB, currency, date string) (ModelData, error) {

	rawDataInstance := create(currency, date)
	if err := validation(rawDataInstance); err != nil {
		logger.Logger.Error("введенные данные не прошли валидацию")
		return ModelData{}, err
	}

	if err := req.mapping(rawDataInstance); err != nil {
		logger.Logger.Error("не удалось подготовить данные для запроса в БД")
		return ModelData{}, err
	}

	/*db, err := sql.Open("mysqlDB", "root:pw@tcp(mysqlDB:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return ModelData{}, err
	}*/

	//использовать sqlx
	result := db.QueryRow("SELECT * FROM currency WHERE currency = ? AND date_updated = ?;", req.currency, req.dateUpdated.Format("2006-01-02 15:04:05"))
	defer db.Close()
	err := result.Scan(&databaseModel.Id, &databaseModel.Currency, &databaseModel.Rate, &databaseModel.DateUpdated, &databaseModel.DateRequested)

	return databaseModel, err
}

//Methods to work with database
//нужно передавать переменную объект БД так как если только передавать JSON то будет неявная зависимость в методе инсерт
//того что метод подключается к БД но она объявлена где-то в другом месте и явно не указана в методе
/*func Insert(db *sql.DB, fileJSON any) error { //M sql.DB добавить во входные параметры
	db, err := sql.Open("mysqlDB", "root:mysqlpw@tcp(localhost:49153)/exchange_rate?parseTime=true")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()
	for key, val := range fileJSON.M("rates") {
		stmt := `INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES (?, ?, ?, UTC_TIMESTAMP());`

		_, err := db.Exec(stmt, key, val, fileJSON["date"])
		if err != nil {
			return err //
		}
	}

	return nil
}*/

//After validation, convert data into correct format and map it to the new structure with proper type for database
func (req *ModelRequest) mapping(rawDataInstance *RawData) error {
	req.currency = rawDataInstance.currency

	newDate, err := time.Parse("2006-01-02", rawDataInstance.date)
	if err != nil {
		logger.Logger.Error("не удалось преобразовать данные в дату для запроса в БД")
		return err
	}

	req.dateUpdated = newDate

	return nil
}

// return raw structure filled with data from the URL
func create(currency, date string) *RawData {
	return &RawData{
		currency: currency,
		date:     date,
	}
}
