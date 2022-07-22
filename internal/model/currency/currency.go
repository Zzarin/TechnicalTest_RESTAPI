package currency

import "time"

type Currency struct {
	Id            int       `json:"-"`
	Currency      string    `json:"name"`
	Rate          float64   `json:"rate"`
	DateUpdated   time.Time `json:"dateUpdated"`
	DateRequested time.Time `json:"-"`
}
