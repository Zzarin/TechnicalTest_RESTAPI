package storage

import (
	"TechnicalTest_RESTAPI/internal/logger"
	"errors"
	"time"
)

var supportedCurrencies = map[string]struct{}{
	"RUB": {},
	"USD": {},
	"EUR": {},
	"JPY": {},
}

//Checking if data is valid before creating structure with raw data
func validation(raw *RawData) error {

	if raw.currency == "" {
		logger.Logger.Error("не введена валюта")
		return errors.New("не введена валюта")
	}
	if len(raw.currency) != 3 {
		logger.Logger.Error("не верный формат ввода названия валюты, ввод должен состоять из 3-х символов")
		return errors.New("не верный формат ввода названия валюты, ввод должен состоять из 3-х символов")
	}
	if _, err := supportedCurrencies[raw.currency]; !err {
		logger.Logger.Error("не верный формат ввода названия валюты или валюта не поддерживается. Поддерживаются символы валют RUB, USD, EUR и JPY")
		return errors.New("не верный формат ввода названия валюты или валюта не поддерживается. Поддерживаются символы валют RUB, USD, EUR и JPY")
	}

	if raw.date == "" {
		logger.Logger.Error("не введена дата")
		return errors.New("не введена дата")
	}
	if len(raw.date) != 10 {
		logger.Logger.Error("не верный формат ввода даты, формат ввода даты: гггг-мм-дд")
		return errors.New("не верный формат ввода даты, формат ввода даты: гггг-мм-дд")
	}
	if _, err := time.Parse("2006-01-02", raw.date); err != nil {
		logger.Logger.Error("формат ввода не является временем, формат ввода даты: гггг-мм-дд")
		return errors.New("формат ввода не является временем, формат ввода даты: гггг-мм-дд")
	}

	if historicalTime, _ := time.Parse("2006-01-02", raw.date); historicalTime.After(time.Now()) {
		logger.Logger.Error("введенная дата должна быть меньше текущей даты")
		return errors.New("введенная дата должна быть меньше текущей даты")
	}

	return nil

}
