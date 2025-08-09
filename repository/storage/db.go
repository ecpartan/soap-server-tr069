package storage

import (
	"context"
	"database/sql"
	"fmt"

	logger "github.com/ecpartan/soap-server-tr069/log"
	dbconf "github.com/ecpartan/soap-server-tr069/repository/db/config"
	"github.com/ecpartan/soap-server-tr069/repository/db/mysql"
	"github.com/ecpartan/soap-server-tr069/repository/db/postgres"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Storage struct {
	DevStorage  *DevStorage
	UserStorage *UserStorage
}

func NewStorage(cfg *dbconf.DatabaseConf) (*Storage, error) {
	db, err := NewDB(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	return &Storage{DevStorage: NewDevStorage(db), UserStorage: NewUserStorage(db)}, nil
}

func NewDB(ctx context.Context, cfg *dbconf.DatabaseConf) (*sqlx.DB, error) {
	logger.LogDebug("NewDB", "New DB")
	switch cfg.Driver {
	case "pgx":
		if db, err := postgres.NewClient(ctx, cfg); err != nil {
			return nil, err
		} else {
			return db, nil
		}
	case "mysql":
		db, err := sqlx.Connect("mysql", mysql.GetURLDB(cfg))
		if err != nil {

			d, err := sql.Open("mysql", mysql.GetlocalURLDB(cfg))
			if err != nil {
				return nil, err
			}
			dbx := sqlx.NewDb(d, "mysql")
			return dbx, nil
		}
		return db, nil
	}

	return nil, fmt.Errorf("database driver not found")
}
