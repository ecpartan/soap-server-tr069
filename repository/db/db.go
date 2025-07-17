package db

import (
	"context"
	"database/sql"
	"fmt"

	logger "github.com/ecpartan/soap-server-tr069/log"
	dbconf "github.com/ecpartan/soap-server-tr069/repository/db/config"
	"github.com/ecpartan/soap-server-tr069/repository/db/dao"
	"github.com/ecpartan/soap-server-tr069/repository/db/mysql"
	"github.com/ecpartan/soap-server-tr069/repository/db/postgres"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	ctx context.Context
	cfg *dbconf.DatabaseConf
	db  *DB
}

type DB struct {
	*sqlx.DB
}

func New(ctx context.Context, cfg *dbconf.DatabaseConf) (*Service, error) {
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

func NewDB(ctx context.Context, cfg *dbconf.DatabaseConf) (*DB, error) {
	logger.LogDebug("NewDB", "New DB")
	switch cfg.Driver {
	case "pgx":
		if db, err := postgres.NewClient(ctx, cfg); err != nil {
			return nil, err
		} else {
			return &DB{db}, nil
		}
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

func (s *Service) GetUsers() ([]dao.User, error) {
	ret := make([]dao.User, 0)
	err := s.db.Select(&ret, "SELECT id, username,password FROM user")
	logger.LogDebug("GetUsers", ret, err)
	if err != nil {
		return nil, err
	}

	logger.LogDebug("GetUsers", ret)

	return ret, err
}

func (s *Service) GetUser(username string) (dao.User, error) {
	ret := dao.User{}
	err := s.db.Get(&ret, "SELECT id, user,password FROM user WHERE username = ?", username)
	if err != nil {
		return dao.User{}, err
	}
	return ret, err
}
