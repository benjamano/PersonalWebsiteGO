package background

import (
	"PersonalWebsiteGO/handlers"
	"time"
)

func StartPlaytimeChecker() {
	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			// fmt.Println("Checking player playtime...")

			handlers.CheckAndUpdatePlaytime()

			<-ticker.C
		}
	}()
}

func StartPublicIpValidator() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			// fmt.Println("Validating public IP address...")

			handlers.CheckToUpdatePublicIp()
			<-ticker.C
		}
	}()
}
