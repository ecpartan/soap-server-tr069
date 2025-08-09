package postgres

import (
	"context"
	"fmt"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/config"
	"github.com/ecpartan/soap-server-tr069/repository/db/postgres/migrations"
	"github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
)

func GetURLDB(cfg *config.DatabaseConf) string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

func GetURLPg(cfg *config.DatabaseConf) string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%d",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
}

func checkDB(cfg *config.DatabaseConf) bool {
	conninfo := "user=" + cfg.UserName + " password=" + cfg.Password + " host=" + cfg.Host + " sslmode=disable database=postgres"

	db, err := sqlx.Open("postgres", conninfo)
	if err != nil {
		return false
	}

	defer db.Close()
	sqlQuery := "SELECT 1 FROM pg_database WHERE datname = '" + cfg.Database + "'"
	rows, err := db.Query(sqlQuery)
	if err != nil {
		return false
	}
	rows.Close()

	return true
}

func addDB(cfg *config.DatabaseConf) {
	conninfo := "user=" + cfg.UserName + " password=" + cfg.Password + " host=" + cfg.Host + " sslmode=disable database=postgres"
	db, err := sqlx.Open("postgres", conninfo)
	if err != nil {
		return
	}

	defer db.Close()
	sqlQuery := "CREATE DATABASE " + cfg.Database
	ret, err := db.Exec(sqlQuery)
	logger.LogDebug("NewDB", ret, err)
}

func NewClient(ctx context.Context, cfg *config.DatabaseConf) (*sqlx.DB, error) {
	stdlib.GetDefaultDriver()

	if !checkDB(cfg) {
		addDB(cfg)
	}

	addDn := GetURLDB(cfg)
	logger.LogDebug("NewCli", addDn)
	db, err := sqlx.Connect("postgres", addDn)

	if err != nil {
		return nil, err
	}
	logger.LogDebug("Migrate")
	goose.SetBaseFS(&migrations.Content)

	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}

	if err := goose.UpContext(ctx, db.DB, "."); err != nil {
		return nil, err
	}

	return db, nil
}
