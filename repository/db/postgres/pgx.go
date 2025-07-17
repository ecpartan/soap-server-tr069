package postgres

import (
	"fmt"

	"github.com/ecpartan/soap-server-tr069/repository/db/config"
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
