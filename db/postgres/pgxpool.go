package postgres

import (
	"context"

	"github.com/ecpartan/soap-server-tr069/db/config"
	"github.com/ecpartan/soap-server-tr069/db/postgres/migrations"
	logger "github.com/ecpartan/soap-server-tr069/log"
	"github.com/jackc/pgx/stdlib"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

type Client interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	AcquireFunc(ctx context.Context, f func(*pgxpool.Conn) error) error
	AcquireAllIdle(ctx context.Context) []*pgxpool.Conn
	Stat() *pgxpool.Stat
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
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
