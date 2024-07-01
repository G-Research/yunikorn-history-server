package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	dbpool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) (*PostgresRepository, error) {
	return &PostgresRepository{dbpool: pool}, nil
}

var _ Repository = &PostgresRepository{}
