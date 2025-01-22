package service

import (
	//"github.com/ArdiSasongko/EwalletProjects-transaction/internal/storage/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
}

func NewService( /*q *sqlc.Queries,*/ db *pgxpool.Pool) Service {
	return Service{}
}
