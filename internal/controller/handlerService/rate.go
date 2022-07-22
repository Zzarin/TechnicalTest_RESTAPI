package handlerService

import (
	"TechnicalTest_RESTAPI/internal/usecase/apikey"
	"TechnicalTest_RESTAPI/internal/usecase/currency"
	"TechnicalTest_RESTAPI/pkg/logger"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//map to parse JSON -------- убрать позже, пустой интерфейс это зло
type Map map[string]interface{}

func (m Map) M(s string) Map {
	return m[s].(map[string]interface{})
}

//????
func (m Map) S(s string) string {
	return m[s].(string)
}

/////////////////////////////////////////

type incomingJSON struct {
	date  string `json:"date"`
	base  string `json:"base"`
	rates struct {
		rub string `json:"RUB"`
		eur string `json:"EUR"`
		usd string `json:"USD"`
		jpy string `json:"JPY"`
	} `json:"rates"`
}

var Currency = map[string]struct{}{
	"RUB": {},
	"USD": {},
	"EUR": {},
	"JPY": {},
}

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

//http handlerService function
func GetExchangeRate(userCase *apikey.UserUseCase, currencyCase *currency.CurrencyUseCase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//parse URL query to get API key and date
		urlString, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			logger.Logger.Error("не удалось распарсить строку параметров")
		}

		//Валидация по API key
		if urlString["apikey"] == nil {
			/*logger.Logger.Info("API ключ не предоставлен, перенаправление на страницу аутентификации",
			zap.String("apikey", urlString["apikey"][0]))*/
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}
		/*if userCase.ApiKeyVerification(context.Background(), urlString["apikey"][0]))*/
		if userCase.ApiKeyValid(context.Background(), urlString["apikey"][0]) != true {
			/*logger.Logger.Info("API ключ не валидный, перенаправление на страницу аутентификации",
			zap.String("apikey", urlString["apikey"][0]))*/
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		for key, _ := range Currency {
			resultSelect, err := currencyCase.GetCurrency(context.Background(), key)

			//check if we have data for all 4 currencies in DB, if not then api request
			if err != nil {
				url := "https://api.currencyfreaks.com/latest?apikey=d4cb5a9843b040e8b2e2b7d85794c18b&symbols=RUB,EUR,USD,JPY"
				method := "GET"

				body, err := apiRequest(url, method)
				if err != nil {
					logger.Logger.Error("запрос на внешнее API не удался")
					return
				}

				//dbModel, err := toDbModel(body)
				//fmt.Println(dbModel)
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

				//исправить нужно отдавать в метод подготовленную структуру с которой сервис может работать
				if err := currencyCase.AddCurrency(context.Background()); err != nil {
					logger.Logger.Error("не удалось записать данные в БД")
				}
				return
			}
			b, err := json.MarshalIndent(resultSelect, "", " ")
			fmt.Fprintf(w, string(b))
			//fmt.Fprintf(w, "Валюта: %v, курс валюты: %v, курс обновлен: %v\n", resultSelect.Name, resultSelect.Rate, resultSelect.DateUpdated.Format("2006-01-02"))
		}
		return
	})
}

func toDbModel(message []byte) (*incomingJSON, error) {
	var m *incomingJSON
	err := json.Unmarshal(message, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func toJSON() {

}

/*
   func GetExchangeRate(w http.ResponseWriter, r *http.Request) {

   	//parse URL query to get API key and date
   	urlString, err := url.ParseQuery(r.URL.RawQuery)
   	if err != nil {
   		logger.Logger.Error("не удалось распарсить строку параметров")
   	}

   	//Валидация по API key
   	if urlString["apikey"] == nil {
	//logger.Logger.Info("API ключ не предоставлен, перенаправление на страницу аутентификации",
	//zap.String("apikey", urlString["apikey"][0]))
http.Redirect(w, r, "/auth", http.StatusFound)
return
}

if CheckKey(db, urlString["apikey"][0]) != nil {
//logger.Logger.Info("API ключ не валидный, перенаправление на страницу аутентификации",
//zap.String("apikey", urlString["apikey"][0]))
http.Redirect(w, r, "/auth", http.StatusFound)
return
}

dateToday := time.Now().UTC().Format("2006-01-02")

dataBaseRequest := storage.ModelRequest{}

for key, _ := range Currency {
resultSelect, err := dataBaseRequest.Select(db, key, dateToday)

//check if we have data for all 4 currencies in DB, if not then api request
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
*/
//not working for unpaid plan CurrencyFreaks API. Getting data from database
func GetExchangeRateByDate(w http.ResponseWriter, r *http.Request) {
	/*
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

		}*/
	return
}

func GetExchangeRatePair(w http.ResponseWriter, r *http.Request) {
	/*
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
	*/
	return
}
