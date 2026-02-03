package httpserver

import (
	"fmt"

	dac "github.com/xinsnake/go-http-digest-auth-client"

	logger "github.com/ecpartan/soap-server-tr069/log"
)

func ExecRequest(url, user, pass string) error {
	if url != "" {
		logger.LogDebug("crURL", url, user, pass)
		dr := dac.NewRequest(user, pass, "GET", url, "")
		_, err := dr.Execute()
		if err != nil {
			return fmt.Errorf("error in execute connection request: %v", err)
		}
	} else {
		return fmt.Errorf("no found addres for this device by SN")
	}
	return nil
}
