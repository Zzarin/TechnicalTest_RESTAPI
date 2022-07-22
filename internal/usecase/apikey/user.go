package apikey

import (
	"context"
)

type UserUseCase struct {
	repository *UserRepository
}

func NewUser(ur *UserRepository) *UserUseCase {
	return &UserUseCase{repository: ur}
}

func (userUC *UserUseCase) ApiKeyValid(ctx context.Context, accessToken string) bool {

	err := userUC.repository.GetApiKey(ctx, accessToken)
	if err != nil {
		return false
	}

	return true
}
