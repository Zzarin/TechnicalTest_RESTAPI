package currency

import (
	"TechnicalTest_RESTAPI/internal/model/currency"
	"context"
	"github.com/jmoiron/sqlx"
	"time"
)

type CurrencyRepository struct {
	*sqlx.DB
}

type CurrencyRepo interface {
	GetOne(ctx context.Context, model *currency.Currency, currencyName, dateToday string) (*currency.Currency, error)
	GetByDate(ctx context.Context, model *currency.Currency, currencyName, historicalDate string) (*currency.Currency, error)
	Insert(ctx context.Context, model *currency.Currency) error
}

type Dbstructure struct {
	Id             int64     `db:"id"`
	Currency       string    `db:"currency"`
	Rate           float64   `db:"rate"`
	Date_updated   time.Time `db:"date_updated"`
	Date_requested time.Time `db:"date_requested"`
}

func GetRepository(db *sqlx.DB) *CurrencyRepository {
	return &CurrencyRepository{db}
}

func (db *CurrencyRepository) GetOne(ctx context.Context, model *currency.Currency, currencyName, dateToday string) (*currency.Currency, error) {
	dbstructure := &Dbstructure{}
	err := db.Get(dbstructure, "SELECT * FROM currency WHERE currency = ? AND date_updated = ? LIMIT 1;", currencyName, dateToday)
	if err != nil {
		return model, err
	}
	toModelstruct(dbstructure, model)
	return model, nil
}

func (db CurrencyRepository) GetByDate(ctx context.Context, model *currency.Currency, currencyName, historicalDate string) (*currency.Currency, error) {
	dbstructure := &Dbstructure{}
	err := db.Get(dbstructure, "SELECT * FROM currency WHERE currency = ? AND date_updated = ?;", currencyName, historicalDate)
	if err != nil {
		return model, err
	}

	toModelstruct(dbstructure, model)
	return model, nil
}

func (db CurrencyRepository) Insert(ctx context.Context, model *currency.Currency) error {

	db.MustExec("INSERT INTO currency (currency, rate, date_updated, date_requested) VALUES (?, ?, ?, UTC_TIMESTAMP());",
		model.Currency, model.Rate, model.DateUpdated)
	//доделать проверку что запись внесена
	return nil
}

func toModelstruct(dbstructure *Dbstructure, modelstruct *currency.Currency) {
	modelstruct.Id = int(dbstructure.Id)
	modelstruct.Currency = dbstructure.Currency
	modelstruct.Rate = dbstructure.Rate
	modelstruct.DateUpdated = dbstructure.Date_updated
	modelstruct.DateRequested = dbstructure.Date_requested
}

//add method toDBstruct
