package currency

import (
	"TechnicalTest_RESTAPI/internal/model/currency"
	"context"
	"time"
)

type CurrencyUseCase struct {
	repository *CurrencyRepository
	model      *currency.Currency
}

type Currency interface {
	GetCurrency(ctx context.Context, currencyName string) (*currency.Currency, error)
	GetHistorical(ctx context.Context, currencyName, historicalDate string) (*currency.Currency, error)
	AddCurrency(ctx context.Context, model *currency.Currency) error
}

func NewCurrency(ur *CurrencyRepository, model *currency.Currency) *CurrencyUseCase {
	return &CurrencyUseCase{
		repository: ur,
		model:      model,
	}
}

func (currency *CurrencyUseCase) GetCurrency(ctx context.Context, currencyName string) (*currency.Currency, error) {
	dateToday := time.Now().UTC().Format("2006-01-02")
	var err error

	currency.model, err = currency.repository.GetOne(ctx, currency.model, currencyName, dateToday)
	if err != nil {
		return nil, err
	}
	return currency.model, err
}

func (currency *CurrencyUseCase) GetHistorical(ctx context.Context, currencyName, historicalDate string) (*currency.Currency, error) {
	var err error

	currency.model, err = currency.repository.GetByDate(ctx, currency.model, currencyName, historicalDate)
	if err != nil {
		return nil, err
	}
	return currency.model, nil
}

func (currency CurrencyUseCase) AddCurrency(ctx context.Context) error {
	var err error

	err = currency.repository.Insert(ctx, currency.model)
	if err != nil {
		return err
	}
	return nil
}
