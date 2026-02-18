package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthUsecase struct {
	db *pgxpool.Pool
}

func NewHealthUsecase(db *pgxpool.Pool) *HealthUsecase {
	return &HealthUsecase{db: db}
}

func (uc *HealthUsecase) Check() error {
	return uc.db.Ping(context.Background())
}
