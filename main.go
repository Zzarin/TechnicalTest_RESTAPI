package main

import (
	"TechnicalTest_RESTAPI/config"
	"TechnicalTest_RESTAPI/internal/controller/handlerService"
	entity "TechnicalTest_RESTAPI/internal/model/currency"
	"TechnicalTest_RESTAPI/internal/model/user"
	"TechnicalTest_RESTAPI/internal/storage"
	"TechnicalTest_RESTAPI/internal/usecase/apikey"
	"TechnicalTest_RESTAPI/internal/usecase/currency"
	"TechnicalTest_RESTAPI/pkg/httpserver"
	"TechnicalTest_RESTAPI/pkg/logger"
	"TechnicalTest_RESTAPI/pkg/mysqlDB"
	"context"
	"encoding/base32"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var db *sqlx.DB

var Currency = map[string]struct{}{
	"RUB": {},
	"USD": {},
	"EUR": {},
	"JPY": {},
}

/*type ModelData struct {
	id            int
	currency      string
	rate          float64
	dateUpdated   time.Time
	dateRequested time.Time
}*/

//var DatabaseModel ModelData

//type to unmarsh json and work with it. Not working for now
/*type IncomingJSON struct {
	Date  string `json:"date"`
	Base  string `json:"base"`
	Rates struct {
		RUB json.RawMessage `json:"RUB"`
		EUR json.RawMessage `json:"EUR"`
		USD json.RawMessage `json:"USD"`
		JPY json.RawMessage `json:"JPY"`
	} `json:"rates"`
}*/
////////////////////////////

//map to parse JSON -------- убрать позже, пустой интерфейс это зло
type Map map[string]interface{}

func (m Map) M(s string) Map {
	return m[s].(map[string]interface{})
}

//????
func (m Map) S(s string) string {
	return m[s].(string)
}

////////////////////////

//api request to third-party API "currencyfreaks.com"
func apiRequest(url, method string) ([]byte, error) {
	//logger.Logger.Info("старт запроса на внешнее API")
	//flag.String()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel() // guarantee what after quit from function or goroutine the context will be cancelled - prevent goroutine memory leak

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, nil)

	if err != nil {
		logger.Logger.Error("не удалось создать http-запрос с контекстом")
		return []byte{}, err
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.Error("не удалось выполнить http-запрос на внешний API \"currencyfreaks.com\"")
		return []byte{}, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Logger.Error("не удалось прочитать тело ответа запроса на внешний API \"currencyfreaks.com\"")
		return []byte{}, err
	}
	return body, nil
}

///////////////////////

func scheduleRequest() {
	//logger.Logger.Info("schedule task is working") -- вылетает с ошибкой
	dateToday := time.Now().UTC().Format("2006-01-02")

	dataBaseRequest := storage.ModelRequest{}

	for key, _ := range Currency {
		resultSelect, err := dataBaseRequest.Select(db, key, dateToday)
		if err != nil { //check if we have all 4 currency, if not then api request

			url := "https://api.currencyfreaks.com/latest?apikey=d4cb5a9843b040e8b2e2b7d85794c18b&symbols=RUB,EUR,USD,JPY"
			method := "GET"

			body, err := apiRequest(url, method)
			if err != nil {
				logger.Logger.Error("запрос на внешнее API не удался")
				return
			}

			//unmarsh json to put data in database
			var fileJSON Map
			json.Unmarshal([]byte(body), &fileJSON)
			date, _, _ := strings.Cut(fileJSON["date"].(string), " ")
			fileJSON["date"] = date

			fmt.Printf("Курс валют на последнюю дату %v\n", fileJSON["date"])

			for key, el := range fileJSON.M("rates") {
				fmt.Println(key, " = ", el)
				fmt.Println("\n")
			}

			if err := Insert(db, fileJSON); err != nil {
				logger.Logger.Error("Не удалось записать данные в БД")
			}

			return
		}

		fmt.Printf("Валюта: %v, курс валюты: %v, курс обновлен: %v\n", resultSelect.Currency, resultSelect.Rate, resultSelect.DateUpdated.Format("2006-01-02"))
	}
	return
}

func Auth(w http.ResponseWriter, r *http.Request) {
	newUser := user.New()
	token, err := getToken(20)
	if err != nil {
		logger.Logger.Error("Не удалось создать токен")
	}
	newUser.ApiKey = token
	if err := InsertUser(db, newUser); err != nil {
		logger.Logger.Error("Не удалось записать данные в БД")
	}
	fmt.Fprintf(w, "Ваш API key для доступа к приложению %v. Для выполнения запросов формат URL должен быть: endpoint?apikey=ваш ключ&date/symbols=запрашиваемая информация", newUser.ApiKey)
}

func getToken(length int) (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		logger.Logger.Error("Не удалось сгенерировать токен")
		return "", err
	}

	return strings.ToLower(base32.StdEncoding.EncodeToString(randomBytes)[:length]), nil
}

func main() {

	//creating logger
	logger := logger.InitializeLoger()
	//flag.String()
	defer logger.Sync()

	//DB connection preparing - configs
	cfgDb := config.GetDBConfigLocal()

	if cfgDb == nil {
		logger.Error("Couldn't get configs for connecting to database")
		return
	}

	//dsn := flag.String("dsn", "root:pw@tcp(mysqlDB:3306)/exchange_rate", "MySQL data source name")
	//flag.Parse()

	//start Database
	db, err := mysqlDB.NewDb(cfgDb)
	if err != nil {
		logger.Fatal("не удалось инициализировать БД")
	}
	logger.Info("Успешное подключение к БД")
	log.Println("Успешное подключение к БД")

	defer db.Shutdown()
	////////////////////////////

	// User usecase
	UserUseCase := apikey.NewUser(apikey.NewRepository(db.Db))

	//currency usecase
	CurrencyUseCase := currency.NewCurrency(currency.GetRepository(db.Db), &entity.Currency{})

	//scheduler for a task
	c := cron.New()
	c.AddFunc("30 * * * *", scheduleRequest) //поменять на раз в сутки в 12:00
	c.Start()
	c.Run()
	defer c.Stop()

	handler := handlerService.InitEndpoints(context.Background(), UserUseCase, CurrencyUseCase)

	// start HTTPserver
	appServer := httpserver.NewServer("4057", handler)
	logger.Info("Starting httpserver...")
	if err := appServer.Start(context.Background()); err != nil {
		logger.Fatal("Couldn't start server on port 4057")
	}
	///////////////////////////

}

//Methods to work with database
//нужно передавать переменную объект БД так как если только передавать JSON то будет неявная зависимость в методе инсерт
//того что метод подключается к БД но она объявлена где-то в другом месте и явно не указана в методе
func Insert(db *sqlx.DB, fileJSON Map) error { //M sql.DB добавить во входные параметры
	/*db, err := sql.Open("mysqlDB", "root:pw@tcp(mysqlDB:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}*/
	defer db.Close()
	for key, val := range fileJSON.M("rates") {
		stmt := `INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES (?, ?, ?, UTC_TIMESTAMP());`

		_, err := db.Exec(stmt, key, val, fileJSON["date"])
		if err != nil {
			logger.Logger.Error("запрос на добавление данных в БД не выполнился")
			return err //
		}
	}

	return nil
}

/*
func Select(currency, dateRequested string) error {
	db, err := sql.Open("mysqlDB", "root:mysqlpw@tcp(localhost:49168)/exchange_rate?parseTime=true")
	if err != nil {

		fmt.Println(err)
	}
	defer db.Close()
	result := db.QueryRow("SELECT * FROM currency WHERE currency = ? AND date_updated = ?;", currency, dateRequested)
	err = result.Scan(&DatabaseModel.id, &DatabaseModel.currency, &DatabaseModel.rate, &DatabaseModel.dateUpdated, &DatabaseModel.dateRequested)
	return err
}*/

////////////////////////////////

func InsertUser(db *sqlx.DB, user *user.User) error { //M sql.DB добавить во входные параметры

	/*db, err := sql.Open("mysqlDB", "root:pw@tcp(mysqlDB:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}*/
	defer db.Close()

	stmt := `INSERT INTO apikey (name, api_key, date_registration) VALUES (?, ?, UTC_TIMESTAMP());`

	_, err := db.Exec(stmt, user.Name, user.ApiKey)
	if err != nil {
		logger.Logger.Error("запрос на добавление данных в БД не выполнился")
		return err
	}

	return nil
}

func CheckKey(db *sqlx.DB, apiKey string) error {
	newUser := user.User{Name: "", ApiKey: ""}
	/*db, err := sql.Open("mysqlDB", "root:pw@tcp(mysqlDB:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}*/
	//defer db.Close()
	result := db.QueryRow("SELECT api_key FROM apikey WHERE api_key = ?;", apiKey)
	err := result.Scan(&newUser.ApiKey)
	return err
}
