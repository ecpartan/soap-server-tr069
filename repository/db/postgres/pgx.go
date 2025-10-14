package postgres

import (
	"context"
	"fmt"
	"os"

	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/ecpartan/soap-server-tr069/repository/db/config"
	"github.com/ecpartan/soap-server-tr069/repository/db/postgres/migrations"
	"github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func GetURLDB(cfg *config.DatabaseConf) string {
	var host string
	if host = os.Getenv("DATABASE_HOST"); host == "" {
		host = cfg.Host
	}
	return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable", cfg.UserName, cfg.Password, host, cfg.Port, cfg.Database)
}

func GetURLPg(cfg *config.DatabaseConf) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)
}

func checkDB(cfg *config.DatabaseConf) bool {
	logger.LogDebug("checkDB", cfg)

	conninfo := GetURLDB(cfg)

	db, err := sqlx.Open("postgres", conninfo)
	if err != nil {
		return false
	}

	defer db.Close()
	sqlQuery := "SELECT 1 FROM pg_database WHERE datname = '" + cfg.Database + "'"

	rows, err := db.Query(sqlQuery)
	logger.LogDebug("checkDB", rows, err)

	if err != nil {
		return false
	}
	rows.Close()

	return true
}

func addDB(cfg *config.DatabaseConf) {

	conninfo := GetURLDB(cfg)
	db, err := sqlx.Open("postgres", conninfo)
	if err != nil {
		logger.LogDebug("NewDB", err)
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
	//db, err := sqlx.Connect("postgres", "user=postgres password=postgres host=postgresdb port=5432 dbname=acsserver sslmode=disable")

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
