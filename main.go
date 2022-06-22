package main

import (
	"TechnicalTest_RESTAPI/internal/logger"
	"TechnicalTest_RESTAPI/internal/model/user"
	"TechnicalTest_RESTAPI/internal/storage"
	"database/sql"
	"encoding/base32"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var db *sql.DB

var Currency = map[string]struct{}{
	"RUB": {},
	"USD": {},
	"EUR": {},
	"JPY": {},
}

type ModelData struct {
	id            int
	currency      string
	rate          float64
	dateUpdated   time.Time
	dateRequested time.Time
}

var DatabaseModel ModelData

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

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		logger.Logger.Error("не удалось составить http запрос")
		return []byte{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		logger.Logger.Error("не удалось выполнить запрос на внешний API \"currencyfreaks.com\"")
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

//http handler function
func GetExchangeRate(w http.ResponseWriter, r *http.Request) {

	//parse URL query to get API key and date
	urlString, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		logger.Logger.Error("не удалось распарсить строку параметров")
	}

	//Валидация по API key
	if urlString["apikey"] == nil {
		/*logger.Logger.Info("API ключ не предоставлен, перенаправление на страницу аутентификации",
		zap.String("apikey", urlString["apikey"][0]))*/
		http.Redirect(w, r, "http://localhost:4050/auth", http.StatusFound)
		return
	}

	if CheckKey(db, urlString["apikey"][0]) != nil {
		/*logger.Logger.Info("API ключ не валидный, перенаправление на страницу аутентификации",
		zap.String("apikey", urlString["apikey"][0]))*/
		http.Redirect(w, r, "http://localhost:4050/auth", http.StatusFound)
		return
	}

	dateToday := time.Now().UTC().Format("2006-01-02")

	dataBaseRequest := storage.ModelRequest{}

	for key, _ := range Currency {
		resultSelect, err := dataBaseRequest.Select(db, key, dateToday)

		//check if we have all 4 currency, if not then api request
		if err != nil {
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
			fmt.Fprintf(w, "Курс валют на последнюю дату %v\n", fileJSON["date"])
			for key, el := range fileJSON.M("rates") {
				fmt.Fprint(w, key, " = ", el)
				fmt.Fprintf(w, "\n")
			}

			if err := Insert(db, fileJSON); err != nil {
				logger.Logger.Error("не удалось записать данные в БД")
			}
			return
		}

		fmt.Fprintf(w, "Валюта: %v, курс валюты: %v, курс обновлен: %v\n", resultSelect.Currency, resultSelect.Rate, resultSelect.DateUpdated.Format("2006-01-02"))
	}

	return
}

//not working for unpaid plan CurrencyFreaks API. Getting data from database
func GetExchangeRateByDate(w http.ResponseWriter, r *http.Request) {

	//parse URL query to get API key and date
	urlString, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		logger.Logger.Error("не удалось распарсить строку параметров для получения API ключа и даты")
	}

	//Валидация по API key
	if urlString["apikey"] == nil {
		logger.Logger.Info("API ключ не предоставлен, перенаправление на страницу аутентификации")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	if CheckKey(db, urlString["apikey"][0]) != nil {
		logger.Logger.Info("API ключ не валидный, перенаправление на страницу аутентификации")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	dateRequested := urlString["date"][0]

	dataBaseRequest := storage.ModelRequest{}

	for key, _ := range Currency {
		resultSelect, err := dataBaseRequest.Select(db, key, dateRequested)
		if err != nil {
			logger.Logger.Info("Курс по запрошенной валюте " + key + " за указанную дату" + dateRequested + "не найден")
			fmt.Fprintf(w, "Курс по запрошенной валюте %v за указанную дату %v не найден\n", key, dateRequested)
		} else {
			fmt.Fprintf(w, "Валюта: %v, курс валюты: %v, курс обновлен: %v\n", resultSelect.Currency, resultSelect.Rate, resultSelect.DateUpdated.Format("2006-01-02"))
		}

	}
	return
}

func GetExchangeRatePair(w http.ResponseWriter, r *http.Request) {

	//parse URL query to get API key and date
	urlString, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		logger.Logger.Error("не удалось распарсить строку параметров")
	}

	//Валидация по API key
	if urlString["apikey"] == nil {
		logger.Logger.Info("API ключ не предоставлен, перенаправление на страницу аутентификации")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	if CheckKey(db, urlString["apikey"][0]) != nil {
		logger.Logger.Info("API ключ не валидный, перенаправление на страницу аутентификации")
		http.Redirect(w, r, "/auth", http.StatusFound)
		return
	}

	//получение пары валют из строки запроса после знака ? "параметры"
	requestedPair := urlString["symbols"][0]
	BaseCurrency, QuoteCurrency, _ := strings.Cut(requestedPair, ",")

	Pair := make(map[string]struct{})
	Pair[BaseCurrency] = struct{}{}
	Pair[QuoteCurrency] = struct{}{}

	dateToday := time.Now().UTC().Format("2006-01-02")

	dataBaseRequest := storage.ModelRequest{}

	for key, _ := range Pair {
		resultSelect, err := dataBaseRequest.Select(db, key, dateToday)
		if err != nil { //check if we have all 4 currency, if not then api request

			url := "https://api.currencyfreaks.com/latest?apikey=d4cb5a9843b040e8b2e2b7d85794c18b&symbols=" + requestedPair
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
			fmt.Fprintf(w, "Курс валют на последнюю дату %v\n", fileJSON["date"])

			for key, el := range fileJSON.M("rates") {
				fmt.Fprint(w, key, " = ", el)
				fmt.Fprintf(w, "\n")
			}
			return
		}

		fmt.Fprintf(w, "Валюта: %v, курс валюты: %v, курс обновлен: %v\n", resultSelect.Currency, resultSelect.Rate, resultSelect.DateUpdated.Format("2006-01-02"))
	}

	return
}

func scheduleRequest() {
	logger.Logger.Info("schedule task is working")
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
	mux := http.NewServeMux()
	//request's handlers to website to get currency rate
	mux.HandleFunc("/getrate", http.HandlerFunc(GetExchangeRate))
	//Получение курсов по указанной дате (все 4 валюты)
	mux.HandleFunc("/getrate/date", http.HandlerFunc(GetExchangeRateByDate))
	//Получение валютных пар из указанных 4х. Т.е. хочу получить курс Рубля к Йене или Доллар к Евро и т.д.
	mux.HandleFunc("/getrate/pair", http.HandlerFunc(GetExchangeRatePair))
	//get API key to use the app
	mux.HandleFunc("/auth", http.HandlerFunc(Auth))

	//creating logger
	logger := logger.InitializeLooger()

	defer logger.Sync()

	//DB connection preparing - configs
	//не вшивай ключ для доступа к API в само приложение,а жди его на вход --- исправить
	//root:mysqlpw@tcp(localhost:49168)/exchange_rate
	dsn := flag.String("dsn", "root:mysqlpw@tcp(mysql:3306)/exchange_rate", "MySQL data source name")
	flag.Parse()

	//start Database
	db, err := InitDB(*dsn)
	if err != nil {
		logger.Fatal("не удалось инициализировать БД")
	}
	logger.Info("Успешное подключение к БД")
	log.Println("Успешное подключение к БД")

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	defer db.Close()
	////////////////////////////

	//scheduler for a task
	c := cron.New()
	c.AddFunc("30 * * * *", scheduleRequest)
	c.Start()
	c.Run()
	defer c.Stop()

	// start web server
	logger.Info("Starting server...")
	if err := http.ListenAndServe(":4057", mux); err != nil {
		logger.Fatal("Не удалось запустить сервер по порту 4057")
	}

	///////////////////////////

}

//DB connection
func InitDB(dsn string) (*sql.DB, error) {
	//initializes a new sql.DB object which is essentially a pool of database connections.
	dbInstance, err := sql.Open("mysql", dsn)
	if err != nil {
		logger.Logger.Panic("не удалось инициализировать базу данных с указанными DN или DSN")
		return nil, err
	}

	//test if we can connect to database
	if err := dbInstance.Ping(); err != nil {
		logger.Logger.Panic("не удалось установить соединение с БД")
		return nil, err
	}

	//seeding the database on initialization - not good, think later
	/*dateRequested := time.Now().Format("2006-01-02")
	dateUpdated := time.Now().Add(time.Duration(-24) * time.Hour).Format("2006-01-02")

	insert1, err := dbInstance.Query("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('RUB', '61.38', ?,?);", dateUpdated, dateRequested)
	insert2, err := dbInstance.Query("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('USD', '1.00', ?,?);", dateUpdated, dateRequested)
	insert3, err := dbInstance.Query("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('EUR', '0.94', ?,?);", dateUpdated, dateRequested)
	insert4, err := dbInstance.Query("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES ('JPY', '132.12', ?,?);", dateUpdated, dateRequested)

	defer insert1.Close()
	defer insert2.Close()
	defer insert3.Close()
	defer insert4.Close()*/

	return dbInstance, nil
}

//Methods to work with database
//нужно передавать переменную объект БД так как если только передавать JSON то будет неявная зависимость в методе инсерт
//того что метод подключается к БД но она объявлена где-то в другом месте и явно не указана в методе
func Insert(db *sql.DB, fileJSON Map) error { //M sql.DB добавить во входные параметры
	db, err := sql.Open("mysql", "root:pw@tcp(mysql:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}
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
	db, err := sql.Open("mysql", "root:mysqlpw@tcp(localhost:49168)/exchange_rate?parseTime=true")
	if err != nil {

		fmt.Println(err)
	}
	defer db.Close()
	result := db.QueryRow("SELECT * FROM currency WHERE currency = ? AND date_updated = ?;", currency, dateRequested)
	err = result.Scan(&DatabaseModel.id, &DatabaseModel.currency, &DatabaseModel.rate, &DatabaseModel.dateUpdated, &DatabaseModel.dateRequested)
	return err
}*/

////////////////////////////////

func InsertUser(db *sql.DB, user *user.User) error { //M sql.DB добавить во входные параметры

	db, err := sql.Open("mysql", "root:pw@tcp(mysql:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}
	defer db.Close()

	stmt := `INSERT INTO user (name, api_key, date_registration) VALUES (?, ?, UTC_TIMESTAMP());`

	_, err = db.Exec(stmt, user.Name, user.ApiKey)
	if err != nil {
		logger.Logger.Error("запрос на добавление данных в БД не выполнился")
		return err
	}
	return nil
}

func CheckKey(db *sql.DB, apiKey string) error {
	newUser := user.User{Name: "", ApiKey: ""}
	db, err := sql.Open("mysql", "root:pw@tcp(mysql:3306)/exchange_rate?parseTime=true")
	if err != nil {
		logger.Logger.Error("не удалось инициализировать базу данных с указанными DN или DSN")
		return err
	}
	defer db.Close()
	result := db.QueryRow("SELECT api_key FROM user WHERE api_key = ?;", apiKey)
	err = result.Scan(&newUser.ApiKey)
	return err
}
