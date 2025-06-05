package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ecpartan/soap-server-tr069/db/mysql"
	"github.com/ecpartan/soap-server-tr069/db/postgres"
	"github.com/ecpartan/soap-server-tr069/internal/config"
	logger "github.com/ecpartan/soap-server-tr069/log"
	_ "github.com/go-sql-driver/mysql"
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
	logger.LogDebug("New", "New DB")
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

func NewDB(ctx context.Context, cfg *config.Config) (*DB, error) {
	logger.LogDebug("NewDB", "New DB")
	switch cfg.Database.Driver {
	case "pgx":
		db, err := sqlx.Connect("pgx", postgres.GetURLDB(cfg))
		if err != nil {
			return nil, err
		}
		return &DB{db}, nil
	case "mysql":
		db, err := sqlx.Connect("mysql", mysql.GetURLDB(cfg))
		if err != nil {

			d, err := sql.Open("mysql", mysql.GetlocalURLDB(cfg))
			if err != nil {
				return nil, err
			}
			dbx := sqlx.NewDb(d, "mysql")
			return &DB{dbx}, nil
		}
		return &DB{db}, nil
	}
	return nil, fmt.Errorf("database driver not found")
}
func (s *Service) GetUsers() ([]User, error) {
	ret := make([]User, 0)
	err := s.db.Select(&ret, "SELECT username,password FROM user")

	if err != nil {
		return nil, err
	}

	logger.LogDebug("GetUsers", ret)

	return ret, err
}
