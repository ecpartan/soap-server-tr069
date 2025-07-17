package mysql

import (
	"fmt"

	"github.com/ecpartan/soap-server-tr069/repository/db/config"
)

func GetURLDB(cfg *config.DatabaseConf) string {
	return fmt.Sprintf(
		"mysql://%s:%s@tcp(%s:%d)/%s",
		cfg.UserName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)
}

func GetlocalURLDB(cfg *config.DatabaseConf) string {
	return fmt.Sprintf(
		"%s:%s@/%s",
		cfg.UserName,
		cfg.Password,
		cfg.Database,
	)
}
