package db

import (
	"context"
	"fmt"

	"github.com/ecpartan/soap-server-tr069/internal/config"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	ctx context.Context
	cfg *config.Config
	db  *DB
}

type DB struct {
	*sqlx.DB
}

func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	db, err := NewDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &Service{
		ctx: ctx,
		cfg: cfg,
		db:  db,
	}, nil
}

func GetURLDB(cfg *config.Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.Database.UserName,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
}
func NewDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	db, err := sqlx.Connect("pgx", GetURLDB(cfg))
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
