package mysql

import (
	"fmt"

	"github.com/ecpartan/soap-server-tr069/internal/config"
)

func GetURLDB(cfg *config.Config) string {
	return fmt.Sprintf(
		"mysql://%s:%s@tcp(%s:%d)/%s",
		cfg.Database.UserName,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
	)
}

func GetlocalURLDB(cfg *config.Config) string {
	return fmt.Sprintf(
		"%s:%s@/%s",
		cfg.Database.UserName,
		cfg.Database.Password,
		cfg.Database.Database,
	)
}
