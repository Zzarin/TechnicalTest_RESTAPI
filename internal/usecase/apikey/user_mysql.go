package apikey

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	*sqlx.DB
}

func NewRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db}
}

func (repo *UserRepository) GetApiKey(ctx context.Context, accessToken string) error {

	var key string
	err := repo.Get(&key, "SELECT api_key FROM user WHERE api_key = ?;", accessToken)
	return err
}
