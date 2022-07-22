package apikey

import "context"

type (
	User interface {
		ApiKeyVerification(ctx context.Context, accessToken string) error
	}

	UserRepo interface {
		GetApiKey(ctx context.Context, accessToken string) error
	}
)
