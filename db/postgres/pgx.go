package postgres

import (
	"fmt"

	"github.com/ecpartan/soap-server-tr069/internal/config"
)

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
