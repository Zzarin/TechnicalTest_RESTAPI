package handlerService

import (
	"TechnicalTest_RESTAPI/internal/usecase/apikey"
	"TechnicalTest_RESTAPI/internal/usecase/currency"
	"context"
	"net/http"
)

type Handler struct {
	handler *http.ServeMux
}

func InitEndpoints(ctx context.Context, userCase *apikey.UserUseCase, currencyCase *currency.CurrencyUseCase) *http.ServeMux {

	newHandler := &Handler{
		handler: http.NewServeMux(),
	}

	//request's handlerService to website to get currency rate
	newHandler.handler.Handle("/getrate", GetExchangeRate(userCase, currencyCase))
	//Получение курсов по указанной дате (все 4 валюты)
	/*newHandler.handler.Handle("/getrate/date", GetExchangeRateByDate(userCase, currencyCase))
	//Получение валютных пар из указанных 4х. Т.е. хочу получить курс Рубля к Йене или Доллар к Евро и т.д.
	newHandler.handler.Handle("/getrate/pair", GetExchangeRatePair(userCase, currencyCase))
	//get API key to use the app
	newHandler.handler.HandleFunc("/auth", http.HandlerFunc(Auth))
	*/

	return newHandler.handler

}

/*
func GetExchangeRate(userCase *apikey.UserUseCase) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
*/
